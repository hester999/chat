package tcp

import (
	"bufio"
	"chat/server/internal/model"
	"encoding/json"
	"fmt"
	"net"
	"sync"

	"chat/server/internal/dto"
	"chat/server/utils"
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

	for scanner.Scan() {
		clientMessage := scanner.Text()

		var msgDTO dto.TCPMessageDTO
		err := utils.JsonToStruct(clientMessage, &msgDTO)
		if err != nil {
			t.sendError(conn, "invalid json format")
			continue
		}

		if msgDTO.Type == "register" {
			if msgDTO.Name == "" {
				t.sendError(conn, "username cannot be empty")
				continue
			}
			t.mu.Lock()
			if _, exists := t.clientsByName[msgDTO.Name]; exists {
				t.mu.Unlock()
				t.sendError(conn, "username already taken")
				return
			}
			if username == "" {
				username = msgDTO.Name
				t.clientsByName[username] = conn
				fmt.Printf("User %s registered from %s\n", username, addr)
			}
			t.mu.Unlock()
			continue
		}

		if msgDTO.Type == "exit" {
			t.mu.Lock()
			delete(t.clients, addr)
			if username != "" {
				delete(t.clientsByName, username)
			}
			t.mu.Unlock()
			return
		}

		incomingMsg := model.IncomingMessage{From: msgDTO.Name, Text: clientMessage}
		if msgDTO.Type == "whisper" {
			t.privateChan <- incomingMsg
		} else if msgDTO.Type == "broadcast" {
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

func (t *Transport) sendError(conn net.Conn, message string) {
	errDTO := dto.ErrorDTO{
		Type:    "error",
		Message: message,
	}
	data, _ := json.Marshal(errDTO)
	conn.Write(append(data, '\n'))
}

func (t *Transport) BroadcastMessage(msg model.IncomingMessage) error {
	var msgDTO dto.TCPMessageDTO
	err := utils.JsonToStruct(msg.Text, &msgDTO)
	if err != nil {
		if sender, ok := t.clientsByName[msg.From]; ok {
			t.sendError(sender, "invalid json format in broadcast")
		}
		return fmt.Errorf("send message error: %v", err)
	}

	responseDTO := dto.TCPMessageDTO{
		Type: "broadcast",
		Name: msgDTO.Name,
		Text: msgDTO.Text,
		Time: msgDTO.Time,
	}
	responseJSON, err := json.Marshal(responseDTO)
	if err != nil {
		if sender, ok := t.clientsByName[msg.From]; ok {
			t.sendError(sender, "marshal error in broadcast")
		}
		return fmt.Errorf("marshal response error: %v", err)
	}

	message := string(responseJSON) + "\n"
	t.mu.RLock()
	for _, client := range t.clientsByName {
		_, err = client.Write([]byte(message))
		if err != nil {
			// Не отправляем ошибку всем, только логируем
			fmt.Printf("send message for client error: %s\n", err)
		}
	}
	t.mu.RUnlock()
	return nil
}

func (t *Transport) SendPrivateMessage(msg model.IncomingMessage) error {
	var msgDTO dto.TCPMessageDTO
	err := utils.JsonToStruct(msg.Text, &msgDTO)
	if err != nil {
		if sender, ok := t.clientsByName[msg.From]; ok {
			t.sendError(sender, "invalid json format in whisper")
		}
		return fmt.Errorf("send private message error: %v", err)
	}

	responseDTO := dto.TCPMessageDTO{
		Type: "whisper",
		Name: msgDTO.Name,
		Text: msgDTO.Text,
		Time: msgDTO.Time,
		Dst:  msgDTO.Dst,
	}
	responseJSON, err := json.Marshal(responseDTO)
	if err != nil {
		if sender, ok := t.clientsByName[msg.From]; ok {
			t.sendError(sender, "marshal error in whisper")
		}
		return fmt.Errorf("marshal response error: %v", err)
	}

	message := string(responseJSON) + "\n"
	t.mu.RLock()
	if toConn, ok := t.clientsByName[msgDTO.Dst]; ok {
		toConn.Write([]byte(message))
	} else {
		if fromConn, ok := t.clientsByName[msgDTO.Name]; ok && msgDTO.Dst != msgDTO.Name {
			t.sendError(fromConn, fmt.Sprintf("%s not found", msgDTO.Dst))
		}
		t.mu.RUnlock()
		return fmt.Errorf("send message error: %v", err)
	}

	if fromConn, ok := t.clientsByName[msgDTO.Name]; ok && msgDTO.Dst != msgDTO.Name {
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
