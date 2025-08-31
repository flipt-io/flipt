package git

import (
	"bytes"
	"context"
	"io"
	"io/fs"
	"os"
	"testing"

	"github.com/go-git/go-git/v6/storage/memory"
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

// Test_RapidNamespaceDeletion tests the scenario that was causing Bitbucket 500 errors
// This simulates rapid deletion of namespaces which was creating invalid tree objects
func Test_RapidNamespaceDeletion(t *testing.T) {
	logger := zaptest.NewLogger(t)
	storer := memory.NewStorage()

	gitFS, err := newFilesystem(logger, storer)
	require.NoError(t, err)

	// Create multiple namespaces with files
	namespaces := []string{"test1", "test2", "test3", "test4"}

	for _, ns := range namespaces {
		// Create namespace directory
		err := gitFS.MkdirAll(ns, 0755)
		require.NoError(t, err)

		// Create a features.yaml file in each namespace
		file, err := gitFS.OpenFile(ns+"/features.yaml", os.O_CREATE|os.O_WRONLY, 0644)
		require.NoError(t, err)

		_, err = file.Write([]byte("namespace: " + ns + "\n"))
		require.NoError(t, err)

		err = file.Close()
		require.NoError(t, err)
	}

	// Verify all namespaces were created
	infos, err := gitFS.ReadDir(".")
	require.NoError(t, err)
	assert.Len(t, infos, 4, "should have 4 namespaces")

	// Now rapidly delete namespaces (simulating the issue scenario)
	for _, ns := range namespaces {
		// First remove the file
		err := gitFS.Remove(ns + "/features.yaml")
		require.NoError(t, err, "failed to remove features.yaml from %s", ns)

		// Note: The directory removal happens in the tree, but the current filesystem
		// view still has the old tree until we create a new filesystem from the commit
		// This simulates what happens in the actual environment where each operation
		// creates a new commit and filesystem
	}

	// Verify all namespaces are gone
	infos, err = gitFS.ReadDir(".")
	require.NoError(t, err)
	assert.Empty(t, infos, "all namespaces should be deleted")

	// Create a commit to ensure the tree is valid
	ctx := context.Background()
	commit, err := gitFS.commit(ctx, "deleted all namespaces")
	require.NoError(t, err, "failed to create commit after deletions")
	require.NotNil(t, commit, "commit should not be nil")

	// Verify we can read the tree from the commit (simulating what Bitbucket would do)
	tree, err := commit.Tree()
	require.NoError(t, err, "failed to get tree from commit")
	require.NotNil(t, tree, "tree should not be nil")

	// Tree should be empty after all deletions
	assert.Empty(t, tree.Entries, "tree should have no entries after all deletions")
}

// Test_NamespaceWithSubdirectories tests deletion of namespaces with nested directories
func Test_NamespaceWithSubdirectories(t *testing.T) {
	logger := zaptest.NewLogger(t)
	storer := memory.NewStorage()

	gitFS, err := newFilesystem(logger, storer)
	require.NoError(t, err)

	// Create a namespace with subdirectories
	err = gitFS.MkdirAll("namespace/subdir1", 0755)
	require.NoError(t, err)

	err = gitFS.MkdirAll("namespace/subdir2", 0755)
	require.NoError(t, err)

	// Add files to subdirectories
	file1, err := gitFS.OpenFile("namespace/subdir1/file1.yaml", os.O_CREATE|os.O_WRONLY, 0644)
	require.NoError(t, err)
	_, err = file1.Write([]byte("content1"))
	require.NoError(t, err)
	err = file1.Close()
	require.NoError(t, err)

	file2, err := gitFS.OpenFile("namespace/subdir2/file2.yaml", os.O_CREATE|os.O_WRONLY, 0644)
	require.NoError(t, err)
	_, err = file2.Write([]byte("content2"))
	require.NoError(t, err)
	err = file2.Close()
	require.NoError(t, err)

	// Delete files from subdirectories
	err = gitFS.Remove("namespace/subdir1/file1.yaml")
	require.NoError(t, err)

	// Note: The filesystem maintains a snapshot view, so directories
	// appear to still exist until a new filesystem is created from a commit
	// The actual removal happens in the tree structure

	// namespace and subdir2 still exist in the snapshot
	info, err := gitFS.Stat("namespace")
	require.NoError(t, err)
	assert.True(t, info.IsDir())

	info, err = gitFS.Stat("namespace/subdir2")
	require.NoError(t, err)
	assert.True(t, info.IsDir())

	// Delete the remaining file
	err = gitFS.Remove("namespace/subdir2/file2.yaml")
	require.NoError(t, err)

	// Verify we can commit successfully
	ctx := context.Background()
	commit, err := gitFS.commit(ctx, "deleted namespace with subdirectories")
	require.NoError(t, err)
	assert.NotNil(t, commit)
}
