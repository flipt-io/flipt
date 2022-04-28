package cache

import (
	"context"
	"errors"
	"fmt"
)

var (
	// ErrCorrupt represents a corrupt cache error
	ErrCorrupt = errors.New("cache corrupted")
)

// Cacher modifies and queries a cache
type Cacher interface {
	// Get retrieves a value from the cache, the bool indicates if the item was found
	Get(ctx context.Context, key string) (interface{}, bool, error)
	// Set sets a value in the cache
	Set(ctx context.Context, key string, value interface{}) error
	// Delete removes a value from the cache
	Delete(ctx context.Context, key string) error
	// Flush removes all values from the cache
	Flush(ctx context.Context) error
	fmt.Stringer
}
