package main

import (
	"chat/server/internal/transport/udp"
)

func main() {
	transport := udp.NewUDPTransport()
	transport.Start()
	
}
