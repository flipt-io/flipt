package s3fs

import (
	"context"
	"errors"
	"io"
	"io/fs"
	"strconv"
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
	// objectChunks simulates truncated responses from S3
	objectChunks [][]string
}

func (fs3 *fakeS3Client) ListObjectsV2(_ context.Context, input *s3.ListObjectsV2Input, _ ...func(*s3.Options)) (*s3.ListObjectsV2Output, error) {
	chunkIndex := 0
	var err error
	// Fake a continuation token by converting the index of the
	// objectChunks to a string
	if input != nil && input.ContinuationToken != nil {
		chunkIndex, err = strconv.Atoi(*input.ContinuationToken)
		if err != nil {
			return nil, err
		}
	}
	if chunkIndex > len(fs3.objectChunks) {
		return nil, errors.New("Invalid ContinuationToken")
	}

	// If there are more chunks after this one, we need to provide
	// a continuation token
	var continuationToken *string
	if chunkIndex+1 < len(fs3.objectChunks) {
		tmp := strconv.Itoa(chunkIndex + 1)
		continuationToken = &tmp
	}

	// create the chunked output
	chunk := fs3.objectChunks[chunkIndex]
	objects := []types.Object{}
	for _, k := range chunk {
		k := k
		v, ok := fs3.objects[k]
		if !ok {
			// this would mean an error when setting up
			// the fake s3 client
			return nil, errors.New("Cannot find object for chunk")
		}
		if input.Prefix == nil || strings.HasPrefix(k, *input.Prefix) {
			objects = append(objects, types.Object{
				Key:          &k,
				Size:         int64(len(v.content)),
				LastModified: &v.lastModified,
			})
		}
	}
	return &s3.ListObjectsV2Output{
		Contents:              objects,
		IsTruncated:           continuationToken != nil,
		ContinuationToken:     input.ContinuationToken,
		NextContinuationToken: continuationToken,
	}, nil
}

func (fs3 *fakeS3Client) GetObject(ctx context.Context, input *s3.GetObjectInput, opts ...func(*s3.Options)) (*s3.GetObjectOutput, error) {
	keyNotFound := "key not found"
	errNotFound := &types.NotFound{
		Message: &keyNotFound,
	}

	if input.Key == nil {
		return nil, errNotFound
	}

	obj, ok := fs3.objects[*input.Key]
	if ok {
		return &s3.GetObjectOutput{
			ContentLength: int64(len(obj.content)),
			Body:          io.NopCloser(strings.NewReader(obj.content)),
			LastModified:  &obj.lastModified,
		}, nil
	}

	return nil, errNotFound
}

func newFakeS3Client() *fakeS3Client {
	t := time.Date(2020, 1, 2, 3, 4, 5, 6, time.UTC)
	f := &fakeS3Client{
		bucket: "mybucket",
		objects: map[string]fakeS3Object{
			"one": {
				content:      "onedata",
				lastModified: t,
			},
			"two": {
				content:      "twodata",
				lastModified: t.Add(time.Hour * 20),
			},
			"prefix/three": {
				content:      "threedata",
				lastModified: t.Add(time.Hour * 30),
			},
			"prefix/four/five": {
				content:      "fivedata",
				lastModified: t.Add(time.Hour * 50),
			},
			"anotherprefix/six": {
				content:      "sixdata",
				lastModified: t.Add(time.Hour * 60),
			},
		},
	}
	// create chunks to test out {non-}truncated responses
	f.objectChunks = [][]string{
		{
			"one", "two",
		},
		{
			"prefix/three", "anotherprefix/six",
		},
		{
			"prefix/four/five",
		},
	}
	return f
}

func Test_FS(t *testing.T) {
	fakeS3Client := newFakeS3Client()
	logger := zaptest.NewLogger(t)
	// run with no prefix, returning all files
	s3fs, err := New(logger, fakeS3Client, fakeS3Client.bucket, "")
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
			"one":               "onedata",
			"two":               "twodata",
			"prefix/three":      "threedata",
			"prefix/four/five":  "fivedata",
			"anotherprefix/six": "sixdata",
		}

		assert.Equal(t, expected, seen)

		assert.Equal(t, map[string]int{
			".": len(expected),
		}, dirs)
	})

}

func Test_FS_Prefix(t *testing.T) {
	fakeS3Client := newFakeS3Client()
	logger := zaptest.NewLogger(t)
	// run with prefix, returning only prefixed files
	s3fs, err := New(logger, fakeS3Client, fakeS3Client.bucket, "prefix/")
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

		_, err = s3fs.Open("one")
		require.Equal(t, &fs.PathError{
			Op:   "Open",
			Path: "one",
			Err:  fs.ErrNotExist,
		}, err)
	})

	t.Run("Ensure Open prepends prefix if not provided", func(t *testing.T) {
		fi, err := s3fs.Open("three")
		require.NoError(t, err)

		contents, err := io.ReadAll(fi)
		require.NoError(t, err)

		require.Equal(t, "threedata", string(contents))
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
			"prefix/three":     "threedata",
			"prefix/four/five": "fivedata",
		}

		assert.Equal(t, expected, seen)

		assert.Equal(t, map[string]int{
			".": len(expected),
		}, dirs)
	})
}
