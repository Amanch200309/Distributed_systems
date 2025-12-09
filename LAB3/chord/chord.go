package chord

// returns true if elt is between start and end, accounting for the right

import (
	"math/big"
	"strings"
	"crypto/sha1"
)

type Chord struct {
	nodelist  map[string]*Node // kanske ändra till linked list sen inte array
	m        int    // hash längd
}


func NewChord(m int) *Chord {
    return &Chord{
        nodelist: make(map[string]*Node),
        m:        m,
    }
}


/* join a Chord ring containing node n′.
	n.join(n′)
	predecessor = nil;
	successor = n′.find successor(n);
*/

func (n *Node) Join(bootstrap *RemoteNode) {
    // We are not part of a ring yet
    n.mu.Lock()
    n.Predecessor = nil
    n.mu.Unlock()

    // Ask the bootstrap node to find our successor
    var reply FindReply
    args := &FindRequest{ID: n.id}

    ok := call(bootstrap.Addr, "Node.FindRPC", args, &reply)
    if !ok || reply.Node == nil {
        // bootstrap node unreachable -> create our own ring
        n.create()
        return
    }

    // Set successor
    n.mu.Lock()
    n.Successor = reply.Node
    n.mu.Unlock()

	n.initFingerTable()
}

// create a new Chord ring when client dont --ja/--jp
/* create a new Chord ring.
n.create()
	predecessor = nil;
	successor = n;
*/
func (n *Node) create() {
	n.mu.Lock()
	defer n.mu.Unlock()

	n.Predecessor = nil
	n.Successor = &RemoteNode{ID: n.id, Addr: n.Address} // noden pekar på sig själv ensam i ringen
}




// AddNode creates a node and either starts a new ring,
// or joins an existing ring via lookupnode
func (c *Chord) AddNode(addr string, port string, lookupnode *Node) *Node {
    n := c.CreateNode(addr, port)

    if lookupnode == nil {
        // First node in the ring
        n.create()
    } else {
        // Must wrap lookupnode in a RemoteNode
        bootstrap := &RemoteNode{
            ID:   lookupnode.id,
            Addr: lookupnode.Address,
        }
        n.Join(bootstrap)
    }

    // Start stabilize, fixFingers, checkPredecessor
    c.StartNode(n)

    return n
}


func (c *Chord) StartNode(n *Node) {
    // Start RPC server  "192.168.1.1:9000"
    parts := strings.Split(n.Address, ":")
    address, port := parts[0], parts[1]
    go startRPCServer(address, port) 
    // Start maintenance loop
    go n.runMaintenance()
}

func (c *Chord) CreateNode(ip string, port string) *Node {
	
	addr := ip + ":" + port
	id := hashKey(addr, c.m)
	
	return NewNode(id, addr, c.m)
}





/* 

ring := NewChord(8)
n1 := ring.AddNode("127.0.0.1:8001", nil)      // create ring
n2 := ring.AddNode("127.0.0.1:8002", n1)       // join
n3 := ring.AddNode("127.0.0.1:8003", n1)       // join
n4 := ring.AddNode("127.0.0.1:8004", n2)       // join using a different bootstrap


*/