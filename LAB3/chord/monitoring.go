package chord

import (
	"fmt"
	"math/big"
	"sort"
	"strings"
)

/*
PrintState returns a formatted string of this node's local state.

	Returns: string containing node ID, address, successor list,
			 stored files, finger table, and predecessor information
	Used for the PrintState command to display node information
*/
func (n *Node) PrintState() string {
	n.mu.RLock()
	defer n.mu.RUnlock()

	var output strings.Builder

	// Node Information
	output.WriteString("==== Node Information ====\n")
	output.WriteString(fmt.Sprintf("ID:      %s\n", n.ID.Text(16)))
	output.WriteString(fmt.Sprintf("Address: %s\n\n", n.Address))

	// Successor List
	output.WriteString("==== Successor List ====\n")
	for i, succ := range n.Successors {
		if succ != nil {
			output.WriteString(fmt.Sprintf(
				"[%02d] ID=%s  Addr=%s\n",
				i,
				succ.ID.Text(16),
				succ.Addr,
			))
		}
	}
	output.WriteString("\n")

	// Stored Files
	output.WriteString("==== Stored Files ====\n")
	if len(n.Data) == 0 {
		output.WriteString("None\n")
	} else {
		for key := range n.Data {
			output.WriteString(fmt.Sprintf("  Hash: %s\n", key))
		}
	}
	output.WriteString("\n")

	// Finger Table
	output.WriteString("==== Finger Table ====\n")
	for i, finger := range n.FingerTable {
		if finger != nil {
			output.WriteString(fmt.Sprintf(
				"  [%02d] ID=%s  Addr=%s\n",
				i,
				finger.ID.Text(16),
				finger.Addr,
			))
		}
	}
	output.WriteString("\n")

	// Predecessor
	output.WriteString("==== Predecessor ====\n")
	if n.Predecessor != nil {
		output.WriteString(fmt.Sprintf(
			"ID=%s  Addr=%s\n",
			n.Predecessor.ID.Text(16),
			n.Predecessor.Addr,
		))
	} else {
		output.WriteString("None\n")
	}

	return output.String()
}

/*
NodeInfo holds comprehensive information about a Chord node.

	ID: Node identifier
	Address: Network address (IP:port)
	Successor: First successor node
	Predecessor: Previous node in ring
	FingerCount: Number of non-nil finger table entries
	StoredFiles: List of filenames stored on this node
	SuccessorIDs: IDs of all successors in successor list
*/
type NodeInfo struct {
	ID           *big.Int
	Address      string
	Successor    *RemoteNode
	Predecessor  *RemoteNode
	FingerCount  int
	StoredFiles  []string
	SuccessorIDs []string
}

/*
GetAllNodesInfo crawls the entire Chord ring and collects node information.

	Returns: array of NodeInfo for all reachable nodes in the ring
	Traverses ring via successor pointers, collecting information
	from each node using RPC calls
*/
func (n *Node) GetAllNodesInfo() []*NodeInfo {
	visited := make(map[string]bool)
	var nodes []*NodeInfo

	current := &RemoteNode{ID: n.ID, Addr: n.Address}
	startAddr := n.Address

	// Crawl the ring via successors
	for {
		if visited[current.Addr] {
			break
		}
		visited[current.Addr] = true

		var info NodeInfoReply
		ok := call(current.Addr, "Node.GetNodeInfoRPC", &NodeInfoRequest{}, &info)
		if !ok {
			break
		}

		// Collect file names
		fileNames := make([]string, 0, len(info.Files))
		for name := range info.Files {
			fileNames = append(fileNames, name)
		}
		sort.Strings(fileNames)

		// Build successor ID list
		succIDs := []string{}
		for _, s := range info.Successors {
			if s != nil {
				idStr := s.ID.Text(16)
				if len(idStr) > 8 {
					idStr = idStr[:8]
				}
				succIDs = append(succIDs, idStr)
			}
		}

		nodeInfo := &NodeInfo{
			ID:           info.ID,
			Address:      info.Address,
			Successor:    info.Successor,
			Predecessor:  info.Predecessor,
			FingerCount:  info.FingerCount,
			StoredFiles:  fileNames,
			SuccessorIDs: succIDs,
		}

		nodes = append(nodes, nodeInfo)

		// Move to next node
		if info.Successor == nil || info.Successor.Addr == startAddr {
			break
		}
		current = info.Successor
	}

	return nodes
}

