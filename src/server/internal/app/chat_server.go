package app

import (
	"chat/server/internal/model"
)

type Transport interface {
	Start() error
	Stop() error
	BroadcastMessage(msg model.IncomingMessage) error
	SendPrivateMessage(msg model.IncomingMessage) error
}

type ChatServer struct {
	transport Transport
	quit      chan struct{}
}

func NewChatServer(tr Transport) *ChatServer {
	return &ChatServer{
		transport: tr,
		quit:      make(chan struct{}),
	}
}

func (s *ChatServer) Start() error {
	err := s.transport.Start()
	if err != nil {
		return err
	}

	return nil
}

// Stop завершает работу сервера
func (s *ChatServer) Stop() error {
	close(s.quit)
	return s.transport.Stop()
}
