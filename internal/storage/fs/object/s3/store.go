package s3

import (
	"context"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"

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
	s3     *s3.Client

	mu   sync.RWMutex
	snap storage.ReadOnlyStore

	endpoint string
	region   string
	bucket   string
	prefix   string

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

// NewSnapshotStore constructs a Store
func NewSnapshotStore(ctx context.Context, logger *zap.Logger, bucket string, opts ...containers.Option[SnapshotStore]) (*SnapshotStore, error) {
	s := &SnapshotStore{
		logger: logger,
		bucket: bucket,
		pollOpts: []containers.Option[storagefs.Poller]{
			storagefs.WithInterval(60 * time.Second),
		},
	}

	containers.ApplyAll(s, opts...)

	cfg, err := config.LoadDefaultConfig(context.Background(),
		config.WithRegion(s.region))
	if err != nil {
		return nil, err
	}

	var s3Opts []func(*s3.Options)
	if s.endpoint != "" {
		s3Opts = append(s3Opts, func(o *s3.Options) {
			o.BaseEndpoint = &s.endpoint
			o.UsePathStyle = true
			o.Region = s.region
		})
	}
	s.s3 = s3.NewFromConfig(cfg, s3Opts...)

	// fetch snapshot at-least once before returning store
	// to ensure we have some state to serve
	if _, err := s.update(ctx); err != nil {
		return nil, err
	}

	s.Poller = storagefs.NewPoller(s.logger, ctx, s.update, s.pollOpts...)

	go s.Poll()

	return s, nil
}

// WithPrefix configures the prefix for s3
func WithPrefix(prefix string) containers.Option[SnapshotStore] {
	return func(s *SnapshotStore) {
		s.prefix = prefix
	}
}

// WithRegion configures the region for s3
func WithRegion(region string) containers.Option[SnapshotStore] {
	return func(s *SnapshotStore) {
		s.region = region
	}
}

// WithEndpoint configures the region for s3
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
	fs, err := NewFS(s.logger, s.s3, s.bucket, s.prefix)
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
	return "s3"
}
