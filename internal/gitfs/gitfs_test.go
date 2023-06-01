package gitfs

import (
	"embed"
	"io"
	"io/fs"
	"testing"

	"github.com/go-git/go-billy/v5/memfs"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/storage/memory"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var expected = map[string]string{
	"one.txt":           "one.txt\n",
	"two/three.txt":     "three.txt\n",
	"four/five/six.txt": "six.txt\n",
}

func Test_FS(t *testing.T) {
	repo := testdataRepo(t)

	filesystem, err := NewFromRepo(repo)
	require.NoError(t, err)

	t.Run("Ensure invalid and non existent paths error", func(t *testing.T) {
		_, err := filesystem.Open("..")
		require.Equal(t, err, &fs.PathError{
			Op:   "Open",
			Path: "..",
			Err:  fs.ErrInvalid,
		})

		_, err = filesystem.Open("zero.txt")
		require.Equal(t, err, &fs.PathError{
			Op:   "Open",
			Path: "zero.txt",
			Err:  fs.ErrNotExist,
		})
	})

	t.Run("Ensure files exist with expected contents", func(t *testing.T) {
		seen := map[string]string{}
		err = fs.WalkDir(filesystem, ".", func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}

			if d.IsDir() {
				return nil
			}

			fi, err := filesystem.Open(path)
			require.NoError(t, err)

			defer fi.Close()

			contents, err := io.ReadAll(fi)
			require.NoError(t, err)

			seen[path] = string(contents)

			return nil
		})
		require.NoError(t, err)

		assert.Equal(t, expected, seen)
	})
}

//go:embed all:testdata/*
var testdata embed.FS

func testdataRepo(t *testing.T) *git.Repository {
	t.Helper()

	workdir := memfs.New()

	repo, err := git.Init(memory.NewStorage(), workdir)
	require.NoError(t, err)

	dir, err := fs.Sub(testdata, "testdata")
	require.NoError(t, err)

	// copy testdata into target tmp dir
	fs.WalkDir(dir, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() {
			err := workdir.MkdirAll(path, 0755)
			require.NoError(t, err)
			return nil
		}

		contents, err := fs.ReadFile(dir, path)
		require.NoError(t, err)

		fi, err := workdir.Create(path)
		require.NoError(t, err)

		_, err = fi.Write(contents)
		require.NoError(t, err)

		require.NoError(t, fi.Close())

		return nil
	})

	tree, err := repo.Worktree()
	require.NoError(t, err)

	err = tree.AddWithOptions(&git.AddOptions{All: true})
	require.NoError(t, err)

	_, err = tree.Commit("feat: add entire contents", &git.CommitOptions{})
	require.NoError(t, err)

	return repo
}
