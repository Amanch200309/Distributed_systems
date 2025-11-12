package main

import (
	"fmt"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("One arg required: (port)")
		return
	}
	port := os.Args[1]

	s := &TCPServer{
		BaseServer{Maxconn: 10}, //  10 connections max
	}

	//Lyssna på (0.0.0.0) + port default
	if err := s.Listen(":" + port); err != nil {
		fmt.Println("error:", err)
	}
}
