package test

import (
	"chat/server/internal/model"
)

type MockTransport struct {
	StartFunc              func(address string) error
	StopFunc               func() error
	BroadcastMessageFunc   func(msg model.IncomingMessage) error
	SendPrivateMessageFunc func(msg model.IncomingMessage) error

	BroadcastCalls []model.IncomingMessage
	PrivateCalls   []model.IncomingMessage
}

func (m *MockTransport) Start(address string) error {
	if m.StartFunc != nil {
		return m.StartFunc(address)
	}
	return nil
}
func (m *MockTransport) Stop() error {
	if m.StopFunc != nil {
		return m.StopFunc()
	}
	return nil
}
func (m *MockTransport) BroadcastMessage(msg model.IncomingMessage) error {
	m.BroadcastCalls = append(m.BroadcastCalls, msg)
	if m.BroadcastMessageFunc != nil {
		return m.BroadcastMessageFunc(msg)
	}
	return nil
}
func (m *MockTransport) SendPrivateMessage(msg model.IncomingMessage) error {
	m.PrivateCalls = append(m.PrivateCalls, msg)
	if m.SendPrivateMessageFunc != nil {
		return m.SendPrivateMessageFunc(msg)
	}
	return nil
}
