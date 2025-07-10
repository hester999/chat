package main

import (
	"chat/client/internal/app"
	"fmt"
	"net"
	"os"
)

//	func main() {
//		conn, err := net.Dial("tcp", "localhost:4545")
//		if err != nil {
//			fmt.Println("Error connecting:", err.Error())
//			os.Exit(1)
//		}
//		defer conn.Close()
//
//		client := app.NewClient(conn)
//
//		client.ConnectToChat()
//
// }
func main() {
	conn, err := net.Dial("udp", "localhost:4545")
	if err != nil {
		fmt.Println("Error connecting:", err.Error())
		os.Exit(1)
	}
	defer conn.Close()

	client := app.NewClient(conn)

	client.ConnectToChat()

}
