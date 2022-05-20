# flowDB

[![Build](https://github.com/tsundata/flowdb/actions/workflows/go.yml/badge.svg)](https://github.com/tsundata/flowdb/actions/workflows/go.yml)

## Setup

```shell
# build
go build -o flowdb-server cmd/server/main.go
go build -o flowdb-client cmd/client/main.go

# run cluster
./flowdb-server --server_config=server1.json
./flowdb-server --server_config=server2.json
./flowdb-server --server_config=server3.json

# run client
./flowdb-client --server_addr=127.0.0.1:7001
```

### Node1 config (server1.json)
```json
{
  "name": "node1",
  "host": "127.0.0.1",
  "port": 7001,
  "version": "0.1",
  "max_packet_size": 4096,
  "max_conn": 1000,
  "raft": {
    "id": "1",
    "addr": "127.0.0.1:5001",
    "cluster": "1/127.0.0.1:5001,2/127.0.0.1:5002,3/127.0.0.1:5003"
  }
}
```

### Node2 config (server2.json)
```json
{
  "name": "node2",
  "host": "127.0.0.1",
  "port": 7002,
  "version": "0.1",
  "max_packet_size": 4096,
  "max_conn": 1000,
  "raft": {
    "id": "2",
    "addr": "127.0.0.1:5002",
    "cluster": "1/127.0.0.1:5001,2/127.0.0.1:5002,3/127.0.0.1:5003"
  }
}
```

### Node3 config (server3.json)
```json
{
  "name": "node3",
  "host": "127.0.0.1",
  "port": 7003,
  "version": "0.1",
  "max_packet_size": 4096,
  "max_conn": 1000,
  "raft": {
    "id": "3",
    "addr": "127.0.0.1:5003",
    "cluster": "1/127.0.0.1:5001,2/127.0.0.1:5002,3/127.0.0.1:5003"
  }
}
```
