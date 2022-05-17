package flowdb

import (
	"errors"
	"fmt"
	"io"
	"io/fs"
	"io/ioutil"
	"log"
	"os"
	"path"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	FM    = os.FileMode(0750)
	FFlag = os.O_RDWR | os.O_APPEND | os.O_CREATE
	// Default max file size
	// 2 << 8 = 512 << 20 = 536870912 kb
	defaultMaxFileSize int64 = 2 << 8 << 20
)

type FlowDB struct {
	mu sync.RWMutex

	activeFile       *os.File
	activeFileOffset int64
	indexMap         map[uint64]*KeyDirRecord
	fileList         map[int64]*os.File
	dataFileVersion  int64

	options Options
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

	fileInfo, _ := f.activeFile.Stat()
	if fileInfo.Size() >= defaultMaxFileSize {
		err := f.closeActiveFile()
		if err != nil {
			return err
		}
		err = f.createActiveFile()
		if err != nil {
			return err
		}
	}

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

	// write hint
	if fd, err := f.openHintFile(f.dataFileVersion); err == nil {
		data, _ := EncodeHint(&Hint{
			Timestamp: uint64(timestamp),
			ValuePos:  uint64(f.activeFileOffset),
			Key:       sum64,
		})
		_, err = fd.Write(data)
		if err != nil {
			log.Fatalln(err)
		}
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

	if err := os.MkdirAll(path.Join(f.options.DatabaseDirectory, "data"), FM); err != nil {
		panic(err)
	}
	if err := os.MkdirAll(path.Join(f.options.DatabaseDirectory, "hint"), FM); err != nil {
		panic(err)
	}

	return f.createActiveFile()
}

func (f *FlowDB) Close() error {
	for _, fd := range f.fileList {
		err := fd.Close()
		if err != nil {
			return err
		}
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

func (f *FlowDB) closeActiveFile() error {
	f.mu.Lock()
	defer f.mu.Unlock()

	err := f.activeFile.Sync()
	if err != nil {
		return err
	}
	err = f.activeFile.Close()
	if err != nil {
		return err
	}
	df := path.Join(f.options.DatabaseDirectory, "data", fmt.Sprintf("%d.data", f.dataFileVersion))
	if fd, err := os.OpenFile(df, os.O_RDONLY, FM); err == nil {
		f.fileList[f.dataFileVersion] = fd
	}

	return errors.New("error close active file")
}

func (f *FlowDB) recoverData() error {
	f.version()
	if fd, err := f.openDataFile(f.dataFileVersion); err == nil {
		fileInfo, _ := fd.Stat()
		if fileInfo.Size() >= defaultMaxFileSize {
			if err := f.createActiveFile(); err != nil {
				return err
			}
			return f.buildIndex()
		}
		f.activeFile = fd
		if offset, err := fd.Seek(0, io.SeekEnd); err == nil {
			f.activeFileOffset = offset
		}
		return f.buildIndex()
	}

	return errors.New("failed to recover data")
}

func (f *FlowDB) buildIndex() error {
	if err := f.readHintFile(); err != nil {
		return err
	}

	for _, record := range f.indexMap {
		if f.fileList[record.fileId] == nil {
			fd, err := f.openDataFile(record.fileId)
			if err != nil {
				return err
			}
			f.fileList[record.fileId] = fd
		}
	}

	return nil
}

func (f *FlowDB) readHintFile() error {
	fid := f.findLatestHintFile()
	for i := int64(1); i <= fid; i++ {
		if fd, err := f.openHintFile(i); err == nil {
			buf := make([]byte, hintHeaderSize)
			for {
				_, err := fd.Read(buf)
				if err != nil && err != io.EOF {
					return err
				}
				if err == io.EOF {
					break
				}
				hint := DecodeHint(buf)
				f.indexMap[hint.Key] = &KeyDirRecord{
					fileId:    i,
					ValueSize: 0, // todo
					ValuePos:  int64(hint.ValuePos),
					Timestamp: int64(hint.Timestamp),
				}
			}
			err = fd.Close()
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (f *FlowDB) version() {
	f.dataFileVersion = f.findLatestDataFile()
}

func (f *FlowDB) findLatestHintFile() int64 {
	files, _ := ioutil.ReadDir(path.Join(f.options.DatabaseDirectory, "hint"))

	var datafiles []fs.FileInfo
	for _, file := range files {
		if path.Ext(file.Name()) == ".hint" {
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

	if len(ids) == 0 {
		return 1
	}

	return int64(ids[len(ids)-1])
}

func (f *FlowDB) findLatestDataFile() int64 {
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

	return int64(ids[len(ids)-1])
}

func (f *FlowDB) openDataFile(dataFileVersion int64) (*os.File, error) {
	df := path.Join(f.options.DatabaseDirectory, "data", fmt.Sprintf("%d.data", dataFileVersion))
	return os.OpenFile(df, FFlag, FM)
}

func (f *FlowDB) openHintFile(hintFileVersion int64) (*os.File, error) {
	df := path.Join(f.options.DatabaseDirectory, "hint", fmt.Sprintf("%d.hint", hintFileVersion))
	return os.OpenFile(df, FFlag, FM)
}
