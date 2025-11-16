PORT=8080
BINARY=./httpserver

echo "Starting server on port $PORT..."
$BINARY $PORT &
SERVER_PID=$!

Give server time to start,
sleep 1

echo "Testing GET existing file..."
echo "Hello World" > index.html
curl -s -o /dev/null -w "%{http_code}\n" http://localhost:$PORT/index.html

echo "Testing GET non-existent file..."
curl -s -o /dev/null -w "%{http_code}\n" http://localhost:$PORT/missing.html

echo "Testing POST new file..."
curl -s -X POST -d "Test data" -H "Content-Type: text/plain" http://localhost:$PORT/test.txt -o /dev/null -w "%{http_code}\n"

echo "Testing GET the POSTed file..."
curl -s http://localhost:$PORT/test.txt

echo "Testing unsupported method (PUT)..."
curl -s -X PUT http://localhost:$PORT/somefile.txt -o /dev/null -w "%{http_code}\n"

Cleanup,
kill $SERVER_PID
rm -f index.html test.txt
echo "Tests completed.
