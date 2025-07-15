package udp

import (
	"chat/server/internal/model"
	"chat/server/utils"
	"encoding/json"
	"fmt"
	"net"
	"strings"
	"sync"
	"time"
)

type ClientInfo struct {
	Addr     *net.UDPAddr
	Name     string
	LastSeen time.Time
}

type Transport struct {
	clients       map[string]*ClientInfo  // ip:port -> ClientInfo
	clientsByName map[string]*net.UDPAddr // имя -> адрес
	publicChan    chan model.IncomingMessage
	privateChan   chan model.IncomingMessage
	quit          chan struct{}
	conn          *net.UDPConn
	mu            sync.RWMutex // Мьютекс для защиты доступа к clients и clientsByName
}

func NewUDPTransport() *Transport {
	return &Transport{
		clients:       make(map[string]*ClientInfo),
		clientsByName: make(map[string]*net.UDPAddr),
		publicChan:    make(chan model.IncomingMessage, 100),
		privateChan:   make(chan model.IncomingMessage, 100),
		quit:          make(chan struct{}),
	}
}

// Очистка неактивных клиентов
func (u *Transport) cleanupInactiveClients(timeout time.Duration) {
	ticker := time.NewTicker(120 * time.Second)
	for {
		select {
		case <-ticker.C:
			now := time.Now()
			u.mu.Lock()
			for ip, client := range u.clients {
				if now.Sub(client.LastSeen) > timeout {
					fmt.Printf("Remove inactive client: %s (%s)\n", client.Name, ip)
					delete(u.clients, ip)
					if client.Name != "" {
						delete(u.clientsByName, client.Name)
					}
				}
			}
			u.mu.Unlock()
		case <-u.quit:
			ticker.Stop()
			return
		}
	}
}

func (u *Transport) Start(address string) error {
	addr, _ := net.ResolveUDPAddr("udp", address)
	conn, _ := net.ListenUDP("udp", addr)
	u.conn = conn
	defer conn.Close()

	go u.cleanupInactiveClients(120 * time.Second)

	go func() {
		for {
			select {
			case msg := <-u.publicChan:
				u.BroadcastMessage(msg)
			case msg := <-u.privateChan:
				u.SendPrivateMessage(msg)
			case <-u.quit:
				return
			}
		}
	}()

	buf := make([]byte, 4096)
	for {
		n, addr, err := conn.ReadFromUDP(buf)
		if err != nil {
			fmt.Printf("%s", err)
			continue
		}
		go u.handleRequest(buf[:n], addr)
	}
}

func (u *Transport) handleRequest(buf []byte, addr *net.UDPAddr) {
	ip := fmt.Sprintf("%s:%d", addr.IP, addr.Port)
	// Обновляем или создаём ClientInfo
	u.mu.Lock()
	if client, ok := u.clients[ip]; ok {
		client.LastSeen = time.Now()
	} else {
		u.clients[ip] = &ClientInfo{Addr: addr, LastSeen: time.Now()}
	}
	u.mu.Unlock()
	strBuf := string(buf)

	var msgDTO struct {
		Type    string `json:"type,omitempty"`
		Name    string `json:"name"`
		Text    string `json:"text,omitempty"`
		Time    string `json:"time,omitempty"`
		Private bool   `json:"private,omitempty"`
	}

	err := utils.JsonToStruct(strBuf, &msgDTO)
	if err != nil {
		fmt.Printf("Error parsing JSON: %s\n", err)
		return
	}

	if msgDTO.Type == "register" {
		if msgDTO.Name != "" {
			// Проверяем, не занято ли имя другим адресом
			u.mu.Lock()
			if existingAddr, ok := u.clientsByName[msgDTO.Name]; ok {
				if existingAddr.String() != addr.String() {
					u.mu.Unlock()
					fmt.Printf("Name %s already taken by %s\n", msgDTO.Name, existingAddr.String())
					return
				}
			}
			u.clientsByName[msgDTO.Name] = addr
			u.clients[ip].Name = msgDTO.Name
			u.mu.Unlock()
			fmt.Printf("User %s registered from %s\n", msgDTO.Name, addr.String())
		}
		return
	}

	// Обработка обычных сообщений (только для зарегистрированных пользователей)
	if msgDTO.Name == "" {
		fmt.Printf("Message from unregistered user %s\n", addr.String())
		return
	}

	// Проверяем, что пользователь зарегистрирован
	u.mu.RLock()
	if _, ok := u.clientsByName[msgDTO.Name]; !ok {
		u.mu.RUnlock()
		fmt.Printf("User %s not registered\n", msgDTO.Name)
		return
	}

	if existingAddr, ok := u.clientsByName[msgDTO.Name]; ok {
		if existingAddr.String() != addr.String() {
			u.mu.RUnlock()
			fmt.Printf("Message from wrong address for user %s\n", msgDTO.Name)
			return
		}
	}
	u.mu.RUnlock()

	incomingMsg := model.IncomingMessage{From: msgDTO.Name, Text: strBuf}
	if strings.HasPrefix(msgDTO.Text, "/w ") {
		u.privateChan <- incomingMsg
	} else if utils.IsExitCommand(msgDTO.Text) {
		ip := fmt.Sprintf("%s:%d", addr.IP, addr.Port)
		var username string
		u.mu.Lock()
		if client, ok := u.clients[ip]; ok {
			username = client.Name
		}

		delete(u.clients, ip)

		if username != "" {
			delete(u.clientsByName, username)
		}
		u.mu.Unlock()
		fmt.Printf("User %s (%s) disconnected\n", username, ip)
		return
	} else {
		u.publicChan <- incomingMsg
	}
}

