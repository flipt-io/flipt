package fs

import (
	"errors"
	"io/fs"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/storage/memory"
	"go.flipt.io/flipt/internal/containers"
	"go.flipt.io/flipt/internal/gitfs"
	"go.uber.org/zap"
)

type GitStore struct {
	*syncedStore

	logger *zap.Logger
	repo   *git.Repository
	conf   GitStoreConfig
}

func (l *GitStore) updateSnapshot(fs fs.FS) error {
	storeSnapshot, _, err := newStoreSnapshot(fs)
	if err != nil {
		return err
	}

	l.mu.Lock()
	l.storeSnapshot = storeSnapshot
	l.mu.Unlock()

	return nil
}

type GitStoreConfig struct {
	url string
	ref plumbing.ReferenceName
}

func NewGitStoreConfig(url string, opts ...containers.Option[GitStoreConfig]) GitStoreConfig {
	c := GitStoreConfig{
		url: url,
		ref: plumbing.NewRemoteReferenceName("origin", "main"),
	}
	containers.ApplyAll(&c)
	return c
}

func NewGitStore(logger *zap.Logger, conf GitStoreConfig) (*GitStore, error) {
	logger = logger.With(
		zap.String("repo", conf.url),
		zap.Stringer("ref", conf.ref),
	)
	store := &GitStore{
		syncedStore: &syncedStore{},
		logger:      logger,
		conf:        conf,
	}

	repo, err := git.Clone(memory.NewStorage(), nil, &git.CloneOptions{
		URL: conf.url,
	})
	if err != nil {
		return nil, err
	}

	fs, err := gitfs.NewFromRepo(repo, gitfs.WithReference(store.conf.ref))
	if err != nil {
		return nil, err
	}

	if err = store.updateSnapshot(fs); err != nil {
		return nil, err
	}

	go func() {
		for range time.Tick(5 * time.Second) {
			logger.Debug("fetching from remote")

			if err := repo.Fetch(&git.FetchOptions{}); err != nil {
				if errors.Is(err, git.NoErrAlreadyUpToDate) {
					logger.Debug("store already up to date")
					continue
				}

				logger.Error("failed fetching remote", zap.Error(err))
				continue
			}

			fs, err := gitfs.NewFromRepo(repo, gitfs.WithReference(store.conf.ref))
			if err != nil {
				logger.Error("failed creating gitfs", zap.Error(err))
				continue
			}

			if err = store.updateSnapshot(fs); err != nil {
				logger.Error("failed updating snapshot", zap.Error(err))
				continue
			}

			logger.Debug("finished fetching from remote")
		}
	}()

	return store, nil
}

func (l *GitStore) Close() error {
	return nil
}

func (l *GitStore) String() string {
	return "git"
}
