# --- Build Stage ---
FROM golang:1.22-alpine AS builder

WORKDIR /app

COPY . .

RUN go build -o proxy main.go

# --- Run Stage ---
FROM alpine:latest

WORKDIR /app

COPY --from=builder /app/proxy .

EXPOSE 8081

CMD ["./proxy", "8081"]