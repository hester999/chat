package test

type MockClient struct {
	ConnectToChatFunc func()
	SendMessageFunc   func()

	ConnectCalled     bool
	SendMessageCalled bool
}

func (m *MockClient) ConnectToChat() {
	m.ConnectCalled = true
	if m.ConnectToChatFunc != nil {
		m.ConnectToChatFunc()
	}
}

func (m *MockClient) SendMessage() {
	m.SendMessageCalled = true
	if m.SendMessageFunc != nil {
		m.SendMessageFunc()
	}
}
