package flowdb

import (
	"encoding/binary"
)

type KeyDirRecord struct {
	fileId    int64
	ValueSize uint32
	ValuePos  int64
	Timestamp int64
}

type Hint struct {
	Timestamp uint64
	ValuePos  uint64
	Key       uint64
}

const hintHeaderSize = uint32(30)

func EncodeHint(h *Hint) ([]byte, uint32) {
	size := hintHeaderSize
	buf := make([]byte, size)

	// | TS 10  | VPOS 10  | KEY 10 |
	binary.BigEndian.PutUint64(buf[:10], h.Timestamp)
	binary.BigEndian.PutUint64(buf[10:20], h.ValuePos)
	binary.BigEndian.PutUint64(buf[20:30], h.Key)

	return buf, size
}

func DecodeHint(data []byte) *Hint {
	// | TS 10  | VPOS 10  | KEY 10 |
	var hint Hint
	hint.Timestamp = binary.BigEndian.Uint64(data[:10])
	hint.ValuePos = binary.BigEndian.Uint64(data[10:20])
	hint.Key = binary.BigEndian.Uint64(data[20:30])

	return &hint
}
