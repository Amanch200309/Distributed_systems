package chord

import (
	"log"
	"math/big"
	"net"
	"net/rpc"
)

/*
FindRPC handles RPC requests to find the successor of an ID.

	Args: 	args (contains ID to find successor for),
			reply (populated with successor node)
	Returns: error (always nil in current implementation)
	Performs complete lookup starting from this node
*/
func (n *Node) FindRPC(args *FindRequest, reply *FindReply) error {
	succ := n.Find(args.ID)
	if succ == nil {
		// Return self as fallback if lookup fails
		n.mu.RLock()
		reply.Node = &RemoteNode{ID: n.ID, Addr: n.Address}
		n.mu.RUnlock()
	} else {
		reply.Node = succ
	}
	return nil
}

/*
FindSuccessorRPC handles one step of iterative lookup.

	Args: 	args (contains ID to find successor for),
			reply (populated with found status and next node)
	Returns: error (always nil)
	Returns true and successor if ID is between this node and successor,
	otherwise returns false and closest preceding node to continue search
*/
func (n *Node) FindSuccessorRPC(args *FindSuccessorRequest, reply *FindSuccessorReply) error {
	found, next := n.findSuccessor(args.ID)
	reply.Found = found
	// Ensure we never return nil
	if next == nil {
		n.mu.RLock()
		reply.Node = &RemoteNode{ID: n.ID, Addr: n.Address}
		n.mu.RUnlock()
	} else {
		reply.Node = next
	}

	return nil
}

/*
GetPredecessorRPC returns this node's predecessor.

	Args: 	arg (empty request struct),
			reply (populated with predecessor node)
	Returns: error (always nil)
	Used during stabilization to verify ring consistency
*/
func (n *Node) GetPredecessorRPC(arg *GetPredecessorRequest, reply *GetPredecessorReply) error {
	n.mu.RLock()
	defer n.mu.RUnlock()
	reply.Node = n.Predecessor
	return nil
}

/*
NotifyRPC handles notification from potential predecessor.

	Args: 	arg (contains notifying node),
			reply (empty reply struct)
	Returns: error (always nil)
	Updates predecessor if notifying node is closer than current predecessor
*/
func (n *Node) NotifyRPC(arg *NotifyRequest, reply *NotifyReply) error {
	n.notify(arg.Node)
	return nil
}

/*
PingRPC checks if this node is alive.

	Args: 	arg (empty request struct),
			reply (populated with alive status)
	Returns: error (always nil)
	Used to detect failed nodes during stabilization
*/
func (n *Node) PingRPC(arg *PingRequest, reply *PingReply) error {
	reply.Alive = true
	return nil
}

/*
StoreFileRPC stores a file on this node.

	Args: 	req (contains filename, hash, and file data),
			rep (populated with success status)
	Returns: error (always nil)
	Stores file data and metadata in node's storage maps
*/
func (n *Node) StoreFileRPC(req *StoreFileRequest, rep *StoreFileReply) error {
	key := req.Hash.String()

	n.mu.Lock()
	n.Data[key] = req.Data
	// Store file metadata
	n.Files[key] = &FileMetadata{
		Filename: req.Filename,
		Data:     req.Data,
		Hash:     req.Hash,
	}
	n.mu.Unlock()

	rep.OK = true
	return nil
}

/*
GetFileRPC retrieves a file from this node.

	Args: 	req (contains file hash),
			rep (populated with file data if found)
	Returns: error (always nil)
	Searches node's storage for file with matching hash
*/
func (n *Node) GetFileRPC(req *GetFileRequest, rep *GetFileReply) error {
	key := req.Hash.String()

	n.mu.RLock()
	data, exists := n.Data[key]
	n.mu.RUnlock()

	if !exists {
		rep.OK = false
		return nil
	}

	rep.OK = true
	rep.Data = data
	return nil
}

/*
StoreFileRequest contains data for storing a file.

	Filename: Original filename
	Hash: Hash of filename (DHT key)
	Data: File contents as bytes
*/
type StoreFileRequest struct {
	Filename string
	Hash     *big.Int
	Data     []byte
}

/*
StoreFileReply indicates success of file storage.

	OK: True if file was stored successfully
*/
type StoreFileReply struct {
	OK bool
}

/*
GetFileRequest requests a file by hash.

	Hash: File hash (DHT key)
*/
type GetFileRequest struct {
	Hash *big.Int
}

/*
GetFileReply returns requested file data.

	Filename: Name of the file
	Data: File contents as bytes
	OK: True if file was found
*/
type GetFileReply struct {
	Filename string
	Data     []byte
	OK       bool
}

