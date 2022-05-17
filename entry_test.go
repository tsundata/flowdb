package flowdb

import (
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func TestEntry(t *testing.T) {
	entry := Entry{
		Timestamp: uint64(time.Now().UnixMicro()),
		Key:       []byte("test"),
		Value:     []byte("test"),
	}
	data, _ := EncodeEntry(&entry)
	expect := DecodeEntry(data)
	require.Equal(t, expect, &entry)
}
