package flowdb

import (
	"encoding/binary"
	"hash/crc32"
)

type Entry struct {
	CRC       uint32
	Timestamp uint64
	KeySize   uint32
	ValueSize uint32
	Key       []byte
	Value     []byte
}

const entryHeaderSize = 24

// EncodeEntry  entry into binary
func EncodeEntry(e *Entry) ([]byte, uint32) {
	e.KeySize = uint32(len(e.Key))
	e.ValueSize = uint32(len(e.Value))
	size := entryHeaderSize + e.KeySize + e.ValueSize
	buf := make([]byte, size)

	// | CRC 4 | TS 10  | KS 5 | VS 5  | KEY ? | VALUE ? |
	binary.BigEndian.PutUint64(buf[4:14], e.Timestamp)
	binary.BigEndian.PutUint32(buf[14:19], e.KeySize)
	binary.BigEndian.PutUint32(buf[19:24], e.ValueSize)

	copy(buf[entryHeaderSize:entryHeaderSize+e.KeySize], e.Key)
	copy(buf[entryHeaderSize+e.KeySize:size], e.Value)

	e.CRC = crc32.ChecksumIEEE(buf[4:])
	binary.BigEndian.PutUint32(buf[:4], e.CRC)
	return buf, size
}

//DecodeEntry binary into entry
func DecodeEntry(data []byte) *Entry {
	if binary.BigEndian.Uint32(data[:4]) != crc32.ChecksumIEEE(data[4:]) {
		return nil
	}

	// | CRC 4 | TS 10  | KS 5 | VS 5  | KEY ? | VALUE ? |
	var entry Entry
	entry.CRC = binary.BigEndian.Uint32(data[:4])
	entry.Timestamp = binary.BigEndian.Uint64(data[4:14])
	entry.KeySize = binary.BigEndian.Uint32(data[14:19])
	entry.ValueSize = binary.BigEndian.Uint32(data[19:24])

	entry.Key = make([]byte, entry.KeySize)
	entry.Value = make([]byte, entry.ValueSize)
	copy(entry.Key, data[entryHeaderSize:entryHeaderSize+entry.KeySize])
	copy(entry.Value, data[entryHeaderSize+entry.KeySize:entryHeaderSize+entry.KeySize+entry.ValueSize])

	return &entry
}
