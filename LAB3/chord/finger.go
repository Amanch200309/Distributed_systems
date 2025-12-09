package chord

func (n *Node) initFingerTable() {
	n.mu.RLock()
	m := len(n.FingerTable)
	n.mu.RUnlock()

	if m == 0{
		return
	}

	for i := 0; i < m; i++ {
		start := computeFingerStart(n.id, i, m)
		succ := n.Find(start)
		if succ != nil {
			n.mu.Lock()
			n.FingerTable[i] = succ
			n.mu.Unlock()
		}
	}


}
