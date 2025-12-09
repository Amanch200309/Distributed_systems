package chord

import (
	"math/big"
)

// update finger table entries.
func (n *Node) fixFingers() {
	for i := range n.FingerTable {
		start := computeFingerStart(n.id, i, len(n.FingerTable))
		successor := n.Find(start)
		if successor != nil {
			n.mu.Lock()
			n.FingerTable[i] = successor
			n.mu.Unlock()
		}
	}
}

/* n′ thinks it might be our predecessor.
n.notify(n′)
	if (predecessor is nil or n′ ∈ (predecessor, n))
	predecessor = n′;
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

func (n *Node) stabilize() {
	n.mu.RLock()
	succ := n.Successor
	id := n.id
	nAddr := n.Address
	n.mu.RUnlock()

	//we are lonely in the ring
	if succ == nil {
		return
	}

	//   x = successor.predecessor; fråga sucessor om sin predecessor
	var reply getPredecessorReply
	ok := call(succ.Addr, "Node.GetPredecessorRPC", &getPredecessorRequest{}, &reply)
	if ok && reply.Node != nil {
		x := reply.Node

		if n.between(id, x.ID, succ.ID) { //  id < x < succ  se figure 7 i papperet
			n.mu.Lock()
			succ = x
			n.Successor = x
			n.mu.Unlock()
		}
	}

	n.mu.RLock()
	succAddr := succ.Addr
	n.mu.RUnlock()

	// successor.notify(n);
	req := &notifyRequest{Node: &RemoteNode{ID: id, Addr: nAddr}}

	var reply_2 notifyReply

	call(succAddr, "Node.NotifyRPC", &req, &reply_2) // notify successor om mig själv
	//ex 21 -> 45 -> 10 då måste 10 notify 21 om 45

}

/*
n.check predecessor()

	if (predecessor has failed)
	predecessor = nil;
*/
func (n *Node) checkPredecessor() {
	n.mu.RLock()
	pred := n.Predecessor
	n.mu.RUnlock()

	if pred != nil {
		var reply pingReply
		ok := call(pred.Addr, "Node.PingRPC", &pingRequest{}, &reply)
		if !ok || !reply.Alive { // if call failed or not alive
			n.mu.Lock()
			n.Predecessor = nil
			n.mu.Unlock()
		}
	}

}

/*

n.stabilize()
	x = successor.predecessor;
	if (x ∈ (n, successor))
	successor = x;
	successor.notify(n);



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
