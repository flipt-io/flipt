package local

import (
	"context"
	"io/fs"
	"os"

	"github.com/fsnotify/fsnotify"
	"go.uber.org/zap"
)

// Source represents an implementation of an fs.FSSource for local
// updates on a FS.
type Source struct {
	logger *zap.Logger

	dir     string
	watcher *fsnotify.Watcher
}

// NewSource constructs a Source.
func NewSource(logger *zap.Logger, dir string) (*Source, error) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}

	err = watcher.Add(dir)
	if err != nil {
		return nil, err
	}
	return &Source{
		logger:  logger,
		dir:     dir,
		watcher: watcher,
	}, nil
}

// Get returns an fs.FS for the local filesystem.
func (s *Source) Get() (fs.FS, error) {
	fs := os.DirFS(s.dir)
	return fs, nil
}

// Subscribe feeds local fs.FS implementations onto the provided channel.
// It blocks until the provided context is cancelled.
func (s *Source) Subscribe(ctx context.Context, ch chan<- fs.FS) {
	defer close(ch)
	s.logger.Debug("watching", zap.Strings("paths", s.watcher.WatchList()))

	for {
		select {
		case <-ctx.Done():
			return
		case event, ok := <-s.watcher.Events:
			if !ok {
				s.logger.Debug("local directory watcher closed")
				return
			}

			if event.Has(fsnotify.Create) || event.Has(fsnotify.Write) || event.Has(fsnotify.Remove) {
				fs, err := s.Get()
				if err != nil {
					s.logger.Error("error getting filesystem from directory", zap.Error(err))
					continue
				}

				s.logger.Debug("updating local store snapshot")
				ch <- fs
			}
		case err, ok := <-s.watcher.Errors:
			if !ok {
				s.logger.Debug("local directory watcher closed")
				return
			}

			s.logger.Error("local directory store watcher", zap.Error(err))
		}
	}
}

// String returns an identifier string for the store type
func (s *Source) String() string {
	return "local"
}
