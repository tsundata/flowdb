package flowdb

import (
	"fmt"
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
	data, size := EncodeHint(&hint)
	fmt.Println(size)
	expect := DecodeHint(data)
	require.Equal(t, expect, &hint)
}
