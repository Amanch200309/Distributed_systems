package chord

import (
	"log"
	"math/big"
	"net"
	"net/rpc"
)

// rpcs method
func (n *Node) FindSuccessorRPC(args *findSuccessorRequest, reply *findSuccessorReply) error {

	found, next := n.findSuccessor(args.ID)

	reply.Found = found
	reply.Node = next

	return nil
}

func (n *Node) GetPredecessorRPC(arg getPredecessorRequest, reply getPredecessorReply) error {
	reply.Node = n.Predecessor
	return nil
}

func (rn *RemoteNode) NotifyRPC(n *RemoteNode) error

func (rn *Node) PingRPC(arg pingRequest, reply pingReply) error {
	reply.Alive = true
	return nil
}

// RPC request and reply types

type findSuccessorRequest struct {
	ID *big.Int
}

type findSuccessorReply struct {
	Found bool
	Node  *RemoteNode
}

type getPredecessorRequest struct {
	//TODO:
}

type getPredecessorReply struct {
	Node *RemoteNode
}

type notifyRequest struct {
	Node *RemoteNode
}

type notifyReply struct {
}
type pingRequest struct {
}
type pingReply struct {
	Alive bool
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

func startRPCServer(address string, port string) error {
	server := rpc.NewServer()

	l, err := net.Listen("tcp", address+":"+port)
	if err != nil {
		log.Fatal("listen error:", err)
	}
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
