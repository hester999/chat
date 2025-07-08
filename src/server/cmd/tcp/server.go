package main

import "server/internal/transport"

func main() {
	transport := transport.NewTransport()
	transport.Listen()


}
