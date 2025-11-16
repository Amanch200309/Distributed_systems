package main

import (
	"fmt"
	"os"

	"github.com/Amanch200309/Distributed_systems/LAB1/base"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("One arg required: (port)")
		return
	}
	port := os.Args[1]

	// Initialize TCP server with connection limit
	s := &TCPServer{
		base.BaseServer{Maxconn: 10},
	}

	// Start server on all interfaces (0.0.0.0) with specified port
	if err := s.Listen(":" + port); err != nil {
		fmt.Println("error:", err)
	}
}
