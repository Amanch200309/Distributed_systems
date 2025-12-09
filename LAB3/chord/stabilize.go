package chord

import "math/big"

func stabilize() {

}

// update finger table entries.
func (n *Node) fixFingers() {
	n.mu.RLock()
	m := len(n.FingerTable)
	selfID := n.id //new(big.Int).Set(n.id) TODO:maybe
	n.mu.RUnlock()

	for i := 0; i < m; i++ {
		start := computeFingerStart(selfID, i, m)
		targetAddr := n.FingerTable[start].Addr
		req := findSuccessorRequest{ID: start}

		var found bool
		var rn RemoteNode
		reply := findSuccessorReply{Found: found, Node: rn}

		found, rn := n.FindSuccessorRPC(start)

		if found {
			n.mu.Lock()
			n.FingerTable[i] = rn
			n.mu.Unlock()
		}
	}
}

/* FRÅN CHATTTITIII
func (n *Node) fixFingers() {
	n.mu.RLock()
	m := len(n.FingerTable)
	selfID := new(big.Int).Set(n.id)
	n.mu.RUnlock()

	// mod = 2^m
	mod := new(big.Int).Exp(big.NewInt(2), big.NewInt(int64(m)), nil)

	for i := 0; i < m; i++ {
		// offset = 2^i
		offset := new(big.Int).Exp(big.NewInt(2), big.NewInt(int64(i)), nil)

		// id = (selfID + 2^i) mod 2^m
		id := new(big.Int).Add(selfID, offset)
		id.Mod(id, mod)

		found, rn := n.findSuccessor(id)
		if found {
			n.mu.Lock()
			n.FingerTable[i] = rn
			n.mu.Unlock()
		}
	}
}

*/

// notify en node att jag kanske är din predecessor
func (n *Node) notify(x *RemoteNode) {
	if n.Predecessor == nil || n.between(n.Predecessor.ID, x.ID, n.id) {
		n.Predecessor = x
	}
}

func computeFingerStart(id *big.Int, i int, m int) *big.Int {
	two := big.NewInt(2)
	//2^i
	twoI := new(big.Int).Exp(two, big.NewInt(int64(i)), nil) // 2^i
	// n+ 2^i                     n = id   i = finger index  m = hash len
	start := new(big.Int).Add(id, twoI)
	// mod 2^m
	mod := new(big.Int).Exp(two, big.NewInt(int64(m)), nil)

	start.Mod(start, mod)

	return start

}

/*
s = successor(n+2^(i-1)), where 1 ≤ i ≤ m
(and all arithmetic is modulo 2^m)
We call node s the ith finger
of node n, and denote it by n.finger[i]

n.finger[i]
n.finger[i]n.finger[i]n.finger[i]
/*


/*

// called periodically. verifies n’s immediate
// successor, and tells the successor about n.
n.stabilize()
	x = successor.predecessor;
	if (x ∈ (n, successor))
	successor = x;
	successor.notify(n);

// n′ thinks it might be our predecessor.
n.notify(n′)
	if (predecessor is nil or n′ ∈ (predecessor, n))
	predecessor = n′;

// called periodically. refreshes finger table entries.
// next stores the index of the next finger to fix.
n.fix fingers()
	next = next + 1 ;
	if (next > m)
	next = 1 ;
	finger[next] = find successor(n + 2next−1 );

// called periodically. checks whether predecessor has failed.
n.check predecessor()
	if (predecessor has failed)
	predecessor = nil;
*/

// while looping maintenance tasks
func runMaintenance() {
	// kör stablize fixFingers jämna mellan rum
}
