package udp

import (
	"chat/server/internal/model"
	"chat/server/utils"
	"fmt"
	"net"
	"strings"
)

type UDPTransport struct {
	clients       map[string]*net.UDPAddr
	clientsByName map[string]*net.UDPAddr // Имя -> соединение2
	publicChan    chan model.IncomingMessage
	privateChan   chan model.IncomingMessage
	quit          chan struct{}
	buff          []byte
}

func NewTCPTransport() *UDPTransport {
	return &UDPTransport{
		clients:       make(map[string]*net.UDPAddr),
		clientsByName: make(map[string]*net.UDPAddr),
		publicChan:    make(chan model.IncomingMessage, 100),
		privateChan:   make(chan model.IncomingMessage, 100),
		quit:          make(chan struct{}),
		buff:          make([]byte, 4096),
	}
}

func (u *UDPTransport) Start() error {
	addr, _ := net.ResolveUDPAddr("udp", ":4545")
	conn, _ := net.ListenUDP("udp", addr)
	defer conn.Close()

	//go func() {
	//	for {
	//		select {
	//		case msg := <-u.publicChan:
	//			//u.BroadcastMessage(msg)
	//		case msg := <-u.privateChan:
	//			//u.SendPrivateMessage(msg)
	//		case <-u.quit:
	//			return
	//		}
	//	}
	//}()

	for {
		n, addr, err := conn.ReadFromUDP(u.buff)
		if err != nil {
			fmt.Errorf("%s", err)
			continue
		}
		go u.handleRequest(u.buff[:n], addr)
	}
}

func (u *UDPTransport) handleRequest(buf []byte, addr *net.UDPAddr) {
	ip := fmt.Sprintf("%s:%d", addr.IP, addr.Port)
	u.clients[ip] = addr
	strBuf := string(buf)
	var outMsg struct {
		Name string `json:"name"`
		Text string `json:"text"`
		Time string `json:"time"`
	}

	var reg struct {
		Type string `json:"type"`
		Name string `json:"name"`
	}

	if strings.Contains(strBuf, "register") {
		err := utils.JsonToStruct(strBuf, &reg)

		if err != nil {
			fmt.Errorf("%s", err)
		}
	} else if !strings.Contains(strBuf, "register") {

		err := utils.JsonToStruct(strBuf, &outMsg)
		if err != nil {
			fmt.Errorf("%s", err)
		}
	}
	
	fmt.Println(outMsg)
}
