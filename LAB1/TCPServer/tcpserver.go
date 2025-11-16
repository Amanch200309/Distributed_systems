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

	"github.com/Amanch200309/Distributed_systems/LAB1/base"
)

/*
TCPServer is an HTTP server that handles requests using raw TCP sockets.

	base: Underlying TCP server managing connections
*/
type TCPServer struct {
	base base.BaseServer
}

/*
Listen starts the HTTP server.

	Args: 	port (e.g., ":8080")
	Returns: error if fails to start
	Delegates to base server with handler for HTTP processing
*/
func (s *TCPServer) Listen(port string) error {
	return s.base.Listen(port, s.handler)
}

/*
handler processes incoming HTTP requests.

	Args: 	conn (client connection)
	Routes GET and POST requests to appropriate handlers, rejects other methods
*/
func (s *TCPServer) handler(conn net.Conn) {
	msg := bufio.NewReader(conn)
	req, err := http.ReadRequest(msg)
	if err != nil {
		resp := newResponse(http.StatusBadRequest, "400 Bad Request\n")
		resp.Write(conn)
		return
	}

	switch req.Method {
	case "GET":
		s.getHandler(conn, req)
	case "POST":
		s.postHandler(conn, req)
	default:
		resp := newResponse(http.StatusNotImplemented, "501 Not Implemented\n")
		resp.Write(conn)
		return
	}
}

/*
getHandler serves files for GET requests.

	Args: 	conn (client connection),
			req (HTTP request)
	Supports: .html, .txt, .gif, .jpeg/.jpg, .css files
	Returns 404 if file not found, 400 for unsupported file types
*/
func (s *TCPServer) getHandler(conn net.Conn, req *http.Request) {
	ext := filepath.Ext(req.URL.Path)
	filename := "." + req.URL.Path

	// Map file extension to content type
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

/*
postHandler saves uploaded files from POST requests.

	Args: 	conn (client connection),
			req (HTTP request)
	Creates directories as needed, saves file to path specified in URL
	Returns 201 on success, 400/500 on errors
*/
func (s *TCPServer) postHandler(conn net.Conn, req *http.Request) {
	defer req.Body.Close()

	ext := filepath.Ext(req.URL.Path)
	filename := "." + req.URL.Path

	// Map file extension to content type
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

	err = os.WriteFile(filename, bodyBytes, 0644) // 0 = base oct 5 = user{4+2+0 = read,write,}, 4 = group{4+0+0=read,,} , 4 = others{4+0+0=read,,} write body to file
	if err != nil {
		resp := newResponse(http.StatusInternalServerError, "Error saving file\n")
		resp.Write(conn)
		return
	}

	msg := "File saved successfully\n"
	resp := http.Response{
		Status:        "201 Created",
		StatusCode:    201,
		Proto:         "HTTP/1.1",
		ProtoMajor:    1,
		ProtoMinor:    1,
		Header:        make(http.Header),
		ContentLength: int64(len(msg)),
		Body:          io.NopCloser(strings.NewReader(msg)),
	}
	resp.Header.Set("Content-Type", contentType)
	resp.Write(conn)
}

/*
newResponse creates a simple HTTP response.

	Args: 	statusCode (HTTP status code),
			body (response body text)
	Returns: http.Response with specified status and body
*/
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
