package flowdb

import (
	"fmt"
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
	data, size := EncodeEntry(&entry)
	fmt.Println(size)
	expect := DecodeEntry(data)
	require.Equal(t, expect, &entry)
}
