package s3

import (
	"context"
	"fmt"
	"net/url"
	"sync"
	"time"

	s3v2 "github.com/aws/aws-sdk-go-v2/service/s3"
	"go.flipt.io/flipt/internal/containers"
	"go.flipt.io/flipt/internal/storage"
	storagefs "go.flipt.io/flipt/internal/storage/fs"
	"go.flipt.io/flipt/internal/storage/fs/object/blob"
	"go.uber.org/zap"
	gcaws "gocloud.dev/aws"
	gcblob "gocloud.dev/blob"
	"gocloud.dev/blob/s3blob"
)

const (
	Schema = "s3i"
)

func init() {
	gcblob.DefaultURLMux().RegisterBucket(Schema, new(urlSessionOpener))
}

type urlSessionOpener struct{}

func (o *urlSessionOpener) OpenBucketURL(ctx context.Context, u *url.URL) (*gcblob.Bucket, error) {
	cfg, err := gcaws.V2ConfigFromURLParams(ctx, u.Query())
	if err != nil {
		return nil, fmt.Errorf("open bucket %v: %w", u, err)
	}
	clientV2 := s3v2.NewFromConfig(cfg, func(o *s3v2.Options) {
		o.UsePathStyle = true
	})
	return s3blob.OpenBucketV2(ctx, clientV2, u.Host, &s3blob.Options{})
}

var _ storagefs.SnapshotStore = (*SnapshotStore)(nil)

// SnapshotStore represents an implementation of storage.SnapshotStore
// This implementation is backed by an S3 bucket
type SnapshotStore struct {
	*storagefs.Poller

	logger *zap.Logger

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
func (s *SnapshotStore) update(ctx context.Context) (bool, error) {
	q := url.Values{}
	q.Set("awssdk", "v2")
	if s.region != "" {
		q.Set("region", s.region)
	}
	if s.endpoint != "" {
		q.Set("endpoint", s.endpoint)
	}

	strurl := fmt.Sprintf("%s://%s?%s", Schema, s.bucket, q.Encode())
	fs, err := blob.NewFS(ctx, strurl, s.bucket, s.prefix)
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
