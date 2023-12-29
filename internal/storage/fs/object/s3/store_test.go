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
	"go.flipt.io/flipt/internal/storage"
	"go.flipt.io/flipt/internal/storage/fs"
	"go.uber.org/zap/zaptest"
)

const testBucket = "testdata"

var minioURL = os.Getenv("TEST_S3_ENDPOINT")

func Test_Store_String(t *testing.T) {
	require.Equal(t, "s3", (&SnapshotStore{}).String())
}

func Test_Store(t *testing.T) {
	ch := make(chan struct{})
	store, skip := testStore(t, WithPollOptions(
		fs.WithInterval(time.Second),
		fs.WithNotify(t, func(modified bool) {
			if modified {
				close(ch)
			}
		}),
	))
	if skip {
		return
	}

	// flag shouldn't be present until we update it
	require.Error(t, store.View(func(s storage.ReadOnlyStore) error {
		_, err := s.GetFlag(context.TODO(), "production", "foo")
		return err
	}), "flag should not be defined yet")

	updated := []byte(`namespace: production
flags:
    - key: foo
      name: Foo`)

	buf := bytes.NewReader(updated)

	s3Client := store.s3
	// update features.yml
	path := "features.yml"
	_, err := s3Client.PutObject(context.TODO(), &s3.PutObjectInput{
		Bucket: &store.bucket,
		Key:    &path,
		Body:   buf,
	})
	require.NoError(t, err)

	// assert matching state
	select {
	case <-ch:
	case <-time.After(time.Minute):
		t.Fatal("timed out waiting for update")
	}

	t.Log("received new snapshot")

	require.NoError(t, store.View(func(s storage.ReadOnlyStore) error {
		_, err = s.GetNamespace(context.TODO(), "production")
		if err != nil {
			return err
		}

		_, err = s.GetFlag(context.TODO(), "production", "foo")
		if err != nil {
			return err
		}

		_, err = s.GetNamespace(context.TODO(), "prefix")
		return err
	}))

}

func Test_Store_WithPrefix(t *testing.T) {
	store, skip := testStore(t, WithPrefix("prefix"))
	if skip {
		return
	}

	// namespace shouldn't exist as it has been filtered out by the prefix
	require.Error(t, store.View(func(s storage.ReadOnlyStore) error {
		_, err := s.GetNamespace(context.TODO(), "production")
		return err
	}), "production namespace shouldn't be retrieavable")
}

func testStore(t *testing.T, opts ...containers.Option[SnapshotStore]) (*SnapshotStore, bool) {
	t.Helper()

	if minioURL == "" {
		t.Skip("Set non-empty TEST_S3_ENDPOINT env var to run this test.")
		return nil, true
	}

	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	source, err := NewSnapshotStore(ctx, zaptest.NewLogger(t), testBucket,
		append([]containers.Option[SnapshotStore]{
			WithEndpoint(minioURL),
		},
			opts...)...,
	)
	require.NoError(t, err)

	t.Cleanup(func() {
		_ = source.Close()
	})

	return source, false
}
