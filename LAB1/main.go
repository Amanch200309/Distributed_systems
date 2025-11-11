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

type TCPServer struct {
	Maxconn int
}

func (s *TCPServer) Listen(addr string) error {
	l, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("failed to listen to %s", addr)
	}
	defer l.Close()

	channel := make(chan struct{}, s.Maxconn)

	for {
		conn, err := l.Accept()
		if err != nil {
			fmt.Println("accept error:", err)
			continue
		}

		channel <- struct{}{} // take spot
		go s.handler(conn, channel)
	}
}

// Handling for one connection
func (s *TCPServer) handler(conn net.Conn, channel chan struct{}) {

	defer func() {
		conn.Close()
		<-channel // release spot
	}()

	msg := bufio.NewReader(conn)
	req, err := http.ReadRequest(msg)
	if err != nil {
		resp := newResponse(http.StatusBadRequest, "400 Bad Request\n")
		resp.Write(conn)
		return
	}

	if req.Method == "GET" {
		s.getHandler(conn, req)
	} else if req.Method == "POST" {
		s.postHandler(conn, req)
	} else {
		resp := newResponse(http.StatusNotImplemented, "501 Not Implemented\n")
		resp.Write(conn)
		return
	}
}

func (s *TCPServer) getHandler(conn net.Conn, req *http.Request) {
	ext := filepath.Ext(req.URL.Path)  // /kiend.html -> .html
	filename := "." + req.URL.Path     // ./kiend.html

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
		resp := newResponse(http.StatusBadRequest, "400 Bad Request (unsupported extension)\n")
		resp.Write(conn)
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

	info, _ := f.Stat()
	size := info.Size()

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
	resp.Header.Set("Content-Type", contentType)
	resp.Write(conn)
}

func (s *TCPServer) postHandler(conn net.Conn, req *http.Request) {
	defer req.Body.Close()

	ext := filepath.Ext(req.URL.Path)
	filename := "." + req.URL.Path

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
		resp := newResponse(http.StatusBadRequest, "Bad Request\n")
		resp.Write(conn)
		return
	}

	bodyBytes, err := io.ReadAll(req.Body)
	if err != nil {
		resp := newResponse(http.StatusBadRequest, "Error reading body\n")
		resp.Write(conn)
		return
	}

	// create all dir and parent dirs (if needed /folder/kiend.html) 0= base oct 7 = user{4+2+1 = read,write,ex}, 5 = group{4+0+1=read,,execute} , 5 = others{4+0+1=read,,execute}
	os.MkdirAll(filepath.Dir(filename), 0755)

	err = os.WriteFile(filename, bodyBytes, 0644) // 0 = base oct 5 = user{4+2+0 = read,write,}, 4 = group{4+0+0=read,,} , 4 = others{4+0+0=read,,}
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
		Body:          io.NopCloser(strings.NewReader(msg)),  // create id.readcloser for the string 
	}
	resp.Header.Set("Content-Type", contentType)
	resp.Write(conn)
}

func newResponse(statusCode int, body string) http.Response {
	return http.Response{
		Status:     fmt.Sprintf("%d %s", statusCode, http.StatusText(statusCode)),
		StatusCode: statusCode,
		Proto:      "HTTP/1.1",
		ProtoMajor: 1,
		ProtoMinor: 1,
		Header:     make(http.Header),
		Body:       io.NopCloser(strings.NewReader(body)),
	}
}


func main() {
	if len(os.Args) < 2 {
		fmt.Println("One arg required: (port)")
		return
	}
	port := os.Args[1]

	s := &TCPServer{
		Maxconn: 10, //  10 connections max 
	}
	if err := s.Listen(":" + port); err != nil {
		fmt.Println("error:", err)
	}
}
