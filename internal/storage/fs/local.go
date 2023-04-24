package fs

import (
	"context"
	"os"
	"path/filepath"
	"sync"

	"github.com/fsnotify/fsnotify"
	"go.flipt.io/flipt/internal/storage"
	"go.flipt.io/flipt/rpc/flipt"
	"go.uber.org/zap"
)

type LocalDirectoryStore struct {
	*storeSnapshot

	logger  *zap.Logger
	mu      sync.RWMutex
	watcher *fsnotify.Watcher
}

func (l *LocalDirectoryStore) updateSnapshot(dir string) (err error) {
	l.mu.Lock()
	defer l.mu.Unlock()

	var dirs []string
	l.storeSnapshot, dirs, err = newStoreSnapshot(os.DirFS(dir))
	if err != nil {
		return
	}

	for _, dir := range dirs {
		l.watcher.Add(dir)
	}

	return
}

func NewLocalDirectoryStore(logger *zap.Logger, dir string) (*LocalDirectoryStore, error) {
	dir, err := filepath.Abs(dir)
	if err != nil {
		return nil, err
	}

	logger = logger.With(zap.String("dir", dir))
	store := &LocalDirectoryStore{
		logger: logger,
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

func (l *LocalDirectoryStore) GetFlag(ctx context.Context, namespaceKey string, key string) (*flipt.Flag, error) {
	l.mu.RLock()
	defer l.mu.RUnlock()

	return l.storeSnapshot.GetFlag(ctx, namespaceKey, key)
}

func (l *LocalDirectoryStore) ListFlags(ctx context.Context, namespaceKey string, opts ...storage.QueryOption) (storage.ResultSet[*flipt.Flag], error) {
	l.mu.RLock()
	defer l.mu.RUnlock()

	return l.storeSnapshot.ListFlags(ctx, namespaceKey, opts...)
}

func (l *LocalDirectoryStore) CountFlags(ctx context.Context, namespaceKey string) (uint64, error) {
	l.mu.RLock()
	defer l.mu.RUnlock()

	return l.storeSnapshot.CountFlags(ctx, namespaceKey)
}

func (l *LocalDirectoryStore) GetRule(ctx context.Context, namespaceKey string, id string) (*flipt.Rule, error) {
	l.mu.RLock()
	defer l.mu.RUnlock()

	return l.storeSnapshot.GetRule(ctx, namespaceKey, id)
}

func (l *LocalDirectoryStore) ListRules(ctx context.Context, namespaceKey string, flagKey string, opts ...storage.QueryOption) (storage.ResultSet[*flipt.Rule], error) {
	l.mu.RLock()
	defer l.mu.RUnlock()

	return l.storeSnapshot.ListRules(ctx, namespaceKey, flagKey, opts...)
}

func (l *LocalDirectoryStore) CountRules(ctx context.Context, namespaceKey string) (uint64, error) {
	l.mu.RLock()
	defer l.mu.RUnlock()

	return l.storeSnapshot.CountRules(ctx, namespaceKey)
}

func (l *LocalDirectoryStore) GetSegment(ctx context.Context, namespaceKey string, key string) (*flipt.Segment, error) {
	l.mu.RLock()
	defer l.mu.RUnlock()

	return l.storeSnapshot.GetSegment(ctx, namespaceKey, key)
}

func (l *LocalDirectoryStore) ListSegments(ctx context.Context, namespaceKey string, opts ...storage.QueryOption) (storage.ResultSet[*flipt.Segment], error) {
	l.mu.RLock()
	defer l.mu.RUnlock()

	return l.storeSnapshot.ListSegments(ctx, namespaceKey, opts...)
}

func (l *LocalDirectoryStore) CountSegments(ctx context.Context, namespaceKey string) (uint64, error) {
	l.mu.RLock()
	defer l.mu.RUnlock()

	return l.storeSnapshot.CountSegments(ctx, namespaceKey)
}

// GetEvaluationRules returns rules applicable to flagKey provided
// Note: Rules MUST be returned in order by Rank
func (l *LocalDirectoryStore) GetEvaluationRules(ctx context.Context, namespaceKey string, flagKey string) ([]*storage.EvaluationRule, error) {
	l.mu.RLock()
	defer l.mu.RUnlock()

	return l.storeSnapshot.GetEvaluationRules(ctx, namespaceKey, flagKey)
}

func (l *LocalDirectoryStore) GetEvaluationDistributions(ctx context.Context, ruleID string) ([]*storage.EvaluationDistribution, error) {
	l.mu.RLock()
	defer l.mu.RUnlock()

	return l.storeSnapshot.GetEvaluationDistributions(ctx, ruleID)
}

func (l *LocalDirectoryStore) GetNamespace(ctx context.Context, key string) (*flipt.Namespace, error) {
	l.mu.RLock()
	defer l.mu.RUnlock()

	return l.storeSnapshot.GetNamespace(ctx, key)
}

func (l *LocalDirectoryStore) ListNamespaces(ctx context.Context, opts ...storage.QueryOption) (storage.ResultSet[*flipt.Namespace], error) {
	l.mu.RLock()
	defer l.mu.RUnlock()

	return l.storeSnapshot.ListNamespaces(ctx, opts...)
}

func (l *LocalDirectoryStore) CountNamespaces(ctx context.Context) (uint64, error) {
	l.mu.RLock()
	defer l.mu.RUnlock()

	return l.storeSnapshot.CountNamespaces(ctx)
}

func (l *LocalDirectoryStore) String() string {
	return "local"
}
