package s3

import (
	"bytes"
	"context"
	"os"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/stretchr/testify/require"
	"go.flipt.io/flipt/internal/containers"
	storagefs "go.flipt.io/flipt/internal/storage/fs"
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

	snap, err := source.Get()
	require.NoError(t, err)

	_, err = snap.GetNamespace(context.TODO(), "production")
	require.NoError(t, err)

	_, err = snap.GetNamespace(context.TODO(), "prefix")
	require.NoError(t, err)
}

func Test_SourceGetPrefix(t *testing.T) {
	source, skip := testSource(t, WithPrefix("prefix/"))
	if skip {
		return
	}

	snap, err := source.Get()
	require.NoError(t, err)

	_, err = snap.GetNamespace(context.TODO(), "production")
	require.Error(t, err, "production namespace should have been skipped")

	_, err = snap.GetNamespace(context.TODO(), "prefix")
	require.NoError(t, err, "prefix namespace should be present in snapshot")
}

func Test_SourceSubscribe(t *testing.T) {
	source, skip := testSource(t)
	if skip {
		return
	}

	snap, err := source.Get()
	require.NoError(t, err)

	_, err = snap.GetNamespace(context.TODO(), "production")
	require.NoError(t, err)

	ctx, cancel := context.WithCancel(context.Background())

	// start subscription
	ch := make(chan *storagefs.StoreSnapshot)
	go source.Subscribe(ctx, ch)

	updated := []byte(`namespace: production
flags:
    - key: foo`)

	buf := bytes.NewReader(updated)

	s3Client := source.s3
	// update features.yml
	path := "features.yml"
	_, err = s3Client.PutObject(ctx, &s3.PutObjectInput{
		Bucket: &source.bucket,
		Key:    &path,
		Body:   buf,
	})
	require.NoError(t, err)

	// assert matching state
	snap = <-ch

	t.Log("received new snapshot")

	_, err = snap.GetFlag(context.TODO(), "production", "foo")
	require.NoError(t, err)

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
