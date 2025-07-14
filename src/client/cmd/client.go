package main

import (
	"chat/client/internal/cfg"
	"log"
)

func main() {
	appInstance, err := cfg.Setup()
	if err != nil {
		log.Fatal("Setup error:", err)
	}
	appInstance.ConnectToChat()
}
