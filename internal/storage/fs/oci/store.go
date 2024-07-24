package oci

import (
	"context"
	"sync"

	"github.com/opencontainers/go-digest"
	"go.flipt.io/flipt/internal/containers"
	"go.flipt.io/flipt/internal/oci"
	"go.flipt.io/flipt/internal/storage"
	storagefs "go.flipt.io/flipt/internal/storage/fs"
	"go.uber.org/zap"
)

var _ storagefs.SnapshotStore = (*SnapshotStore)(nil)

// SnapshotStore is an implementation storage.SnapshotStore backed by OCI repositories.
// It fetches instances of OCI manifests and uses them to build snapshots from their contents.
type SnapshotStore struct {
	*storagefs.Poller

	logger *zap.Logger

	store *oci.Store
	ref   oci.Reference

	mu         sync.RWMutex
	snap       storage.ReadOnlyStore
	lastDigest digest.Digest

	pollOpts []containers.Option[storagefs.Poller]
}

// View accepts a function which takes a *StoreSnapshot.
// The SnapshotStore will supply a snapshot which is valid
// for the lifetime of the provided function call.
func (s *SnapshotStore) View(_ context.Context, fn func(storage.ReadOnlyStore) error) error {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return fn(s.snap)
}

// NewSnapshotStore constructs and configures a Store.
// The store uses the connection and credential details provided to build
// *storagefs.StoreSnapshot implementations around a target OCI repository.
func NewSnapshotStore(ctx context.Context, logger *zap.Logger, store *oci.Store, ref oci.Reference, opts ...containers.Option[SnapshotStore]) (_ *SnapshotStore, err error) {
	s := &SnapshotStore{
		logger: logger,
		store:  store,
		ref:    ref,
	}

	containers.ApplyAll(s, opts...)

	if _, err := s.update(ctx); err != nil {
		return nil, err
	}

	s.Poller = storagefs.NewPoller(logger, ctx, s.update, s.pollOpts...)

	go s.Poller.Poll()

	return s, nil
}

// WithPollOptions configures the options used periodically invoke the update procedure
func WithPollOptions(opts ...containers.Option[storagefs.Poller]) containers.Option[SnapshotStore] {
	return func(s *SnapshotStore) {
		s.pollOpts = append(s.pollOpts, opts...)
	}
}

func (s *SnapshotStore) String() string {
	return "oci"
}

// update attempts to fetch the latest state for the target OCi repository and tag.
// If the state has not change sinced the last observed image digest it skips
// updating the snapshot and returns false (not modified).
func (s *SnapshotStore) update(ctx context.Context) (bool, error) {
	resp, err := s.store.Fetch(ctx, s.ref, oci.IfNoMatch(s.lastDigest))
	if err != nil {
		return false, err
	}

	// return not modified as the last observed digest matched
	// the remote digest
	if resp.Matched {
		return false, nil
	}

	snap, err := storagefs.SnapshotFromFiles(s.logger, resp.Files, storagefs.WithEtag(resp.Digest.Hex()))
	if err != nil {
		return false, err
	}

	s.mu.Lock()
	s.lastDigest = resp.Digest
	s.snap = snap
	s.mu.Unlock()

	return true, nil
}
