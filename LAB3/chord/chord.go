package chord

import (
	"fmt"
	"time"
)

type Chord struct {
	M int // number of hash bits (usually 160)
}

func NewChord(m int) *Chord {
	return &Chord{M: m}
}


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
			// we found our successor
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


func (n *Node) Create() {
	n.mu.Lock()
	defer n.mu.Unlock()

	n.Predecessor = nil
	n.Successors[0] = &RemoteNode{ID: n.ID, Addr: n.Address}
}

