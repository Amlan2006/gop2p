package algorithm

import (
	"encoding/hex"
	"fmt"
	"math/big"
	"net"
	"sort"
	"time"
)

const Idlength = 16
const Idbits = 8

type NodeId [Idlength / Idbits]byte

func (id NodeId) String() string {
	return hex.EncodeToString(id[:])
}

func (id NodeId) XOR(otherId NodeId) *big.Int {
	var result NodeId
	for i := 0; i < len(id); i++ {
		result[i] = id[i] ^ otherId[i]
	}
	return new(big.Int).SetBytes(result[:][:])
}

const contactsize = 2

type Contacts struct {
	Id           NodeId
	Address      string
	last_seen_at time.Time
}
type KBuckets struct {
	contacts []Contacts
}

func (kb *KBuckets) Find(id NodeId) string {
	for _, c := range kb.contacts {
		if c.Id == id {
			return c.Address
		}

	}
	return ""
}
func (kb *KBuckets) Add(contact Contacts) bool {
	for _, c := range kb.contacts {
		if c.Id == contact.Id {
			return true
		}
	}

	if len(kb.contacts) < contactsize {
		kb.contacts = append(kb.contacts, contact)
		return true
	} else {
		evicted := kb.Evict()
		if evicted {
			kb.contacts = append(kb.contacts, contact)
			return true
		} else {
			fmt.Println("There is no space in this bucket")
			return false
		}
	}
}
func (kb *KBuckets) Evict() bool {
	sort.Slice(kb.contacts, func(i, j int) bool {
		return kb.contacts[i].last_seen_at.Before(kb.contacts[j].last_seen_at)
	})

	contact := kb.contacts[0]
	_, err := net.Dial("tcp", contact.Address)
	if err != nil {
		kb.contacts = append(kb.contacts[:0], kb.contacts[1:]...)
		return true
	}
	return false
}

type RoutingTable struct {
	buckets []KBuckets
	selfId  NodeId
}

func GetMSBIndex(distance *big.Int) int {
	if distance.Sign() == 0 {
		return -1
	}
	return distance.BitLen() - 1
}
func (rt *RoutingTable) Add(contact Contacts) {
	distance := rt.selfId.XOR(contact.Id)
	index := GetMSBIndex(distance)

	if index >= 0 {
		isAdded := rt.buckets[index].Add(contact)
		if isAdded {
			fmt.Printf("The contact is added to bucket : %d \n", index)

		} else {
			fmt.Println("Not added")
		}
	} else {
		fmt.Println("Same peer cant connect with each other")
	}
}
