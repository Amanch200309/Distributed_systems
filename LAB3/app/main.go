package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"math/big"
	"os"
	"strings"

	"github.com/Amanch200309/Distributed_systems/LAB3/chord"
)

/*

The Chord client will be a command-line utility which takes the following arguments:
1. -a <String> = The IP address that the Chord client will bind to, as well as advertise to other nodes. Represented as an ASCII string (e.g., 128.8.126.63). Must be specified.
2. -p <Number> = The port that the Chord client will bind to and listen on. Represented as a base-10 integer. Must be specified.
3. --ja <String> = The IP address of the machine running a Chord node. The Chord client will join this node’s ring. Represented as an ASCII string (e.g., 128.8.126.63). Must be specified if --jp is specified.
4. --jp <Number> = The port that an existing Chord node is bound to and listening on. The Chord client will join this node’s ring. Represented as a base-10 integer. Must be specified if --ja is specified.
5. --ts <Number> = The time in milliseconds between invocations of ‘stabilize’. Represented as a base-10 integer. Must be specified, with a value in the range of [1,60000].
6. --tff <Number> = The time in milliseconds between invocations of ‘fix fingers’. Represented as a base-10 integer. Must be specified, with a value in the range of [1,60000].
7. --tcp <Number> = The time in milliseconds between invocations of ‘check predecessor’.
Represented as a base-10 integer. Must be specified, with a value in the range of [1,60000].
8. -r <Number> = The number of successors maintained by the Chord client. Represented as a base-10 integer. Must be specified, with a value in the range of [1,32].
9. -i <String> = The identifier (ID) assigned to the Chord client which will override the ID computed by the SHA1 sum of the client’s IP address and port number. Represented as a string of 40 characters matching [0-9a-fA-F]. Optional parameter.
*/

var (
	address = flag.String("a", "", "IP address to listen on")
	port    = flag.String("p", "", "Port to listen on")

	ja = flag.String("ja", "", "IP address of existing Chord node to join")
	jp = flag.String("jp", "", "Port of existing Chord node to join")

	ts  = flag.Int("ts", 0, "Time in milliseconds between stabilize calls")
	tff = flag.Int("tff", 0, "Time in milliseconds between fix fingers calls")
	tcp = flag.Int("tcp", 0, "Time in milliseconds between check predecessor calls")

	r = flag.Int("r", 0, "Number of successors to maintain")
	i = flag.String("i", "", "Custom node ID (40 hex characters)")
)

/*
main initializes and runs a Chord node.

	Parses command-line flags, validates arguments, creates node,
	starts RPC server, joins or creates ring, runs maintenance tasks,
	and handles user commands
*/
func main() {
	flag.Parse()

	// Validate required flags
	if *address == "" || *port == "" || *ts <= 0 || *tff <= 0 || *tcp <= 0 || *r <= 0 {
		log.Fatal("Missing required flags")
		os.Exit(1)
	}

	if !isInRange(*ts, 1, 60000) || !isInRange(*tff, 1, 60000) || !isInRange(*tcp, 1, 60000) {
		log.Fatal("--ts, --tff, and --tcp must be in the range [1,60000]")
	}

	if !isInRange(*r, 1, 32) {
		log.Fatal("-r must be in the range [1,32]")
	}

	// xor validation for ja and jp (ja and not jp) || (not ja and jp) = ja xor jp
	if (*ja == "" && *jp != "") || (*ja != "" && *jp == "") {
		log.Fatal("--ja and --jp must both be specified together")
	}

	// Validate -i flag format if provided
	if *i != "" {
		if len(*i) != 40 {
			log.Fatal("-i must be exactly 40 characters")
		}
		for _, c := range *i {
			if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f') || (c >= 'A' && c <= 'F')) {
				log.Fatal("-i must contain only hex characters [0-9a-fA-F]")
			}
		}
	}

	// create node
	addr := *address + ":" + *port

	hashLen := 20 // 20 bytes = 160 bits for SHA-1

	var nodeID *big.Int
	if *i != "" {
		// Use custom ID if provided
		nodeID = new(big.Int)
		nodeID.SetString(*i, 16)
	} else {
		nodeID = chord.HashKey(addr, hashLen) //sha1 hash of addr
	}
	node := chord.NewNode(nodeID, addr, hashLen, *r) // Create node
	// start server
	go chord.StartRPCServer(node, *address, *port)

	if *ja == "" {
		fmt.Printf("Creating new Chord ring at: %s \n", addr)
		node.Create()
	} else {
		bootstrap := &chord.RemoteNode{
			Addr: *ja + ":" + *jp,
			ID:   chord.HashKey(*ja+":"+*jp, hashLen),
		}
		fmt.Printf("Joining ring via:%s \n", bootstrap.Addr)

		err := node.Join(bootstrap)
		if err != nil {
			log.Fatal("Join failed:", err)
		}
	}

	go node.RunMaintenance(*ts, *tff, *tcp)

	fmt.Printf("Node is running. ID: %s, Address: %s\n", node.ID.Text(16), node.Address)
	handleCommands(node)

}

