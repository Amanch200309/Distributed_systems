package chord

import "math/big"

/*
findSuccessor finds the successor node for a given ID.

	Args: 	id (identifier to find successor for)
	Returns: bool (true if successor found, false if need to continue search),
			 RemoteNode (successor if found, or next node to query)
	If ID is between this node and successor, returns successor
	Otherwise returns closest preceding node to continue search
*/
func (n *Node) findSuccessor(id *big.Int) (bool, *RemoteNode) {
	n.mu.RLock()
	succ := n.Successors[0]
	selfID := n.ID
	n.mu.RUnlock()

	// If ID lies between n and successor, successor is correct
	if n.between(selfID, id, succ.ID) || id.Cmp(succ.ID) == 0 {
		return true, succ
	}

	// Otherwise return the next node to contact (not the answer)
	next := n.closestPrecedingNode(id)
	return false, next
}

/*
closestPrecedingNode finds the closest finger preceding the given ID.

	Args: 	id (target identifier)
	Returns: RemoteNode closest to id in finger table
	Searches finger table from highest to lowest, returns first node
	between this node and target ID
*/
func (n *Node) closestPrecedingNode(id *big.Int) *RemoteNode {
	n.mu.RLock()
	defer n.mu.RUnlock()

	for i := len(n.FingerTable) - 1; i >= 0; i-- {
		f := n.FingerTable[i]
		if f != nil && n.between(n.ID, f.ID, id) {
			return f
		}
	}
	return n.Successors[0]
}

/*
between checks if an ID lies in the interval (start, end) on the Chord ring.

	Args: 	start (interval start),
			id (identifier to check),
			end (interval end)
	Returns: bool (true if start < id < end, considering ring wraparound)
	Handles modulo arithmetic for circular identifier space
*/
func (n *Node) between(start, id, end *big.Int) bool {
	if start.Cmp(end) < 0 { // start < end (no wraparound)
		return id.Cmp(start) > 0 && id.Cmp(end) < 0 // start < id < end
	} else { // Wraparound case
		return id.Cmp(start) > 0 || id.Cmp(end) < 0 // id > start OR id < end
	}
}

/*
Find performs iterative Chord lookup starting from this node.

	Args: 	id (identifier to find successor for)
	Returns: RemoteNode that is successor of id, nil if lookup fails
	Performs iterative lookup by repeatedly calling FindSuccessorRPC
	on the closest preceding node until successor is found
	Used for local lookups (Lookup, StoreFile) when node is in the ring
*/

// Find finds the successor of a given ID using iterative lookup.
func (n *Node) Find(id *big.Int) *RemoteNode {
	current := &RemoteNode{ID: n.ID, Addr: n.Address}

	// m is the size of the finger table (equal to hash length)
	n.mu.RLock()
	maxSteps := len(n.FingerTable)
	n.mu.RUnlock()

	for i := 0; i < maxSteps; i++ {
		var reply FindSuccessorReply
		ok := call(current.Addr, "Node.FindSuccessorRPC", &FindSuccessorRequest{ID: id}, &reply)
		if !ok {
			return nil
		}

		if reply.Found {
			return reply.Node
		}

		current = reply.Node
	}

	return nil // Failed after m steps
}
