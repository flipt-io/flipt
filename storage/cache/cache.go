package cache

import (
	"errors"
	"fmt"
)

// ErrCacheCorrupt represents a corrupt cache error
var ErrCacheCorrupt = errors.New("cache corrupted")

// Cacher modifies and queries a cache
type Cacher interface {
	Get(key string) (interface{}, bool)
	Set(key string, value interface{})
	Delete(key string)
	Flush()
	fmt.Stringer
}
