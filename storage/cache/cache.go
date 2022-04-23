package cache

import (
	"context"
	"errors"
	"fmt"
)

var (
	// ErrCorrupt represents a corrupt cache error
	ErrCorrupt = errors.New("cache corrupted")
	// ErrMiss represents a cache miss error
	ErrMiss = errors.New("cache miss")
)

// Cacher modifies and queries a cache
type Cacher interface {
	Get(ctx context.Context, key string) (interface{}, error)
	Set(ctx context.Context, key string, value interface{}) error
	Delete(ctx context.Context, key string) error
	Flush(ctx context.Context) error
	fmt.Stringer
}
