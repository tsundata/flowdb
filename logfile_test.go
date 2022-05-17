package flowdb

import (
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func TestHint(t *testing.T) {
	hint := Hint{
		Timestamp: uint64(time.Now().UnixMicro()),
		ValuePos:  100,
		Key:       Hash([]byte("test")),
	}
	data, _ := EncodeHint(&hint)
	expect := DecodeHint(data)
	require.Equal(t, expect, &hint)
}
