package osfs

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNew(t *testing.T) {
	fs := New("/tmp")
	assert.NotNil(t, fs)
}

func TestFilesystem_Operations(t *testing.T) {
	// Create a temporary directory for testing
	tmpDir := t.TempDir()
	fs := New(tmpDir)

	t.Run("MkdirAll", func(t *testing.T) {
		err := fs.MkdirAll("testdir", 0755)
		require.NoError(t, err)

		info, err := fs.Stat("testdir")
		require.NoError(t, err)
		assert.True(t, info.IsDir())
	})

	t.Run("OpenFile for writing", func(t *testing.T) {
		f, err := fs.OpenFile("testfile.txt", os.O_CREATE|os.O_WRONLY, 0644)
		require.NoError(t, err)

		content := []byte("test content")
		_, err = f.Write(content)
		require.NoError(t, err)
		require.NoError(t, f.Close())

		info, err := fs.Stat("testfile.txt")
		require.NoError(t, err)
		assert.Equal(t, int64(len(content)), info.Size())
	})

	t.Run("Open for reading", func(t *testing.T) {
		f, err := fs.Open("testfile.txt")
		require.NoError(t, err)

		buf := make([]byte, 100)
		n, err := f.Read(buf)
		require.NoError(t, err)
		assert.Equal(t, "test content", string(buf[:n]))
		require.NoError(t, f.Close())
	})

	t.Run("Stat", func(t *testing.T) {
		info, err := fs.Stat("testfile.txt")
		require.NoError(t, err)
		assert.Equal(t, "testfile.txt", info.Name())
		assert.False(t, info.IsDir())
	})

	t.Run("ReadDir", func(t *testing.T) {
		// Create a few files in testdir
		require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "testdir", "file1.txt"), []byte("content1"), 0644))
		require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "testdir", "file2.txt"), []byte("content2"), 0644))

		entries, err := fs.ReadDir("testdir")
		require.NoError(t, err)
		assert.Len(t, entries, 2)

		// Verify entries are sorted by filename
		assert.Equal(t, "file1.txt", entries[0].Name())
		assert.Equal(t, "file2.txt", entries[1].Name())
	})

	t.Run("Remove", func(t *testing.T) {
		err := fs.Remove("testfile.txt")
		require.NoError(t, err)

		_, err = fs.Stat("testfile.txt")
		assert.True(t, os.IsNotExist(err))
	})

	t.Run("error cases", func(t *testing.T) {
		// Test opening non-existent file
		_, err := fs.Open("nonexistent.txt")
		assert.Error(t, err)

		// Test stat on non-existent file
		_, err = fs.Stat("nonexistent.txt")
		assert.Error(t, err)

		// Test reading non-existent directory
		_, err = fs.ReadDir("nonexistentdir")
		assert.Error(t, err)

		// Test removing non-existent file
		err = fs.Remove("nonexistent.txt")
		assert.Error(t, err)
	})
}

func TestFile_Operations(t *testing.T) {
	tmpDir := t.TempDir()
	fs := New(tmpDir)

	t.Run("file stat matches filesystem stat", func(t *testing.T) {
		// Create a test file
		f, err := fs.OpenFile("stattest.txt", os.O_CREATE|os.O_WRONLY, 0644)
		require.NoError(t, err)
		_, err = f.Write([]byte("test content"))
		require.NoError(t, err)
		require.NoError(t, f.Close())

		// Open the file and check its stat
		f, err = fs.Open("stattest.txt")
		require.NoError(t, err)
		defer f.Close()

		fstat, err := f.Stat()
		require.NoError(t, err)

		fsstat, err := fs.Stat("stattest.txt")
		require.NoError(t, err)

		assert.Equal(t, fsstat.Name(), fstat.Name())
		assert.Equal(t, fsstat.Size(), fstat.Size())
		assert.Equal(t, fsstat.Mode(), fstat.Mode())
		assert.Equal(t, fsstat.ModTime(), fstat.ModTime())
		assert.Equal(t, fsstat.IsDir(), fstat.IsDir())
	})
}
