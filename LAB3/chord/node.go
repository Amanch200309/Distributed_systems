package chord

import "sync"

// Node represents a node in the Chord DHT
type Node struct {
	id string // fix later TODO:   // hashed id, vi definerar vad id är sen kanske ip? vem vet // det är hash(ip)
	mu sync.RWMutex

	Address     string   // ip string
	Predecessor string   // node innan oss kanske Node typ istället så de blir linked list?
	Successors  []string // varför lista inte vet
	FingerTable []string // borde va lista på närmsta noder m lång ränkar ut index med 2 pow(n->m)
}

func NewNode(id string, addr string, m int) *Node {
	fingers := make([]string, m)
	return &Node{
		id:          id,
		Address:     addr,
		FingerTable: fingers,
	}
}
