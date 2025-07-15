package udp

import (
	"chat/server/internal/dto"
	"chat/server/internal/model"
	"chat/server/utils"
	"encoding/json"
	"fmt"
	"net"
	"sync"
	"time"
)

type ClientInfo struct {
	Addr     *net.UDPAddr
	Name     string
	LastSeen time.Time
}

type Transport struct {
	clients       map[string]*ClientInfo  
	clientsByName map[string]*net.UDPAddr 
	publicChan    chan model.IncomingMessage
	privateChan   chan model.IncomingMessage
	quit          chan struct{}
	conn          *net.UDPConn
	mu            sync.RWMutex 
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
	
	u.mu.Lock()
	if client, ok := u.clients[ip]; ok {
		client.LastSeen = time.Now()
	} else {
		u.clients[ip] = &ClientInfo{Addr: addr, LastSeen: time.Now()}
	}
	u.mu.Unlock()
	strBuf := string(buf)

	var msgDTO dto.UDPMessageDTO
	err := utils.JsonToStruct(strBuf, &msgDTO)
	if err != nil {
		u.sendError(addr, "invalid json format")
		return
	}

	if msgDTO.Type == "register" {
		if msgDTO.Name == "" {
			u.sendError(addr, "username cannot be empty")
			return
		}
		u.mu.Lock()
		if existingAddr, ok := u.clientsByName[msgDTO.Name]; ok {
			if existingAddr.String() != addr.String() {
				u.mu.Unlock()
				u.sendError(addr, "username already taken")
				return
			}
		}
		u.clientsByName[msgDTO.Name] = addr
		u.clients[ip].Name = msgDTO.Name
		u.mu.Unlock()
		fmt.Printf("User %s registered from %s\n", msgDTO.Name, addr.String())
		return
	}

	if msgDTO.Name == "" {
		u.sendError(addr, "message from unregistered user")
		return
	}

	u.mu.RLock()
	if _, ok := u.clientsByName[msgDTO.Name]; !ok {
		u.mu.RUnlock()
		u.sendError(addr, "user not registered")
		return
	}
	if existingAddr, ok := u.clientsByName[msgDTO.Name]; ok {
		if existingAddr.String() != addr.String() {
			u.mu.RUnlock()
			u.sendError(addr, "message from wrong address for user")
			return
		}
	}
	u.mu.RUnlock()

	incomingMsg := model.IncomingMessage{From: msgDTO.Name, Text: strBuf}
	if msgDTO.Type == "whisper" {
		u.privateChan <- incomingMsg
	} else if msgDTO.Type == "exit" {
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
	} else if msgDTO.Type == "broadcast" {
		u.publicChan <- incomingMsg
	}
}

func (u *Transport) sendError(addr *net.UDPAddr, message string) {
	errDTO := dto.ErrorDTO{
		Type:    "error",
		Message: message,
	}
	data, _ := json.Marshal(errDTO)
	u.conn.WriteTo(data, addr)
}

func (u *Transport) BroadcastMessage(msg model.IncomingMessage) error {
	var msgDTO dto.UDPMessageDTO
	err := utils.JsonToStruct(msg.Text, &msgDTO)
	if err != nil {
		return fmt.Errorf("send message error: %v", err)
	}

	responseDTO := dto.UDPMessageDTO{
		Type: "broadcast",
		Name: msgDTO.Name,
		Text: msgDTO.Text,
		Time: msgDTO.Time,
	}
	responseJSON, err := json.Marshal(responseDTO)
	if err != nil {
		return fmt.Errorf("marshal response error: %v", err)
	}

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
	var msgDTO dto.UDPMessageDTO
	err := utils.JsonToStruct(msg.Text, &msgDTO)
	if err != nil {
		if fromAddr, ok := u.clientsByName[msgDTO.Name]; ok {
			u.sendError(fromAddr, "invalid json format in whisper")
		}
		return fmt.Errorf("не удалось распарсить JSON приватного сообщения: %v", err)
	}

	if msgDTO.Dst == "" {
		if fromAddr, ok := u.clientsByName[msgDTO.Name]; ok {
			u.sendError(fromAddr, "destination user not specified")
		}
		return fmt.Errorf("destination user not specified")
	}

	u.mu.RLock()
	toAddr, ok := u.clientsByName[msgDTO.Dst]
	u.mu.RUnlock()
	if !ok {
		if fromAddr, ok := u.clientsByName[msgDTO.Name]; ok {
			u.sendError(fromAddr, fmt.Sprintf("user %s not found", msgDTO.Dst))
		}
		return fmt.Errorf("user %s not found", msgDTO.Dst)
	}

	responseDTO := dto.UDPMessageDTO{
		Type: "whisper",
		Name: msgDTO.Name,
		Text: msgDTO.Text,
		Time: msgDTO.Time,
		Dst:  msgDTO.Dst,
	}
	responseJSON, err := json.Marshal(responseDTO)
	if err != nil {
		if fromAddr, ok := u.clientsByName[msgDTO.Name]; ok {
			u.sendError(fromAddr, "marshal error in whisper")
		}
		return fmt.Errorf("marshal private message error: %v", err)
	}

	u.conn.WriteTo(responseJSON, toAddr)
	if fromAddr, ok := u.clientsByName[msgDTO.Name]; ok && msgDTO.Dst != msgDTO.Name {
		u.conn.WriteTo(responseJSON, fromAddr)
	}
	return nil
}
