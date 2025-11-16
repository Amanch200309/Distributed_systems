package main

import (
	"fmt"
	"os"
	"sync"

	"distributed-systems/lab1/base"
)

func main() {

	if len(os.Args) < 2 {
		fmt.Println("One arg required: (port)")
		return
	}
	port := os.Args[1]

	p := &ProxyServer{make(map[string]*CacheEntry), base.BaseServer{Maxconn: 10}, &sync.Mutex{}}

	//Lyssna på (0.0.0.0) + port default
	if err := p.Listen(":" + port); err != nil {
		fmt.Println("error:", err)
	}

}
