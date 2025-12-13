package chord

import (
	"fmt"
	"math/big"
	"sort"
	"strings"
)

// GetNodeInfo returns comprehensive info about this node
type NodeInfo struct {
	ID           *big.Int
	Address      string
	Successor    *RemoteNode
	Predecessor  *RemoteNode
	FingerCount  int
	StoredFiles  []string
	SuccessorIDs []string // For successor list
}

// GetAllNodesInfo crawls the ring starting from this node
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

// PrintRingStatus shows a visual representation of the ring
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

// PrintFileDistribution shows which files are stored where
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

// PrintCompactRing shows a one-line ring visualization
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

// VerifyFileOwnership checks if files are stored on the correct nodes
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

// Static helper for ownership checking
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
