package tcp

import (
	"bufio"
	"fmt"
	"net"
	"chat/server/internal/model"
	
	"chat/server/utils"
	"strings"
)



type TCPTransport struct {
	clients       map[string]net.Conn
	clientsByName map[string]net.Conn // имя -> соединение
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
	// Ждём регистрацию
	if scanner.Scan() {
		regMsg := scanner.Text()
		var reg struct {
			Type string `json:"type"`
			Name string `json:"name"`
		}
		err := utils.JsonToStruct(regMsg, &reg)
		if err != nil || reg.Type != "register" || reg.Name == "" {
			return // невалидная регистрация
		}
		username = reg.Name
		t.clientsByName[username] = conn
	}
	// Теперь читаем обычные сообщения
	for scanner.Scan() {
		clientMessage := scanner.Text()
		var outMsg struct {
			Name string `json:"name"`
			Text string `json:"text"`
			Time string `json:"time"`
		}
		err := utils.JsonToStruct(clientMessage, &outMsg)
		if err != nil {
			continue // или логировать ошибку
		}
		msg := model.IncomingMessage{From: outMsg.Name, Text: clientMessage}
		if strings.HasPrefix(outMsg.Text, "/w ") {
			t.privateChan <- msg
		} else {
			t.publicChan <- msg
		}
	}
	delete(t.clients, addr)
	delete(t.clientsByName, username)
}

func (t *TCPTransport) BroadcastMessage(msg model.IncomingMessage) error {
	var err error
	outgoingMessage := struct {
		Name string `json:"name"`
		Text string `json:"text"`
		Time string `json:"time"`
	}{}
	err = utils.JsonToStruct(msg.Text, &outgoingMessage)
	if err != nil {
		err = fmt.Errorf("send message error")
	}
	message := msg.Text + "\n"
	for _, client := range t.clientsByName {
		_, err = client.Write([]byte(message))
		if err != nil {
			err = fmt.Errorf("send message for client error: %s", err)
		}
	}
	return err
}

func (t *TCPTransport) SendPrivateMessage(msg model.IncomingMessage) error {
	var outMsg struct {
		Name string `json:"name"`
		Text string `json:"text"`
		Time string `json:"time"`
	}
	err := utils.JsonToStruct(msg.Text, &outMsg)
	if err != nil {
		return fmt.Errorf("не удалось распарсить JSON приватного сообщения: %v", err)
	}
	parts := strings.SplitN(outMsg.Text, " ", 3)
	if len(parts) < 3 {
		return fmt.Errorf("неправильный формат whisper")
	}
	toName := parts[1]
	whisperText := parts[2]
	privateMsg := model.OutgoingMessage{
		Name:    outMsg.Name,
		Text:    whisperText,
		Time:    outMsg.Time,
		Private: true,
	}
	jsonMsg := utils.ClientMessageToJsonStr(privateMsg)
	jsonMsg = append(jsonMsg, '\n')
	if toConn, ok := t.clientsByName[toName]; ok {
		toConn.Write(jsonMsg)
	}
	if fromConn, ok := t.clientsByName[outMsg.Name]; ok {
		fromConn.Write(jsonMsg)
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
