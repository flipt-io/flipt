package fs

import (
	"context"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/timestamppb"
)

const (
	revisionOne = "revision-one"
	revisionTwo = "revision-two"

	referenceFixed = "main"
	referenceA     = "reference-A"
	referenceB     = "reference-B"
)

func Test_SnapshotCache(t *testing.T) {
	cache, err := NewSnapshotCache[string](2)
	require.NoError(t, err)

	ctx := context.Background()
	snap := &Snapshot{ns: map[string]*namespace{
		"default": newNamespace("default", "Default", timestamppb.Now()),
	}}

	t.Run("References", func(t *testing.T) {
		assert.Empty(t, cache.References())

		cache.AddFixed(ctx, referenceFixed, revisionOne, snap)

		assert.Equal(t, []string{referenceFixed}, cache.References())
	})

	t.Run("Get fixed entry", func(t *testing.T) {
		found, ok := cache.Get(referenceFixed)
		require.True(t, ok, "snapshot not found")

		assert.Equal(t, snap, found)
	})

	builder := newSnapshotBuilder(snap)

	t.Run("AddOrBuild new reference with existing revision", func(t *testing.T) {
		found, err := cache.AddOrBuild(ctx, referenceA, revisionOne, builder.build)
		require.NoError(t, err)

		assert.Equal(t, snap, found, "AddOrBuild returned unexpected snapshot")
		assert.Zero(t, builder.hashes[revisionOne], "Revision was rebuilt when it should have been cached")

		var ok bool
		found, ok = cache.Get(referenceA)
		require.True(t, ok, "New reference is not retrievable from cache")
		assert.Equal(t, snap, found)

		assert.Equal(t, []string{referenceFixed, referenceA}, cache.References())
	})

	t.Run("AddOrBuild new reference with new revision", func(t *testing.T) {
		found, err := cache.AddOrBuild(ctx, referenceB, revisionTwo, builder.build)
		require.NoError(t, err)

		assert.Equal(t, snap, found, "AddOrBuild returned unexpected snapshot")
		assert.Equal(t, 1, builder.hashes[revisionTwo], "Revision was not built as expected")

		var ok bool
		found, ok = cache.Get(referenceB)
		require.True(t, ok, "New reference is not retrievable from cache")
		assert.Equal(t, snap, found)

		assert.Equal(t, []string{referenceFixed, referenceA, referenceB}, cache.References())
	})

	t.Run("AddOrBuild existing reference with existing revision", func(t *testing.T) {
		found, err := cache.AddOrBuild(ctx, referenceB, revisionTwo, builder.build)
		require.NoError(t, err)

		assert.Equal(t, snap, found, "AddOrBuild returned unexpected snapshot")
		assert.Equal(t, 1, builder.hashes[revisionTwo], "Revision was rebuilt unexpectedly")

		var ok bool
		found, ok = cache.Get(referenceB)
		require.True(t, ok, "Existing reference is no longer retrievable from cache")
		assert.Equal(t, snap, found)

		assert.Equal(t, []string{referenceFixed, referenceA, referenceB}, cache.References())
	})

	t.Run("AddOrBuild existing reference with new revision", func(t *testing.T) {
		// from: referenceB -> revisionTwo
		// to:   referenceB -> revisionOne
		found, err := cache.AddOrBuild(ctx, referenceB, revisionOne, builder.build)
		require.NoError(t, err)

		assert.Equal(t, snap, found, "AddOrBuild returned unexpected snapshot")
		// revision already exists as referenceFixed and referenceA point to it
		// it was inserted via AddFixed and has not been built by the builder
		assert.Zero(t, builder.hashes[revisionOne], "Revision was rebuilt unexpectedly")

		// ensure still retrievable
		var ok bool
		found, ok = cache.Get(referenceB)
		require.True(t, ok, "Existing reference is no longer retrievable from cache")
		assert.Equal(t, snap, found)
	})
}

type snapshotBuiler struct {
	mu     sync.Mutex
	hashes map[string]int
	snap   *Snapshot
}

func newSnapshotBuilder(snap *Snapshot) *snapshotBuiler {
	return &snapshotBuiler{hashes: map[string]int{}, snap: snap}
}

func (b *snapshotBuiler) build(_ context.Context, hash string) (*Snapshot, error) {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.hashes[hash] += 1

	return b.snap, nil
}
