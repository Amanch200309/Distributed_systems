# --- Build Stage ---
FROM golang:1.22-alpine AS builder

WORKDIR /app

# Copy go.mod and base package first
COPY go.mod ./
COPY base ./base
COPY proxy ./proxy

# Build proxy
WORKDIR /app/proxy
RUN go build -o proxyserver .

# --- Run Stage ---
FROM alpine:latest

WORKDIR /app

COPY --from=builder /app/proxy/proxyserver .

EXPOSE 8081

CMD ["./proxyserver", "8081"]