package base

import (
	"fmt"
	"net"
)

type BaseServer struct {
	Maxconn int
}

func (b *BaseServer) Listen(port string, handler func(net.Conn)) error {

	// start tcp-socket on addr
	l, err := net.Listen("tcp", port)
	if err != nil {
		return fmt.Errorf("failed to listen to %s", port)
	}
	defer l.Close()

	//// Create one buffered channel that can hold up to Maxconn empty signals (struct{} values) if full block until a spot is free
	channel := make(chan struct{}, b.Maxconn)

	//always accept new connections
	for {
		conn, err := l.Accept() // accept new client connection
		if err != nil {
			fmt.Println("accept error:", err)
			continue // do not stop server on accept error
		}

		channel <- struct{}{} // take spot
		go func(c net.Conn) {
			defer func() { // <--- detta kommer köras efter  handler(c) har kört klart
				c.Close()
				<-channel
			}()
			handler(c)
		}(conn)
	}
}