/*
isInRange checks if a value is within a range.

	Args: 	x (value to check),
			low (minimum value inclusive),
			high (maximum value inclusive)
	Returns: bool (true if low <= x <= high)
*/
func isInRange(x int, low int, high int) bool {
	return x >= low && x <= high
}

/*

Commands:

The Chord client will handle commands by reading from stdin and writing to stdout. There are three command that the Chord client must support: ‘Lookup’, 'StoreFile', and ‘PrintState’.

‘Lookup’ takes as input the name of a file to be searched (e.g., “Hello.txt”). The Chord client takes this string, hashes it to a key in the identifier space, and performs a search for the node that is the successor to the key (i.e., the owner of the key). The Chord client then outputs that node’s identifier, IP address, port, and the contents of the file.
'StoreFile' takes the location of a file on a local disk, then performs a lookup to find the Chord node to store the file at, then uploading the file to the Chord ring.
‘PrintState’ requires no input. The Chord client outputs its local state information at the current time, which consists of:
The Chord client’s own node information and its stored files,
The node information for all nodes in the successor list,
The node information for all nodes in the finger table,
where “node information” corresponds to the identifier, IP address, and port for a given node.
*/

func handleCommands(n *chord.Node) {
	sc := bufio.NewScanner(os.Stdin)

	printHelp() // Show available commands on start

	for {
		fmt.Print("> ")

		if !sc.Scan() {
			break
		}

		line := strings.TrimSpace(sc.Text())
		if line == "" {
			continue
		}

		parts := strings.Fields(line)
		cmd := strings.ToLower(parts[0])

		switch cmd {

		// ----------------------------------------------------
		// Standard assignment commands
		// ----------------------------------------------------
		case "printstate":
			fmt.Println(n.PrintState())

		case "lookup":
			if len(parts) != 2 {
				fmt.Println("Usage: lookup <filename>")
				continue
			}
			remote, _, err := n.Lookup(parts[1])
			if err != nil {
				fmt.Printf("Lookup failed: %v\n", err)
				continue
			}
			fmt.Printf("Owned by Node %s @ %s\n",
				remote.ID.Text(16), remote.Addr)

		case "storefile":
			if len(parts) != 2 {
				fmt.Println("Usage: storefile <filepath>")
				continue
			}
			if err := n.StoreFile(parts[1]); err != nil {
				fmt.Printf("Error: %v\n", err)
			} else {
				fmt.Println("File stored successfully.")
			}

		case "ring":
			fmt.Println(n.PrintRingStatus())

		case "files":
			fmt.Println(n.PrintFileDistribution())

		case "compact":
			fmt.Println(n.PrintCompactRing())

		case "verify":
			fmt.Println(n.VerifyFileOwnership())

		case "clear":
			fmt.Print("\033[H\033[2J") // ANSI clear screen

		case "help":
			printHelp()

		case "exit", "quit":
			fmt.Println("Exiting.")
			return

		default:
			fmt.Printf("Unknown command: %s\n", cmd)
			printHelp()
		}
	}
}

/*
printHelp displays available commands to the user.

	Shows command syntax and descriptions for all supported operations
*/
func printHelp() {
	fmt.Println("\nAvailable commands:")
	fmt.Println("  lookup <filename>        - Find which node owns a file")
	fmt.Println("  storefile <path>         - Store a file into the Chord ring")
	fmt.Println("  printstate               - Print local node state")
	fmt.Println()
	fmt.Println("  ring                     - Full ring visualization")
	fmt.Println("  files                    - File distribution table")
	fmt.Println("  verify                   - Validate file ownership")
	fmt.Println("  compact                  - One-line ring view")
	fmt.Println()
	fmt.Println("  clear                    - Clear the terminal")
	fmt.Println("  help                     - Show this help message")
	fmt.Println("  exit / quit              - Exit the client")
	fmt.Println()
}
