package http

import (
	"chat/server/internal/model"
	"chat/server/utils"
	"encoding/json"
	"fmt"
	"github.com/gorilla/websocket"
	_ "github.com/gorilla/websocket"
	"log"
	"net/http"
	"strings"
	"sync"
)

type RawMessage struct {
	Conn *websocket.Conn
	Data []byte
}
type HTTPTransport struct {
	clients       map[*websocket.Conn]bool
	clientsByName map[string]*websocket.Conn
	publicChan    chan model.IncomingMessage
	privateChan   chan model.IncomingMessage
	messageChan   chan RawMessage
	quit          chan struct{}
	mu            sync.RWMutex
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:     func(r *http.Request) bool { return true },
}

func NewHTTPTransport() *HTTPTransport {
	return &HTTPTransport{
		clients:       make(map[*websocket.Conn]bool),
		clientsByName: make(map[string]*websocket.Conn),
		publicChan:    make(chan model.IncomingMessage, 100),
		privateChan:   make(chan model.IncomingMessage, 100),
		messageChan:   make(chan RawMessage, 100),
		quit:          make(chan struct{}),
	}
}

func (h *HTTPTransport) Start() error {

	go func() {
		for {
			select {
			case msg := <-h.publicChan:
				h.BroadcastMessage(msg)
			case msg := <-h.privateChan:
				h.SendPrivateMessage(msg)
			case <-h.quit:
				return
			}
		}
	}()

	http.HandleFunc("/ws", h.handleConnections)
	return http.ListenAndServe(":8000", nil)
}
func (h *HTTPTransport) Stop() error {

}
func (h *HTTPTransport) BroadcastMessage(msg model.IncomingMessage) error {

	var msgDTO struct {
		Name string `json:"name"`
		Text string `json:"text"`
		Time string `json:"time"`
	}

	err := utils.JsonToStruct(msg.Text, &msgDTO)
	if err != nil {
		return fmt.Errorf("send message error: %v", err)
	}

	outgoingMsg := model.OutgoingMessage{
		Name:    msgDTO.Name,
		Text:    msgDTO.Text,
		Time:    msgDTO.Time,
		Private: false,
	}

	var responseDTO struct {
		Name    string `json:"name"`
		Text    string `json:"text"`
		Time    string `json:"time"`
		Private bool   `json:"private,omitempty"`
	}

	responseDTO.Name = outgoingMsg.Name
	responseDTO.Text = outgoingMsg.Text
	responseDTO.Time = outgoingMsg.Time
	responseDTO.Private = outgoingMsg.Private

	h.mu.RLock()
	for _, client := range h.clientsByName {
		err = client.WriteJSON(responseDTO)
		if err != nil {
			return fmt.Errorf("send message for client error: %s", err)
		}
	}
	h.mu.RUnlock()
	return nil
}

func (h *HTTPTransport) SendPrivateMessage(msg model.IncomingMessage) error {

}

func (h *HTTPTransport) handleConnections(w http.ResponseWriter, r *http.Request) {
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Fatal(err)
	}
	defer ws.Close()

	h.mu.Lock()
	h.clients[ws] = true
	h.mu.Unlock()

	var username string

	for {
		messageType, message, err := ws.ReadMessage()
		if err != nil {
			log.Println(err)
			break
		}
		if messageType != websocket.TextMessage {
			continue
		}

		var msgDTO struct {
			Type    string `json:"type,omitempty"`
			Name    string `json:"name"`
			Text    string `json:"text,omitempty"`
			Time    string `json:"time,omitempty"`
			Private bool   `json:"private,omitempty"`
		}

		if err := utils.JsonToStruct(string(message), &msgDTO); err != nil {
			log.Println("JSON decode error:", err)
			continue
		}

		if msgDTO.Type == "register" && username == "" {
			username = msgDTO.Name
			h.mu.Lock()
			if _, ok := h.clients[ws]; ok {
				h.mu.Unlock()
				continue
			}
			h.clientsByName[username] = ws
			h.mu.Unlock()
			fmt.Printf("User %s registered from %s\n", username, ws.RemoteAddr())
			continue
		}

		incoming := model.IncomingMessage{From: msgDTO.Name, Text: string(message)}
		if strings.HasPrefix(msgDTO.Text, "/w ") {
			h.privateChan <- incoming
		} else if utils.IsExitCommand(msgDTO.Text) {
			h.mu.Lock()
			delete(h.clients, ws)
			if username != "" {
				delete(h.clientsByName, username)
			}
			h.mu.Unlock()
			return
		} else {
			h.publicChan <- incoming
		}
	}

	h.mu.Lock()
	delete(h.clients, ws)
	if username != "" {
		delete(h.clientsByName, username)
	}
	h.mu.Unlock()
}
