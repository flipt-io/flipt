package fs

import (
	"os"
	"path/filepath"

	"github.com/fsnotify/fsnotify"
	"go.uber.org/zap"
)

type LocalDirectoryStore struct {
	*syncedStore

	logger  *zap.Logger
	watcher *fsnotify.Watcher
}

func (l *LocalDirectoryStore) updateSnapshot(dir string) error {
	storeSnapshot, dirs, err := newStoreSnapshot(os.DirFS(dir))
	if err != nil {
		return err
	}

	l.mu.Lock()
	defer l.mu.Unlock()

	l.storeSnapshot = storeSnapshot
	for _, dir := range dirs {
		l.watcher.Add(dir)
	}

	return nil
}

func NewLocalDirectoryStore(logger *zap.Logger, dir string) (*LocalDirectoryStore, error) {
	dir, err := filepath.Abs(dir)
	if err != nil {
		return nil, err
	}

	logger = logger.With(zap.String("dir", dir))
	store := &LocalDirectoryStore{
		syncedStore: &syncedStore{},
		logger:      logger,
	}

	store.watcher, err = fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}

	if err = store.updateSnapshot(dir); err != nil {
		return nil, err
	}

	go func() {
		for {
			select {
			case event, ok := <-store.watcher.Events:
				logger.Debug("watch event", zap.Stringer("event", event))
				if !ok {
					logger.Info("local directory store watcher closed")
					return
				}

				if event.Has(fsnotify.Write) ||
					event.Has(fsnotify.Remove) ||
					event.Has(fsnotify.Create) {
					logger.Debug("updating local store snapshot")
					if err := store.updateSnapshot(dir); err != nil {
						logger.Error("failed to update local store snapshot", zap.Error(err))
					}
					logger.Debug("finished update of local store snapshot")
				}
			case err, ok := <-store.watcher.Errors:
				if !ok {
					logger.Info("local directory store watcher closed")
					return
				}

				logger.Error("local directory store watcher", zap.Error(err))
			}
		}
	}()

	logger.Debug("watching", zap.Strings("paths", store.watcher.WatchList()))

	return store, nil
}

func (l *LocalDirectoryStore) Close() error {
	return l.watcher.Close()
}

func (l *LocalDirectoryStore) String() string {
	return "local"
}
