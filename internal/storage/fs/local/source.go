package local

import (
	"context"
	"os"
	"time"

	"go.flipt.io/flipt/internal/containers"
	storagefs "go.flipt.io/flipt/internal/storage/fs"
	"go.uber.org/zap"
)

// Source represents an implementation of an fs.FSSource for local
// updates on a FS.
type Source struct {
	logger *zap.Logger

	dir      string
	interval time.Duration
}

// NewSource constructs a Source.
func NewSource(logger *zap.Logger, dir string, opts ...containers.Option[Source]) (*Source, error) {
	s := &Source{
		logger:   logger,
		dir:      dir,
		interval: 10 * time.Second,
	}

	containers.ApplyAll(s, opts...)

	return s, nil
}

// WithPollInterval configures the interval in which we will restore
// the local fs.
func WithPollInterval(tick time.Duration) containers.Option[Source] {
	return func(s *Source) {
		s.interval = tick
	}
}

// Get returns an fs.FS for the local filesystem.
func (s *Source) Get(context.Context) (*storagefs.StoreSnapshot, error) {
	return storagefs.SnapshotFromFS(s.logger, os.DirFS(s.dir))
}

// Subscribe feeds local fs.FS implementations onto the provided channel.
// It blocks until the provided context is cancelled.
func (s *Source) Subscribe(ctx context.Context, ch chan<- *storagefs.StoreSnapshot) {
	defer close(ch)

	ticker := time.NewTicker(s.interval)
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			snap, err := s.Get(ctx)
			if err != nil {
				s.logger.Error("error getting file system from directory", zap.Error(err))
				continue
			}

			s.logger.Debug("updating local store snapshot")

			ch <- snap
		}
	}
}

// String returns an identifier string for the store type.
func (s *Source) String() string {
	return "local"
}
