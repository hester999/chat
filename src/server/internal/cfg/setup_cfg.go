package cfg

import (
	"chat/server/internal/app"
	"chat/server/internal/transport/http"
	"chat/server/internal/transport/tcp"
	"chat/server/internal/transport/udp"
	"fmt"
	"net"
)

func Setup() (*app.ChatServer, error) {
	flags := NewFlagsFromArgs()
	address := net.JoinHostPort(flags.IP, flags.Port)

	switch flags.ProtoType {
	case "tcp":
		return setupTCP(address), nil

	case "udp":
		return setupUDP(address), nil

	case "http":
		return setupHTTP(address), nil

	default:
		return nil, fmt.Errorf("unsupported protocol type: %s (expected: tcp, udp, http)", flags.ProtoType)
	}
}

func setupTCP(address string) *app.ChatServer {
	t := tcp.NewTCPTransport()
	server := app.NewChatServer(t, address)
	return server
}

func setupUDP(address string) *app.ChatServer {
	u := udp.NewUDPTransport()
	server := app.NewChatServer(u, address)
	return server
}

func setupHTTP(address string) *app.ChatServer {
	h := http.NewHTTPTransport()
	server := app.NewChatServer(h, address)
	return server
}