/*
PrintRingStatus returns a formatted visualization of the Chord ring.

	Returns: string with formatted ring structure showing all nodes,
			 their successors, predecessors, and stored files
	Provides comprehensive view of ring topology and file distribution
*/
func (n *Node) PrintRingStatus() string {
	nodes := n.GetAllNodesInfo()
	if len(nodes) == 0 {
		return "Error: Could not fetch ring information"
	}

	// Sort nodes by ID
	sort.Slice(nodes, func(i, j int) bool {
		return nodes[i].ID.Cmp(nodes[j].ID) < 0
	})

	var sb strings.Builder

	sb.WriteString("\n")
	sb.WriteString("╔════════════════════════════════════════════════════════════════╗\n")
	sb.WriteString("║                      CHORD RING STATUS                         ║\n")
	sb.WriteString("╚════════════════════════════════════════════════════════════════╝\n")
	sb.WriteString("\n")

	// Summary stats
	totalFiles := 0
	for _, node := range nodes {
		totalFiles += len(node.StoredFiles)
	}

	sb.WriteString(fmt.Sprintf("  Total Nodes:  %d\n", len(nodes)))
	sb.WriteString(fmt.Sprintf("  Total Files:  %d\n", totalFiles))
	sb.WriteString("\n")

	// Ring visualization
	sb.WriteString("┌─────────────────────────────────────────────────────────────┐\n")
	sb.WriteString("│ RING STRUCTURE (ordered by ID)                              │\n")
	sb.WriteString("└─────────────────────────────────────────────────────────────┘\n")
	sb.WriteString("\n")

	for i, node := range nodes {
		// Node header
		idStr := node.ID.Text(16)
		if len(idStr) > 12 {
			idStr = idStr[:12]
		}
		idShort := idStr + "..."
		sb.WriteString(fmt.Sprintf("  [Node %d] ID: %s\n", i+1, idShort))
		sb.WriteString(fmt.Sprintf("           Addr: %s\n", node.Address))

		// Successor/Predecessor
		if node.Successor != nil {
			succStr := node.Successor.ID.Text(16)
			if len(succStr) > 12 {
				succStr = succStr[:12]
			}
			succShort := succStr + "..."
			sb.WriteString(fmt.Sprintf("           Succ: %s (%s)\n", succShort, node.Successor.Addr))
		} else {
			sb.WriteString("           Succ: nil\n")
		}

		if node.Predecessor != nil {
			predStr := node.Predecessor.ID.Text(16)
			if len(predStr) > 12 {
				predStr = predStr[:12]
			}
			predShort := predStr + "..."
			sb.WriteString(fmt.Sprintf("           Pred: %s (%s)\n", predShort, node.Predecessor.Addr))
		} else {
			sb.WriteString("           Pred: nil\n")
		}

		// Files stored
		fileCount := len(node.StoredFiles)
		if fileCount > 0 {
			sb.WriteString(fmt.Sprintf("           Files: %d stored\n", fileCount))
			for _, file := range node.StoredFiles {
				sb.WriteString(fmt.Sprintf("             • %s\n", file))
			}
		} else {
			sb.WriteString("           Files: none\n")
		}

		// Arrow to next node
		if i < len(nodes)-1 {
			sb.WriteString("           │\n")
			sb.WriteString("           ↓\n")
		} else {
			sb.WriteString("           │\n")
			sb.WriteString("           ↓ (wraps back to Node 1)\n")
		}
		sb.WriteString("\n")
	}

	return sb.String()
}

/*
PrintFileDistribution returns a table showing file locations across the ring.

	Returns: string with formatted table mapping filenames to nodes
	Shows which node stores each file in the ring
*/
func (n *Node) PrintFileDistribution() string {
	nodes := n.GetAllNodesInfo()
	if len(nodes) == 0 {
		return "Error: Could not fetch ring information"
	}

	var sb strings.Builder

	sb.WriteString("\n")
	sb.WriteString("╔════════════════════════════════════════════════════════════════╗\n")
	sb.WriteString("║                    FILE DISTRIBUTION MAP                       ║\n")
	sb.WriteString("╚════════════════════════════════════════════════════════════════╝\n")
	sb.WriteString("\n")

	// Collect all files across the ring
	type FileLocation struct {
		Filename string
		NodeAddr string
		NodeID   string
	}

	var allFiles []FileLocation
	for _, node := range nodes {
		for _, file := range node.StoredFiles {
			idStr := node.ID.Text(16)
			if len(idStr) > 12 {
				idStr = idStr[:12]
			}
			allFiles = append(allFiles, FileLocation{
				Filename: file,
				NodeAddr: node.Address,
				NodeID:   idStr + "...",
			})
		}
	}

	if len(allFiles) == 0 {
		sb.WriteString("  No files stored in the ring yet.\n\n")
		return sb.String()
	}

	// Sort by filename
	sort.Slice(allFiles, func(i, j int) bool {
		return allFiles[i].Filename < allFiles[j].Filename
	})

	// Print table
	sb.WriteString("  ┌─────────────────────────┬──────────────────┬─────────────────┐\n")
	sb.WriteString("  │ Filename                │ Node Address     │ Node ID         │\n")
	sb.WriteString("  ├─────────────────────────┼──────────────────┼─────────────────┤\n")

	for _, fl := range allFiles {
		filename := fl.Filename
		if len(filename) > 23 {
			filename = filename[:20] + "..."
		}
		sb.WriteString(fmt.Sprintf("  │ %-23s │ %-16s │ %-15s │\n",
			filename, fl.NodeAddr, fl.NodeID))
	}

	sb.WriteString("  └─────────────────────────┴──────────────────┴─────────────────┘\n")
	sb.WriteString(fmt.Sprintf("\n  Total: %d files distributed across %d nodes\n\n", len(allFiles), len(nodes)))

	return sb.String()
}

