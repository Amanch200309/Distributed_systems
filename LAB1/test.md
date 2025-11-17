# Distributed Systems Lab 1

This repository contains implementations for a basic HTTP server and an advanced HTTP proxy server, built using Go's `net` and `net/http` libraries. Additionally, Docker-based deployment instructions are included for AWS.

## Structure

* **base/** – shared TCP server abstraction
* **TCPServer/** – Basic HTTP server (GET/POST)
* **proxy/** – Advanced caching proxy server (GET only)
* **Dockerfiles** – used only for AWS deployment
* **test scripts** – for local and AWS testing

---

# Part 1 — Testing the Basic HTTP Server

## Build

```bash
go build -o http_server ./TCPServer
```

## Run

```bash
./http_server 8080
```

## Test: GET

```bash
curl -X GET http://localhost:8080/index.html
```

## Test: Missing File

```bash
curl -X GET http://localhost:8080/nope.html -v
```

## Test: Invalid Extension

```bash
curl -X GET http://localhost:8080/file.exe -v
```

## Test: POST Upload

```bash
echo "hello world" > test.txt
curl -X POST --data-binary @test.txt http://localhost:8080/upload/test.txt
```

## Test: GET After POST

```bash
curl http://localhost:8080/upload/test.txt
```

---

# Part 2 — Testing the Proxy Server

## Build

```bash
go build -o proxyserver ./proxy
```

## Run

```bash
./proxyserver 8081
```

## Basic Proxy GET

```bash
curl -X GET http://localhost:8080/index.html -x http://localhost:8081
```

## Cached GET (should be instant)

```bash
curl -X GET http://localhost:8080/index.html -x http://localhost:8081
```

## Check Cache Expiry (30s)

```bash
sleep 31
curl -X GET http://localhost:8080/index.html -x http://localhost:8081
```

## Test Unsupported Method

```bash
curl -X POST http://localhost:8080/index.html -x http://localhost:8081 -v
```

---

# Part 3 — AWS Deployment & Testing

## Build Docker Image

Inside **TCPServer/**:

```bash
docker build -t http_server .
```

Inside **proxy/**:

```bash
docker build -t proxy
```
