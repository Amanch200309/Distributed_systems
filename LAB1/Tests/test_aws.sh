#!/bin/bash

SERVER_IP="$1"
PROXY_IP="$2"

if [ -z "$SERVER_IP" ] || [ -z "$PROXY_IP" ]; then
    echo "Usage: ./test_aws.sh <server_ip> <proxy_ip>"
    exit 1
fi

SERVER_PORT=8080
PROXY_PORT=8081

# Test proxy GET
echo -e "\e[32mTesting GET via proxy...\e[0m"
curl -v -X GET http://$SERVER_IP:$SERVER_PORT/index.html -x $PROXY_IP:$PROXY_PORT
echo -e "\n"

# Test POST to server
echo -e "\e[32mTesting POST to server...\e[0m"
echo "hello cloud" > cloud.txt
curl -v -X POST --data-binary @cloud.txt $SERVER_IP:$SERVER_PORT/upload/cloud.txt
echo -e "\n"

# Test GET of uploaded file through proxy
echo -e "\e[32mTesting GET via proxy after upload...\e[0m"
curl -v -X GET $SERVER_IP:$SERVER_PORT/upload/cloud.txt -x $PROXY_IP:$PROXY_PORT
echo -e "\n"

# Cache expiration
echo -e "\e[33mWaiting for cache to expire (31s)...\e[0m"
sleep 31

echo -e "\e[32mRe-uploading to update server file...\e[0m"
curl -v -X POST --data-binary "Hello world" \
    $SERVER_IP:$SERVER_PORT/upload/cloud.txt
echo -e "\n"

echo -e "\e[32mChecking that proxy cache is updated (should show 'Hello world')...\e[0m"
curl -v -X GET $SERVER_IP:$SERVER_PORT/upload/cloud.txt -x $PROXY_IP:$PROXY_PORT
echo -e "\n"
