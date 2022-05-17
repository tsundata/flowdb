package flowdb

import (
	"github.com/stretchr/testify/require"
	"strconv"
	"testing"
)

func TestFlowDB(t *testing.T) {
	db := New("./testdata")
	err := db.Load()
	if err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 100; i++ {
		err = db.Put([]byte("k:"+strconv.Itoa(i)), []byte("test"))
		if err != nil {
			t.Fatal(err)
		}
		value, err := db.Get([]byte("k:" + strconv.Itoa(i)))
		if err != nil {
			t.Fatal(err)
		}
		require.Equal(t, []byte("test"), value)
	}
	err = db.Sync()
	if err != nil {
		t.Fatal(err)
	}
	err = db.Merge()
	if err != nil {
		t.Fatal(err)
	}
	err = db.Close()
	if err != nil {
		t.Fatal(err)
	}
}
