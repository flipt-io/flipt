package azblob

import (
	"context"
	"net/url"
	"os"
	"strconv"
	"sync"
	"time"

	"go.flipt.io/flipt/internal/containers"
	"go.flipt.io/flipt/internal/storage"
	storagefs "go.flipt.io/flipt/internal/storage/fs"
	"go.uber.org/zap"
)

var _ storagefs.SnapshotStore = (*SnapshotStore)(nil)

// SnapshotStore represents an implementation of storage.SnapshotStore
// This implementation is backed by an S3 bucket
type SnapshotStore struct {
	*storagefs.Poller

	logger *zap.Logger

	mu   sync.RWMutex
	snap storage.ReadOnlyStore

	endpoint  string
	container string
	pollOpts  []containers.Option[storagefs.Poller]
}

// View accepts a function which takes a *StoreSnapshot.
// The SnapshotStore will supply a snapshot which is valid
// for the lifetime of the provided function call.
func (s *SnapshotStore) View(fn func(storage.ReadOnlyStore) error) error {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return fn(s.snap)
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

	if s.endpoint != "" {
		url, err := url.Parse(s.endpoint)
		if err != nil {
			return nil, err
		}
		os.Setenv("AZURE_STORAGE_PROTOCOL", url.Scheme)
		os.Setenv("AZURE_STORAGE_IS_LOCAL_EMULATOR", strconv.FormatBool(url.Scheme == "http"))
		os.Setenv("AZURE_STORAGE_DOMAIN", url.Host)
	}

	// fetch snapshot at-least once before returning store
	// to ensure we have some state to serve
	if _, err := s.update(ctx); err != nil {
		return nil, err
	}

	s.Poller = storagefs.NewPoller(s.logger, ctx, s.update, s.pollOpts...)

	go s.Poll()

	return s, nil
}

// WithEndpoint configures the endpoint
func WithEndpoint(endpoint string) containers.Option[SnapshotStore] {
	return func(s *SnapshotStore) {
		s.endpoint = endpoint
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
	fs, err := NewFS(s.logger, "azblob", s.container)
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
