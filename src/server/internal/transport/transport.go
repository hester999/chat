package transport

import (
	"bufio"

	"fmt"
	"server/internal/model"
	// "fmt"
	"net"
)

// type Transport interface {
// 	// Запуск прослушивания порта/сервера (обычно в отдельной горутине)
// 	Listen() error

// 	// Канал, из которого ChatServer будет читать входящие сообщения от клиентов
// 	MessageChannel() <-chan IncomingMessage

// 	// Разослать сообщение всем клиентам (broadcast)
// 	BroadcastMessage(msg IncomingMessage) error

// 	// Отправить сообщение конкретному клиенту (по адресу)
// 	SendMessage(msg string, toAddr string) error

// 	// Завершить работу транспорта, закрыть соединения
// 	Close() error
// }

type Transport struct {
	clients  map[string]net.Conn
	messages chan model.IncomingMessage
}

func NewTransport() *Transport {
	return &Transport{
		clients:  make(map[string]net.Conn),
		messages: make(chan model.IncomingMessage, 100),
	}
}

func (t *Transport) Listen() error {
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

func (t *Transport) handleRequest(conn net.Conn) {
	defer conn.Close()
	addr := conn.RemoteAddr().String()
	t.clients[addr] = conn
	scanner := bufio.NewScanner(conn)
	for scanner.Scan() {
		clientMessage := scanner.Text()
		t.messages <- model.IncomingMessage{From: addr, Text: clientMessage}
	}

	delete(t.clients, addr)
}

func (t *Transport) SendClientMessage(conn net.Conn) {
	for msg := range t.messages {
		msg.Text += "\n"
		conn.Write([]byte(msg.Text))
	}

}

func (t *Transport) BroadcastMessage(msg model.IncomingMessage) error {
	var err error
	message := msg.Text + "\n"
	//outgoingMessage := struct {
	//	Name string `json:"name"`
	//	Text string `json:"text"`
	//	Time string `json:"time"`
	//}{}
	//err = utils.JsonToStruct(msg.Text, &outgoingMessage)
	//if err != nil {
	//	err = fmt.Errorf("send message error")
	//}
	//message := fmt.Sprintf("[%s] %s: %s\n", outgoingMessage.Time, outgoingMessage.Name, outgoingMessage.Text)
	for addr, client := range t.clients {
		_, err = client.Write([]byte(message))
		if err != nil {
			err = fmt.Errorf("send message for client:%s error:%s", addr, err)
		}
	}
	return err
}
