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

type KeyOptions struct {
	prefix string
}

type KeyOption func(*KeyOptions)

func WithPrefix(prefix string) KeyOption {
	return func(o *KeyOptions) {
		o.prefix = prefix
	}
}

func Key(k string, opts ...KeyOption) string {
	ko := KeyOptions{
		prefix: "flipt",
	}

	for _, opt := range opts {
		opt(&ko)
	}

	return fmt.Sprintf("%s:%x", ko.prefix, md5.Sum([]byte(k)))
}
