package tcp

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
	conn     net.Conn
	username string
}

func NewClient(connect net.Conn) *Client {
	return &Client{
		conn: connect,
	}
}

func (cl *Client) ConnectToChat() {
	cl.registration()

	go func() {
		serverReader := bufio.NewReader(cl.conn)
		for {
			msg, err := serverReader.ReadString('\n')
			if err != nil {
				fmt.Println("Disconnected from server.")
				os.Exit(0)
			}

			msg = strings.TrimSpace(msg)
			if msg == "" {
				continue
			}

			var dtoMsg dto.TCPMessageDTO
			if err := json.Unmarshal([]byte(msg), &dtoMsg); err == nil && dtoMsg.Type != "" && dtoMsg.Type != "error" {
				cl.Print(dtoMsg)
				continue
			}

			var errMsg dto.ErrorDTO
			if err := json.Unmarshal([]byte(msg), &errMsg); err == nil && errMsg.Type == "error" {
				cl.Print(dto.TCPMessageDTO{Type: "error", Text: errMsg.Message})
				continue
			}

			fmt.Println(msg)
		}
	}()

	go cl.SendMessage()

	select {}
}

func (cl *Client) Print(msg dto.TCPMessageDTO) {
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

func (cl *Client) SendMessage() {
	consoleScanner := bufio.NewScanner(os.Stdin)
	fmt.Println("Enter text to send:")
	for consoleScanner.Scan() {
		text := consoleScanner.Text()

		if text == "/exit" {
			exitMsg := dto.TCPMessageDTO{
				Type: "exit",
				Name: cl.username,
			}
			data, _ := json.Marshal(exitMsg)
			cl.conn.Write(append(data, '\n'))
			return
		}

		if len(text) > 3 && text[:3] == "/w " {
			// Формат: /w username message
			parts := strings.SplitN(text[3:], " ", 2)
			if len(parts) == 2 {
				dst := parts[0]
				whisperText := parts[1]
				whisperMsg := dto.TCPMessageDTO{
					Type: "whisper",
					Name: cl.username,
					Text: whisperText,
					Dst:  dst,
					Time: time.Now().Format("2006/01/02 15:04:05"),
				}
				data, _ := json.Marshal(whisperMsg)
				cl.conn.Write(append(data, '\n'))
				continue
			}
		}

		broadcastMsg := dto.TCPMessageDTO{
			Type: "broadcast",
			Name: cl.username,
			Text: text,
			Time: time.Now().Format("2006/01/02 15:04:05"),
		}
		data, _ := json.Marshal(broadcastMsg)
		cl.conn.Write(append(data, '\n'))
	}
}

func (cl *Client) registration() {
	fmt.Print("Enter your name: ")
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Scan()
	cl.username = scanner.Text()

	regMsg := dto.TCPMessageDTO{
		Type: "register",
		Name: cl.username,
	}
	data, _ := json.Marshal(regMsg)
	cl.conn.Write(append(data, '\n'))
}
