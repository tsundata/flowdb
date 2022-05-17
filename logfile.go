package flowdb

import "os"

type KeyDirRecord struct {
	fileId    int64
	ValueSize uint32
	ValuePos  int64
	Timestamp int64
}

type HintFile struct {
	Timestamp int64
	KeySize   uint32
	ValuePos  int64
	Key       []byte
}

func pathExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

// todo hintFile encode/decode
