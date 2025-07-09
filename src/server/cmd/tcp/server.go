package main

import (
	"chat/server/internal/transport/tcp"
	"chat/server/internal/app"
)

func main() {
	transport := tcp.NewTCPTransport()
	server := app.NewChatServer(transport)
	server.Start()
}
