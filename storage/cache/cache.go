package cache

import "errors"

// ErrCacheCorrupt represents a corrupt cache error
var ErrCacheCorrupt = errors.New("cache corrupted")

// Cacher modifies and queries a cache
type Cacher interface {
	Get(key interface{}) (interface{}, bool)
	Add(key, value interface{}) bool
	Remove(key interface{})
}
