package algorithm

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"sort"
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
	err = json.Unmarshal(buf[:n], &msg)
	if err != nil {
		return
	}
	mp.Messages = append(mp.Messages, msg)
}
func (mp *MessagingPeer) FindNode(targetId NodeId) []Contacts {
	return mp.RoutingTable.FindClosestContacts(targetId, contactsize)
}
func (mp *MessagingPeer) IterativeFindNode(targetId NodeId, knownPeers map[string]*MessagingPeer) []Contacts {
	visited := make(map[string]bool)
	shortList := mp.RoutingTable.FindClosestContacts(targetId, contactsize)
	closest := shortList
	for {
		newShortList := []Contacts{}
		for _, contact := range shortList {
			idStr := contact.Id.String()
			if visited[idStr] {
				continue
			}
			visited[idStr] = true
			peer := knownPeers[idStr]
			if peer == nil {
				continue
			}
			closerContacts := peer.FindNode((targetId))
			for _, c := range closerContacts {
				if !visited[c.Id.String()] {
					newShortList = append(newShortList, c)
				}
			}
		}
		all := append(closest, newShortList...)
		sort.Slice(all, func(i, j int) bool {
			return targetId.XOR(all[i].Id).Cmp(targetId.XOR(all[j].Id)) < 0
		})
		if len(all) > contactsize {
			all = all[:contactsize]
		}
		if EqualContacts(closest, all) {
			break
		}
		closest = all
		shortList = newShortList

	}
	return closest
}
func (mp *MessagingPeer) SendMessage(content string, peerId NodeId, network map[string]*MessagingPeer) (string, error) {
	if mp.ID == peerId {
		return "", fmt.Errorf("same peer can't send message to itself")
	}
	distance := mp.ID.XOR(peerId)
	index := GetMSBIndex(distance)
	peerAddress := mp.RoutingTable.buckets[index].Find(peerId)

	if peerAddress == "" {
		closest := mp.IterativeFindNode(peerId, network)
		for _, con := range closest {
			if con.Id == peerId {
				peerAddress = con.Address
				break
			}
		}
	}
	if peerAddress == "" {
		return "", fmt.Errorf("could not find peer address")
	}

	conn, err := net.Dial("tcp", peerAddress)
	if err != nil {
		return "", fmt.Errorf("failed to connect to peer: %w", err)
	}
	defer conn.Close()
	msg := &Messages{
		SenderId:       mp.ID,
		RecieverId:     peerId,
		MessageContent: content,
		MessageId:      "Random",
	}
	msgBytes, err := json.Marshal(msg)
	if err != nil {
		return "", fmt.Errorf("failed to serialize message: %w", err)
	}
	_, err = conn.Write(msgBytes)
	if err != nil {
		return "", fmt.Errorf("failed to send message: %w", err)

	}
	return "message sent successfully", nil
}
func NewMessagingPeer(port int, address string) *MessagingPeer {
	add := fmt.Sprintf("%s:%d", address, port)
	selfID := NewNodeId([]byte(add))
	return &MessagingPeer{
		ID:           selfID,
		Messages:     make([]Messages, 0),
		Port:         port,
		Address:      add,
		RoutingTable: NewRoutingTable(selfID),
	}
}
