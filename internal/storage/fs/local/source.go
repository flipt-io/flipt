package local

import (
	"context"
	"io/fs"
	"os"
	"time"

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
func NewSource(logger *zap.Logger, dir string, duration time.Duration) (*Source, error) {
	return &Source{
		logger:   logger,
		dir:      dir,
		interval: duration,
	}, nil
}

// Get returns an fs.FS for the local filesystem.
func (s *Source) Get() (fs.FS, error) {
	return os.DirFS(s.dir), nil
}

// Subscribe feeds local fs.FS implementations onto the provided channel.
// It blocks until the provided context is cancelled.
func (s *Source) Subscribe(ctx context.Context, ch chan<- fs.FS) {
	defer close(ch)

	ticker := time.NewTicker(s.interval)
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			fs, err := s.Get()
			if err != nil {
				s.logger.Error("error getting file system from directory", zap.Error(err))
				continue
			}

			s.logger.Debug("updating local store snapshot")
			ch <- fs
		}
	}
}

// String returns an identifier string for the store type.
func (s *Source) String() string {
	return "local"
}
