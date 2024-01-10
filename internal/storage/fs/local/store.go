package local

import (
	"context"
	"os"
	"sync"

	"go.flipt.io/flipt/internal/containers"
	"go.flipt.io/flipt/internal/storage"
	storagefs "go.flipt.io/flipt/internal/storage/fs"
	"go.uber.org/zap"
)

var _ storagefs.SnapshotStore = (*SnapshotStore)(nil)

// SnapshotStore implements storagefs.SnapshotStore which
// is backed by the local filesystem through os.DirFS
type SnapshotStore struct {
	*storagefs.Poller

	logger *zap.Logger
	dir    string

	mu   sync.RWMutex
	snap storage.ReadOnlyStore

	pollOpts []containers.Option[storagefs.Poller]
}

// NewSnapshotStore constructs a new SnapshotStore
func NewSnapshotStore(ctx context.Context, logger *zap.Logger, dir string, opts ...containers.Option[SnapshotStore]) (*SnapshotStore, error) {
	s := &SnapshotStore{
		logger: logger,
		dir:    dir,
	}

	containers.ApplyAll(s, opts...)

	// seed initial state an ensure we have state
	// before returning
	if _, err := s.update(ctx); err != nil {
		return nil, err
	}

	s.Poller = storagefs.NewPoller(logger, ctx, s.update, s.pollOpts...)

	go s.Poll()

	return s, nil
}

// WithPollOptions configures poller options on the store.
func WithPollOptions(opts ...containers.Option[storagefs.Poller]) containers.Option[SnapshotStore] {
	return func(s *SnapshotStore) {
		s.pollOpts = append(s.pollOpts, opts...)
	}
}

// View passes the current snapshot to the provided function
// while holding a read lock.
func (s *SnapshotStore) View(_ context.Context, fn func(storage.ReadOnlyStore) error) error {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return fn(s.snap)
}

// update fetches a new snapshot from the local filesystem
// and updates the current served reference via a write lock
func (s *SnapshotStore) update(context.Context) (bool, error) {
	snap, err := storagefs.SnapshotFromFS(s.logger, os.DirFS(s.dir))
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
	return "local"
}
