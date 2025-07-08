package app

import (
	"bufio"
	"client/internal/model"
	"client/internal/utils"
	"fmt"
	"net"
	"os"
	"time"
)

type Client struct {
	conn     net.Conn
	userData []byte
}

func NewClient(connect net.Conn) *Client {
	return &Client{
		conn: connect,
	}
}

func (cl *Client) ConnectToChat() {
	fmt.Print("Enter your name: ")
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Scan()
	username := scanner.Text()

	// Горутина для чтения всех сообщений от сервера
	go func() {
		serverReader := bufio.NewReader(cl.conn)
		for {
			msg, err := serverReader.ReadString('\n')
			if err != nil {
				fmt.Println("Disconnected from server.")
				os.Exit(0)
			}
			msgStruct := model.OutgoingMessage{}
			err = utils.JsonToStruct(msg, &msgStruct)
			cl.Print(msgStruct)
			fmt.Print("enter message\n")
		}
	}()

	// Основной поток — отправка сообщений
	consoleScanner := bufio.NewScanner(os.Stdin)
	fmt.Println("Enter text to send:")
	for consoleScanner.Scan() {
		text := consoleScanner.Text()
		msg := model.OutgoingMessage{Name: username, Text: text, Time: time.Now().Format("2006/01/02 15:04:05")}
		message := utils.ClientMessageToJsonStr(msg)
		cl.conn.Write(append(message, '\n'))
	}

}

func (cl *Client) Print(msg model.OutgoingMessage) {
	const (
		ColorReset = "\033[0m"
		ColorGreen = "\033[32m"
		ColorBlue  = "\033[34m"
	)

	fmt.Printf("%s[%s]%s %s%s: %s%s\n", ColorBlue, msg.Time, ColorReset, ColorGreen, msg.Name, ColorReset, msg.Text)
}
