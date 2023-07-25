package s3fs

import (
	"context"
	"io"
	"io/fs"
	"strings"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"
)

type fakeS3Object struct {
	lastModified time.Time
	content      string
}

type fakeS3Client struct {
	bucket  string
	objects map[string]fakeS3Object
}

func (fs3 *fakeS3Client) ListObjectsV2(context.Context, *s3.ListObjectsV2Input, ...func(*s3.Options)) (*s3.ListObjectsV2Output, error) {
	objects := make([]types.Object, len(fs3.objects))
	i := 0
	for k, v := range fs3.objects {
		k := k
		v := v
		objects[i] = types.Object{
			Key:          &k,
			Size:         int64(len(v.content)),
			LastModified: &v.lastModified,
		}
		i++
	}
	return &s3.ListObjectsV2Output{
		Contents: objects,
	}, nil
}

func (fs3 *fakeS3Client) GetObject(ctx context.Context, input *s3.GetObjectInput, opts ...func(*s3.Options)) (*s3.GetObjectOutput, error) {
	obj, ok := fs3.objects[*input.Key]
	if ok {
		return &s3.GetObjectOutput{
			ContentLength: int64(len(obj.content)),
			Body:          io.NopCloser(strings.NewReader(obj.content)),
			LastModified:  &obj.lastModified,
		}, nil
	}

	keyNotFound := "key not found"
	return nil, &types.NotFound{
		Message: &keyNotFound,
	}
}

func newFakeS3Client() *fakeS3Client {
	t := time.Date(2020, 1, 2, 3, 4, 5, 6, time.UTC)
	return &fakeS3Client{
		bucket: "mybucket",
		objects: map[string]fakeS3Object{
			"one": {
				content:      "onedata",
				lastModified: t,
			},
			"two": {
				content:      "twodata",
				lastModified: t.Add(time.Hour * 24),
			},
		},
	}
}

func Test_FS(t *testing.T) {
	fakeS3Client := newFakeS3Client()
	logger := zaptest.NewLogger(t)
	s3fs, err := New(logger, fakeS3Client, fakeS3Client.bucket)
	require.NoError(t, err)

	t.Run("Ensure invalid and non existent paths produce an error", func(t *testing.T) {
		_, err := s3fs.Open("..")
		require.Equal(t, &fs.PathError{
			Op:   "Open",
			Path: "..",
			Err:  fs.ErrInvalid,
		}, err)

		_, err = s3fs.Open("zero.txt")
		require.Equal(t, &fs.PathError{
			Op:   "Open",
			Path: "zero.txt",
			Err:  fs.ErrNotExist,
		}, err)
	})

	t.Run("Ensure files exist with expected contents", func(t *testing.T) {
		seen := map[string]string{}
		dirs := map[string]int{}

		err := fs.WalkDir(s3fs, ".", func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}

			fi, err := s3fs.Open(path)
			require.NoError(t, err)

			defer fi.Close()

			if d.IsDir() {
				entries, err := s3fs.ReadDir(d.Name())
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

		expected := map[string]string{
			"one": "onedata",
			"two": "twodata",
		}

		assert.Equal(t, expected, seen)

		assert.Equal(t, map[string]int{
			".": len(expected),
		}, dirs)
	})

}
