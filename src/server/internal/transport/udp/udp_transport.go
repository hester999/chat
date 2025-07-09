package udp
import (
	"net"
	"chat/server/internal/model"
)
type UDPTransport struct {
	clients       map[string]net.Conn
	clientsByName map[string]net.Conn // имя -> соединение
	publicChan    chan model.IncomingMessage
	privateChan   chan model.IncomingMessage
	quit          chan struct{}
}

func NewTCPTransport() *UDPTransport {
	return &UDPTransport{
		clients:       make(map[string]net.Conn),
		clientsByName: make(map[string]net.Conn),
		publicChan:    make(chan model.IncomingMessage, 100),
		privateChan:   make(chan model.IncomingMessage, 100),
		quit:          make(chan struct{}),
	}
}