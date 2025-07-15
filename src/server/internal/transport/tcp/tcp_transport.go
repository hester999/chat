package tcp

import (
	"bufio"
	"chat/server/internal/model"
	"encoding/json"
	"fmt"
	"net"
	"sync"

	"chat/server/utils"
	"strings"
)

type Transport struct {
	clients       map[string]net.Conn
	clientsByName map[string]net.Conn // Имя -> соединение
	publicChan    chan model.IncomingMessage
	privateChan   chan model.IncomingMessage
	quit          chan struct{}
	mu            sync.RWMutex
}

func NewTCPTransport() *Transport {
	return &Transport{
		clients:       make(map[string]net.Conn),
		clientsByName: make(map[string]net.Conn),
		publicChan:    make(chan model.IncomingMessage, 100),
		privateChan:   make(chan model.IncomingMessage, 100),
		quit:          make(chan struct{}),
	}
}

func (t *Transport) Start(address string) error {
	listener, err := net.Listen("tcp", address)
	if err != nil {
		return err
	}
	defer listener.Close()

	go func() {
		for {
			select {
			case msg := <-t.publicChan:
				t.BroadcastMessage(msg)
			case msg := <-t.privateChan:
				t.SendPrivateMessage(msg)
			case <-t.quit:
				return
			}
		}
	}()

	for {
		conn, err := listener.Accept()
		if err != nil {
			continue
		}
		go t.handleRequest(conn)
	}
}

func (t *Transport) handleRequest(conn net.Conn) {
	defer conn.Close()
	addr := conn.RemoteAddr().String()
	t.mu.Lock()
	t.clients[addr] = conn
	t.mu.Unlock()

	scanner := bufio.NewScanner(conn)
	var username string

	// Читаем все сообщения
	for scanner.Scan() {
		clientMessage := scanner.Text()

		// DTO для парсинга входящих сообщений
		var msgDTO struct {
			Type    string `json:"type,omitempty"`
			Name    string `json:"name"`
			Text    string `json:"text,omitempty"`
			Time    string `json:"time,omitempty"`
			Private bool   `json:"private,omitempty"`
		}

		err := utils.JsonToStruct(clientMessage, &msgDTO)
		if err != nil {
			continue // пропускаем невалидный JSON
		}

		// Обработка регистрации
		if msgDTO.Type == "register" {
			if msgDTO.Name != "" && username == "" {
				username = msgDTO.Name
				t.mu.Lock()
				t.clientsByName[username] = conn
				t.mu.Unlock()
				fmt.Printf("User %s registered from %s\n", username, addr)
			}
			continue
		}

		incomingMsg := model.IncomingMessage{From: msgDTO.Name, Text: clientMessage}
		if strings.HasPrefix(msgDTO.Text, "/w ") {
			t.privateChan <- incomingMsg
		} else if utils.IsExitCommand(msgDTO.Text) {
			t.mu.Lock()
			delete(t.clients, addr)
			if username != "" {
				delete(t.clientsByName, username)
			}
			t.mu.Unlock()
			return
		} else {
			t.publicChan <- incomingMsg
		}
	}

	t.mu.Lock()
	delete(t.clients, addr)
	if username != "" {
		delete(t.clientsByName, username)
	}
	t.mu.Unlock()
}

func (t *Transport) BroadcastMessage(msg model.IncomingMessage) error {
	// DTO для парсинга входящего сообщения
	var msgDTO struct {
		Name string `json:"name"`
		Text string `json:"text"`
		Time string `json:"time"`
	}

	err := utils.JsonToStruct(msg.Text, &msgDTO)
	if err != nil {
		return fmt.Errorf("send message error: %v", err)
	}

	// Создаём бизнес-модель
	outgoingMsg := model.OutgoingMessage{
		Name:    msgDTO.Name,
		Text:    msgDTO.Text,
		Time:    msgDTO.Time,
		Private: false,
	}

	// DTO для отправки
	var responseDTO struct {
		Name    string `json:"name"`
		Text    string `json:"text"`
		Time    string `json:"time"`
		Private bool   `json:"private,omitempty"`
	}

	// Конвертируем бизнес-модель в DTO
	responseDTO.Name = outgoingMsg.Name
	responseDTO.Text = outgoingMsg.Text
	responseDTO.Time = outgoingMsg.Time
	responseDTO.Private = outgoingMsg.Private

	responseJSON, err := json.Marshal(responseDTO)
	if err != nil {
		return fmt.Errorf("marshal response error: %v", err)
	}

	message := string(responseJSON) + "\n"
	t.mu.RLock()
	for _, client := range t.clientsByName {
		_, err = client.Write([]byte(message))
		if err != nil {
			return fmt.Errorf("send message for client error: %s", err)
		}
	}
	t.mu.RUnlock()
	return nil
}

func (t *Transport) SendPrivateMessage(msg model.IncomingMessage) error {

	var msgDTO struct {
		Name string `json:"name"`
		Text string `json:"text"`
		Time string `json:"time"`
	}

	err := utils.JsonToStruct(msg.Text, &msgDTO)
	if err != nil {
		return fmt.Errorf("error comvert json: %v", err)
	}

	parts := strings.SplitN(msgDTO.Text, " ", 3)
	if len(parts) < 3 {
		return fmt.Errorf("invalid whisper")
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

	message := string(responseJSON) + "\n"

	// Отправляем получателю
	t.mu.RLock()
	if toConn, ok := t.clientsByName[toName]; ok {
		toConn.Write([]byte(message))
	}
	t.mu.RUnlock()

	// Отправляем отправителю (подтверждение)
	t.mu.RLock()
	if fromConn, ok := t.clientsByName[msgDTO.Name]; ok {
		fromConn.Write([]byte(message))
	}
	t.mu.RUnlock()

	return nil
}

func (t *Transport) Stop() error {
	close(t.quit)
	t.mu.Lock()
	for addr, conn := range t.clients {
		conn.Close()
		delete(t.clients, addr)
	}
	for name, conn := range t.clientsByName {
		conn.Close()
		delete(t.clientsByName, name)
	}
	t.mu.Unlock()
	close(t.publicChan)
	close(t.privateChan)
	return nil
}
