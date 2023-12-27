package blob

import (
	"io"
	"io/fs"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNewDir(t *testing.T) {
	fi := &FileInfo{}
	d := NewDir(fi)
	s, err := d.Stat()
	require.NoError(t, err)
	require.Equal(t, fi, s)
	require.True(t, d.IsDir())
	require.Equal(t, fs.ModeDir, d.Mode())
	n, err := d.Read([]byte{})
	require.Equal(t, 0, n)
	require.Error(t, io.EOF, err)
	require.NoError(t, d.Close())

}
