package object

import (
	"context"
	"encoding/hex"
	"errors"
	"io"
	"io/fs"
	"strings"
	"sync"
	"time"

	"go.flipt.io/flipt/internal/containers"
	"go.flipt.io/flipt/internal/storage"
	storagefs "go.flipt.io/flipt/internal/storage/fs"
	"go.uber.org/zap"
	gcblob "gocloud.dev/blob"
	"gocloud.dev/gcerrors"
)

var _ storagefs.SnapshotStore = (*SnapshotStore)(nil)

type SnapshotStore struct {
	*storagefs.Poller

	logger   *zap.Logger
	scheme   string
	bucket   *gcblob.Bucket
	prefix   string
	pollOpts []containers.Option[storagefs.Poller]

	mu   sync.RWMutex
	snap storage.ReadOnlyStore
}

func NewSnapshotStore(ctx context.Context, logger *zap.Logger, scheme string, bucket *gcblob.Bucket, opts ...containers.Option[SnapshotStore]) (*SnapshotStore, error) {
	s := &SnapshotStore{
		logger: logger,
		scheme: scheme,
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

// WithPrefix configures the prefix for object store
func WithPrefix(prefix string) containers.Option[SnapshotStore] {
	return func(s *SnapshotStore) {
		s.prefix = prefix
	}
}

// WithPollOptions configures the poller options used when periodically updating snapshot state
func WithPollOptions(opts ...containers.Option[storagefs.Poller]) containers.Option[SnapshotStore] {
	return func(s *SnapshotStore) {
		s.pollOpts = append(s.pollOpts, opts...)
	}
}

// View accepts a function which takes a *StoreSnapshot.
// The SnapshotStore will supply a snapshot which is valid
// for the lifetime of the provided function call.
func (s *SnapshotStore) View(_ context.Context, fn func(storage.ReadOnlyStore) error) error {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return fn(s.snap)
}

func (s *SnapshotStore) String() string {
	return s.scheme
}

// Update fetches a new snapshot and swaps it out for the current one.
func (s *SnapshotStore) update(ctx context.Context) (bool, error) {
	snap, err := s.build(ctx)
	if err != nil {
		return false, err
	}

	s.mu.Lock()
	s.snap = snap
	s.mu.Unlock()

	return true, nil
}

func (s *SnapshotStore) build(ctx context.Context) (*storagefs.Snapshot, error) {
	idx, err := s.getIndex(ctx)
	if err != nil {
		return nil, err
	}

	iterator := s.bucket.List(&gcblob.ListOptions{
		Prefix: s.prefix,
	})

	var files []fs.File
	for {
		item, err := iterator.Next(ctx)
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return nil, err
		}

		key := strings.TrimPrefix(item.Key, s.prefix)
		if !idx.Match(key) {
			continue
		}

		rd, err := s.bucket.NewReader(ctx, s.prefix+key, &gcblob.ReaderOptions{})
		if err != nil {
			return nil, err
		}

		files = append(files, NewFile(
			key,
			item.Size,
			rd,
			item.ModTime,
			hex.EncodeToString(item.MD5),
		))
	}

	return storagefs.SnapshotFromFiles(s.logger, files)
}

func (s *SnapshotStore) getIndex(ctx context.Context) (*storagefs.FliptIndex, error) {
	rd, err := s.bucket.NewReader(ctx, s.prefix+storagefs.IndexFileName, &gcblob.ReaderOptions{})

	if err != nil {
		if gcerrors.Code(err) != gcerrors.NotFound {
			return nil, err
		}
		s.logger.Debug("using default index",
			zap.String("file", storagefs.IndexFileName),
			zap.Error(fs.ErrNotExist))
		return storagefs.DefaultFliptIndex()
	}

	defer rd.Close()
	idx, err := storagefs.ParseFliptIndex(rd)
	if err != nil {
		return nil, err
	}

	return idx, nil

}
