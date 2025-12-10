package chord

import (
	"fmt"
	"io/ioutil"
)

// Lookup finds the node responsible for a given filename and retrieves the file
func (n *Node) Lookup(filename string) {
	// Hash the filename to get the key
	key := hashKey(filename, n.M)

	// Find the successor node responsible for this key
	succ := n.Find(key)

	if succ == nil {
		fmt.Printf("Error: Could not find successor for file '%s'\n", filename)
		return
	}

	// Query that node for the file
	var reply GetFileReply
	args := &GetFileRequest{Filename: filename}

	ok := call(succ.Addr, "Node.GetFileRPC", args, &reply)
	if !ok {
		fmt.Printf("Error: Could not contact node %s\n", succ.Addr)
		return
	}

	if reply.Found {
		fmt.Printf("File: %s\n", filename)
		fmt.Printf("Owner: Node %s (ID: %s)\n", succ.Addr, succ.ID.Text(16))
		fmt.Printf("Content:\n%s\n", reply.Content)
	} else {
		fmt.Printf("File '%s' not found in the Chord ring\n", filename)
	}
}

// StoreFile reads a file from local disk and stores it in the Chord ring
func (n *Node) StoreFile(filepath string, filename string) {
	// Read file content from local disk
	content, err := ioutil.ReadFile(filepath)
	if err != nil {
		fmt.Printf("Error reading file: %v\n", err)
		return
	}

	// Hash the filename to get the key
	key := hashKey(filename, n.M)

	// Find the successor node responsible for this key
	succ := n.Find(key)

	if succ == nil {
		fmt.Printf("Error: Could not find successor for file '%s'\n", filename)
		return
	}

	// Store the file at that node
	var reply StoreFileReply
	args := &StoreFileRequest{
		Filename: filename,
		Content:  string(content),
	}

	ok := call(succ.Addr, "Node.StoreFileRPC", args, &reply)
	if !ok {
		fmt.Printf("Error: Could not contact node %s\n", succ.Addr)
		return
	}

	if reply.Success {
		fmt.Printf("File '%s' stored successfully at node %s (ID: %s)\n",
			filename, succ.Addr, succ.ID.Text(16))
	} else {
		fmt.Printf("Error: Failed to store file '%s'\n", filename)
	}
}

// PrintState outputs the current state of the node
func (n *Node) PrintState() {
	n.mu.RLock()
	defer n.mu.RUnlock()

	fmt.Println("========================================")
	fmt.Println("Node State")
	fmt.Println("========================================")

	// Node's own information
	fmt.Printf("Self:\n")
	fmt.Printf("  ID: %s\n", n.id.Text(16))
	fmt.Printf("  Address: %s\n", n.Address)

	// Stored files
	fmt.Printf("\nStored Files:\n")
	if len(n.Data) == 0 {
		fmt.Printf("  (none)\n")
	} else {
		for filename := range n.Data {
			fmt.Printf("  - %s\n", filename)
		}
	}

	// Predecessor
	fmt.Printf("\nPredecessor:\n")
	if n.Predecessor == nil {
		fmt.Printf("  (none)\n")
	} else {
		fmt.Printf("  ID: %s\n", n.Predecessor.ID.Text(16))
		fmt.Printf("  Address: %s\n", n.Predecessor.Addr)
	}

	// Successor list
	fmt.Printf("\nSuccessor List:\n")
	if n.Successor != nil {
		fmt.Printf("  [0] ID: %s, Address: %s\n", n.Successor.ID.Text(16), n.Successor.Addr)
	}
	for i, succ := range n.Successors {
		if succ != nil {
			fmt.Printf("  [%d] ID: %s, Address: %s\n", i+1, succ.ID.Text(16), succ.Addr)
		}
	}

	// Finger table - show only unique entries or important ones
	fmt.Printf("\nFinger Table (unique entries):\n")
	seen := make(map[string]bool)
	count := 0
	for i, finger := range n.FingerTable {
		if finger != nil {
			key := finger.ID.Text(16) + finger.Addr
			// Show first occurrence of each unique node, or important indices
			if !seen[key] || i == 0 || i == len(n.FingerTable)/4 || i == len(n.FingerTable)/2 || i == len(n.FingerTable)-1 {
				if !seen[key] {
					count++
					seen[key] = true
				}
				start := computeFingerStart(n.id, i, n.M)
				fmt.Printf("  [%d] start: %s..., node: %s... (%s)\n",
					i, start.Text(16)[:16], finger.ID.Text(16)[:16], finger.Addr)
			}
		}
	}
	fmt.Printf("  Total: %d entries, %d unique nodes\n", len(n.FingerTable), count)

	fmt.Println("========================================")
}

// RPC methods for file operations

type GetFileRequest struct {
	Filename string
}

type GetFileReply struct {
	Found   bool
	Content string
}

type StoreFileRequest struct {
	Filename string
	Content  string
}

type StoreFileReply struct {
	Success bool
}

func (n *Node) GetFileRPC(args *GetFileRequest, reply *GetFileReply) error {
	n.mu.RLock()
	content, found := n.Data[args.Filename]
	n.mu.RUnlock()

	reply.Found = found
	reply.Content = content
	return nil
}

func (n *Node) StoreFileRPC(args *StoreFileRequest, reply *StoreFileReply) error {
	n.mu.Lock()
	n.Data[args.Filename] = args.Content
	n.mu.Unlock()

	reply.Success = true
	return nil
}
