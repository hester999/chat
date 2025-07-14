package tcp

import (
	"bufio"
	"chat/server/internal/model"
	"encoding/json"
	"fmt"
	"net"

	"chat/server/utils"
	"strings"
)

type TCPTransport struct {
	clients       map[string]net.Conn
	clientsByName map[string]net.Conn // Имя -> соединение
	publicChan    chan model.IncomingMessage
	privateChan   chan model.IncomingMessage
	quit          chan struct{}
}

func NewTCPTransport() *TCPTransport {
	return &TCPTransport{
		clients:       make(map[string]net.Conn),
		clientsByName: make(map[string]net.Conn),
		publicChan:    make(chan model.IncomingMessage, 100),
		privateChan:   make(chan model.IncomingMessage, 100),
		quit:          make(chan struct{}),
	}
}

func (t *TCPTransport) Start() error {
	listener, err := net.Listen("tcp", "localhost:4545")
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

func (t *TCPTransport) handleRequest(conn net.Conn) {
	defer conn.Close()
	addr := conn.RemoteAddr().String()
	t.clients[addr] = conn

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
				t.clientsByName[username] = conn
				fmt.Printf("User %s registered from %s\n", username, addr)
			}
			continue
		}

		incomingMsg := model.IncomingMessage{From: msgDTO.Name, Text: clientMessage}
		if strings.HasPrefix(msgDTO.Text, "/w ") {
			t.privateChan <- incomingMsg
		} else if utils.IsExitCommand(msgDTO.Text) {
			delete(t.clients, addr)
			if username != "" {
				delete(t.clientsByName, username)
			}
			return
		} else {
			t.publicChan <- incomingMsg
		}
	}

	delete(t.clients, addr)
	if username != "" {
		delete(t.clientsByName, username)
	}
}

func (t *TCPTransport) BroadcastMessage(msg model.IncomingMessage) error {
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
	for _, client := range t.clientsByName {
		_, err = client.Write([]byte(message))
		if err != nil {
			return fmt.Errorf("send message for client error: %s", err)
		}
	}
	return nil
}

func (t *TCPTransport) SendPrivateMessage(msg model.IncomingMessage) error {

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
	if toConn, ok := t.clientsByName[toName]; ok {
		toConn.Write([]byte(message))
	}

	// Отправляем отправителю (подтверждение)
	if fromConn, ok := t.clientsByName[msgDTO.Name]; ok {
		fromConn.Write([]byte(message))
	}

	return nil
}

func (t *TCPTransport) Stop() error {
	close(t.quit)
	for addr, conn := range t.clients {
		conn.Close()
		delete(t.clients, addr)
	}
	for name, conn := range t.clientsByName {
		conn.Close()
		delete(t.clientsByName, name)
	}
	close(t.publicChan)
	close(t.privateChan)
	return nil
}
