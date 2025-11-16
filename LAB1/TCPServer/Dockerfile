# --- Build Stage ---
FROM golang:1.22-alpine AS builder

# Create app directory
WORKDIR /app

# Copy go.mod and base package first
COPY go.mod ./
COPY base ./base
COPY TCPServer ./TCPServer

# Build the server binary
WORKDIR /app/TCPServer
RUN go build -o http_server .

# --- Run Stage ---
FROM alpine:latest

WORKDIR /app

# Copy only the final binary
COPY --from=builder /app/TCPServer/http_server .

# Expose default port (optional — you pass your own)
EXPOSE 8080

# Run the server with a port argument
CMD ["./http_server", "8080"]