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
	id *big.Int // hashed id
	mu sync.RWMutex

	Address     string            // ip:port string
	Predecessor *RemoteNode       // predecessor node
	Successor   *RemoteNode       // primary successor (for compatibility)
	Successors  []*RemoteNode     // successor list
	FingerTable []*RemoteNode     // finger table (m entries)
	Data        map[string]string // filename -> file content
	M           int               // hash bit length

	// Maintenance timing (milliseconds)
	StabilizeInterval        int
	FixFingersInterval       int
	CheckPredecessorInterval int
}

func NewNode(id *big.Int, addr string, m int, r int, ts int, tff int, tcp int) *Node {
	fingers := make([]*RemoteNode, m)
	successors := make([]*RemoteNode, r)
	self := &RemoteNode{ID: id, Addr: addr}
	return &Node{
		id:                       id,
		Address:                  addr,
		Predecessor:              nil,
		Successor:                self,
		Successors:               successors,
		FingerTable:              fingers,
		Data:                     make(map[string]string),
		M:                        m,
		StabilizeInterval:        ts,
		FixFingersInterval:       tff,
		CheckPredecessorInterval: tcp,
	}
}

func (n *Node) GetID() *big.Int {
	n.mu.RLock()
	defer n.mu.RUnlock()
	return n.id
}
