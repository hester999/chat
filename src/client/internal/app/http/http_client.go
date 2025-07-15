package http

import (
	"bufio"
	"chat/client/internal/dto"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/gorilla/websocket"
)

type Client struct {
	ws       *websocket.Conn
	username string
}

func NewClient(ws *websocket.Conn) *Client {
	return &Client{
		ws: ws,
	}
}

func (c *Client) ConnectToChat() {
	c.registration()

	go func() {
		for {
			var msg dto.HTTPMessageDTO
			err := c.ws.ReadJSON(&msg)
			fmt.Println(msg)
			if err != nil {
				fmt.Println("Disconnected from server:", err)
				os.Exit(0)
			}
			c.print(msg)
		}
	}()

	go c.SendMessage()

	select {}
}

func (c *Client) print(msg dto.HTTPMessageDTO) {
	const (
		ColorReset   = "\033[0m"
		ColorGreen   = "\033[32m"
		ColorBlue    = "\033[34m"
		ColorMagenta = "\033[35m"
	)

	timeStr := fmt.Sprintf("%s[%s]%s", ColorBlue, msg.Time, ColorReset)
	nameStr := fmt.Sprintf("%s%s%s", ColorGreen, msg.Name, ColorReset)

	if msg.Type == "whisper" {
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
}

func (c *Client) SendMessage() {
	consoleScanner := bufio.NewScanner(os.Stdin)
	fmt.Println("Enter text to send:")
	for consoleScanner.Scan() {
		text := consoleScanner.Text()

		if text == "/exit" {
			exitMsg := dto.HTTPMessageDTO{
				Type: "exit",
				Name: c.username,
			}
			c.ws.WriteJSON(exitMsg)
			c.ws.Close()
			return
		}

		if len(text) > 3 && text[:3] == "/w " {

			parts := bufio.NewScanner(bufio.NewReader(strings.NewReader(text[3:])))
			parts.Split(bufio.ScanWords)
			if parts.Scan() {
				dst := parts.Text()
				whisperText := text[3+len(dst)+1:] // +1 для пробела

				whisperMsg := dto.HTTPMessageDTO{
					Type: "whisper",
					Name: c.username,
					Text: whisperText,
					Dst:  dst,
					Time: time.Now().Format("2006/01/02 15:04:05"),
				}
				c.ws.WriteJSON(whisperMsg)
			}
		} else {

			broadcastMsg := dto.HTTPMessageDTO{
				Type: "broadcast",
				Name: c.username,
				Text: text,
				Time: time.Now().Format("2006/01/02 15:04:05"),
			}
			c.ws.WriteJSON(broadcastMsg)
		}
	}
}

func (c *Client) registration() {
	fmt.Print("Enter your name: ")
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Scan()
	c.username = scanner.Text()

	regMsg := dto.HTTPMessageDTO{
		Type: "register",
		Name: c.username,
	}
	c.ws.WriteJSON(regMsg)
}