func (u *Transport) BroadcastMessage(msg model.IncomingMessage) error {

	var msgDTO struct {
		Name string `json:"name"`
		Text string `json:"text"`
		Time string `json:"time"`
	}

	err := utils.JsonToStruct(msg.Text, &msgDTO)
	if err != nil {
		return fmt.Errorf("send message error: %v", err)
	}

	outgoingMsg := model.OutgoingMessage{
		Name:    msgDTO.Name,
		Text:    msgDTO.Text,
		Time:    msgDTO.Time,
		Private: false,
	}

	var responseDTO struct {
		Name    string `json:"name"`
		Text    string `json:"text"`
		Time    string `json:"time"`
		Private bool   `json:"private,omitempty"`
	}

	responseDTO.Name = outgoingMsg.Name
	responseDTO.Text = outgoingMsg.Text
	responseDTO.Time = outgoingMsg.Time
	responseDTO.Private = outgoingMsg.Private

	responseJSON, err := json.Marshal(responseDTO)
	if err != nil {
		return fmt.Errorf("marshal response error: %v", err)
	}

	responseJSON = append(responseJSON, '\n')

	u.mu.RLock()
	for _, client := range u.clients {
		if _, err = u.conn.WriteTo(responseJSON, client.Addr); err != nil {
			fmt.Errorf("send message for clients error: %s", err)
			continue
		}
	}
	u.mu.RUnlock()
	return nil
}

func (u *Transport) Stop() error {
	close(u.quit)
	u.mu.Lock()
	for ip, client := range u.clients {
		fmt.Printf("Disconnecting client: %s (%s)\n", client.Name, ip)
	}
	u.clients = make(map[string]*ClientInfo)
	u.clientsByName = make(map[string]*net.UDPAddr)
	u.mu.Unlock()
	close(u.publicChan)
	close(u.privateChan)
	if u.conn != nil {
		u.conn.Close()
	}
	return nil
}

func (u *Transport) SendPrivateMessage(msg model.IncomingMessage) error {
	// DTO для парсинга входящего сообщения
	var msgDTO struct {
		Name string `json:"name"`
		Text string `json:"text"`
		Time string `json:"time"`
	}

	err := utils.JsonToStruct(msg.Text, &msgDTO)
	if err != nil {
		return fmt.Errorf("не удалось распарсить JSON приватного сообщения: %v", err)
	}

	parts := strings.SplitN(msgDTO.Text, " ", 3)
	if len(parts) < 3 {
		return fmt.Errorf("неправильный формат whisper")
	}

	toName := parts[1]
	whisperText := parts[2]

	// Создаём бизнес-модель
	privateMsg := model.OutgoingMessage{
		Name:    msgDTO.Name,
		Text:    whisperText,
		Time:    msgDTO.Time,
		Private: true,
	}

	// DTO для отправки
	var responseDTO struct {
		Name    string `json:"name"`
		Text    string `json:"text"`
		Time    string `json:"time"`
		Private bool   `json:"private,omitempty"`
	}

	// Конвертируем бизнес-модель в DTO
	responseDTO.Name = privateMsg.Name
	responseDTO.Text = privateMsg.Text
	responseDTO.Time = privateMsg.Time
	responseDTO.Private = privateMsg.Private

	// Сериализуем
	responseJSON, err := json.Marshal(responseDTO)
	if err != nil {
		return fmt.Errorf("marshal private message error: %v", err)
	}

	responseJSON = append(responseJSON, '\n')

	u.mu.RLock()
	for _, v := range u.clients {
		if v.Name == toName {
			if _, err = u.conn.WriteTo(responseJSON, v.Addr); err != nil {
				u.mu.RUnlock()
				return fmt.Errorf("send message for clients error: %s", err)
			}
			break
		}
	}

	if fromConn, ok := u.clientsByName[msgDTO.Name]; ok {
		u.conn.WriteTo(responseJSON, fromConn)
	}
	u.mu.RUnlock()

	return nil
}
