package git

import (
	"bytes"
	"context"
	"io"
	"os"
	"testing"

	"github.com/go-git/go-git/v5/storage/memory"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"
)

func Test_filesystem(t *testing.T) {
	logger := zaptest.NewLogger(t)
	storer := memory.NewStorage()

	fs, err := newFilesystem(logger, storer)
	require.NoError(t, err)

	infos, err := fs.ReadDir(".")
	require.NoError(t, err)

	assert.Empty(t, infos, "unexpected returned set of infos")

	// Test MkdirAll
	t.Run("MkdirAll", func(t *testing.T) {
		err := fs.MkdirAll("test", 0755)
		require.NoError(t, err)

		infos, err := fs.ReadDir(".")
		require.NoError(t, err)

		assert.Len(t, infos, 1, "unexpected returned set of infos")
		assert.Equal(t, "test", infos[0].Name(), "unexpected name")
	})

	// Test OpenFile
	t.Run(`OpenFile("file.txt", os.O_CREATE|os.O_RDWR|os.O_TRUNC, 0755)`, func(t *testing.T) {
		file, err := fs.OpenFile("test.txt", os.O_CREATE|os.O_RDWR|os.O_TRUNC, 0755)
		require.NoError(t, err)
		require.NotNil(t, file)

		_, err = io.Copy(file, bytes.NewBufferString("hello world"))
		require.NoError(t, err)

		// filesystem does not update the tree until file is closed
		err = file.Close()
		require.NoError(t, err)

		infos, err := fs.ReadDir(".")
		require.NoError(t, err)

		assert.Len(t, infos, 2, "unexpected returned set of infos")
		assert.Equal(t, "test.txt", infos[0].Name(), "unexpected name")
		assert.Equal(t, "test", infos[1].Name(), "unexpected name")
	})

	commit, err := fs.commit(context.Background(), "add first file")
	require.NoError(t, err)

	fs, err = newFilesystem(logger, storer, withBaseCommit(commit.Hash))
	require.NoError(t, err)

	// Test OpenFile
	t.Run(`OpenFile("file.txt", os.O_RDONLY, 0755)`, func(t *testing.T) {
		file, err := fs.OpenFile("test.txt", os.O_RDONLY, 0755)
		require.NoError(t, err)
		require.NotNil(t, file)

		defer file.Close()

		data, err := io.ReadAll(file)
		require.NoError(t, err)

		assert.Equal(t, "hello world", string(data), "unexpected data")
	})

	// Test Stat
	t.Run("Stat", func(t *testing.T) {
		info, err := fs.Stat("test.txt")
		require.NoError(t, err)

		assert.Equal(t, "test.txt", info.Name(), "unexpected name")
		assert.Equal(t, int64(11), info.Size(), "unexpected size")
		assert.True(t, info.Mode().IsRegular(), "unexpected mode")
		assert.Nil(t, info.Sys(), "unexpected sys")
		assert.False(t, info.ModTime().IsZero(), "unexpected mod time")
		assert.False(t, info.IsDir(), "unexpected is dir")
	})

	// Test Remove
	t.Run("Remove", func(t *testing.T) {
		err := fs.Remove("test.txt")
		require.NoError(t, err)

		infos, err := fs.ReadDir(".")
		require.NoError(t, err)

		assert.Len(t, infos, 1, "unexpected returned set of infos")
		assert.Equal(t, "test", infos[0].Name(), "unexpected name")
	})
}
