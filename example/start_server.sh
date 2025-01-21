#!/bin/bash

# Start Comical-KV server peers
# rm server binary if it exists
trap "rm server;kill 0" EXIT

# build server binary
go build -o server

# start server peers
./server -port 8080 &
./server -port 8081 &
./server -port 8082 -api=1 -apiPort=9999 -peers="http://localhost:8080,http://localhost:8081,http://localhost:8082" &

# wait for server peers to start
sleep 2

# start test
echo "Starting test..."
curl "http://localhost:9999/api/Tom" &
curl "http://localhost:9999/api/Tom" &
curl "http://localhost:9999/api/Tom" &

# wait for test to finish
wait

# clean up & stop server peers
trap "rm server;kill 0" EXIT
