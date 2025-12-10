package main

import (
	"flag"
	"log"
	"os"
	"fmt"
	"bufio"
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
	address = flag.String("a", "", "IP address to listen on") // a- for address, "" for default,
	port    = flag.String("p", "", "Port to listen on")       // p- for port, "" for default

	ja = flag.String("ja", "", "The IP addres of the machine runnnig Chord node. The Chord client will join this node's ring")                    // ja
	jp = flag.String("jp", "", "The port that an existing Chord node is bound to and listening on. The Chord client will join this nodes ring. ") // jp

	ts  = flag.Int("ts", 0, "Time in milliseconds between invocations of 'stabilize'")          // ts, time stabilize
	tff = flag.Int("tff", 0, "Time in milliseconds between invocations of 'fix fingers'")       // tff, time fix fingers
	tcp = flag.Int("tcp", 0, "Time in milliseconds between invocations of 'check predecessor'") // tcp, time check predecessor

	r = flag.Int("r", 0, "Number of successors maintained by the Chord client") // r, number of successors

)

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

	// create node 
	addr := *address + ":" + *port
	nodeID := chord.HashKey(addr, 160) //sha1 hash of addr
	node := chord.NewNode(nodeID,addr,160,*r) // Create node
	// start server 
	go chord.StartRPCServer(node, *address, *port)

	if *ja == "" {
		fmt.Printf("Creating new Chord ring at: %s \n",addr)
		node.Create()
	} else {
		bootstrap := &chord.RemoteNode{
			Addr:	*ja + ":" + *jp,
			ID:		chord.HashKey(*ja + ":" + *jp,160),
		}
		fmt.Printf("Joining ring via:%s \n",bootstrap.Addr)

		err := node.Join(bootstrap)
		if err != nil {
			log.Fatal("Join failed:",err)
	}
	}

	go node.RunMaintenance(*ts,*tff,*tcp)

	fmt.Printf("Node is running. ID: %s, Address: %s\n", node.ID.Text(16), node.Address)
	handleCommands(node)

	

}
func isInRange(x int, low int, high int) bool {
	return x >= low && x <= high // low <= x <= high
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
	for {

		fmt.Print("> ")
		
		//ctrl + D to exit
		if !sc.Scan() {
			break
		}

		line := sc.Text() // read the whole line

		parts := strings.Fields(line) // split by whitespace

		if len(parts) == 0 {
			continue // empty line eg enter
		}
		cmd := strings.ToLower(parts[0])

		switch cmd{
			case "printstate":
				fmt.Println(n.PrintState())
			
			case "lookup":
				if len(parts) != 2 {
					fmt.Println("Usage: Lookup <filename>")
					continue
				}
				remote, _, err := n.Lookup(parts[1]) //lookup filename
				if err != nil  {
					fmt.Printf("Look up failed for Node: %s with error: %s \n",n.ID.Text(16), err)
					continue
				}
				fmt.Printf("File owned by Node ID=%s, Addr=%s \n",remote.ID.Text(16),remote.Addr)
			case "storefile":
				if len(parts) != 2 {
					fmt.Println("Usage: StoreFile <filepath>")
					continue
				}
				err := n.StoreFile(parts[1]) // store file at filepath
				if err != nil {
					fmt.Printf("StoreFile error: %s \n",err)
				} else {
					fmt.Println("File stored successfully.")
				}
			default:
				fmt.Printf("Unknown command: %s. Available commands: Lookup, StoreFile, PrintState",cmd)
		}		
	}

}
