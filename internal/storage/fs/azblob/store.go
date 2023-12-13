package azblob

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob"
	"go.flipt.io/flipt/internal/containers"
	"go.flipt.io/flipt/internal/storage"
	storagefs "go.flipt.io/flipt/internal/storage/fs"
	"go.uber.org/zap"
)

var _ storagefs.SnapshotStore = (*SnapshotStore)(nil)

// SnapshotStore represents an implementation of storage.SnapshotStore
// This implementation is backed by an S3 bucket
type SnapshotStore struct {
	logger *zap.Logger
	client *azblob.Client

	mu   sync.RWMutex
	snap storage.ReadOnlyStore

	endpoint  string
	container string
	account   string
	sharedKey string

	pollOpts []containers.Option[storagefs.Poller]
}

// View accepts a function which takes a *StoreSnapshot.
// The SnapshotStore will supply a snapshot which is valid
// for the lifetime of the provided function call.
func (s *SnapshotStore) View(fn func(storage.ReadOnlyStore) error) error {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return fn(s.snap)
}

// newClient provides the client based on the provided option
// There are 2 options supported for Azure Blob Storage authentication:
//   - Any system using environment variables supported by the Azure SDK for Go.
//   - Secret Access Key, using the account_shared_key attribute to specify
//     the primary or secondary account key for your Azure Blob Storage account
func newClient(account, sharedKey, endpoint string) (*azblob.Client, error) {
	if endpoint == "" {
		endpoint = fmt.Sprintf("https://%s.blob.core.windows.net/", account)
	}

	if sharedKey != "" {
		credentials, err := azblob.NewSharedKeyCredential(account, sharedKey)
		if err != nil {
			return nil, err
		}
		client, err := azblob.NewClientWithSharedKeyCredential(endpoint, credentials, nil)
		if err != nil {
			return nil, err
		}
		return client, nil
	}
	credentials, err := azidentity.NewDefaultAzureCredential(nil)
	if err != nil {
		return nil, err
	}
	client, err := azblob.NewClient(endpoint, credentials, nil)
	if err != nil {
		return nil, err
	}
	return client, nil
}

// NewSnapshotStore constructs a Store
func NewSnapshotStore(ctx context.Context, logger *zap.Logger, container string, opts ...containers.Option[SnapshotStore]) (*SnapshotStore, error) {
	s := &SnapshotStore{
		logger:    logger,
		container: container,
		pollOpts: []containers.Option[storagefs.Poller]{
			storagefs.WithInterval(60 * time.Second),
		},
	}

	containers.ApplyAll(s, opts...)

	client, err := newClient(s.account, s.sharedKey, s.endpoint)
	if err != nil {
		return nil, err
	}
	s.client = client
	// fetch snapshot at-least once before returning store
	// to ensure we have some state to serve
	if _, err := s.update(ctx); err != nil {
		return nil, err
	}

	go storagefs.NewPoller(s.logger, s.pollOpts...).Poll(ctx, s.update)

	return s, nil
}

// WithEndpoint configures the endpoint
func WithEndpoint(endpoint string) containers.Option[SnapshotStore] {
	return func(s *SnapshotStore) {
		s.endpoint = endpoint
	}
}

// WithAccount configures the account
func WithAccount(account string) containers.Option[SnapshotStore] {
	return func(s *SnapshotStore) {
		s.account = account
	}
}

// WithAccount configures the account
func NewWithSharedKey(sharedKey string) containers.Option[SnapshotStore] {
	return func(s *SnapshotStore) {
		s.sharedKey = sharedKey
	}
}

// WithPollOptions configures the poller options used when periodically updating snapshot state
func WithPollOptions(opts ...containers.Option[storagefs.Poller]) containers.Option[SnapshotStore] {
	return func(s *SnapshotStore) {
		s.pollOpts = append(s.pollOpts, opts...)
	}
}

// Update fetches a new snapshot and swaps it out for the current one.
func (s *SnapshotStore) update(context.Context) (bool, error) {
	fs, err := NewFS(s.logger, s.client, s.container)
	if err != nil {
		return false, err
	}

	snap, err := storagefs.SnapshotFromFS(s.logger, fs)
	if err != nil {
		return false, err
	}

	s.mu.Lock()
	s.snap = snap
	s.mu.Unlock()

	return true, nil
}

// String returns an identifier string for the store type.
func (s *SnapshotStore) String() string {
	return "azblob"
}
