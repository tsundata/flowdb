package main

import (
	"flag"
	"fmt"
	"github.com/tsundata/flowdb"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"
)

var (
	dbDir       string
	serverAddr  string
	raftAddr    string
	raftId      string
	raftCluster string
	raftDir     string
)

func main() {
	flag.StringVar(&serverAddr, "server_addr", "127.0.0.1:7000", "server listen addr")
	flag.StringVar(&raftAddr, "raft_addr", "127.0.0.1:5000", "raft listen addr")
	flag.StringVar(&raftId, "raft_id", "1", "raft id")
	flag.StringVar(&raftCluster, "raft_cluster",
		"1/127.0.0.1:5000,2/127.0.0.1:5001,3/127.0.0.1:5002",
		"raft cluster info")
	flag.Parse()

	if serverAddr == "" || raftAddr == "" || raftId == "" || raftCluster == "" {
		panic("server config error")
	}

	dbDir = fmt.Sprintf("node/db_%s", raftId)
	err := os.MkdirAll(dbDir, os.FileMode(0700))
	if err != nil {
		panic(err)
	}

	raftDir = fmt.Sprintf("node/raft_%s", raftId)
	err = os.MkdirAll(raftDir, os.FileMode(0700))
	if err != nil {
		panic(err)
	}

	rf, fsm, err := flowdb.NewRaft(raftAddr, raftId, raftDir)
	if err != nil {
		panic(err)
	}
	fmt.Println(fsm)

	// bootstrap raft
	flowdb.Bootstrap(rf, raftAddr, raftId, raftCluster)

	// tcp server
	listen, err := net.Listen("tcp", serverAddr)
	if err != nil {
		panic(err)
	}
	go func() {
		for {
			conn, err := listen.Accept()
			if err != nil {
				log.Println(err)
				return
			}
			go func() {
				data := make([]byte, 200)
				_, err = conn.Read(data)
				if err != nil {
					log.Println(err)
					return
				}
				fmt.Println(string(data))
				future := rf.Apply([]byte("test"), 5*time.Second)
				if err := future.Error(); err != nil {
					_, _ = conn.Write([]byte(err.Error()))
					return
				}
				//fsm.DataBase.Put([]byte("test"), data)
				//fsm.DataBase.Get([]byte("test"))
				_, _ = conn.Write([]byte("ok\n"))
			}()
		}
	}()

	// close
	sig := make(chan os.Signal)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	<-sig

	shutdownFuture := rf.Shutdown()
	if err := shutdownFuture.Error(); err != nil {
		log.Println("raft shutdown error", err)
	}
	log.Println("kv server shutdown")
}
