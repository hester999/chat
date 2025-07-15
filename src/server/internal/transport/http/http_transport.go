package http

import (
	"chat/server/internal/dto"
	"chat/server/internal/model"
	"chat/server/utils"
	"fmt"
	"github.com/gorilla/websocket"
	"log"
	"net/http"
	"sync"
)

type RawMessage struct {
	Conn *websocket.Conn
	Data []byte
}
type Transport struct {
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

func NewHTTPTransport() *Transport {
	return &Transport{
		clients:       make(map[*websocket.Conn]bool),
		clientsByName: make(map[string]*websocket.Conn),
		publicChan:    make(chan model.IncomingMessage, 100),
		privateChan:   make(chan model.IncomingMessage, 100),
		messageChan:   make(chan RawMessage, 100),
		quit:          make(chan struct{}),
	}
}

func (h *Transport) Start(address string) error {
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
	log.Println("http server started on ", address)
	err := http.ListenAndServe(address, nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
	return nil
}

func (h *Transport) Stop() error {
	close(h.quit)
	return nil
}

func (h *Transport) BroadcastMessage(msg model.IncomingMessage) error {

	var dtoMsg dto.HTTPMessageDTO
	err := utils.JsonToStruct(msg.Text, &dtoMsg)
	if err != nil {
		return fmt.Errorf("send message error: %v", err)
	}

	outgoingMsg := model.HTTPMessage{
		Name: dtoMsg.Name,
		Text: dtoMsg.Text,
		Time: dtoMsg.Time,
	}

	responseDTO := dto.HTTPMessageDTO{
		Name: outgoingMsg.Name,
		Text: outgoingMsg.Text,
		Time: outgoingMsg.Time,
		Type: "broadcast",
	}

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

func (h *Transport) SendPrivateMessage(msg model.IncomingMessage) error {
	var dtoMsg dto.HTTPMessageDTO
	err := utils.JsonToStruct(msg.Text, &dtoMsg)
	if err != nil {
		return fmt.Errorf("send private message error: %v", err)
	}

	privateMsg := model.HTTPMessage{
		Name: dtoMsg.Name,
		Text: dtoMsg.Text,
		Time: dtoMsg.Time,
	}

	responseDTO := dto.HTTPMessageDTO{
		Name: privateMsg.Name,
		Text: privateMsg.Text,
		Time: privateMsg.Time,
		Type: "whisper",
		Dst:  dtoMsg.Dst,
	}

	h.mu.RLock()
	if toConn, ok := h.clientsByName[dtoMsg.Dst]; ok {
		toConn.WriteJSON(responseDTO)
	}
	if fromConn, ok := h.clientsByName[dtoMsg.Name]; ok && dtoMsg.Dst != dtoMsg.Name {
		fromConn.WriteJSON(responseDTO)
	}
	h.mu.RUnlock()
	return nil
}

func (h *Transport) handleConnections(w http.ResponseWriter, r *http.Request) {
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
		var msg dto.HTTPMessageDTO
		if err := ws.ReadJSON(&msg); err != nil {
			log.Println("ReadJSON error:", err)
			break
		}

		switch msg.Type {
		case "register":
			fmt.Println(123)
			h.handleRegister(ws, &username, msg)
		case "exit":
			h.handleExit(ws, username)
			return
		case "whisper":
			h.handleWhisper(msg)
		case "broadcast":
			fmt.Println(123)
			h.handleBroadcast(msg)
		}
	}

	h.mu.Lock()
	delete(h.clients, ws)
	if username != "" {
		delete(h.clientsByName, username)
	}
	h.mu.Unlock()
}

func (h *Transport) handleRegister(ws *websocket.Conn, username *string, msg dto.HTTPMessageDTO) {
	*username = msg.Name
	h.mu.Lock()

	h.clientsByName[*username] = ws
	h.mu.Unlock()
	fmt.Printf("User %s registered from %s\n", *username, ws.RemoteAddr())
}

func (h *Transport) handleExit(ws *websocket.Conn, username string) {
	h.mu.Lock()
	delete(h.clients, ws)
	if username != "" {
		delete(h.clientsByName, username)
	}
	h.mu.Unlock()
}

func (h *Transport) handleBroadcast(msg dto.HTTPMessageDTO) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	for _, client := range h.clientsByName {

		client.WriteJSON(msg)
	}
}

func (h *Transport) handleWhisper(msg dto.HTTPMessageDTO) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	if toConn, ok := h.clientsByName[msg.Dst]; ok {
		toConn.WriteJSON(msg)
	}
	if fromConn, ok := h.clientsByName[msg.Name]; ok && msg.Dst != msg.Name {
		fromConn.WriteJSON(msg)
	}
}
