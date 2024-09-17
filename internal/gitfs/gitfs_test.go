package gitfs

import (
	"embed"
	"io"
	"io/fs"
	"path"
	"strings"
	"testing"

	"github.com/go-git/go-billy/v5/memfs"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/storage/memory"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"
)

var expected = map[string]string{
	"one.txt":           "one.txt\n",
	"two/three.txt":     "three.txt\n",
	"four/five/six.txt": "six.txt\n",
}

func Test_FS(t *testing.T) {
	repo := testdataRepo(t, "simple")

	filesystem, err := NewFromRepo(zaptest.NewLogger(t), repo)
	require.NoError(t, err)

	t.Run("Ensure invalid and non existent paths produce an error", func(t *testing.T) {
		_, err := filesystem.Open("..")
		require.Equal(t, &fs.PathError{
			Op:   "Open",
			Path: "..",
			Err:  fs.ErrInvalid,
		}, err)

		_, err = filesystem.Open("zero.txt")
		require.Equal(t, &fs.PathError{
			Op:   "Open",
			Path: "zero.txt",
			Err:  fs.ErrNotExist,
		}, err)
	})

	t.Run("Ensure files exist with expected contents", func(t *testing.T) {
		seen := map[string]string{}
		dirs := map[string]int{}

		err := fs.WalkDir(filesystem, ".", func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}

			fi, err := filesystem.Open(path)
			require.NoError(t, err)

			defer fi.Close()

			if d.IsDir() {
				dir := requireCast[*Dir](t, fi)

				entries, err := dir.ReadDir(0)
				require.NoError(t, err)

				dirs[path] = len(entries)

				return nil
			}

			contents, err := io.ReadAll(fi)
			require.NoError(t, err)

			seen[path] = string(contents)

			return nil
		})
		require.NoError(t, err)

		assert.Equal(t, expected, seen)
		assert.Equal(t, map[string]int{
			".":         3,
			"two":       1,
			"four":      1,
			"four/five": 1,
		}, dirs)
	})

	t.Run("Walk root directory using ReadDir", func(t *testing.T) {
		fi, err := filesystem.Open(".")
		require.NoError(t, err)
		defer fi.Close()

		dir := requireCast[*Dir](t, fi)

		var ent, all []fs.DirEntry
		for ent, err = dir.ReadDir(1); err == nil; ent, err = dir.ReadDir(1) {
			require.NoError(t, err)
			assert.Len(t, ent, 1)

			all = append(all, ent...)
		}
		assert.Equal(t, io.EOF, err, "expected ReadDir to end in io.EOF")
		assert.Len(t, all, 3, "expected root dir to contain three entries")

		var (
			infos    []fs.FileInfo
			expected = []fs.FileInfo{
				FileInfo{name: "four", size: int64(0), mode: fs.ModeDir | fs.ModePerm},
				FileInfo{name: "one.txt", size: int64(8), mode: fs.FileMode(0644)},
				FileInfo{name: "two", size: int64(0), mode: fs.ModeDir | fs.ModePerm},
			}
		)
		for _, e := range all {
			info, err := e.Info()
			require.NoError(t, err)
			infos = append(infos, info)
		}

		assert.Equal(t, expected, infos)
	})

	t.Run("File.Stat returns as expected", func(t *testing.T) {
		fi, err := filesystem.Open("one.txt")
		require.NoError(t, err)
		defer fi.Close()

		stat, err := fi.Stat()
		require.NoError(t, err)

		assert.Equal(t, "one.txt", stat.Name())
		assert.Equal(t, int64(8), stat.Size())
		assert.Equal(t, fs.FileMode(0644), stat.Mode())
		assert.False(t, stat.IsDir(), "file stat reports expected file is a directory")
	})

	t.Run("File.Seek", func(t *testing.T) {
		fi := &File{ReadCloser: readCloser("cannot be seeked")}
		n, err := fi.Seek(4, io.SeekStart)
		require.Error(t, err, "seeker cannot seek")
		assert.Zero(t, n)

		fi = &File{ReadCloser: closer{strings.NewReader("seeker can seek")}}
		n, err = fi.Seek(7, io.SeekStart)
		require.NoError(t, err)
		assert.Equal(t, int64(7), n)

		contents, err := io.ReadAll(fi)
		require.NoError(t, err)
		assert.Equal(t, "can seek", string(contents))
	})
}

func Test_FS_Submodule(t *testing.T) {
	store := memory.NewStorage()
	work := memfs.New()
	repo, err := git.Clone(store, work, &git.CloneOptions{
		URL: "https://github.com/flipt-io/flipt-gitops-test.git",
	})
	require.NoError(t, err)

	// build gitfs instance on parent repo
	filesystem, err := NewFromRepo(zaptest.NewLogger(t), repo)
	require.NoError(t, err)

	require.NoError(t, fs.WalkDir(filesystem, ".", func(path string, d fs.DirEntry, err error) error {
		t.Log("open", path)
		return err
	}))
}

type closer struct {
	io.ReadSeeker
}

func (c closer) Close() error { return nil }

// readCloser is a strings reader which does not implement Seek
type readCloser string

func (r readCloser) Read(d []byte) (int, error) {
	copy(d, []byte(r))
	return len(r), nil
}

func (r readCloser) Close() error { return nil }

func requireCast[T any](t *testing.T, v any) (c T) {
	c, ok := v.(T)
	require.True(t, ok, "expected %T, found %T", c, v)
	return c
}

//go:embed all:testdata/*
var testdata embed.FS

func testdataRepo(t *testing.T, sub string) *git.Repository {
	t.Helper()

	workdir := memfs.New()

	repo, err := git.Init(memory.NewStorage(), workdir)
	require.NoError(t, err)

	dir, err := fs.Sub(testdata, path.Join("testdata", sub))
	require.NoError(t, err)

	// copy testdata into target tmp dir
	require.NoError(t, fs.WalkDir(dir, ".", func(path string, d fs.DirEntry, err error) error {
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
	}))

	tree, err := repo.Worktree()
	require.NoError(t, err)

	err = tree.AddWithOptions(&git.AddOptions{All: true})
	require.NoError(t, err)

	_, err = tree.Commit("feat: add entire contents", &git.CommitOptions{
		Author: &object.Signature{Email: "dev@flipt.io", Name: "dev"},
	})
	require.NoError(t, err)

	return repo
}
