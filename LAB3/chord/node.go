package chord

import (
	"math/big"
	"sync"
)

type RemoteNode struct {
	ID   *big.Int
	Addr string
}

// FileMetadata stores information about a file
type FileMetadata struct {
	Filename string
	Data     []byte
	Hash     *big.Int
}

// Node represents a node in the Chord DHT
type Node struct {
	ID *big.Int
	mu sync.RWMutex

	Address     string
	Predecessor *RemoteNode
	Successors  []*RemoteNode // successor list of size r
	FingerTable []*RemoteNode // size m
	m           int
	r           int
	Data        map[string][]byte        // hash -> data (for backward compatibility)
	Files       map[string]*FileMetadata // hash -> metadata (NEW!)
}

func NewNode(id *big.Int, addr string, m int, numSucc int) *Node {
	self := &RemoteNode{ID: id, Addr: addr}

	// successor list (size numSucc)
	succList := make([]*RemoteNode, numSucc)
	succList[0] = self // first element always self until join()

	// finger table (size m)
	fingers := make([]*RemoteNode, m)

	return &Node{
		ID:          id,
		Address:     addr,
		Predecessor: nil,
		Successors:  succList,
		FingerTable: fingers,
		m:           m,
		r:           numSucc,
		Data:        make(map[string][]byte),
		Files:       make(map[string]*FileMetadata),
	}
}
