package main

import (
	"chat/server/internal/cfg"
	"log"
)

func main() {

	serverInstance, err := cfg.Setup()

	if err != nil {
		log.Fatal(err)
	}
	serverInstance.Start()
}
