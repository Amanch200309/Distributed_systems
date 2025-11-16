package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"
	"sync"

	"github.com/Amanch200309/Distributed_systems/LAB1/base"
)

type ProxyServer struct {
	cache map[string]*CacheEntry // spara data på proxyn
	base  base.BaseServer        // en bas tcp server som den kommunicerar med
	mu    *sync.Mutex            // för att skydda cache map vid samtidiga accesser
}

// vad som ska sparas i cachen
type CacheEntry struct {
	body        []byte // hämtade datan
	contentType string // content type för filen html,txt,jpeg etc
	statusCode  int    // http status kod
}

func (p *ProxyServer) Listen(port string) error {
	if p.cache == nil {
		p.cache = make(map[string]*CacheEntry)
	}
	if p.mu == nil {
		p.mu = &sync.Mutex{}
	}
	return p.base.Listen(port, p.handler)
}

func sendCachedResponse(conn net.Conn, cached *CacheEntry) {
	resp := http.Response{
		Status:        fmt.Sprintf("%d %s", cached.statusCode, http.StatusText(cached.statusCode)),
		StatusCode:    cached.statusCode,
		Proto:         "HTTP/1.1",
		ProtoMajor:    1,
		ProtoMinor:    1,
		Header:        make(http.Header),
		Body:          io.NopCloser(bytes.NewReader(cached.body)),
		ContentLength: int64(len(cached.body)),
	}
	resp.Header.Set("Content-Type", cached.contentType)
	resp.Write(conn)
}

func (p *ProxyServer) forward(req *http.Request) (*http.Response, error) {
	// TCPServer
	target := req.URL.Host               // target server address
	conn, err := net.Dial("tcp", target) // connect to target server loclalhost:80
	if err != nil {
		return nil, err
	}
	//defer conn.Close()
	req.Write(conn) // forward the request to target server

	return http.ReadResponse(bufio.NewReader(conn), req)
}

func (p *ProxyServer) handler(conn net.Conn) {

	msg := bufio.NewReader(conn)
	req, err := http.ReadRequest(msg) // read http request from client
	if err != nil {
		resp := newResponse(http.StatusBadRequest, "400 Bad Request\n") // create 400 response
		resp.Write(conn)                                                // send response to client
		return
	}
	if req.Method != "GET" {
		resp := newResponse(http.StatusNotImplemented, "501 Not Implemented\n") // create 501 response as mentioned in the lab pm
		resp.Write(conn)                                                        // send response to client
		return
	}

	//kolla om filen finns i cachen

	key := req.URL.String() // använd url som nyckel ex /index.html

	p.mu.Lock()
	cached, found := p.cache[key] // om filen finns i cachen found = true och cahched innehåller datan annars found = false och cached = nil
	p.mu.Unlock()

	if found {

		// någon function som sickar tbx till client func(conn,cached), return
		sendCachedResponse(conn, cached)

	} else {
		resp, err := p.forward(req)
		if err != nil {
			resp := newResponse(http.StatusBadGateway, "502 Bad Gateway\n") // create 502 response
			resp.Write(conn)                                                // send response to client
			return
		}
		//  resp = p.sendtoclient(req)
		// resp is valid
		// cache resp

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			resp := newResponse(http.StatusInternalServerError, "500 Internal Server Error\n") // create 500 response
			resp.Write(conn)                                                                   // send response to client
			return
		}
		resp.Body.Close()

		/// spara i cache
		p.mu.Lock()
		p.cache[key] = &CacheEntry{
			body:        body,
			contentType: resp.Header.Get("Content-Type"),
			statusCode:  resp.StatusCode,
		}

		p.mu.Unlock()

		sendCachedResponse(conn, p.cache[key])
	}

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

/*
func main() {

	if len(os.Args) < 2 {
		fmt.Println("One arg required: (port)")
		return
	}
	port := os.Args[1]

	p := &ProxyServer{}

	_, err := p.Listen(":" + port)
	if err != nil {
		fmt.Println("Error starting server:", err)
	}

}
*/
