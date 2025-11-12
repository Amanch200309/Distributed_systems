package main

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"
)

// startServer starts the compiled binary and returns a cleanup function.
func startServer(t *testing.T, port string) func() {
	t.Helper()

	// Build the binary
	cmd := exec.Command("go", "build", "-o", "http_server", "main.go")
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("failed to build server: %v\n%s", err, string(out))
	}

	// Start the server
	server := exec.Command("./http_server", port)
	server.Stdout = os.Stdout
	server.Stderr = os.Stderr
	if err := server.Start(); err != nil {
		t.Fatalf("failed to start server: %v", err)
	}

	// Wait for startup
	time.Sleep(1 * time.Second)

	return func() {
		server.Process.Kill()
		server.Wait()
		os.Remove("./http_server")
		os.Remove("./test.html")
		os.Remove("./test.txt")
	}
}

func TestPostAndGetHTML(t *testing.T) {
	cleanup := startServer(t, "8085")
	defer cleanup()

	body := "<h1>Hello</h1>"
	resp, err := http.Post("http://localhost:8085/test.html", "text/html", strings.NewReader(body))
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("expected 201 Created, got %d", resp.StatusCode)
	}

	// Ensure file was created
	if _, err := os.Stat("./test.html"); err != nil {
		t.Fatalf("expected file test.html to exist: %v", err)
	}

	// GET same file
	getResp, err := http.Get("http://localhost:8085/test.html")
	if err != nil {
		t.Fatal(err)
	}
	defer getResp.Body.Close()

	if getResp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d", getResp.StatusCode)
	}

	b, _ := io.ReadAll(getResp.Body)
	if string(b) != body {
		t.Fatalf("file content mismatch: got %q, want %q", string(b), body)
	}
}

func TestPostAndGetText(t *testing.T) {
	cleanup := startServer(t, "8086")
	defer cleanup()

	data := "plain text"
	resp, err := http.Post("http://localhost:8086/test.txt", "text/plain", strings.NewReader(data))
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("expected 201 Created, got %d", resp.StatusCode)
	}

	getResp, err := http.Get("http://localhost:8086/test.txt")
	if err != nil {
		t.Fatal(err)
	}
	defer getResp.Body.Close()
	if getResp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d", getResp.StatusCode)
	}
	b, _ := io.ReadAll(getResp.Body)
	if string(b) != data {
		t.Fatalf("expected %q, got %q", data, string(b))
	}
}

func TestUnsupportedExtension(t *testing.T) {
	cleanup := startServer(t, "8087")
	defer cleanup()

	resp, err := http.Get("http://localhost:8087/test.xyz")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400 Bad Request, got %d", resp.StatusCode)
	}
}

func TestMissingFile(t *testing.T) {
	cleanup := startServer(t, "8088")
	defer cleanup()

	resp, err := http.Get("http://localhost:8088/missing.html")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("expected 404 Not Found, got %d", resp.StatusCode)
	}
}

func TestUnsupportedMethod(t *testing.T) {
	cleanup := startServer(t, "8089")
	defer cleanup()

	client := &http.Client{}
	req, err := http.NewRequest("PUT", "http://localhost:8089/test.html", strings.NewReader("data"))
	if err != nil {
		t.Fatal(err)
	}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNotImplemented {
		t.Fatalf("expected 501 Not Implemented, got %d", resp.StatusCode)
	}
}

func TestMalformedRequest(t *testing.T) {
	// Direct TCP connection to send a bad HTTP request
	cleanup := startServer(t, "8090")
	defer cleanup()

	conn, err := net.Dial("tcp", "localhost:8090")
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()

	io.WriteString(conn, "BAD REQUEST\r\n\r\n")

	resp, err := http.ReadResponse(bufio.NewReader(conn), nil)
	if err != nil {
		t.Fatal(err)
	}

	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400 Bad Request, got %d", resp.StatusCode)
	}
}

func TestConcurrentRequests(t *testing.T) {
	cleanup := startServer(t, "8091")
	defer cleanup()

	// Pre-create a file
	os.WriteFile("test.html", []byte("<h1>Hello</h1>"), 0644)

	const workers = 10
	errCh := make(chan error, workers)
	for i := 0; i < workers; i++ {
		go func() {
			resp, err := http.Get("http://localhost:8091/test.html")
			if err != nil {
				errCh <- err
				return
			}
			defer resp.Body.Close()
			if resp.StatusCode != http.StatusOK {
				errCh <- fmt.Errorf("expected 200, got %d", resp.StatusCode)
				return
			}
			io.Copy(io.Discard, resp.Body)
			errCh <- nil
		}()
	}
	for i := 0; i < workers; i++ {
		if err := <-errCh; err != nil {
			t.Fatal(err)
		}
	}
}
