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

func Key(k string) string {
	return fmt.Sprintf("flipt:%x", md5.Sum([]byte(k)))
}

type contextCacheKey string

// do not store informs the cache to not store the value in the cache
var doNotStore contextCacheKey = "no-store"

func WithDoNotStore(ctx context.Context) context.Context {
	return context.WithValue(ctx, doNotStore, true)
}

func IsDoNotStore(ctx context.Context) bool {
	return ctx.Value(doNotStore) != nil
}
