package main

import (
	"chat/client/internal/cfg"
	"log"
)

func main() {
	clientInstance, err := cfg.Setup()
	if err != nil {
		log.Fatal("Setup error:", err)
	}
	clientInstance.ConnectToChat()
}
