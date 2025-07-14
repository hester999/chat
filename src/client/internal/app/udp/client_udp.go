package udp

import (
	"bufio"
	"chat/client/internal/model"
	"chat/client/internal/utils"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"time"
)

type Client struct {
	addr     *net.UDPAddr
	conn     *net.UDPConn
	username string
}

func NewClient(addr *net.UDPAddr) *Client {
	return &Client{
		addr: addr,
	}
}

func (c *Client) ConnectToChat() {
	c.registration()
	fmt.Printf("Enter message: \n")

	go func() {
		buf := make([]byte, 4096)
		for {
			n, _, err := c.conn.ReadFromUDP(buf)
			if err != nil {
				fmt.Println("Error reading from server:", err)
				return
			}
			msgStruct := model.OutgoingMessage{}
			err = utils.JsonToStruct(string(buf[:n]), &msgStruct)
			c.print(msgStruct)
		}
	}()

	c.SendMessage()
}

func (c *Client) print(msg model.OutgoingMessage) {
	const (
		ColorReset   = "\033[0m"
		ColorGreen   = "\033[32m"
		ColorBlue    = "\033[34m"
		ColorMagenta = "\033[35m"
	)

	timeStr := fmt.Sprintf("%s[%s]%s", ColorBlue, msg.Time, ColorReset)
	nameStr := fmt.Sprintf("%s%s%s", ColorGreen, msg.Name, ColorReset)

	if msg.Private {
		fmt.Printf("%s[whisper]%s %s %s: %s\n",
			ColorMagenta, ColorReset,
			timeStr,
			nameStr,
			msg.Text,
		)
	} else {
		fmt.Printf("%s %s: %s\n",
			timeStr,
			nameStr,
			msg.Text,
		)
	}
	fmt.Printf("Enter message: ")
}

func (c *Client) SendMessage() {
	consoleScanner := bufio.NewScanner(os.Stdin)
	for consoleScanner.Scan() {
		text := consoleScanner.Text()

		msg := model.OutgoingMessage{Name: c.username, Text: text, Time: time.Now().Format("2006/01/02 15:04:05")}
		msgData := utils.ClientMessageToJsonStr(msg)
		_, err := c.conn.Write(msgData)
		if err != nil {
			fmt.Println("Error sending message:", err)
			break
		}
		fmt.Printf("Enter message: \n")
	}
}

func (c *Client) registration() {
	fmt.Print("Enter your name: ")
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Scan()
	c.username = scanner.Text()

	conn, err := net.DialUDP("udp", nil, c.addr)
	if err != nil {
		fmt.Println("Error connecting:", err)
		os.Exit(1)
	}
	c.conn = conn

	regMsg := struct {
		Type string `json:"type"`
		Name string `json:"name"`
	}{"register", c.username}

	regData, _ := json.Marshal(regMsg)
	c.conn.Write(regData)
	fmt.Printf("Registered as %s\n", c.username)
}
