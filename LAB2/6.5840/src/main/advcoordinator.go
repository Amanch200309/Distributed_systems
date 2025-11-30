package main

import (
	"fmt"
	"os"
	"time"

	"6.5840/yeet"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "Usage: advcoordinator inputfiles...\n")
		os.Exit(1)
	}

	files := os.Args[1:]

	const nReduce = 10

	m := yeet.MakeCoordinator(files, nReduce)

	for m.Done() == false {
		time.Sleep(time.Second)
	}

	time.Sleep(time.Second)
}
