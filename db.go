package flowdb

import (
	"errors"
	"fmt"
	"io"
	"io/fs"
	"io/ioutil"
	"os"
	"path"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

type FlowDB struct {
	mu sync.RWMutex

	activeFile       *os.File
	activeFileOffset int64
	indexMap         map[uint64]*KeyDirRecord
	fileList         map[int64]*os.File
	options          Options
	dataFileVersion  int64
}

type Options struct {
	DatabaseDirectory string
}

func DefaultOptions(directory string) Options {
	return Options{
		DatabaseDirectory: directory,
	}
}

func New(directory string) *FlowDB {
	return &FlowDB{
		mu:              sync.RWMutex{},
		activeFile:      nil,
		indexMap:        make(map[uint64]*KeyDirRecord),
		fileList:        make(map[int64]*os.File),
		options:         DefaultOptions(directory),
		dataFileVersion: 0,
	}
}

func (f *FlowDB) Get(key []byte) ([]byte, error) {
	f.mu.RLock()
	defer f.mu.RUnlock()

	sum64 := Hash(key)
	if f.indexMap[sum64] == nil {
		return nil, errors.New("key not exist")
	}
	if fd, ok := f.fileList[f.indexMap[sum64].fileId]; ok {
		data := make([]byte, f.indexMap[sum64].ValueSize)
		_, err := fd.ReadAt(data, f.indexMap[sum64].ValuePos)
		if err != nil {
			return nil, err
		}
		entry := DecodeEntry(data)
		return entry.Value, nil
	}

	return []byte{}, errors.New("failed to read")
}

func (f *FlowDB) Put(key, value []byte) error {
	f.mu.Lock()
	defer f.mu.Unlock()

	sum64 := Hash(key)
	timestamp := time.Now().UnixMicro()
	entryData, size := EncodeEntry(&Entry{
		Timestamp: uint64(timestamp),
		Key:       key,
		Value:     value,
	})
	_, err := f.activeFile.Write(entryData)
	if err != nil {
		return err
	}

	f.indexMap[sum64] = &KeyDirRecord{
		fileId:    f.dataFileVersion,
		ValueSize: size,
		ValuePos:  f.activeFileOffset,
		Timestamp: timestamp,
	}
	f.activeFileOffset += int64(size)

	return nil
}

func (f *FlowDB) Load() error {
	if ok, err := pathExists(f.options.DatabaseDirectory); ok {
		return f.recoverData()
	} else if err != nil {
		panic(err)
	}

	if err := os.MkdirAll(path.Join(f.options.DatabaseDirectory, "data"), os.FileMode(0750)); err != nil {
		panic(err)
	}
	if err := os.MkdirAll(path.Join(f.options.DatabaseDirectory, "index"), os.FileMode(0750)); err != nil {
		panic(err)
	}

	return f.createActiveFile()
}

func (f *FlowDB) createActiveFile() error {
	f.mu.Lock()
	defer f.mu.Unlock()

	f.dataFileVersion++
	if fd, err := f.openDataFile(f.dataFileVersion); err == nil {
		f.activeFile = fd
		f.fileList[f.dataFileVersion] = fd
		return nil
	}

	return errors.New("failed to create active file")
}

func (f *FlowDB) recoverData() error {
	f.version()
	if fd, err := f.openDataFile(f.dataFileVersion); err == nil {
		f.activeFile = fd
		if offset, err := fd.Seek(0, io.SeekEnd); err == nil {
			f.activeFileOffset = offset
		}
		return f.buildIndex()
	}

	return errors.New("failed to recover data")
}

func (f *FlowDB) buildIndex() error {
	// todo read index dir
	files, _ := ioutil.ReadDir(path.Join(f.options.DatabaseDirectory, "data"))

	var datafiles []fs.FileInfo
	for _, file := range files {
		if path.Ext(file.Name()) == ".data" {
			datafiles = append(datafiles, file)
		}
	}

	for _, info := range datafiles {
		id := strings.Split(info.Name(), ".")[0]
		i, _ := strconv.Atoi(id)

		fd, err := f.openDataFile(int64(i))
		if err != nil {
			return err
		}
		f.fileList[int64(i)] = fd
	}

	return nil
}

func (f *FlowDB) version() {
	files, _ := ioutil.ReadDir(path.Join(f.options.DatabaseDirectory, "data"))

	var datafiles []fs.FileInfo
	for _, file := range files {
		if path.Ext(file.Name()) == ".data" {
			datafiles = append(datafiles, file)
		}
	}
	var ids []int
	for _, info := range datafiles {
		id := strings.Split(info.Name(), ".")[0]
		i, _ := strconv.Atoi(id)
		ids = append(ids, i)
	}
	sort.Ints(ids)
	f.dataFileVersion = int64(ids[len(ids)-1])
}

func (f *FlowDB) Close() error {
	// todo fileList close
	err := f.activeFile.Close()
	if err != nil {
		return err
	}
	return nil
}

func (f *FlowDB) Sync() error {
	err := f.activeFile.Sync()
	if err != nil {
		return err
	}
	return nil
}

func (f *FlowDB) Merge() error {
	// todo merge gc
	return nil
}

func (f *FlowDB) openDataFile(dataFileVersion int64) (*os.File, error) {
	df := path.Join(f.options.DatabaseDirectory, "data", fmt.Sprintf("%d.data", dataFileVersion))
	return os.OpenFile(df, os.O_RDWR|os.O_APPEND|os.O_CREATE, os.FileMode(0750))
}
