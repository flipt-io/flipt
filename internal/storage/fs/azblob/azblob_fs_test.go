package azblob

import (
	"context"
	"io"
	"io/fs"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"
	"gocloud.dev/blob"
	_ "gocloud.dev/blob/fileblob"
	_ "gocloud.dev/blob/memblob"
)

func Test_FS(t *testing.T) {
	containerName := "test-container"
	logger := zaptest.NewLogger(t)
	// run with no prefix, returning all files
	t.Run("Ensure invalid and non existent paths produce an error", func(t *testing.T) {
		azfs, err := NewFS(logger, "mem", containerName)
		require.NoError(t, err)

		_, err = azfs.Open("..")
		require.Equal(t, &fs.PathError{
			Op:   "Open",
			Path: "..",
			Err:  fs.ErrInvalid,
		}, err)

		_, err = azfs.Open("zero.txt")
		require.Equal(t, &fs.PathError{
			Op:   "open",
			Path: "zero.txt",
			Err:  fs.ErrNotExist,
		}, err)
	})

	t.Run("Ensure files exist with expected contents", func(t *testing.T) {
		// setup the mock
		ctx, cancel := context.WithCancel(context.Background())
		t.Cleanup(cancel)

		dir, err := os.MkdirTemp("/tmp", "flipt-io-test*")
		require.NoError(t, err)
		t.Cleanup(func() { os.RemoveAll(dir) })

		bucket, err := blob.OpenBucket(ctx, "file://"+dir)
		require.NoError(t, err)

		objectChunks := []string{"one", "two"}
		for _, ob := range objectChunks {
			err = bucket.WriteAll(ctx, ob, []byte(ob+"data"), nil)
			require.NoError(t, err)
		}
		require.NoError(t, bucket.Close())

		azfs, err := NewFS(logger, "file", dir)
		require.NoError(t, err)

		// running test
		seen := map[string]string{}
		err = fs.WalkDir(azfs, ".", func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}

			fi, err := azfs.Open(path)
			require.NoError(t, err)

			defer fi.Close()

			contents, err := io.ReadAll(fi)
			require.NoError(t, err)

			seen[path] = string(contents)

			return nil
		})
		require.NoError(t, err)

		expected := map[string]string{
			".":   "",
			"one": "onedata",
			"two": "twodata",
		}
		require.Equal(t, expected, seen)
	})
}
