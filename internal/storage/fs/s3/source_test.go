package s3

import (
	"bytes"
	"context"
	"io"
	"io/fs"
	"os"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.flipt.io/flipt/internal/containers"
	"go.uber.org/zap/zaptest"
)

const testBucket = "testdata"

var minioURL = os.Getenv("TEST_S3_ENDPOINT")

func Test_SourceString(t *testing.T) {
	require.Equal(t, "s3", (&Source{}).String())
}

func Test_SourceGet(t *testing.T) {
	source, skip := testSource(t)
	if skip {
		return
	}

	s3fs, err := source.Get()
	require.NoError(t, err)

	fi, err := s3fs.Open("features.yml")
	require.NoError(t, err)

	data, err := io.ReadAll(fi)
	require.NoError(t, err)
	require.NoError(t, fi.Close())

	assert.Equal(t, []byte("namespace: production\n"), data)

	fi, err = s3fs.Open("prefix/prefix_features.yml")
	require.NoError(t, err)

	data, err = io.ReadAll(fi)
	require.NoError(t, err)
	require.NoError(t, fi.Close())

	assert.Equal(t, []byte("namespace: prefix\n"), data)
}

func Test_SourceGetPrefix(t *testing.T) {
	source, skip := testSource(t, WithPrefix("prefix/"))
	if skip {
		return
	}

	s3fs, err := source.Get()
	require.NoError(t, err)

	_, err = s3fs.Open("features.yml")
	require.Error(t, err)

	// Open without a prefix path prepends the prefix
	fi, err := s3fs.Open("prefix_features.yml")
	require.NoError(t, err)
	require.NoError(t, fi.Close())

	fi, err = s3fs.Open("prefix/prefix_features.yml")
	require.NoError(t, err)

	data, err := io.ReadAll(fi)
	require.NoError(t, err)
	require.NoError(t, fi.Close())

	assert.Equal(t, []byte("namespace: prefix\n"), data)
}

func Test_SourceSubscribe(t *testing.T) {
	source, skip := testSource(t)
	if skip {
		return
	}
	sourceFs, err := source.Get()
	require.NoError(t, err)

	filename := "features.yml"
	_, err = sourceFs.Open(filename)
	require.NoError(t, err)

	ctx, cancel := context.WithCancel(context.Background())

	// start subscription
	ch := make(chan fs.FS)
	go source.Subscribe(ctx, ch)

	updated := []byte(`namespace: production
flags:
    - key: foo`)

	buf := bytes.NewReader(updated)

	s3Client := source.s3
	// update features.yml
	_, err = s3Client.PutObject(ctx, &s3.PutObjectInput{
		Bucket: &source.bucket,
		Key:    &filename,
		Body:   buf,
	})
	require.NoError(t, err)

	// assert matching state
	fs := <-ch

	t.Log("received new FS")

	found, err := fs.Open("features.yml")
	require.NoError(t, err)

	data, err := io.ReadAll(found)
	require.NoError(t, err)

	assert.Equal(t, string(updated), string(data))

	cancel()
	_, open := <-ch
	require.False(t, open, "expected channel to be closed after cancel")
}

func testSource(t *testing.T, opts ...containers.Option[Source]) (*Source, bool) {
	t.Helper()

	if minioURL == "" {
		t.Skip("Set non-empty TEST_S3_ENDPOINT env var to run this test.")
		return nil, true
	}

	source, err := NewSource(zaptest.NewLogger(t), testBucket,
		append([]containers.Option[Source]{
			WithEndpoint(minioURL),
			WithPollInterval(5 * time.Second),
		},
			opts...)...,
	)
	require.NoError(t, err)

	return source, false
}
