package gcs

import (
	"context"
	"os"
	"testing"
	"time"

	gstorage "cloud.google.com/go/storage"
	"github.com/stretchr/testify/require"
	"go.flipt.io/flipt/internal/containers"
	"go.flipt.io/flipt/internal/storage"
	"go.flipt.io/flipt/internal/storage/fs"
	"go.uber.org/zap/zaptest"
)

const testBucket = "testdata"

var emulatorURL = os.Getenv("STORAGE_EMULATOR_HOST")

func Test_Store_String(t *testing.T) {
	require.Equal(t, "gcs", (&SnapshotStore{}).String())
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
	gcsClient := testClient(t)
	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)
	// update features.yml
	path := "features.yml"
	w := gcsClient.Bucket(store.bucket).Object(path).NewWriter(ctx)
	_, err := w.Write(updated)
	require.NoError(t, err)
	err = w.Close()
	require.NoError(t, err)

	// assert matching state
	select {
	case <-ch:
	case <-time.After(time.Minute):
		t.Fatal("timed out waiting for update")
	}

	t.Log("received new snapshot")

	require.NoError(t, store.View(func(s storage.ReadOnlyStore) error {
		_, err := s.GetNamespace(ctx, "production")
		if err != nil {
			return err
		}

		_, err = s.GetFlag(ctx, "production", "foo")
		if err != nil {
			return err
		}

		return err
	}))

}

func testStore(t *testing.T, opts ...containers.Option[SnapshotStore]) (*SnapshotStore, bool) {
	t.Helper()

	if emulatorURL == "" {
		t.Skip("Set non-empty STORAGE_EMULATOR_HOST env var to run this test.")
		return nil, true
	}

	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	gcsClient := testClient(t)
	bucket := gcsClient.Bucket(testBucket)
	err := bucket.Create(ctx, "", nil)
	require.NoError(t, err)

	t.Cleanup(func() {
		objects := bucket.Objects(ctx, nil)
		for {
			o, err := objects.Next()
			if err != nil {
				break
			}
			err = bucket.Object(o.Name).Delete(ctx)
			require.NoError(t, err)
		}
		err := bucket.Delete(ctx)
		require.NoError(t, err)
		err = gcsClient.Close()
		require.NoError(t, err)
	})
	w := bucket.Object(".flipt.yml").NewWriter(ctx)
	_, err = w.Write([]byte(`namespace: production`))
	require.NoError(t, err)
	err = w.Close()
	require.NoError(t, err)
	source, err := NewSnapshotStore(ctx, zaptest.NewLogger(t), testBucket, opts...)
	require.NoError(t, err)
	return source, false
}

func testClient(t *testing.T) *gstorage.Client {
	t.Helper()
	client, err := gstorage.NewClient(context.Background())
	require.NoError(t, err)
	return client
}
