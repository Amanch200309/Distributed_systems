#!/bin/bash

SERVER_PORT=8080
BINARY=./http_server

# Build
echo "Building server..."
go build -o http_server ../TCPServer || exit 1

# Start server in background
echo "Starting server on port $SERVER_PORT..."
$BINARY $SERVER_PORT &
PID=$!
sleep 1



# Test POST
echo -e "\n"
echo -e "\e[32mTesting POST...\n\e[0m"
curl -v -X POST --data-binary "<html><body><h1>test</h1></body></html>" localhost:$SERVER_PORT/test.html
echo -e "\n"

# Test GET
echo -e "\e[32mTesting GET...\n\e[0m"
curl -v -X GET localhost:$SERVER_PORT/test.html
echo -e "\n"

# Test missing file
echo -e "\e[32mTesting 404...\n\e[0m"
curl -v -X GET localhost:$SERVER_PORT/does_not_exist.html
echo -e "\n"

# Test invalid extension
echo -e "\e[32mTesting 400 invalid extension...\n\e[0m"
curl -v -X GET localhost:$SERVER_PORT/bad.exe
echo -e "\n"

# Test unimplemented method 501
echo -e "\e[32mTesting 501 unimplemented method...\n\e[0m"
curl -v -X DELETE localhost:$SERVER_PORT/test.html
echo -e "\n"

# Kill server
kill $PID
rm test.html
