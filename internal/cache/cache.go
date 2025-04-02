package cache

import (
	"context"
	"crypto/md5"
	"fmt"
)

// Cacher modifies and queries a cache
type Cacher interface {
	// Get retrieves a value from the cache, the bool indicates if the item was found
	Get(ctx context.Context, key string) ([]byte, bool, error)
	// Set sets a value in the cache
	Set(ctx context.Context, key string, value []byte) error
	// Delete removes a value from the cache
	Delete(ctx context.Context, key string) error
	fmt.Stringer
}

var DefaultKeyer = NewKeyer("flipt")

type Keyer struct {
	prefix string
}

func NewKeyer(prefix string) *Keyer {
	return &Keyer{prefix: prefix}
}

func (k *Keyer) Key(key string) string {
	return fmt.Sprintf("%s:%x", k.prefix, md5.Sum([]byte(key)))
}
