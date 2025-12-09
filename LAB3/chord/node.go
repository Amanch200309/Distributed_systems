package chord

import (
	"math/big"
	"sync"
)

/*
type RemoteNode struct {
    ID   *big.Int
    Addr string
}

type Node struct {
    mu           sync.RWMutex
    ID           *big.Int
    Address      string
    Predecessor  *RemoteNode
    Successors   []*RemoteNode
    FingerTable  []*RemoteNode
    Data         map[string]string   // filename → content
    M            int
}


*/

type RemoteNode struct {
	ID   *big.Int
	Addr string
}

// Node represents a node in the Chord DHT
type Node struct {
	id *big.Int // fix later TODO:   // hashed id, vi definerar vad id är sen kanske ip? vem vet // det är hash(ip)
	mu sync.RWMutex

	Address     string        // ip string
	Predecessor *RemoteNode   // node innan oss kanske Node typ istället så de blir linked list?
	Successor   *RemoteNode   // varför lista inte vet
	FingerTable []*RemoteNode // borde va lista på närmsta noder m lång ränkar ut index med 2 pow(n->m)
}

func NewNode(id *big.Int, addr string, m int) *Node {
	fingers := make([]*RemoteNode, m)
	self := &RemoteNode{ID: id, Addr: addr}
	return &Node{
		id:          id,
		Address:     addr,
		Predecessor: nil,
		Successor:   self,
		FingerTable: fingers,
	}
}
