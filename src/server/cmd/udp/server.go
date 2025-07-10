package main

import "chat/server/internal/transport/udp"

func main() {
	t := udp.NewTCPTransport()
	err := t.Start()
	if err != nil {
		return
	}

}
