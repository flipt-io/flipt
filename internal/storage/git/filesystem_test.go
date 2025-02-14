package git

import (
	"bytes"
	"context"
	"io"
	"io/fs"
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

	gitFS, err := newFilesystem(logger, storer)
	require.NoError(t, err)

	infos, err := gitFS.ReadDir(".")
	require.NoError(t, err)

	assert.Empty(t, infos, "unexpected returned set of infos")

	t.Run(`Stat(".")`, func(t *testing.T) {
		info, err := gitFS.Stat(".")
		require.NoError(t, err)

		assert.Equal(t, ".", info.Name(), "unexpected name")
		assert.True(t, info.Mode().IsDir(), "unexpected mode")
		assert.Nil(t, info.Sys(), "unexpected sys")
		assert.True(t, info.IsDir(), "unexpected is dir")
	})

	// Test MkdirAll
	t.Run("MkdirAll", func(t *testing.T) {
		err := gitFS.MkdirAll("test", 0755)
		require.NoError(t, err)

		infos, err := gitFS.ReadDir(".")
		require.NoError(t, err)

		assert.Len(t, infos, 1, "unexpected returned set of infos")
		assert.Equal(t, "test", infos[0].Name(), "unexpected name")
	})

	// Test OpenFile
	t.Run(`OpenFile("file.txt", os.O_CREATE|os.O_RDWR|os.O_TRUNC, 0755)`, func(t *testing.T) {
		file, err := gitFS.OpenFile("test.txt", os.O_CREATE|os.O_RDWR|os.O_TRUNC, 0755)
		require.NoError(t, err)
		require.NotNil(t, file)

		_, err = io.Copy(file, bytes.NewBufferString("hello world"))
		require.NoError(t, err)

		// filesystem does not update the tree until file is closed
		err = file.Close()
		require.NoError(t, err)

		infos, err := gitFS.ReadDir(".")
		require.NoError(t, err)

		assert.Len(t, infos, 2, "unexpected returned set of infos")
		assert.Equal(t, "test.txt", infos[0].Name(), "unexpected name")
		assert.Equal(t, "test", infos[1].Name(), "unexpected name")
	})

	commit, err := gitFS.commit(context.Background(), "add first file")
	require.NoError(t, err)

	gitFS, err = newFilesystem(logger, storer, withBaseCommit(commit.Hash))
	require.NoError(t, err)

	// Test OpenFile
	t.Run(`OpenFile("file.txt", os.O_RDONLY, 0755)`, func(t *testing.T) {
		file, err := gitFS.OpenFile("test.txt", os.O_RDONLY, 0755)
		require.NoError(t, err)
		require.NotNil(t, file)

		defer file.Close()

		data, err := io.ReadAll(file)
		require.NoError(t, err)

		assert.Equal(t, "hello world", string(data), "unexpected data")
	})

	// Test Stat
	t.Run(`Stat("test.txt")`, func(t *testing.T) {
		info, err := gitFS.Stat("test.txt")
		require.NoError(t, err)

		assert.Equal(t, "test.txt", info.Name(), "unexpected name")
		assert.Equal(t, int64(11), info.Size(), "unexpected size")
		assert.True(t, info.Mode().IsRegular(), "unexpected mode")
		assert.Nil(t, info.Sys(), "unexpected sys")
		assert.False(t, info.ModTime().IsZero(), "unexpected mod time")
		assert.False(t, info.IsDir(), "unexpected is dir")
	})

	t.Run("ReadDirFile", func(t *testing.T) {
		fi, err := gitFS.OpenFile(".", os.O_RDONLY, 0755)
		require.NoError(t, err)

		dir, ok := fi.(fs.ReadDirFile)
		require.True(t, ok)

		entries, err := dir.ReadDir(0)
		require.NoError(t, err)

		assert.Len(t, entries, 2, "unexpected returned set of infos")
		assert.Equal(t, "test.txt", entries[0].Name(), "unexpected name")
		assert.Equal(t, "test", entries[1].Name(), "unexpected name")
	})

	// Test Remove
	t.Run("Remove", func(t *testing.T) {
		err := gitFS.Remove("test.txt")
		require.NoError(t, err)

		infos, err := gitFS.ReadDir(".")
		require.NoError(t, err)

		assert.Len(t, infos, 1, "unexpected returned set of infos")
		assert.Equal(t, "test", infos[0].Name(), "unexpected name")
	})
}
