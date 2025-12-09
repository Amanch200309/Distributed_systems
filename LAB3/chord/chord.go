package chord

// returns true if elt is between start and end, accounting for the right

type Chord struct {
	nodelist []Node // kanske ändra till linked list sen inte array
	m        int    // hash längd
}

func join() {

}

// create a new Chord ring when client dont --ja/--jp
func (n *Node) create() {
	n.mu.Lock()
	defer n.mu.Unlock()

	n.Predecessor = nil
	n.Successor = &RemoteNode{ID: n.id, Addr: n.Address} // noden pekar på sig själv ensam i ringen

}

/* create a new Chord ring.
n.create()
	predecessor = nil;
	successor = n;

// join a Chord ring containing node n′.
	n.join(n′)
	predecessor = nil;
	successor = n′.find successor(n);
*/
