package cache

import (
	"fmt"

	"github.com/sirupsen/logrus"
	"go.flipt.io/flipt/storage"
)

var _ storage.Store = &Store{}

// Store wraps an existing storage.Store and provides caching
type Store struct {
	logger *logrus.Entry
	cache  Cacher
	store  storage.Store
}

// NewStore creates a new *CacheStore
func NewStore(logger *logrus.Entry, cache Cacher, store storage.Store) *Store {
	return &Store{
		logger: logger,
		cache:  cache,
		store:  store,
	}
}

func (c *Store) String() string {
	return fmt.Sprintf("[cached] %s", c.store.String())
}