/*
PrintCompactRing returns a concise one-line ring visualization.

	Returns: string showing node IDs and file counts in ring order
	Provides quick overview of ring structure
*/
func (n *Node) PrintCompactRing() string {
	nodes := n.GetAllNodesInfo()
	if len(nodes) == 0 {
		return "Ring: [Error fetching nodes]"
	}

	// Sort by ID
	sort.Slice(nodes, func(i, j int) bool {
		return nodes[i].ID.Cmp(nodes[j].ID) < 0
	})

	var parts []string
	for _, node := range nodes {
		idStr := node.ID.Text(16)
		if len(idStr) > 8 {
			idStr = idStr[:8]
		}
		fileCount := len(node.StoredFiles)
		parts = append(parts, fmt.Sprintf("%s(%d)", idStr, fileCount))
	}

	return fmt.Sprintf("Ring: [%s] → (loops back)", strings.Join(parts, " → "))
}

/*
VerifyFileOwnership validates that files are stored on correct nodes.

	Returns: string showing verification results for each file
	Checks if each file's hash places it in the correct node's range
	Used for debugging and testing ring consistency
*/
func (n *Node) VerifyFileOwnership() string {
	nodes := n.GetAllNodesInfo()
	if len(nodes) == 0 {
		return "Error: Could not fetch ring information"
	}

	var sb strings.Builder

	sb.WriteString("\n")
	sb.WriteString("╔════════════════════════════════════════════════════════════════╗\n")
	sb.WriteString("║                  FILE OWNERSHIP VERIFICATION                   ║\n")
	sb.WriteString("╚════════════════════════════════════════════════════════════════╝\n")
	sb.WriteString("\n")

	correct := 0
	total := 0

	for _, node := range nodes {
		for _, filename := range node.StoredFiles {
			total++

			// Hash the filename to get the key
			key := HashKey(filename, 160)

			// Find who SHOULD own this key
			var shouldOwn *NodeInfo
			for _, candidate := range nodes {
				succ := candidate.Successor
				if succ != nil && betweenRightInclStatic(candidate.ID, key, succ.ID) {
					shouldOwn = candidate
					break
				}
			}

			// Check if current owner is correct
			if shouldOwn != nil && shouldOwn.ID.Cmp(node.ID) == 0 {
				correct++
				sb.WriteString(fmt.Sprintf("  ✓ %s: correctly stored at %s\n", filename, node.Address))
			} else if shouldOwn != nil {
				sb.WriteString(fmt.Sprintf("  ✗ %s: stored at %s, should be at %s\n",
					filename, node.Address, shouldOwn.Address))
			} else {
				sb.WriteString(fmt.Sprintf("  ? %s: could not determine owner\n", filename))
			}
		}
	}

	sb.WriteString("\n")
	sb.WriteString(fmt.Sprintf("  Result: %d/%d files correctly placed", correct, total))
	if correct == total {
		sb.WriteString(" ✓\n")
	} else {
		sb.WriteString(" ✗\n")
	}
	sb.WriteString("\n")

	return sb.String()
}

/*
betweenRightInclStatic checks if an ID is in interval (start, end].

	Args: 	start (interval start, exclusive),
			id (identifier to check),
			end (interval end, inclusive)
	Returns: bool (true if start < id <= end on the ring)
	Helper function for ownership verification with right-inclusive interval
*/
func betweenRightInclStatic(start, id, end *big.Int) bool {
	if id.Cmp(end) == 0 {
		return true
	}

	if start.Cmp(end) == 0 {
		return false
	}

	if start.Cmp(end) < 0 {
		return id.Cmp(start) > 0 && id.Cmp(end) < 0
	} else {
		return id.Cmp(start) > 0 || id.Cmp(end) < 0
	}
}
