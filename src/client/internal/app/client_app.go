package app

type Client interface {
	ConnectToChat()
	SendMessage()
}

type App struct {
	client Client
}

func NewApp(c Client) *App {
	return &App{
		client: c,
	}
}

func (a *App) ConnectToChat() {
	a.client.ConnectToChat()
}

func (a *App) SendMessage() {
	a.client.SendMessage()
}
