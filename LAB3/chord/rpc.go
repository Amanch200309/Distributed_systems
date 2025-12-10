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

func (n *Node) PingRPC(arg *PingRequest, reply *PingReply) error {
	reply.Alive = true
	return nil
}

// RPC request and reply types

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

// CallPublic is an exported version of call for use in main package
func CallPublic(address string, rpcname string, args interface{}, reply interface{}) bool {
	return call(address, rpcname, args, reply)
}

func startRPCServer(n *Node, address string, port string) error {
	server := rpc.NewServer()
	err := server.Register(n)
	if err != nil {
		log.Fatal("RPC registration error:", err)
		return err
	}

	l, err := net.Listen("tcp", address+":"+port)
	if err != nil {
		log.Fatal("listen error:", err)
		return err
	}
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
