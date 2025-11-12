package main

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

// struct for TCP server
type TCPServer struct {
	base BaseServer
}

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

// s *TCPServer metood för structen samma klass metod i andra språk
func (s *TCPServer) Listen(port string) error {
	return s.base.Listen(port, s.handler)
}

// Handling for one connection
func (s *TCPServer) handler(conn net.Conn) {

	msg := bufio.NewReader(conn)
	req, err := http.ReadRequest(msg) // read http request from client
	if err != nil {
		resp := newResponse(http.StatusBadRequest, "400 Bad Request\n") // create 400 response
		resp.Write(conn)                                                // send response to client
		return
	}

	if req.Method == "GET" {
		s.getHandler(conn, req)
	} else if req.Method == "POST" {
		s.postHandler(conn, req)
	} else {
		resp := newResponse(http.StatusNotImplemented, "501 Not Implemented\n") // create 501 response as mentioned in the lab pm
		resp.Write(conn)                                                        // send response to client
		return
	}
}

// Handler for GET requests inputs: connection and request
func (s *TCPServer) getHandler(conn net.Conn, req *http.Request) {
	ext := filepath.Ext(req.URL.Path) // /index.html -> .html
	filename := "." + req.URL.Path    // ./index.html

	var contentType string
	switch ext {
	case ".html":
		contentType = "text/html"
	case ".txt":
		contentType = "text/plain"
	case ".gif":
		contentType = "image/gif"
	case ".jpeg", ".jpg":
		contentType = "image/jpeg"
	case ".css":
		contentType = "text/css"
	default:
		resp := newResponse(http.StatusBadRequest, "400 Bad Request (unsupported extension)\n") // create 400 response as mentioned in the lab pm
		resp.Write(conn)                                                                        // send response to client
		return
	}
	// open filename if error respond with 404
	f, err := os.Open(filename)
	if err != nil {
		resp := newResponse(http.StatusNotFound, "404 Not Found\n")
		resp.Write(conn)
		return
	}
	defer f.Close()

	info, _ := f.Stat() // get file info to obtain size
	size := info.Size() // get file size

	resp := http.Response{
		Status:        "200 OK",
		StatusCode:    200,
		Proto:         "HTTP/1.1",
		ProtoMajor:    1,
		ProtoMinor:    1,
		Header:        make(http.Header),
		Body:          f,
		ContentLength: size,
	}
	resp.Header.Set("Content-Type", contentType) // set content-type header
	resp.Write(conn)                             // send response to client
}

// Handler for POST requests inputs: connection and request
func (s *TCPServer) postHandler(conn net.Conn, req *http.Request) {
	defer req.Body.Close()

	ext := filepath.Ext(req.URL.Path) // /index.html -> .html
	filename := "." + req.URL.Path    // ./index.html

	var contentType string
	switch ext {
	case ".html":
		contentType = "text/html"
	case ".txt":
		contentType = "text/plain"
	case ".gif":
		contentType = "image/gif"
	case ".jpeg", ".jpg":
		contentType = "image/jpeg"
	case ".css":
		contentType = "text/css"
	default:
		resp := newResponse(http.StatusBadRequest, "Bad Request\n") // create 400 response as mentioned in the lab pm
		resp.Write(conn)                                            // send response to client
		return
	}

	bodyBytes, err := io.ReadAll(req.Body) // read entire body of the clients post request
	if err != nil {
		resp := newResponse(http.StatusBadRequest, "Error reading body\n") // create 400 response as mentioned in the lab pm
		resp.Write(conn)                                                   // send response to client
		return
	}

	// create all dir and parent dirs (if needed /folder/kiend.html) 0= base oct 7 = user{4+2+1 = read,write,ex}, 5 = group{4+0+1=read,,execute} , 5 = others{4+0+1=read,,execute}
	os.MkdirAll(filepath.Dir(filename), 0755)

	err = os.WriteFile(filename, bodyBytes, 0644) // 0 = base oct 5 = user{4+2+0 = read,write,}, 4 = group{4+0+0=read,,} , 4 = others{4+0+0=read,,} write body to file
	if err != nil {
		resp := newResponse(http.StatusInternalServerError, "Error saving file\n")
		resp.Write(conn)
		return
	}

	// respond
	msg := "File saved successfully\n"
	resp := http.Response{
		Status:        "201 Created",
		StatusCode:    201,
		Proto:         "HTTP/1.1",
		ProtoMajor:    1,
		ProtoMinor:    1,
		Header:        make(http.Header),
		ContentLength: int64(len(msg)),
		Body:          io.NopCloser(strings.NewReader(msg)), // create id.readcloser for the string
	}
	resp.Header.Set("Content-Type", contentType)
	resp.Write(conn)
}

// create new http response with status code and body=innehåll
func newResponse(statusCode int, body string) http.Response {
	return http.Response{
		Status:     fmt.Sprintf("%d %s", statusCode, http.StatusText(statusCode)),
		StatusCode: statusCode,
		Proto:      "HTTP/1.1",
		ProtoMajor: 1,
		ProtoMinor: 1,
		Header:     make(http.Header),
		Body:       io.NopCloser(strings.NewReader(body)), // Gör om body-strängen till en io.ReadCloser så att http.Response kan läsa den.
		// strings.NewReader(body) skapar en io.Reader, och io.NopCloser "wrappar" den
		// så att den även har en tom Close()-metod (krävs för Response.Body).
	}
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("One arg required: (port)")
		return
	}
	port := os.Args[1]

	s := &TCPServer{
		BaseServer{Maxconn: 10}, //  10 connections max
	}
	p := €

	//Lyssna på (0.0.0.0) + port default
	if err := s.Listen(":" + port); err != nil {
		fmt.Println("error:", err)
	}
}
