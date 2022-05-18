# flowdb

[![Build](https://github.com/tsundata/flowdb/actions/workflows/go.yml/badge.svg)](https://github.com/tsundata/flowdb/actions/workflows/go.yml)

## Setup

```shell
# build
go build -o flowdb-server cmd/server/main.go
go build -o flowdb-client cmd/client/main.go

# run cluster
./flowdb-server --server_addr=127.0.0.1:7001 --raft_addr=127.0.0.1:5001 --raft_id=1 --raft_cluster=1/127.0.0.1:5001,2/127.0.0.1:5002,3/127.0.0.1:5003
./flowdb-server --server_addr=127.0.0.1:7002 --raft_addr=127.0.0.1:5002 --raft_id=2 --raft_cluster=1/127.0.0.1:5001,2/127.0.0.1:5002,3/127.0.0.1:5003
./flowdb-server --server_addr=127.0.0.1:7003 --raft_addr=127.0.0.1:5003 --raft_id=3 --raft_cluster=1/127.0.0.1:5001,2/127.0.0.1:5002,3/127.0.0.1:5003

# run client
./flowdb-client --server_addr=127.0.0.1:7001
```
