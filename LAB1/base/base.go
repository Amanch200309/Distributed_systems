package base

import (
	"fmt"
	"net"
)

/*
BaseServer is a concurrent TCP server with connection limits.

	Maxconn: Maximum simultaneous connections allowed
*/
type BaseServer struct {
	Maxconn int
}

/*
Listen starts the TCP server and handles incoming connections concurrently.

	Args: 	port (e.g., ":8080"),
			handler function for each connection
	Returns: error if fails to start
	Limits connections to Maxconn, runs handler in separate goroutines
*/
func (b *BaseServer) Listen(port string, handler func(net.Conn)) error {
	l, err := net.Listen("tcp", port) // TCP-listener on given porten
	if err != nil {
		return fmt.Errorf("failed to listen to %s", port)
	}
	defer l.Close()

	channel := make(chan struct{}, b.Maxconn) // Semaphore for connection limiting

	for {
		conn, err := l.Accept()
		if err != nil {
			fmt.Println("accept error:", err)
			continue
		}

		channel <- struct{}{} // Acquire slot (blocks if at max), may be better in go func(c net.Conn)
		go func(c net.Conn) {
			defer func() {
				c.Close()
				<-channel // Release slot
			}()
			handler(c)
		}(conn)
	}
}
