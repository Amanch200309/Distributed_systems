#!/bin/bash
set -euo pipefail

# ----------------------------
# Configuration
# ----------------------------
PORT=8080
BINARY="./http_server"
TMPDIR="./test_files"
LOGFILE="./server.log"

# Cleanup old files
rm -rf "$TMPDIR" "$LOGFILE" test.html
mkdir -p "$TMPDIR"

# ----------------------------
# 1. Build the server
# ----------------------------
echo "[BUILD] Compiling Go server..."
go build -o http_server main.go

# ----------------------------
# 2. Start the server
# ----------------------------
echo "[START] Launching server on port $PORT..."
$BINARY $PORT >"$LOGFILE" 2>&1 &
SERVER_PID=$!
sleep 1

if ! ps -p $SERVER_PID >/dev/null; then
  echo "[ERROR] Failed to start server."
  cat "$LOGFILE"
  exit 1
fi

echo "[INFO] Server running (PID: $SERVER_PID)"

# ----------------------------
# 3. Test POST request (create file)
# ----------------------------
echo "[TEST] POST /test.html"
STATUS=$(curl -s -o /dev/null -w "%{http_code}" \
  -X POST "http://localhost:$PORT/test.html" \
  -H "Content-Type: text/html" \
  -d "<h1>Hello</h1>")
if [[ "$STATUS" != "201" ]]; then
  echo "[FAIL] POST returned HTTP $STATUS, expected 201."
  kill $SERVER_PID
  exit 1
fi
echo "[PASS] POST returned 201 Created"

if [[ ! -f "./test.html" ]]; then
  echo "[FAIL] File test.html was not created."
  kill $SERVER_PID
  exit 1
fi
echo "[PASS] File saved successfully"

# ----------------------------
# 4. Test GET request
# ----------------------------
echo "[TEST] GET /test.html"
STATUS=$(curl -s -o /dev/null -w "%{http_code}" "http://localhost:$PORT/test.html")
if [[ "$STATUS" != "200" ]]; then
  echo "[FAIL] GET returned HTTP $STATUS, expected 200."
  kill $SERVER_PID
  exit 1
fi
echo "[PASS] GET returned 200 OK"

# ----------------------------
# 5. Test unsupported file extension
# ----------------------------
echo "[TEST] GET /invalid.xyz"
STATUS=$(curl -s -o /dev/null -w "%{http_code}" "http://localhost:$PORT/invalid.xyz")
if [[ "$STATUS" != "400" ]]; then
  echo "[FAIL] Unsupported extension returned HTTP $STATUS, expected 400."
  kill $SERVER_PID
  exit 1
fi
echo "[PASS] Unsupported extension returned 400 Bad Request"

# ----------------------------
# 6. Test missing file (404)
# ----------------------------
echo "[TEST] GET /missing.html"
STATUS=$(curl -s -o /dev/null -w "%{http_code}" "http://localhost:$PORT/missing.html")
if [[ "$STATUS" != "404" ]]; then
  echo "[FAIL] Missing file returned HTTP $STATUS, expected 404."
  kill $SERVER_PID
  exit 1
fi
echo "[PASS] Missing file returned 404 Not Found"

# ----------------------------
# 7. Test unsupported method (501)
# ----------------------------
echo "[TEST] PUT /test.html"
STATUS=$(curl -s -o /dev/null -w "%{http_code}" -X PUT "http://localhost:$PORT/test.html")
if [[ "$STATUS" != "501" ]]; then
  echo "[FAIL] Unsupported method returned HTTP $STATUS, expected 501."
  kill $SERVER_PID
  exit 1
fi
echo "[PASS] Unsupported method returned 501 Not Implemented"


# ----------------------------
# 9. Verify file content
# ----------------------------
EXPECTED_CONTENT="<h1>Hello</h1>"
ACTUAL_CONTENT=$(cat test.html)
if [[ "$ACTUAL_CONTENT" != "$EXPECTED_CONTENT" ]]; then
  echo "[FAIL] File content does not match expected value."
  kill $SERVER_PID
  exit 1
fi
echo "[PASS] File content verified"

# ----------------------------
# 10. Cleanup
# ----------------------------
echo "[CLEANUP] Stopping server..."
kill $SERVER_PID
wait $SERVER_PID 2>/dev/null || true
rm -f test.html

echo "[DONE] All tests passed successfully."
