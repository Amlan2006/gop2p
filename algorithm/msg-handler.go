package algorithm

import (
	"fmt"
	"log"
	"net"
)

type Messages struct {
	SenderId       NodeId
	RecieverId     NodeId
	MessageContent string
	MessageId      string
}
type MessagingPeer struct {
	ID           NodeId
	Messages     []Messages
	Port         int
	Address      string
	RoutingTable *RoutingTable
}

func (mp *MessagingPeer) StartServer() {
	ln, err := net.Listen("tcp", fmt.Sprintf(":%d", mp.Port))
	if err != nil {
		log.Fatal("Failed to start the server: %v", err)

	}
	defer ln.Close()
	for {
		conn, err := ln.Accept()
		if err != nil {
			continue
		}
		go mp.handleConnection(conn)
	}
}
func (mp *MessagingPeer) handleConnection(conn net.Conn) {
	defer conn.Close()
	buf := make([]byte, 1024)
	n, err := conn.Read(buf)
	if err != nil {
		return
	}
	var msg Messages
}
