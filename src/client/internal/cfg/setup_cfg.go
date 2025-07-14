package cfg

import (
	"chat/client/internal/app"
	"chat/client/internal/app/tcp"
	"chat/client/internal/app/udp"
	"fmt"
	"net"
)

func Setup() (*app.App, error) {
	flags := NewFlagsFromArgs()
	address := net.JoinHostPort(flags.IP, flags.Port)

	switch flags.ProtoType {
	case "tcp":
		return setupTCP(address)

	case "udp":
		return setupUDP(address)

	//case "http":
	//	// на будущее: реализуй http.NewClient(address), реализующий app.Client
	//	client := http.NewClient(address)
	//	return app.NewApp(client), nil

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
