package azblob

import (
	"context"
	"encoding/base64"
	"os"
	"testing"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob/bloberror"
	"github.com/stretchr/testify/require"
	"go.flipt.io/flipt/internal/containers"
	"go.flipt.io/flipt/internal/storage"
	"go.flipt.io/flipt/internal/storage/fs"
	"go.uber.org/zap/zaptest"
)

const testContainer = "testdata"

var aruziteURL = os.Getenv("TEST_AZURE_ENDPOINT")

func Test_Store_String(t *testing.T) {
	require.Equal(t, "azblob", (&SnapshotStore{}).String())
}

func TestNewClient(t *testing.T) {
	testSharedKey := base64.StdEncoding.EncodeToString([]byte("sharedkey"))
	tests := []struct {
		account     string
		sharedkey   string
		endpoint    string
		expectedURL string
	}{
		{"devaccount1", testSharedKey, "", "https://devaccount1.blob.core.windows.net/"},
		{"devaccount1", testSharedKey, "http://localhost:10000/", "http://localhost:10000/"},
		{"devaccount1", "", "", "https://devaccount1.blob.core.windows.net/"},
		{"devaccount1", "", "http://localhost:10000/", "http://localhost:10000/"},
	}
	for _, tt := range tests {
		client, err := newClient(tt.account, tt.sharedkey, tt.endpoint)
		require.NoError(t, err)
		require.Equal(t, tt.expectedURL, client.URL())
	}
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

	azbClient := store.client

	// flag shouldn't be present until we update it
	require.Error(t, store.View(func(s storage.ReadOnlyStore) error {
		_, err := s.GetFlag(context.TODO(), "production", "foo")
		return err
	}), "flag should not be defined yet")

	updated := []byte(`namespace: production
flags:
    - key: foo
      name: Foo`)

	// update features.yml
	_, err := azbClient.UploadBuffer(context.TODO(), store.container, "features.yml", updated, nil)
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

		return err
	}))

}

func testStore(t *testing.T, opts ...containers.Option[SnapshotStore]) (*SnapshotStore, bool) {
	t.Helper()

	if aruziteURL == "" {
		t.Skip("Set non-empty TEST_AZURE_ENDPOINT env var to run this test.")
		return nil, true
	}

	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)
	account := os.Getenv("TEST_AZURE_ACCOUNT")
	sharedKey := os.Getenv("TEST_AZURE_SHARED_KEY")
	client, err := newClient(account, sharedKey, aruziteURL)
	require.NoError(t, err)
	// create container
	_, err = client.CreateContainer(ctx, testContainer, nil)
	if err != nil {
		if !bloberror.HasCode(err, bloberror.ContainerAlreadyExists) {
			require.NoError(t, err)
		}
	}
	_, err = client.UploadBuffer(ctx, testContainer, ".flipt.yml", []byte(`namespace: production`), nil)
	require.NoError(t, err)

	source, err := NewSnapshotStore(ctx, zaptest.NewLogger(t), testContainer,
		append([]containers.Option[SnapshotStore]{
			WithEndpoint(aruziteURL),
			WithAccount(account),
			NewWithSharedKey(sharedKey),
		},
			opts...)...,
	)
	require.NoError(t, err)
	return source, false
}
