package chord

import (
	"log"
	"math/big"
	"net"
	"net/rpc"
)

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

// FIXED: Use pointer for both arguments
func (n *Node) GetPredecessorRPC(arg *GetPredecessorRequest, reply *GetPredecessorReply) error {
	n.mu.RLock()
	defer n.mu.RUnlock()
	reply.Node = n.Predecessor
	return nil
}

// FIXED: Use pointer for both arguments
func (n *Node) NotifyRPC(arg *NotifyRequest, reply *NotifyReply) error {
	n.notify(arg.Node)
	return nil
}

// FIXED: Use pointer for both arguments
func (n *Node) PingRPC(arg *PingRequest, reply *PingReply) error {
	reply.Alive = true
	return nil
}

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

// RPC request and reply types

type StoreFileRequest struct {
	Filename string
	Hash     *big.Int
	Data     []byte
}

type StoreFileReply struct {
	OK bool
}

type GetFileRequest struct {
	Hash *big.Int
}

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
	Files       map[string]bool // Just filenames
}

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

type GetSuccessorListRequest struct {
}

type GetSuccessorListReply struct {
	Successors []*RemoteNode
}

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

func call(address string, rpcname string, args interface{}, reply interface{}) bool {
	c, err := rpc.Dial("tcp", address)
	if err != nil {
		return false
	}
	defer c.Close()

	err = c.Call(rpcname, args, reply)
	return err == nil
}

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
