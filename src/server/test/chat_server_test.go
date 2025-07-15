package test

import (
	"chat/server/internal/app"
	"chat/server/internal/model"
	"testing"
)

func TestChatServer_BroadcastMessage(t *testing.T) {
	mock := &MockTransport{}
	server := app.NewChatServer(mock, "localhost:1234")

	msg := model.IncomingMessage{From: "alice", Text: "hello"}
	err := server.BroadcastMessage(msg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(mock.BroadcastCalls) != 1 {
		t.Fatalf("expected 1 broadcast call, got %d", len(mock.BroadcastCalls))
	}
	if mock.BroadcastCalls[0] != msg {
		t.Errorf("broadcast call mismatch: want %+v, got %+v", msg, mock.BroadcastCalls[0])
	}
}

func TestChatServer_SendPrivateMessage(t *testing.T) {
	mock := &MockTransport{}
	server := app.NewChatServer(mock, "localhost:1234")

	msg := model.IncomingMessage{From: "bob", Text: "secret"}
	err := server.SendPrivateMessage(msg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(mock.PrivateCalls) != 1 {
		t.Fatalf("expected 1 private call, got %d", len(mock.PrivateCalls))
	}
	if mock.PrivateCalls[0] != msg {
		t.Errorf("private call mismatch: want %+v, got %+v", msg, mock.PrivateCalls[0])
	}
}
