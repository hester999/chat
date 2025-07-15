package app

import (
	"chat/server/internal/model"
)

type Transport interface {
	Start(address string) error
	Stop() error
	BroadcastMessage(msg model.IncomingMessage) error
	SendPrivateMessage(msg model.IncomingMessage) error
}

type ChatServer struct {
	transport  Transport
	quit       chan struct{}
	serverAddr string
}

func NewChatServer(tr Transport, addr string) *ChatServer {
	return &ChatServer{
		transport:  tr,
		quit:       make(chan struct{}),
		serverAddr: addr,
	}
}

func (s *ChatServer) Start() error {
	err := s.transport.Start(s.serverAddr)
	if err != nil {
		return err
	}

	return nil
}

func (s *ChatServer) Stop() error {
	close(s.quit)
	return s.transport.Stop()
}

func (s *ChatServer) BroadcastMessage(msg model.IncomingMessage) error {
	return s.transport.BroadcastMessage(msg)
}

func (s *ChatServer) SendPrivateMessage(msg model.IncomingMessage) error {
	return s.transport.SendPrivateMessage(msg)
}
