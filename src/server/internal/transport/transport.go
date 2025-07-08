package transport

import (
	"bufio"
	"sync"

	"fmt"
	"server/internal/model"

	"net"
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
	clients  map[string]net.Conn
	messages chan model.IncomingMessage
	mu       sync.Mutex
}

func NewTransport() *TransportIMPL {
	return &TransportIMPL{
		clients:  make(map[string]net.Conn),
		messages: make(chan model.IncomingMessage, 100),
	}
}

func (t *TransportIMPL) Listen() error {
	listener, err := net.Listen("tcp", "localhost:4545")
	if err != nil {
		return err
	}
	defer listener.Close()

	// Запускаем одну горутину для рассылки сообщений всем клиентам
	go func() {
		for msg := range t.messages {
			t.BroadcastMessage(msg)
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
	t.mu.Lock()
	t.clients[addr] = conn
	t.mu.Unlock()
	scanner := bufio.NewScanner(conn)
	for scanner.Scan() {
		clientMessage := scanner.Text()
		t.messages <- model.IncomingMessage{From: addr, Text: clientMessage}
	}
	t.mu.Lock()
	delete(t.clients, addr)
	t.mu.Unlock()
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
	message := msg.Text + "\n"
	for addr, client := range t.clients {
		_, err = client.Write([]byte(message))
		if err != nil {
			err = fmt.Errorf("send message for client:%s error:%s", addr, err)
		}
	}
	return err
}

func (t *TransportIMPL) MessageChannel() <-chan model.IncomingMessage {
	return t.messages
}

func (t *TransportIMPL) Close() error {
	// Закрыть все клиентские соединения
	t.mu.Lock()
	for addr, conn := range t.clients {
		conn.Close()
		delete(t.clients, addr)
	}
	t.mu.Unlock()

	// Закрыть канал сообщений (если нужно завершить горутину рассылки)
	close(t.messages)

	// (Если listener хранится в структуре — закрыть его)
	// t.listener.Close()

	return nil
}
