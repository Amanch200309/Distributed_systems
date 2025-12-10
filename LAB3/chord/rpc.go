package chord

import (
	"log"
	"math/big"
	"net"
	"net/rpc"
)

func (n *Node) FindRPC(args *FindRequest, reply *FindReply) error {
	// Use the node's full local Find() implementation
	succ := n.Find(args.ID)
	reply.Node = succ
	return nil
}

// rpcs method
func (n *Node) FindSuccessorRPC(args *FindSuccessorRequest, reply *FindSuccessorReply) error {

	found, next := n.findSuccessor(args.ID)

	reply.Found = found
	reply.Node = next

	return nil
}

func (n *Node) GetPredecessorRPC(arg *GetPredecessorRequest, reply *GetPredecessorReply) error {
	n.mu.RLock()
	defer n.mu.RUnlock()
	reply.Node = n.Predecessor
	return nil
}

func (n *Node) NotifyRPC(arg *NotifyRequest, reply *NotifyReply) error {
	n.notify(arg.Node)
	return nil
}

func (rn *Node) PingRPC(arg *PingRequest, reply *PingReply) error {
	reply.Alive = true
	return nil
}

func (n *Node) StoreFileRPC(req *StoreFileRequest, rep *StoreFileReply) error {
	key := req.Hash.String()

	n.mu.Lock()
	n.Data[key] = req.Data
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
	//TODO:
}

type GetPredecessorReply struct {
	Node *RemoteNode
}

type NotifyRequest struct {
	Node *RemoteNode
}

type NotifyReply struct {
}
type PingRequest struct {
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
			go server.ServeConn(conn) // handle connections concurrently from multiple clients
		}
	}()
	return nil

}

