package chord

import (
	"math/big"
	"time"
)

// update finger table entries.
func (n *Node) fixFinger(i int) {
	n.mu.RLock()
	m := len(n.FingerTable)
	n.mu.RUnlock()

	if m == 0 {
		return
	}

	start := computeFingerStart(n.ID, i, m)
	succ := n.Find(start)
	if succ != nil {
		n.mu.Lock()
		n.FingerTable[i] = succ
		n.mu.Unlock()
	}
}

/* n′ thinks it might be our predecessor.
n.notify(n′)
	if (predecessor is nil or n′ ∈ (predecessor, n))
	predecessor = n′;
*/
// notify en node att jag kanske är din predecessor
func (n *Node) notify(x *RemoteNode) {
	n.mu.Lock()
	defer n.mu.Unlock()

	if n.Predecessor == nil || n.between(n.Predecessor.ID, x.ID, n.ID) {
		n.Predecessor = x
	}
}

// start = (n + 2ⁱ) mod 2ᵐ
func computeFingerStart(id *big.Int, i int, m int) *big.Int {
	two := big.NewInt(2)                                     // 2
	twoI := new(big.Int).Exp(two, big.NewInt(int64(i)), nil) // 2^i
	// n+ 2^i                     n = id   i = finger index  m = hash len
	start := new(big.Int).Add(id, twoI) // n + 2^i
	// mod 2^m
	mod := new(big.Int).Exp(two, big.NewInt(int64(m)), nil) // 2^m

	start.Mod(start, mod) // n+ 2^i mod 2^m

	return start

}

func (n *Node) stabilize() {
	n.mu.RLock()
	succ := n.Successors[0]
	id := n.ID
	nAddr := n.Address
	n.mu.RUnlock()

	//we are lonely in the ring
	if succ == nil {
		return
	}

	//   x = successor.predecessor; fråga sucessor om sin predecessor
	var reply GetPredecessorReply
	ok := call(succ.Addr, "Node.GetPredecessorRPC", &GetPredecessorRequest{}, &reply)

	// If successor is dead, use backup from successor list
	if !ok {
		n.mu.Lock()
		// Shift successors left, removing failed one
		for i := 0; i < len(n.Successors)-1; i++ {
			n.Successors[i] = n.Successors[i+1]
		}
		n.Successors[len(n.Successors)-1] = nil

		// If all successors failed, we're alone
		if n.Successors[0] == nil {
			self := &RemoteNode{ID: n.ID, Addr: n.Address}
			n.Successors[0] = self
		}
		n.mu.Unlock()
		return
	}

	if reply.Node != nil {
		x := reply.Node

		if n.between(id, x.ID, succ.ID) { //  id < x < succ  se figure 7 i papperet
			n.mu.Lock()
			succ = x
			n.Successors[0] = x
			n.mu.Unlock()
		}
	}

	n.mu.RLock()
	succAddr := succ.Addr
	n.mu.RUnlock()

	// successor.notify(n);
	req := &NotifyRequest{Node: &RemoteNode{ID: id, Addr: nAddr}}

	var reply_2 NotifyReply

	call(succAddr, "Node.NotifyRPC", &req, &reply_2) // notify successor om mig själv
	//ex 21 -> 45 -> 10 då måste 10 notify 21 om 45

	// Update successor list by copying from successor
	// Successors[1..r-1] = successor's Successors[0..r-2]
	var succListReply GetSuccessorListReply
	if call(succAddr, "Node.GetSuccessorListRPC", &GetSuccessorListRequest{}, &succListReply) {
		n.mu.Lock()
		// Copy successor's list to fill our remaining slots
		maxCopy := len(n.Successors) - 1
		for i := 0; i < len(succListReply.Successors) && i < maxCopy; i++ {
			if succListReply.Successors[i] != nil {
				n.Successors[i+1] = succListReply.Successors[i]
			}
		}
		n.mu.Unlock()
	}

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
		var reply PingReply
		ok := call(pred.Addr, "Node.PingRPC", &PingRequest{}, &reply)
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
func (n *Node) RunMaintenance(ts int, tff int, tcp int) {
	// kör stablize fixFingers jämna mellan rum

	go func() {
		for {
			n.stabilize()
			time.Sleep(time.Duration(ts) * time.Millisecond)
		}
	}()

	go func() {
		for {
			n.checkPredecessor()
			time.Sleep(time.Duration(tcp) * time.Millisecond)
		}
	}()

	go func() {
		nextFix := 0

		for {
			n.mu.RLock()
			m := len(n.FingerTable)
			n.mu.RUnlock()

			if m > 0 {
				n.fixFinger(nextFix)
				nextFix = (nextFix + 1) % m
			}

			time.Sleep(time.Duration(tff) * time.Millisecond)
		}
	}()
}
