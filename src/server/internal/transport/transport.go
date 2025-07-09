package transport

import (
	"bufio"
	"fmt"
	"server/internal/model"
	"server/utils"

	"net"
	"strings"
	"time"
)

type Transport interface {
	// Запуск прослушивания порта/сервера (обычно в отдельной горутине)
	Listen() error

	// Канал, из которого ChatServer будет читать входящие сообщения от клиентов
	MessageChannel() <-chan model.IncomingMessage

	// Разослать сообщение всем клиентам (broadcast)
	BroadcastMessage(msg model.IncomingMessage) error

	// Отправить сообщение конкретному клиенту (по адресу)
	SendMessage(msg string, toAddr string) error

	// Завершить работу транспорта, закрыть соединения
	Close() error
}

type TransportIMPL struct {
	clients       map[string]net.Conn
	clientsByName map[string]net.Conn // имя -> соединение
	publicChan    chan model.IncomingMessage
	privateChan   chan model.IncomingMessage
}

func NewTransport() *TransportIMPL {
	return &TransportIMPL{
		clients:       make(map[string]net.Conn),
		clientsByName: make(map[string]net.Conn),
		publicChan:    make(chan model.IncomingMessage, 100),
		privateChan:   make(chan model.IncomingMessage, 100),
	}
}

func (t *TransportIMPL) Listen() error {
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

func (t *TransportIMPL) handleRequest(conn net.Conn) {
	defer conn.Close()
	addr := conn.RemoteAddr().String()
	t.clients[addr] = conn

	scanner := bufio.NewScanner(conn)
	//Сначала ждём имя пользователя
	var username string
	if scanner.Scan() {
		username = scanner.Text()
		t.clientsByName[username] = conn
	}
	for scanner.Scan() {
		clientMessage := scanner.Text()
		msg := model.IncomingMessage{From: username, Text: clientMessage}
		if strings.HasPrefix(clientMessage, "/w ") {
			t.privateChan <- msg
		} else {
			t.publicChan <- msg
		}
	}
	delete(t.clients, addr)
	delete(t.clientsByName, username)
}

func (t *TransportIMPL) SendMessage(msg string, toAddr string) error {
	//for msg := range t.messages {
	//	msg.Text += "\n"
	//	conn.Write([]byte(msg.Text))
	//}
	panic("implement me")

}

func (t *TransportIMPL) BroadcastMessage(msg model.IncomingMessage) error {
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
	//message := fmt.Sprintf("[%s] %s: %s\n", outgoingMessage.Time, outgoingMessage.Name, outgoingMessage.Text)
	message := msg.Text + "\n"
	for _, client := range t.clientsByName {
		_, err = client.Write([]byte(message))
		if err != nil {
			err = fmt.Errorf("send message for client error: %s", err)
		}
	}
	return err
}

func (t *TransportIMPL) SendPrivateMessage(msg model.IncomingMessage) error {
	parts := strings.SplitN(msg.Text, " ", 3)
	if len(parts) < 3 {
		return fmt.Errorf("неправильный формат whisper")
	}
	toName := parts[1]
	whisperText := parts[2]
	privateMsg := fmt.Sprintf("[whisper][%s] %s: %s\n", time.Now().Format("15:04:05"), msg.From, whisperText)
	if toConn, ok := t.clientsByName[toName]; ok {
		toConn.Write([]byte(privateMsg))
	}
	if fromConn, ok := t.clientsByName[msg.From]; ok {
		fromConn.Write([]byte(privateMsg))
	}
	return nil
}

func (t *TransportIMPL) MessageChannel() <-chan model.IncomingMessage {
	return t.publicChan
}

func (t *TransportIMPL) Close() error {
	// Закрыть все клиентские соединения
	for addr, conn := range t.clients {
		conn.Close()
		delete(t.clients, addr)
	}
	for name, conn := range t.clientsByName {
		conn.Close()
		delete(t.clientsByName, name)
	}
	// Закрыть каналы сообщений
	close(t.publicChan)
	close(t.privateChan)
	return nil
}
