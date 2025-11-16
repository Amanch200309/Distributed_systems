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
	"time"

	"github.com/Amanch200309/Distributed_systems/LAB1/base"
)

/*
ProxyServer is a caching HTTP proxy that forwards GET requests and stores responses.

	cache: Map of URL -> cached response data
	base: Underlying TCP server handling connections
	mu: Mutex protecting cache from concurrent access
*/
type ProxyServer struct {
	cache map[string]*CacheEntry
	base  base.BaseServer
	mu    *sync.Mutex
}

/*
CacheEntry stores a cached HTTP response.

	body: Response body data
	contentType: HTTP Content-Type header value
	statusCode: HTTP status code
	timestamp: When this entry was cached
*/
type CacheEntry struct {
	body        []byte
	contentType string
	statusCode  int
	timestamp   time.Time
}

/*
Listen starts the proxy server.

	Args: 	port (e.g., ":8080")
	Returns: error if fails to start
	Initializes cache and mutex if needed, delegates to base server
*/
func (p *ProxyServer) Listen(port string) error {
	if p.cache == nil {
		p.cache = make(map[string]*CacheEntry)
	}
	if p.mu == nil {
		p.mu = &sync.Mutex{}
	}
	return p.base.Listen(port, p.handler)
}

/*
sendCachedResponse sends a cached HTTP response to the client.

	Args: 	conn (client connection),
			cached (cached response data)
*/
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

/*
forward sends the HTTP request to the target server and returns the response.

	Args: 	req (HTTP request to forward)
	Returns: HTTP response from target server, error if connection fails
*/
func (p *ProxyServer) forward(req *http.Request) (*http.Response, error) {
	target := req.URL.Host
	conn, err := net.Dial("tcp", target)
	if err != nil {
		return nil, err
	}

	req.Write(conn) // Forward request to target server
	return http.ReadResponse(bufio.NewReader(conn), req)
}

/*
handler processes incoming proxy requests.

	Args: 	conn (client connection)
	Handles GET requests: checks cache first, forwards to target if not cached,
	caches response, and sends to client
*/
func (p *ProxyServer) handler(conn net.Conn) {
	msg := bufio.NewReader(conn)
	req, err := http.ReadRequest(msg)
	if err != nil {
		resp := newResponse(http.StatusBadRequest, "400 Bad Request\n")
		resp.Write(conn)
		return
	}

	if req.Method != "GET" {
		resp := newResponse(http.StatusNotImplemented, "501 Not Implemented\n")
		resp.Write(conn)
		return
	}

	// Check cache for requested URL
	key := req.URL.String()

	p.mu.Lock()
	cached, found := p.cache[key]
	p.mu.Unlock()

	// Cache expiration time: 30 seconds
	cacheExpiration := 30 * time.Second

	// Check if cache entry is valid and not expired
	cacheValid := found && time.Since(cached.timestamp) < cacheExpiration

	if cacheValid {
		// Send cached response to client
		sendCachedResponse(conn, cached)
	} else {
		// Forward request to target server (not cached or expired)
		resp, err := p.forward(req)
		if err != nil {
			resp := newResponse(http.StatusBadGateway, "502 Bad Gateway\n")
			resp.Write(conn)
			return
		}

		// Read and cache response body
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			resp := newResponse(http.StatusInternalServerError, "500 Internal Server Error\n")
			resp.Write(conn)
			return
		}
		resp.Body.Close()

		// Store/update response in cache with current timestamp
		p.mu.Lock()
		p.cache[key] = &CacheEntry{
			body:        body,
			contentType: resp.Header.Get("Content-Type"),
			statusCode:  resp.StatusCode,
			timestamp:   time.Now(),
		}
		p.mu.Unlock()

		sendCachedResponse(conn, p.cache[key])
	}
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
