package cache

import (
	"github.com/markphelps/flipt/storage"
	"github.com/sirupsen/logrus"
)

var _ storage.Store = &Store{}

// Store wraps an existing storage.Store and provides caching
type Store struct {
	logger *logrus.Entry
	cache  Cacher
	store  storage.Store
}

// NewCacheStore creates a new *CacheStore
func NewStore(logger *logrus.Entry, cache Cacher, store storage.Store) *Store {
	return &Store{
		logger: logger,
		cache:  cache,
		store:  store,
	}
}
