package chord

import (
	"strings"
    "fmt"
)

// Local state
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
    if len(n.Successors) > 0 && n.Successors[0] != nil {
        output.WriteString(fmt.Sprintf(
            "Primary: ID=%s  Addr=%s\n",
            n.Successors[0].ID.Text(16),
            n.Successors[0].Addr,
        ))
    }

    for i, succ := range n.Successors {
        if succ != nil {
            output.WriteString(fmt.Sprintf(
                "  [%02d] ID=%s  Addr=%s\n",
                i,
                succ.ID.Text(16),
                succ.Addr,
            ))
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


