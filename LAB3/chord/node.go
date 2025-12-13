package chord

import (
	"fmt"
	"math/big"
	"sync"
	"time"
)

/*
RemoteNode represents a remote node in the Chord ring.

	ID: Node identifier in the hash space
	Addr: Network address (IP:port)
*/
type RemoteNode struct {
	ID   *big.Int
	Addr string
}

/*
FileMetadata stores information about a file in the DHT.

	Filename: Original filename
	Data: File contents as byte array
	Hash: Hash of the filename (key in DHT)
*/
type FileMetadata struct {
	Filename string
	Data     []byte
	Hash     *big.Int
}

/*
Node represents a node in the Chord DHT.

	ID: This node's identifier in the hash space
	mu: Mutex protecting concurrent access to node state
	Address: Network address of this node (IP:port)
	Predecessor: Previous node in the ring
	Successors: List of successor nodes for fault tolerance
	FingerTable: Routing table for efficient lookups
	m: Number of bits in hash space (finger table size)
	r: Number of successors to maintain
	Data: Storage for file data (hash -> data)
	Files: Storage for file metadata (hash -> metadata)
*/
type Node struct {
	ID *big.Int
	mu sync.RWMutex

	Address     string
	Predecessor *RemoteNode
	Successors  []*RemoteNode // successor list of size r
	FingerTable []*RemoteNode // size m
	m           int
	r           int
	Data        map[string][]byte // hash -> data (for backward compatibility)
	Files       map[string]*FileMetadata
}

/*
NewNode creates a new Chord node.

	Args: 	id (node identifier in hash space),
			addr (network address IP:port),
			m (number of hash bits),
			numSucc (number of successors to maintain)
	Returns: Pointer to new Node
	Initializes successor list with self, creates empty finger table
*/
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

/*
Join adds this node to an existing Chord ring.

	Args: 	bootstrap (existing node to join via)
	Returns: error if join fails
	Contacts bootstrap node to find successor, sets predecessor to nil,
	stabilization will handle finger table initialization
*/
func (n *Node) Join(bootstrap *RemoteNode) error {
	n.mu.Lock()
	n.Predecessor = nil
	n.mu.Unlock()

	var reply FindReply
	args := &FindRequest{ID: n.ID}

	// Try contacting bootstrap multiple times
	for i := 0; i < 5; i++ {
		ok := call(bootstrap.Addr, "Node.FindRPC", args, &reply)

		if ok && reply.Node != nil {
			// We found our successor
			n.mu.Lock()
			n.Successors[0] = reply.Node
			n.mu.Unlock()

			fmt.Printf("Joined ring successfully. Successor: %s\n", reply.Node.Addr)

			return nil
		}
		time.Sleep(200 * time.Millisecond)
	}

	return fmt.Errorf("Join failed: bootstrap %s unreachable", bootstrap.Addr)
}

/*
Create initializes a new Chord ring with this node as the only member.

	Sets predecessor to nil, successor to self, and initializes finger table
	All entries point to self initially until other nodes join
*/
func (n *Node) Create() {
	n.mu.Lock()
	defer n.mu.Unlock()

	n.Predecessor = nil
	n.Successors[0] = &RemoteNode{ID: n.ID, Addr: n.Address}

	self := &RemoteNode{ID: n.ID, Addr: n.Address}
	n.Successors[0] = self

	// Initialize all finger table entries to self
	for i := range n.FingerTable {
		n.FingerTable[i] = self
	}
	fmt.Println("Created new Chord ring")
}
