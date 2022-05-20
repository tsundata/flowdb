package main

import (
	"flag"
	"fmt"
	"github.com/hashicorp/raft"
	"github.com/tsundata/flowdb"
	"github.com/tsundata/flowdb/network"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"
)

var (
	dbDir        string
	raftDir      string
	serverConfig string
)

func main() {
	flag.StringVar(&serverConfig, "server_config", "server.json", "server config path")
	flag.Parse()

	if serverConfig == "" {
		panic("server config error")
	}

	network.Setting.Reload(serverConfig)

	dbDir = fmt.Sprintf("node/db_%s", network.Setting.Raft.Id)
	err := os.MkdirAll(dbDir, os.FileMode(0700))
	if err != nil {
		panic(err)
	}

	raftDir = fmt.Sprintf("node/raft_%s", network.Setting.Raft.Id)
	err = os.MkdirAll(raftDir, os.FileMode(0700))
	if err != nil {
		panic(err)
	}

	rf, _, err := flowdb.NewRaft(network.Setting.Raft.Addr, network.Setting.Raft.Id, raftDir)
	if err != nil {
		panic(err)
	}

	// bootstrap raft
	flowdb.Bootstrap(rf, network.Setting.Raft.Addr, network.Setting.Raft.Id, network.Setting.Raft.Cluster)

	server := network.NewServer()

	server.AddRouter(0, &GetRouter{rf: rf})
	server.AddRouter(1, &PutRouter{rf: rf})

	server.Serve()

	// close
	sig := make(chan os.Signal)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	<-sig

	server.Close()

	shutdownFuture := rf.Shutdown()
	if err := shutdownFuture.Error(); err != nil {
		log.Println("raft shutdown error", err)
	}
	log.Println("kv server shutdown")
}

type GetRouter struct {
	rf *raft.Raft
	network.BaseRouter
}

func (r *GetRouter) Handle(req network.IRequest) {
	log.Println("call get router Handle")
	err := req.GetConnection().SendMessage(1, []byte("ping\n"))
	if err != nil {
		log.Println(err)
	}
}

type PutRouter struct {
	rf *raft.Raft
	network.BaseRouter
}

func (r *PutRouter) Handle(req network.IRequest) {
	log.Println("call put router Handle")
	future := r.rf.Apply([]byte("test"), 5*time.Second)
	if err := future.Error(); err != nil {
		err := req.GetConnection().SendMessage(1, []byte(err.Error()+"\n"))
		if err != nil {
			log.Println(err)
		}
		return
	}
	err := req.GetConnection().SendMessage(1, []byte("ping\n"))
	if err != nil {
		log.Println(err)
	}
}
