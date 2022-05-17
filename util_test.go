package flowdb

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func TestPathExits(t *testing.T) {
	b, err := pathExists("/tmp")
	if err != nil {
		t.Fatal(err)
	}
	require.True(t, b)
}
