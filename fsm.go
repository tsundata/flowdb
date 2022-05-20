package flowdb

import (
	"fmt"
	"github.com/hashicorp/raft"
	"io"
	"log"
	"time"
)

type FSM struct {
	DataBase database
}

func NewFSM() *FSM {
	return &FSM{
		DataBase: newDatabase(),
	}
}

func (f *FSM) Apply(log *raft.Log) interface{} {
	fmt.Println("apply data", string(log.Data))
	f.DataBase.Put([]byte("test"), log.Data) //todo
	return nil
}

func (f *FSM) Snapshot() (raft.FSMSnapshot, error) {
	return &f.DataBase, nil
}

func (f *FSM) Restore(snapshot io.ReadCloser) error {
	return nil
}

type database struct {
}

func newDatabase() database {
	return database{}
}

func (d *database) Get(key []byte) []byte {
	return key
}

func (d *database) Put(key, value []byte) {
	fmt.Println(key, value)
}

func (d *database) Persist(sink raft.SnapshotSink) error {
	_, err := sink.Write([]byte(time.Now().String()))
	if err != nil {
		log.Println(err)
	}
	err = sink.Close()
	if err != nil {
		log.Println(err)
	}
	return nil
}

func (d *database) Release() {}
