package chord

import "math/big"

// ask node n to find the successor of id
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

// närmsta nod åt höger
func (n *Node) closestPrecedingNode(id *big.Int) *RemoteNode {
	n.mu.RLock()
	defer n.mu.RUnlock()

	// TODO
	for i := len(n.FingerTable) - 1; i >= 0; i-- {
		f := n.FingerTable[i]
		if f != nil && n.between(n.ID, f.ID, id) {
			return f
		}
	}
	return n.Successors[0]// 
}

func (n *Node) between(start, id, end *big.Int) bool {
	if start.Cmp(end) < 0 { // start < end
		/*
		   -1 if x < y;
		   0 if x == y;
		   +1 if x > y.
		*/
		return id.Cmp(start) > 0 && id.Cmp(end) < 0 // start < id < end
	} else {
		// modulo läge
		return id.Cmp(start) > 0 || id.Cmp(end) < 0 // id > start OR id < end
	}
}

// iterativt kolla alla noder i chord efter id och returna succ
// Find(id)
// --------------------
// Full iterative Chord lookup that always STARTS FROM THIS NODE.
// Used when WE are already part of the ring and want to resolve a key.
// This calls FindSuccessorRPC repeatedly starting from our own address.
//
// Why needed:
// - Local lookups (Lookup, StoreFile, PrintState) must begin at our node.
// - Correct because our node is already in the ring.
// - Cannot be used for Join(), because Join must start from the bootstrap node.
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

	return nil // failed after m steps
}

// FindFrom(start, id)
// --------------------
// Same iterative Chord lookup, BUT starts from ANY node we choose.
// Used during Join(), because we are NOT in the ring yet.
// We must begin lookup from the bootstrap node (the one given by --ja/--jp),
// not from ourselves.
//
// Why needed:
// - New nodes have no successor or finger table, so starting from themselves fails.
// - Join must ask the EXISTING ring who the successor of our ID is.
// - Allows Chord lookup to begin from a remote node instead of self.
//
//____
//func (n *Node) FindFrom(start *RemoteNode, id *big.Int) *RemoteNode {
//	current := start
//
//	n.mu.RLock()
//	maxSteps := len(n.FingerTable)
//	n.mu.RUnlock()
//
//	for i := 0; i < maxSteps; i++ {
//		var reply findSuccessorReply
//		ok := call(current.Addr, "Node.FindSuccessorRPC",
//			&findSuccessorRequest{ID: id}, &reply)
//		if !ok {
//			return nil
//		}
//		if reply.Found {
//			return reply.Node
//		}
//		current = reply.Node
//	}
//	return nil
//}

//____

/*

// ask node n to find the successor of id
n.find successor(id)
    if (id ∈ (n, successor])
        return successor;
    else
        n′ = closest preceding node(id );
        return n′.find successor(id);


// search the local table for the highest predecessor of id
n.closest preceding node(id)
    for i = m downto 1
        if (finger[i] ∈ (n, id))
            return finger[i];
    return n;

*/

/*
n.closest preceding node(id)
    for i = m downto 1
        if (finger[i] ∈ (n, id))
        return finger[i];
    return n;
*/
//pseudo code iterative for findSuccessor

/*
   // ask node n to find the successor of id
    // or a better node to continue the search with
    n.find_successor(id)
        if (id ∈ (n, successor])
            return true, successor;
        else
            return false, closest_preceding_node(id);

    // search the local table for the highest predecessor of id




    n.closest_preceding_node(id)
        // skip this loop if you do not have finger tables implemented yet
        for i = m downto 1
            if (finger[i] ∈ (n,id))
                return finger[i];
        return successor;

    // find the successor of id
    find(id, start)
        found, nextNode = false, start;
        i = 0
        while not found and i < maxSteps
            found, nextNode = nextNode.find_successor(id);
            i += 1            //https://www.cs.utahtech.edu/directorylisting.php?u=cs/3410/chord-starter/
        if found
            return nextNode;
        else
            report error;
*/
