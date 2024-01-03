package git

import (
	"context"
	"sync"

	lru "github.com/hashicorp/golang-lru/v2"
	"golang.org/x/exp/maps"
)

type buildFunc[K comparable, V any] func(context.Context, K) (V, error)

// cache contains a fixed set of non-evictable entries along
// with additional capacity stored in an LRU.
// The cache is keyed by reference (type R) with an indirect index
// (type K for key) through into the stored values (type V)
// The cache is provided with a resolver on creation, which is
// used to dynamically resolve the key (K) from provided (R).
type cache[R, K comparable, V any] struct {
	mu sync.RWMutex

	fixed map[R]K
	extra *lru.Cache[R, K]

	store map[K]V
}

func newCache[R, K comparable, V any](extra int) (_ *cache[R, K, V], err error) {
	c := &cache[R, K, V]{
		fixed: map[R]K{},
	}

	c.extra, err = lru.NewWithEvict[R, K](extra, c.evict)
	if err != nil {
		return nil, err
	}

	return c, nil
}

// AddFixed forcibly adds the revision, key and value tuple to fixed storage.
// The supplied revision will never be evicted.
// Subsequent calls to Add with the save value for r will update the fixed entries.
func (c *cache[R, K, V]) AddFixed(ctx context.Context, r R, k K, v V) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.fixed[r] = k
	c.store[k] = v

	return
}

// AddOrBuild adds the revision, key and value tuple.
// If the reference r is already tracked in the fixed set, then it is updated there.
// Otherwise, the entry is added to the LRU cache.
func (c *cache[R, K, V]) AddOrBuild(ctx context.Context, r R, k K, build buildFunc[K, V]) (V, error) {
	v, ok, err := c.getOrBuild(ctx, r, k, build)
	if err != nil {
		return v, err
	}

	// fast path: r, k and v are all as expected already
	if ok {
		return v, nil
	}

	// Otherwise, either r did not resolve to k or there was
	// no v stored for k
	// In all cases we update both atomically to ensure that
	// we don't get a race condition leading to a broken
	// reference to key to value link

	c.mu.Lock()
	defer c.mu.Unlock()

	if _, ok := c.fixed[r]; ok {
		c.fixed[r] = k
	} else {
		c.extra.Add(r, k)
	}
	c.store[k] = v

	return v, nil
}

// Get attempts to resolve the value V for a given reference r.
func (c *cache[R, K, V]) Get(r R) (v V, ok bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	k, ok := c.fixed[r]
	if ok {
		v, ok = c.store[k]
		return
	}

	k, ok = c.extra.Get(r)
	if !ok {
		return v, ok
	}

	v, ok = c.store[k]
	return
}

func (c *cache[R, K, V]) getOrBuild(ctx context.Context, r R, k K, build buildFunc[K, V]) (v V, ok bool, err error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	v, ok = c.store[k]
	if !ok {
		// if there is no V for the provided k then we
		// return early with ok is false to let the caller
		// know to update all the references
		v, err = build(ctx, k)
		return
	}

	// check whether r is present in the fixed set
	// if it is and it point to the provided k then
	// we return true to signify everything is consistent
	if fk, present := c.fixed[r]; present {
		ok = fk == k
		return
	}

	// same as before except this time we check in the LRU
	if ek, present := c.extra.Get(r); present {
		ok = ek == k
	}

	return
}

// References returns all the references currently tracked within the cache.
func (c *cache[R, K, V]) References() []R {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return append(maps.Keys(c.fixed), c.extra.Keys()...)
}

// evict is used for garbage collection when keys are evicted from
// the LRU cache.
// It checks to see if the target key for the evicted revision is
// still being referenced by other existing references in either the
// fixed set or the remaining LRU entries.
// If the key is dangling then it removes the entry from the store.
// NOTE: calls to evict must be made while holding a write lock
// the LRU implementation inlines calls to evict on calls to cache.Add
// we only call Add in AddOrBuild while holding a write lock
func (c *cache[R, K, V]) evict(r R, k K) {
	for _, key := range append(maps.Values(c.fixed), c.extra.Values()...) {
		if key == k {
			return
		}
	}

	delete(c.store, k)
}
