#!/bin/bash

SERVER_PORT=8080
PROXY_PORT=8081
SERVER_BINARY=./http_server
PROXY_BINARY=./proxyserver

# Build server and proxy
echo "Building server and proxy..."
go build -o http_server ../TCPServer || exit 1
go build -o proxyserver ../proxy || exit 1

# Start server and proxy
echo "Starting server..."
$SERVER_BINARY $SERVER_PORT &
SERVER_PID=$!

sleep 1
echo "Starting proxy..."
$PROXY_BINARY $PROXY_PORT &
PROXY_PID=$!

sleep 1

# upload test file
echo -e "\e[32mUploading test file...\e[0m"
curl -v -X POST --data-binary "<html><body><h1>test</h1></body></html>" \
    localhost:$SERVER_PORT/test.html
echo -e "\n"

# Test proxy GET
echo -e "\e[32mTesting GET via proxy...\e[0m"
curl -v -X GET localhost:$SERVER_PORT/test.html -x localhost:$PROXY_PORT
echo -e "\n"

# Cached GET
echo -e "\e[32mTesting cached GET...\e[0m"
curl -v -X GET localhost:$SERVER_PORT/test.html -x localhost:$PROXY_PORT
echo -e "\n"

# Cache expiration test
echo -e "\e[33mWaiting for cache to expire (31s)...\e[0m"
sleep 31

echo -e "\e[32mPost 'Hello world' to the test.html on the httpserver\e[0m"
curl -v -X POST --data-binary "Hello world" \
     localhost:$SERVER_PORT/test.html
echo -e "\n"

echo -e "\e[32mCheck that cache is updated (should show 'Hello world')\e[0m"
curl -v -X GET localhost:$SERVER_PORT/test.html -x localhost:$PROXY_PORT
echo -e "\n"

# Unsupported method
echo -e "\e[32mTesting unsupported method (POST via proxy → should be 501)...\e[0m"
curl -v -X POST localhost:$SERVER_PORT/test.html -x localhost:$PROXY_PORT
echo -e "\n"


kill $SERVER_PID
kill $PROXY_PID

rm test.html