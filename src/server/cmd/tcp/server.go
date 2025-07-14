package main

import (
	"chat/server/internal/app"
	"chat/server/internal/transport/tcp"
)

func main() {
	transport := tcp.NewTCPTransport()
	server := app.NewChatServer(transport)
	server.Start()
}
