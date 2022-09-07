#!/bin/bash

trap "rm server; kill 0" EXIT

go build -o server

./server -port=8081 -api=1 &
./server -port=8082 &
./server -port=8083 &

sleep 2


echo ">>> start test"

curl "http://localhost:9090/api?key=Jack" &
curl "http://localhost:9090/api?key=Jack" &
curl "http://localhost:9090/api?key=Jack" &

wait
