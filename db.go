package flowdb

import (
	"os"
	"sync"
)

type FlowDB struct {
	mu sync.RWMutex

	activeFile *os.File
	indexMap   map[uint32]*os.File
	options    Options
}

type Options struct {
	DbDirectory string
}

func DefaultOptions(directory string) Options {
	return Options{
		DbDirectory: directory,
	}
}

func New(directory string) *FlowDB {
	return &FlowDB{
		mu:         sync.RWMutex{},
		activeFile: nil,
		indexMap:   make(map[uint32]*os.File),
		options:    DefaultOptions(directory),
	}
}

func (f *FlowDB) Load() error {
	return nil
}

func (f *FlowDB) Close() error {
	return nil
}

func (f *FlowDB) Sync() error {
	return nil
}

func (f *FlowDB) Merge() error {
	return nil
}
