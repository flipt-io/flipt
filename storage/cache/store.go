package cache

import (
	"github.com/markphelps/flipt/storage"
	"github.com/sirupsen/logrus"
)

var _ storage.Store = &CacheStore{}

// CacheStore wraps an existing storage.Store and provides caching
type CacheStore struct {
	logger *logrus.Entry
	cache  Cacher
	store  storage.Store
}

// NewCacheStore creates a new *CacheStore
func NewCacheStore(logger *logrus.Entry, cache Cacher, store storage.Store) *CacheStore {
	return &CacheStore{
		logger: logger,
		cache:  cache,
		store:  store,
	}
}