type FindSuccessorRequest struct {
	ID *big.Int
}

type FindSuccessorReply struct {
	Found bool
	Node  *RemoteNode
}

type GetPredecessorRequest struct {
	// Empty struct for RPC
}

type GetPredecessorReply struct {
	Node *RemoteNode
}

type NotifyRequest struct {
	Node *RemoteNode
}

type NotifyReply struct {
	// Empty struct for RPC
}

type PingRequest struct {
	// Empty struct for RPC
}

type PingReply struct {
	Alive bool
}

type FindRequest struct {
	ID *big.Int
}

type FindReply struct {
	Node *RemoteNode
}

// RPC for getting node information
type NodeInfoRequest struct {
	// Empty
}

type NodeInfoReply struct {
	ID          *big.Int
	Address     string
	Successor   *RemoteNode
	Predecessor *RemoteNode
	Successors  []*RemoteNode
	FingerCount int
	Files       map[string]bool
}

/*
GetNodeInfoRPC returns comprehensive information about this node.

	Args: 	req (empty request struct),
			reply (populated with node information)
	Returns: error (always nil)
	Provides node state including ID, address, successor, predecessor,
	finger table size, and stored file names for monitoring
*/
func (n *Node) GetNodeInfoRPC(req *NodeInfoRequest, reply *NodeInfoReply) error {
	n.mu.RLock()
	defer n.mu.RUnlock()

	reply.ID = n.ID
	reply.Address = n.Address
	reply.Predecessor = n.Predecessor

	// Filter out nil elements from Successors to avoid gob encoding errors
	reply.Successors = make([]*RemoteNode, 0, len(n.Successors))
	for _, s := range n.Successors {
		if s != nil {
			reply.Successors = append(reply.Successors, s)
		}
	}

	if len(reply.Successors) > 0 {
		reply.Successor = reply.Successors[0]
	}

	// Count non-nil fingers
	count := 0
	for _, f := range n.FingerTable {
		if f != nil {
			count++
		}
	}
	reply.FingerCount = count

	// Collect file names
	reply.Files = make(map[string]bool)
	for _, meta := range n.Files {
		reply.Files[meta.Filename] = true
	}

	return nil
}

/*
GetSuccessorListRequest requests node's successor list.
*/
type GetSuccessorListRequest struct {
}

/*
GetSuccessorListReply returns successor list.

	Successors: Array of successor nodes
*/
type GetSuccessorListReply struct {
	Successors []*RemoteNode
}

/*
GetSuccessorListRPC returns this node's successor list.

	Args: 	args (empty request struct),
			reply (populated with successor list)
	Returns: error (always nil)
	Filters out nil entries to avoid encoding errors
*/
func (n *Node) GetSuccessorListRPC(args *GetSuccessorListRequest, reply *GetSuccessorListReply) error {
	n.mu.RLock()
	defer n.mu.RUnlock()
	// Filter out nil entries to avoid gob encoding errors
	reply.Successors = make([]*RemoteNode, 0, len(n.Successors))
	for _, succ := range n.Successors {
		if succ != nil {
			reply.Successors = append(reply.Successors, succ)
		}
	}
	return nil
}

/*
call performs a remote procedure call to another node.

	Args: 	address (target node address),
			rpcname (method to call),
			args (request arguments),
			reply (response struct to populate)
	Returns: bool (true if call succeeded, false on error)
	Establishes TCP connection, makes RPC call, and closes connection
*/
func call(address string, rpcname string, args interface{}, reply interface{}) bool {
	c, err := rpc.Dial("tcp", address)
	if err != nil {
		return false
	}
	defer c.Close()

	err = c.Call(rpcname, args, reply)
	return err == nil
}

/*
StartRPCServer starts the RPC server for a Chord node.

	Args: 	node (Chord node to serve RPC requests for),
			address (IP address to bind to),
			port (port to listen on)
	Returns: error if server fails to start
	Registers node with RPC server and handles incoming connections
	in separate goroutines
*/
func StartRPCServer(node *Node, address string, port string) error {
	server := rpc.NewServer()
	err := server.Register(node)
	if err != nil {
		log.Fatal("register error:", err)
	}

	l, err := net.Listen("tcp", address+":"+port)
	if err != nil {
		log.Fatal("listen error:", err)
	}
	log.Printf("RPC server listening on %s:%s\n", address, port)
	go func() {
		for {
			conn, err := l.Accept()
			if err != nil {
				log.Fatal("accept error:", err)
			}
			go server.ServeConn(conn)
		}
	}()
	return nil
}
