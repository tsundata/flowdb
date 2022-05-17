package flowdb

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func TestHash(t *testing.T) {
	require.Equal(t, uint64(0xf9e6e6ef197c2b25), Hash([]byte("test")))
}
