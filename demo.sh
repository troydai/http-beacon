#!/bin/bash

# Start server in background and capture PID
go run main.go &
SERVER_PID=$!
echo "Server PID: $SERVER_PID"

# Wait for server to start
sleep 1

# Run curl commands...

# HTTP/1.1 TLS
curl -v --cacert certs/cert.pem  https://localhost:8443/example --http1.1

# HTTP/2 TLS
curl -v --cacert certs/cert.pem  https://localhost:8443/example --http2

# Kill server at end
lsof -t -i :8443 | xargs kill -9