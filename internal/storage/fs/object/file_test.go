package object

import (
	"io"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestNewFile(t *testing.T) {
	modTime := time.Now()
	r := io.NopCloser(strings.NewReader("hello"))
	f := NewFile("f.txt", 5, r, modTime)
	fi, err := f.Stat()
	require.NoError(t, err)
	require.Equal(t, "f.txt", fi.Name())
	require.Equal(t, int64(5), fi.Size())
	require.Equal(t, modTime, fi.ModTime())
	buf := make([]byte, fi.Size())
	n, err := f.Read(buf)
	require.NoError(t, err)
	require.Equal(t, 5, n)
	require.Equal(t, []byte("hello"), buf)
	err = f.Close()
	require.NoError(t, err)
}
