package network

import (
	"encoding/json"
	"io/ioutil"
)

type Config struct {
	Name    string `json:"name"`
	Host    string `json:"host"`
	TCPPort int    `json:"port"`
	Version string `json:"version"`

	MaxPacketSize     uint32 `json:"max_packet_size"`
	MaxConn           int    `json:"max_conn"`
	WorkerPoolSize    uint32 `json:"worker_pool_size"`
	MaxWorkerTaskLen  uint32 `json:"max_worker_task_len"`
	MaxMessageChanLen uint32 `json:"max_message_chan_len"`

	Raft struct {
		Id      string `json:"id"`
		Addr    string `json:"addr"`
		Cluster string `json:"cluster"`
	} `json:"raft"`
}

var Setting *Config

func (c *Config) Reload(path string) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		panic(err)
	}

	err = json.Unmarshal(data, &Setting)
	if err != nil {
		panic(err)
	}
}

func init() {
	// default config
	Setting = &Config{
		Name:              "server",
		Host:              "0.0.0.0",
		TCPPort:           5678,
		Version:           "0.1",
		MaxPacketSize:     4096,
		MaxConn:           1000,
		WorkerPoolSize:    10,
		MaxWorkerTaskLen:  1024,
		MaxMessageChanLen: 1024,
	}
}
