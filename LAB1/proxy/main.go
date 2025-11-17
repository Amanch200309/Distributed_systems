package main

import (
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/Amanch200309/Distributed_systems/LAB1/base"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("One arg required: (port)")
		return
	}
	port := os.Args[1]

	// Initialize proxy with cache, connection limit, mutex, and how long entries should be cached
	p := &ProxyServer{make(map[string]*CacheEntry), base.BaseServer{Maxconn: 10}, &sync.Mutex{}, 30 * time.Second}

	// Start proxy server on all interfaces (0.0.0.0) with specified port
	if err := p.Listen(":" + port); err != nil {
		fmt.Println("error:", err)
	}
}
