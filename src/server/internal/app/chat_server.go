package app

import (
	"chat/server/internal/transport"
)

// ChatServer реализует бизнес-логику чата
// и работает с Transport через интерфейс

type ChatServer struct {
	transport transport.Transport
	quit      chan struct{}
}

func NewChatServer(tr transport.Transport) *ChatServer {
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
