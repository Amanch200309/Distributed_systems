package main

import (
	"bufio"
	"flag"
	"fmt"
	"math/big"
	"os"
	"strings"

	"github.com/Amanch200309/Distributed_systems/LAB3/chord"
)

func main() {
	// Define command-line flags
	var (
		address    = flag.String("a", "", "IP address to bind to (required)")
		port       = flag.String("p", "", "Port to bind to (required)")
		joinAddr   = flag.String("ja", "", "IP address of node to join")
		joinPort   = flag.String("jp", "", "Port of node to join")
		ts         = flag.Int("ts", 0, "Time in ms between stabilize invocations (1-60000)")
		tff        = flag.Int("tff", 0, "Time in ms between fix_fingers invocations (1-60000)")
		tcp        = flag.Int("tcp", 0, "Time in ms between check_predecessor invocations (1-60000)")
		r          = flag.Int("r", 0, "Number of successors to maintain (1-32)")
		idOverride = flag.String("i", "", "Override ID (40 hex chars, optional)")
	)

	flag.Parse()

	// Validate required arguments
	if *address == "" || *port == "" {
		fmt.Println("Error: -a and -p are required")
		flag.Usage()
		os.Exit(1)
	}

	if *ts < 1 || *ts > 60000 {
		fmt.Println("Error: --ts must be in range [1, 60000]")
		os.Exit(1)
	}

	if *tff < 1 || *tff > 60000 {
		fmt.Println("Error: --tff must be in range [1, 60000]")
		os.Exit(1)
	}

	if *tcp < 1 || *tcp > 60000 {
		fmt.Println("Error: --tcp must be in range [1, 60000]")
		os.Exit(1)
	}

	if *r < 1 || *r > 32 {
		fmt.Println("Error: -r must be in range [1, 32]")
		os.Exit(1)
	}

	// Check if both --ja and --jp are provided together
	if (*joinAddr != "" && *joinPort == "") || (*joinAddr == "" && *joinPort != "") {
		fmt.Println("Error: --ja and --jp must both be specified together")
		os.Exit(1)
	}

	// Create Chord ring with m=160 (SHA-1 produces 160-bit hashes)
	m := 160
	ring := chord.NewChord(m)

	// Create this node
	var node *chord.Node

	if *joinAddr == "" && *joinPort == "" {
		// Create new ring
		fmt.Printf("Creating new Chord ring...\n")
		node = ring.AddNode(*address, *port, *r, *ts, *tff, *tcp, nil)

		// Override ID if specified
		if *idOverride != "" {
			id := new(big.Int)
			id.SetString(*idOverride, 16)
			// This would require modifying the node's ID, which we don't expose
			// For simplicity, we'll skip this feature for now
			fmt.Printf("Warning: ID override not fully implemented\n")
		}

		fmt.Printf("Node created with ID: %s\n", node.GetID().Text(16))
		fmt.Printf("Listening on: %s:%s\n", *address, *port)
	} else {
		// Join existing ring
		fmt.Printf("Joining existing Chord ring at %s:%s...\n", *joinAddr, *joinPort)

		// Create a temporary bootstrap reference
		// We need to contact this node, but we don't have a full Node object
		// So we'll create the node and have it join
		node = ring.CreateNode(*address, *port, *r, *ts, *tff, *tcp)

		// Start RPC server first
		ring.StartNode(node)

		// Join via bootstrap node
		bootstrapAddr := *joinAddr + ":" + *joinPort
		bootstrap := &chord.RemoteNode{
			ID:   nil, // Will be discovered
			Addr: bootstrapAddr,
		}

		// We need to discover the bootstrap node's ID first
		var reply chord.GetPredecessorReply
		ok := chord.CallPublic(bootstrapAddr, "Node.GetPredecessorRPC", &chord.GetPredecessorRequest{}, &reply)
		if !ok {
			// Try a different RPC to get node info
			// For now, just use a dummy ID - the Join will fix it
			bootstrap.ID = big.NewInt(0)
		}

		node.Join(bootstrap)

		fmt.Printf("Node created with ID: %s\n", node.GetID().Text(16))
		fmt.Printf("Listening on: %s:%s\n", *address, *port)
	}

	// Print initial state
	fmt.Println("\nInitial state:")
	node.PrintState()

	// Start command loop
	fmt.Println("\nAvailable commands:")
	fmt.Println("  Lookup <filename>")
	fmt.Println("  StoreFile <filepath> <filename>")
	fmt.Println("  PrintState")
	fmt.Println("  quit")
	fmt.Println()

	scanner := bufio.NewScanner(os.Stdin)
	for {
		fmt.Print("> ")
		if !scanner.Scan() {
			break
		}

		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		parts := strings.Fields(line)
		cmd := parts[0]

		switch cmd {
		case "Lookup":
			if len(parts) < 2 {
				fmt.Println("Usage: Lookup <filename>")
				continue
			}
			filename := parts[1]
			node.Lookup(filename)

		case "StoreFile":
			if len(parts) < 3 {
				fmt.Println("Usage: StoreFile <filepath> <filename>")
				continue
			}
			filepath := parts[1]
			filename := parts[2]
			node.StoreFile(filepath, filename)

		case "PrintState":
			node.PrintState()

		case "quit", "exit":
			fmt.Println("Exiting...")
			os.Exit(0)

		default:
			fmt.Printf("Unknown command: %s\n", cmd)
		}
	}
}
