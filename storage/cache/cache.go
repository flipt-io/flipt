package cache

import "errors"

var ErrCacheCorrupt = errors.New("cache corrupted")

type Cacher interface {
	Get(key interface{}) (interface{}, bool)
	Add(key, value interface{}) bool
	Remove(key interface{})
}
