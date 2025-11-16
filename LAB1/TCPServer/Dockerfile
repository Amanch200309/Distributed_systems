# --- Build Stage ---
FROM golang:1.22-alpine AS builder

# Create app directory
WORKDIR /app

# Copy source
COPY . .

# Build the server binary
RUN go build -o http_server main.go

# --- Run Stage ---
FROM alpine:latest

WORKDIR /app

# Copy only the final binary
COPY --from=builder /app/http_server .

# Expose default port (optional — you pass your own)
EXPOSE 8080

# Run the server with a port argument
CMD ["./http_server", "8080"]