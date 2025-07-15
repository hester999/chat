package test

import (
	"chat/client/internal/app"
	"testing"
)

func TestApp_ConnectToChat(t *testing.T) {
	mock := &MockClient{}
	app := app.NewApp(mock)

	app.ConnectToChat()

	if !mock.ConnectCalled {
		t.Error("ConnectToChat was not called on client")
	}
}

func TestApp_SendMessage(t *testing.T) {
	mock := &MockClient{}
	app := app.NewApp(mock)

	app.SendMessage()

	if !mock.SendMessageCalled {
		t.Error("SendMessage was not called on client")
	}
}
