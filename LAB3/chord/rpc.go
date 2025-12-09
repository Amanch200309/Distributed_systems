package chord

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"net/rpc"
	"strings"
	"time"

	"google.golang.org/grpc"
)

func (c *Chord) server(address string) {
	rpc.Register(c)
	rpc.HandleHTTP()

	l, e := net.Listen("tcp", ":8080")
	if e != nil {
		log.Fatal("listen error:", e)
	}
	go http.Serve(l, nil)
}

// resolveAddress handles :port format by adding the local address
func resolveAddress(address string) string {
	if strings.HasPrefix(address, ":") {
		return net.JoinHostPort(localaddress, address[1:])
	} else if !strings.Contains(address, ":") {
		return net.JoinHostPort(address, defaultPort)
	}
	return address
}

// StartServer starts the gRPC server for this node
func StartServer(address string, nprime string) (*Node, error) {
	address = resolveAddress(address)

	node := &Node{
		Address:     address,
		FingerTable: make([]string, keySize+1),
		Predecessor: "",
		Successors:  nil,
		Bucket:      make(map[string]string),
	}

	// Are we the first node?
	if nprime == "" {
		log.Print("StartServer: creating new ring")
		node.Successors = []string{node.Address}
	} else {
		log.Print("StartServer: joining existing ring using ", nprime)
		// For now use the given address as our successor
		nprime = resolveAddress(nprime)
		node.Successors = []string{nprime}
		// TODO: use a GetAll request to populate our bucket
	}

	// Start listening for RPC calls
	grpcServer := grpc.NewServer()
	pb.RegisterChordServer(grpcServer, node)

	lis, err := net.Listen("tcp", node.Address)
	if err != nil {
		return nil, fmt.Errorf("failed to listen: %v", err)
	}

	// Start server in goroutine
	log.Printf("Starting Chord node server on %s", node.Address)
	go func() {
		if err := grpcServer.Serve(lis); err != nil {
			log.Fatalf("failed to serve: %v", err)
		}
	}()

	// Start background tasks
	go func() {
		nextFinger := 0
		for {
			time.Sleep(time.Second / 3)
			node.stabilize()

			time.Sleep(time.Second / 3)
			nextFinger = node.fixFingers(nextFinger)

			time.Sleep(time.Second / 3)
			node.checkPredecessor()
		}
	}()

	return node, nil
}
