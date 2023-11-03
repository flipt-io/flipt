package oci

import (
	"context"
	"errors"
	"time"

	"github.com/opencontainers/go-digest"
	"go.flipt.io/flipt/internal/containers"
	"go.flipt.io/flipt/internal/oci"
	storagefs "go.flipt.io/flipt/internal/storage/fs"
	"go.uber.org/zap"
)

var ErrDigestSeen = errors.New("digest already seen")

type Source struct {
	logger   *zap.Logger
	interval time.Duration

	curSnap   *storagefs.StoreSnapshot
	curDigest digest.Digest
	store     *oci.Store
}

// NewSource constructs and configures a Source.
// The source uses the connection and credential details provided to build
// fs.FS implementations around a target git repository.
func NewSource(logger *zap.Logger, store *oci.Store, opts ...containers.Option[Source]) (_ *Source, err error) {
	src := &Source{
		logger:   logger,
		interval: 30 * time.Second,
		store:    store,
	}
	containers.ApplyAll(src, opts...)

	return src, nil
}

// WithPollInterval configures the interval in which origin is polled to
// discover any updates to the target reference.
func WithPollInterval(tick time.Duration) containers.Option[Source] {
	return func(s *Source) {
		s.interval = tick
	}
}

func (s *Source) String() string {
	return "oci"
}

// Get builds a single instance of an *storagefs.StoreSnapshot
func (s *Source) Get(context.Context) (*storagefs.StoreSnapshot, error) {
	resp, err := s.store.Fetch(context.Background(), oci.IfNoMatch(s.curDigest))
	if err != nil {
		return nil, err
	}

	if resp.Matched {
		return s.curSnap, nil
	}

	if s.curSnap, err = storagefs.SnapshotFromFiles(resp.Files...); err != nil {
		return nil, err
	}

	s.curDigest = resp.Digest

	return s.curSnap, nil
}

// Subscribe feeds implementations of fs.FS onto the provided channel.
// It should block until the provided context is cancelled (it will be called in a goroutine).
// It should close the provided channel before it returns.
func (s *Source) Subscribe(ctx context.Context, ch chan<- *storagefs.StoreSnapshot) {
	defer close(ch)

	ticker := time.NewTicker(s.interval)
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			current := s.curDigest
			snap, err := s.Get(ctx)
			if err != nil {
				s.logger.Error("failed resolving upstream", zap.Error(err))
				continue
			}

			if current == s.curDigest {
				s.logger.Debug("store already up to date")
				continue
			}

			ch <- snap

			s.logger.Debug("fetched new reference from remote")
		}
	}
}
