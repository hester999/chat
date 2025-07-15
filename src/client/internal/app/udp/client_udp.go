package udp

import (
	"bufio"
	"chat/client/internal/dto"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"strings"
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

	go func() {
		buf := make([]byte, 4096)
		for {
			n, _, err := c.conn.ReadFromUDP(buf)
			if err != nil {
				fmt.Println("Error reading from server:", err)
				return
			}
			msg := strings.TrimSpace(string(buf[:n]))
			if msg == "" {
				continue
			}

			var dtoMsg dto.UDPMessageDTO
			if err := json.Unmarshal([]byte(msg), &dtoMsg); err == nil && dtoMsg.Type != "" && dtoMsg.Type != "error" {
				c.Print(dtoMsg)
				continue
			}

			var errMsg struct {
				Type    string `json:"type"`
				Message string `json:"message"`
			}
			if err := json.Unmarshal([]byte(msg), &errMsg); err == nil && errMsg.Type == "error" {
				c.Print(dto.UDPMessageDTO{Type: "error", Text: errMsg.Message})
				continue
			}

			fmt.Println(msg)
		}
	}()

	c.SendMessage()
}

func (c *Client) Print(msg dto.UDPMessageDTO) {
	const (
		ColorReset   = "\033[0m"
		ColorGreen   = "\033[32m"
		ColorBlue    = "\033[34m"
		ColorMagenta = "\033[35m"
		ColorRed     = "\033[31m"
	)

	timeStr := ""
	if msg.Time != "" {
		timeStr = fmt.Sprintf("%s[%s]%s ", ColorBlue, msg.Time, ColorReset)
	}
	nameStr := ""
	if msg.Name != "" {
		nameStr = fmt.Sprintf("%s%s%s", ColorGreen, msg.Name, ColorReset)
	}

	switch msg.Type {
	case "whisper":
		fmt.Println(msg.Type)
		fmt.Printf("%s%s[whisper]%s %s: %s\n",
			timeStr,
			ColorMagenta, ColorReset,
			nameStr,
			msg.Text,
		)
	case "broadcast":

		fmt.Printf("%s%s: %s\n",
			timeStr,
			nameStr,
			msg.Text,
		)
	case "error":
		fmt.Printf("%s[error]%s %s\n", ColorRed, ColorReset, msg.Text)
	default:
		fmt.Println(msg.Text)
	}
	fmt.Print("Enter text to send:\n")
}

func (c *Client) SendMessage() {
	consoleScanner := bufio.NewScanner(os.Stdin)
	for consoleScanner.Scan() {
		text := consoleScanner.Text()

		if text == "/exit" {
			exitMsg := dto.UDPMessageDTO{
				Type: "exit",
				Name: c.username,
			}
			data, _ := json.Marshal(exitMsg)
			c.conn.Write(data)
			return
		}

		if len(text) > 3 && text[:3] == "/w " {

			parts := strings.SplitN(text[3:], " ", 2)
			if len(parts) == 2 {
				dst := parts[0]
				whisperText := parts[1]
				whisperMsg := dto.UDPMessageDTO{
					Type: "whisper",
					Name: c.username,
					Text: whisperText,
					Dst:  dst,
					Time: time.Now().Format("2006/01/02 15:04:05"),
				}
				data, _ := json.Marshal(whisperMsg)
				c.conn.Write(data)
				continue
			}
		}

		broadcastMsg := dto.UDPMessageDTO{
			Type: "broadcast",
			Name: c.username,
			Text: text,
			Time: time.Now().Format("2006/01/02 15:04:05"),
		}
		data, _ := json.Marshal(broadcastMsg)
		c.conn.Write(data)
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

	regMsg := dto.UDPMessageDTO{
		Type: "register",
		Name: c.username,
	}
	regData, _ := json.Marshal(regMsg)
	c.conn.Write(regData)
	fmt.Printf("Registered as %s\n", c.username)
}
