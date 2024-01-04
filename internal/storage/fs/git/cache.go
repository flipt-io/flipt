package git

import (
	"context"
	"sync"

	lru "github.com/hashicorp/golang-lru/v2"
	"go.flipt.io/flipt/internal/storage/fs"
	"golang.org/x/exp/maps"
)

type buildFunc[K comparable] func(context.Context, K) (*fs.Snapshot, error)

// cache contains a fixed set of non-evictable entries along
// with additional capacity stored in an LRU.
// The cache is keyed by reference with an indirect index
// (type K for key) through into the stored snapshots.
// The type K is generic to support the different kinds of content
// address types we expect to support (commit SHA and OCI digest).
type cache[K comparable] struct {
	mu sync.RWMutex

	fixed map[string]K
	extra *lru.Cache[string, K]

	store map[K]*fs.Snapshot
}

func newCache[K comparable](extra int) (_ *cache[K], err error) {
	c := &cache[K]{
		fixed: map[string]K{},
		store: map[K]*fs.Snapshot{},
	}

	c.extra, err = lru.NewWithEvict[string, K](extra, c.evict)
	if err != nil {
		return nil, err
	}

	return c, nil
}

// AddFixed forcibly adds the reference, key and value tuple to fixed storage.
// The supplied reference will never be evicted.
// Subsequent calls to Add with the same value for ref will update the fixed entries.
func (c *cache[K]) AddFixed(ctx context.Context, ref string, k K, s *fs.Snapshot) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.fixed[ref] = k
	c.store[k] = s

	return
}

// AddOrBuild adds the reference, key and value tuple.
// If the reference r is already tracked in the fixed set, then it is updated there.
// Otherwise, the entry is added to the LRU cache.
func (c *cache[K]) AddOrBuild(ctx context.Context, ref string, k K, build buildFunc[K]) (*fs.Snapshot, error) {
	s, ok, err := c.getOrBuild(ctx, ref, k, build)
	if err != nil {
		return s, err
	}

	// fast path: r, k and v are all as expected already
	if ok {
		return s, nil
	}

	if s == nil {
		s, err = build(ctx, k)
		if err != nil {
			return s, err
		}
	}

	// Otherwise, either r did not resolve to k or there was
	// no v stored for k
	// In all cases we update both atomically to ensure that
	// we don't get a race condition leading to a broken
	// reference to key to value link

	c.mu.Lock()
	defer c.mu.Unlock()

	if _, ok := c.fixed[ref]; ok {
		c.fixed[ref] = k
	} else {
		c.extra.Add(ref, k)
	}
	c.store[k] = s

	return s, nil
}

// Get attempts to resolve a snapshot for a given reference r.
func (c *cache[K]) Get(ref string) (s *fs.Snapshot, ok bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	k, ok := c.fixed[ref]
	if ok {
		s, ok = c.store[k]
		return
	}

	k, ok = c.extra.Get(ref)
	if !ok {
		return s, ok
	}

	s, ok = c.store[k]
	return
}

func (c *cache[K]) getOrBuild(ctx context.Context, ref string, k K, build buildFunc[K]) (s *fs.Snapshot, ok bool, err error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	s, ok = c.store[k]
	if !ok {
		// if there is no V for the provided k then we
		// return early with ok is false to let the caller
		// know to update all the references
		s, err = build(ctx, k)
		return
	}

	// check whether r is present in the fixed set
	// if it is and it point to the provided k then
	// we return true to signify everything is consistent
	if fk, present := c.fixed[ref]; present {
		ok = fk == k
		return
	}

	// same as before except this time we check in the LRU
	if ek, present := c.extra.Get(ref); present {
		ok = ek == k
	}

	return
}

// References returns all the references currently tracked within the cache.
func (c *cache[K]) References() []string {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return append(maps.Keys(c.fixed), c.extra.Keys()...)
}

// evict is used for garbage collection when keys are evicted from
// the LRU cache.
// It checks to see if the target key for the evicted reference is
// still being referenced by other existing references in either the
// fixed set or the remaining LRU entries.
// If the key is dangling then it removes the entry from the store.
// NOTE: calls to evict must be made while holding a write lock
// the LRU implementation inlines calls to evict on calls to cache.Add
// we only call Add in AddOrBuild while holding a write lock
func (c *cache[K]) evict(_ string, k K) {
	for _, key := range append(maps.Values(c.fixed), c.extra.Values()...) {
		if key == k {
			return
		}
	}

	delete(c.store, k)
}
