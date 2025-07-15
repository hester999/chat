package cfg

import (
	"chat/client/internal/app"
	"chat/client/internal/app/http"
	"chat/client/internal/app/tcp"
	"chat/client/internal/app/udp"
	"fmt"
	"net"

	"github.com/gorilla/websocket"
)

func Setup() (*app.App, error) {
	flags := NewFlagsFromArgs()
	address := net.JoinHostPort(flags.IP, flags.Port)

	switch flags.ProtoType {
	case "tcp":
		return setupTCP(address)

	case "udp":
		return setupUDP(address)

	case "http":
		return setupHTTP(address)

	default:
		return nil, fmt.Errorf("unsupported protocol type: %s (expected: tcp, udp, http)", flags.ProtoType)
	}
}

func setupTCP(address string) (*app.App, error) {
	conn, err := net.Dial("tcp", address)
	if err != nil {
		fmt.Println("Error connecting (TCP):", err.Error())
		return nil, err
	}
	client := tcp.NewClient(conn)
	return app.NewApp(client), nil
}

func setupUDP(address string) (*app.App, error) {
	addr, err := net.ResolveUDPAddr("udp", address)
	if err != nil {
		fmt.Println("Error resolving UDP address:", err.Error())
		return nil, err
	}
	client := udp.NewClient(addr)
	return app.NewApp(client), nil
}

func setupHTTP(address string) (*app.App, error) {
	
	wsURL := fmt.Sprintf("ws://%s/ws", address)
	ws, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		fmt.Println("Error connecting (HTTP/WebSocket):", err.Error())
		return nil, err
	}
	client := http.NewClient(ws)
	return app.NewApp(client), nil
}
