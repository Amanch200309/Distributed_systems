# Distributed Systems Lab 1

This repository contains implementations for a basic HTTP server and an HTTP proxy server, built using Go's `net` and `net/http` libraries. Additionally, Docker-based deployment instructions are included for AWS.

## Structure

* **base/** – shared TCP server abstraction
* **TCPServer/** – Basic HTTP server (GET/POST)
* **proxy/** – Advanced caching proxy server (GET only) (NOTE cache entries only updates if a new POST happens and the cacheExpiration time elapsed)
* **Dockerfiles** – used for AWS deployment
* **test scripts** – for local and AWS testing

---
## Error Codes and Responses

#### Success Responses

| Code | Status | When Triggered | Response Body |
|------|--------|----------------|---------------|
| **200** | OK | Successful GET request for existing file | File contents |
| **201** | Created | Successful POST request (file saved) | "File saved successfully" |

#### Client Error Responses

| Code | Status | When Triggered | Response Body |
|------|--------|----------------|---------------|
| **400** | Bad Request | - Malformed HTTP request<br>- Unsupported file extension (not .html, .txt, .gif, .jpeg, .jpg, .css)<br>- Error reading POST body | "400 Bad Request" or<br>"400 Bad Request (unsupported extension)" or<br>"Error reading body" |
| **404** | Not Found | GET request for non-existent file | "404 Not Found" |

#### Server Error Responses

| Code | Status | When Triggered | Response Body |
|------|--------|----------------|---------------|
| **500** | Internal Server Error | Error saving file during POST operation | "Error saving file" |
| **501** | Not Implemented | HTTP methods other than GET or POST (e.g., PUT, DELETE, HEAD, PATCH) | "501 Not Implemented" |

---

### Proxy Server

#### Success Responses

| Code | Status | When Triggered | Response Body |
|------|--------|----------------|---------------|
| **200** | OK | Successful GET request (cached or forwarded) | Cached or forwarded response body |

#### Client Error Responses

| Code | Status | When Triggered | Response Body |
|------|--------|----------------|---------------|
| **400** | Bad Request | Malformed HTTP request | "400 Bad Request" |

#### Server Error Responses

| Code | Status | When Triggered | Response Body |
|------|--------|----------------|---------------|
| **500** | Internal Server Error | Error reading response body from target server | "500 Internal Server Error" |
| **501** | Not Implemented | HTTP methods other than GET (POST, PUT, DELETE, etc.) | "501 Not Implemented" |
| **502** | Bad Gateway | Failed to connect to target server or forward request | "502 Bad Gateway" |

---

### Supported Content Types

Both the HTTP server and proxy support the following file extensions with their corresponding Content-Type headers:

| Extension | Content-Type |
|-----------|--------------|
| `.html` | `text/html` |
| `.txt` | `text/plain` |
| `.gif` | `image/gif` |
| `.jpeg`, `.jpg` | `image/jpeg` |
| `.css` | `text/css` |

**Note:** Any other file extension will result in a **400 Bad Request** error.

---

# Part 1 — Testing the Basic HTTP Server

## Build

```bash
cd LAB1/TCPServer
go build -o http_server 
```

## Run

```bash
./http_server 8080
```

## Test: POST

```bash
curl -X POST http://localhost:8080/index.html -d "<html><body><h1>test</h1></body></html>"        
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



---

# Part 2 — Testing the Proxy Server

## Build

```bash
cd LAB1/proxy
go build -o proxyserver
```

## Run

```bash
./proxyserver 8081
```

## Basic Proxy GET

```bash
curl -X GET http://localhost:8080/index.html -x http://localhost:8081
```

## Check Cache Expiry (30s)
```bash
sleep 31
curl -X POST http://localhost:8080/index.html -d "<html><body><h1>new_test</h1></body></html>"        
```

```bash
curl -X GET http://localhost:8080/index.html -x http://localhost:8081 // should print new_test
```
should return ```<html><body><h1>new_test</h1></body></html>%```


## Test Unsupported Method - POST not allowed through proxy

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

## Run on EC2
```bash   
docker run -p 8080:8080 http_server
docker run -p 8081:8081 proxy_server

```

## Test from Local Machine 

```bash 
curl http://<EC2-IP>:8080/index.html
curl -x http://<EC2-IP-server>:8081 http://<EC2-IP-proxy>:8080/index.html
```

# Test Scripts 

* `./test_basic.sh – test local HTTP server`

* `./test_proxy.sh – test local proxy`

* `./test_aws.sh <server-ip> <proxy-ip> – test cloud deployment` 


## Genereate Docker Image Tar files
```bash 
go build -v -o TCPServer/httpserver ./TCPServer
go build -v -o proxy/proxyserver ./proxy
```

## Setting up AWS
1. Launch EC2 instance (Ubuntu) or Amazon Linux
2. use vockey key
3. Add inbound rules:
   - To create an inbound rule Open the Amazon EC2 console at https://console.aws.amazon.com/ec2/.
   - On the navigation pane, choose Security Groups.
   - Under the Inbound rules tab, choose Edit inbound rules.
   - On Inbound rules page, choose Add rules.
   - For Type, choose Custom TCP.
   - For Port range enter port1-port2.
4. download SSH Key labuser.pem from canvas under AWS Details "Download PEM"
5. Set permissions: `chmod 400 labuser.pem`
6. SSH into the instance:
   - `ssh -i "labuser.pem" ubuntu@<EC2-IP>` (for Ubuntu)
   - `ssh -i "labuser.pem" ec2-user@<EC2-IP>` (for Amazon Linux)
7. Install Docker:
   For Ubuntu:
   ```bash
   sudo apt update
   sudo apt install docker.io
   sudo systemctl start docker
   sudo systemctl enable docker
   exit
   ```
   For Amazon Linux:
   ```bash
   sudo yum update -y
   sudo yum install docker -y
   sudo systemctl enable docker
   sudo systemctl start docker
   sudo usermod -aG docker ec2-user
   exit
   ```
8. From your local machine:
   ```bash
   scp -i labsuser.pem httpserver.tar ec2-user@<server-ip>:/home/ec2-user/
   ```
9. SSH back into the instance and load the docker image:
   ```bash
   ssh -i "labuser.pem" ec2-user@<EC2-IP>
   ```
   ```bash
   docker load -i httpserver.tar
   ```
   Note the image names printed after loading (e.g., httpserver:latest).
10. Run the Docker container:
    ```bash
    docker run -d -p 8080:8080 httpserver:latest
    ```
