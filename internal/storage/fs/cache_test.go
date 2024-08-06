package fs

import (
	"context"
	"errors"
	"math/rand"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.flipt.io/flipt/internal/ext"
	"go.uber.org/zap/zaptest"
	"golang.org/x/sync/errgroup"
	"google.golang.org/protobuf/types/known/timestamppb"
)

const (
	referenceFixed = "main"
	referenceA     = "reference-A"
	referenceB     = "reference-B"
	referenceC     = "reference-C"

	revisionOne   = "revision-one"
	revisionTwo   = "revision-two"
	revisionThree = "revision-three"
)

var (
	snapshotOne   = newMockSnapshot("one", "One", timestamppb.Now())
	snapshotTwo   = newMockSnapshot("two", "Two", timestamppb.Now())
	snapshotThree = newMockSnapshot("three", "Three", timestamppb.Now())
)

func newMockSnapshot(key, name string, created *timestamppb.Timestamp) *Snapshot {
	n := newNamespace(&ext.NamespaceEmbed{IsNamespace: &ext.Namespace{Key: key, Name: name}}, created)
	return &Snapshot{ns: map[string]*namespace{key: n}}
}

func Test_SnapshotCache(t *testing.T) {
	cache, err := NewSnapshotCache[string](zaptest.NewLogger(t), 2)
	require.NoError(t, err)

	ctx := context.Background()
	t.Run("References", func(t *testing.T) {
		assert.Empty(t, cache.References())

		cache.AddFixed(ctx, referenceFixed, revisionOne, snapshotOne)

		assert.Equal(t, []string{referenceFixed}, cache.References())
	})

	t.Run("Get fixed entry", func(t *testing.T) {
		found, ok := cache.Get(referenceFixed)
		require.True(t, ok, "snapshot not found")

		assert.Equal(t, snapshotOne, found)
	})

	builder := newSnapshotBuilder(map[string]*Snapshot{
		revisionOne: snapshotOne,
		revisionTwo: snapshotTwo,
	})

	t.Run("AddOrBuild new reference with existing revision", func(t *testing.T) {
		found, err := cache.AddOrBuild(ctx, referenceA, revisionOne, builder.build)
		require.NoError(t, err)

		assert.Equal(t, snapshotOne, found, "AddOrBuild returned unexpected snapshot")
		assert.Zero(t, builder.builds[revisionOne], "Revision was rebuilt when it should have been cached")

		var ok bool
		found, ok = cache.Get(referenceA)
		require.True(t, ok, "New reference is not retrievable from cache")
		assert.Equal(t, snapshotOne, found)

		assert.Equal(t, []string{referenceFixed, referenceA}, cache.References())
	})

	t.Run("AddOrBuild new reference with new revision", func(t *testing.T) {
		found, err := cache.AddOrBuild(ctx, referenceB, revisionTwo, builder.build)
		require.NoError(t, err)

		assert.Equal(t, snapshotTwo, found, "AddOrBuild returned unexpected snapshot")
		assert.Equal(t, 1, builder.builds[revisionTwo], "Revision was not built as expected")

		var ok bool
		found, ok = cache.Get(referenceB)
		require.True(t, ok, "New reference is not retrievable from cache")
		assert.Equal(t, snapshotTwo, found)

		assert.Equal(t, []string{referenceFixed, referenceA, referenceB}, cache.References())
	})

	t.Run("AddOrBuild existing reference with existing revision", func(t *testing.T) {
		found, err := cache.AddOrBuild(ctx, referenceB, revisionTwo, builder.build)
		require.NoError(t, err)

		assert.Equal(t, snapshotTwo, found, "AddOrBuild returned unexpected snapshot")
		assert.Equal(t, 1, builder.builds[revisionTwo], "Revision was rebuilt unexpectedly")

		var ok bool
		found, ok = cache.Get(referenceB)
		require.True(t, ok, "Existing reference is no longer retrievable from cache")
		assert.Equal(t, snapshotTwo, found)

		assert.Equal(t, []string{referenceFixed, referenceA, referenceB}, cache.References())
	})

	t.Run("AddOrBuild existing reference with new revision", func(t *testing.T) {
		// from: referenceB -> revisionTwo -> snapshotTwo
		// to:   referenceB -> revisionOne -> snapshotOne
		found, err := cache.AddOrBuild(ctx, referenceB, revisionOne, builder.build)
		require.NoError(t, err)

		assert.Equal(t, snapshotOne, found, "AddOrBuild returned unexpected snapshot")
		// revision already exists as referenceFixed and referenceA point to it
		// it was inserted via AddFixed and has not been built by the builder
		assert.Zero(t, builder.builds[revisionOne], "Revision was rebuilt unexpectedly")

		// ensure still retrievable
		var ok bool
		found, ok = cache.Get(referenceB)
		require.True(t, ok, "Existing reference is no longer retrievable from cache")
		assert.Equal(t, snapshotOne, found)
	})

	t.Run("AddOrBuild new reference with previously evicted revision", func(t *testing.T) {
		// attempt to reference to a revision which should have been evicted
		// referenceB was updated from revisionTwo to revisionOne leaving the revision dangling
		// the cache should have identified this and evicted it
		found, err := cache.AddOrBuild(ctx, referenceC, revisionTwo, builder.build)
		require.NoError(t, err)

		assert.Equal(t, snapshotTwo, found, "AddOrBuild returned unexpected snapshot")
		// revisionTwo should be built again because it was dropped from the cache store
		// when referenceB was pointed at referenceOne
		assert.Equal(t, 2, builder.builds[revisionTwo], "Revision wasn't rebuilt as expected")

		// ensure still retrievable
		var ok bool
		found, ok = cache.Get(referenceC)
		require.True(t, ok, "Existing reference is no longer retrievable from cache")
		assert.Equal(t, snapshotTwo, found)

		// now that we're exceeding the cache extra capacity (2) the oldest
		// reference (in terms of recency) should be evicted
		found, ok = cache.Get(referenceA)
		require.False(t, ok)
		assert.Nil(t, found)
	})

	t.Run("AddOrBuild fixed reference with different but existing revision", func(t *testing.T) {
		// from: referenceFixed -> revisionOne -> snapshotOne
		// to:   referenceFixed -> revisionTwo -> snapshotTwo
		found, err := cache.AddOrBuild(ctx, referenceFixed, revisionTwo, builder.build)
		require.NoError(t, err)

		assert.Equal(t, snapshotTwo, found, "AddOrBuild returned unexpected snapshot")
		// revisionTwo's snapshot should currently exist in cache and so we shouldn't
		// see its build count exceed the previous value 2
		assert.Equal(t, 2, builder.builds[revisionTwo], "Revision wasn't rebuilt as expected")

		// ensure still retrievable
		var ok bool
		found, ok = cache.Get(referenceFixed)
		require.True(t, ok, "Existing reference is no longer retrievable from cache")
		assert.Equal(t, snapshotTwo, found)
	})
}

func Test_SnapshotCache_Concurrently(t *testing.T) {
	cache, err := NewSnapshotCache[string](zaptest.NewLogger(t), 2)
	require.NoError(t, err)

	ctx := context.Background()
	snaps := map[string]*Snapshot{
		revisionOne:   snapshotOne,
		revisionTwo:   snapshotTwo,
		revisionThree: snapshotThree,
	}

	builder := newSnapshotBuilder(snaps)

	cache.AddFixed(ctx, referenceFixed, revisionOne, snapshotOne)

	var group errgroup.Group
	for _, ref := range []string{referenceA, referenceB, referenceC} {
		for _, rev := range []string{revisionOne, revisionTwo, revisionThree} {
			var (
				ref = ref
				rev = rev
			)

			group.Go(func() error {
				// each goroutine will make a bunch of attempts
				// to store their respective ref / revision pair
				for i := 0; i < 10; i++ {
					// add a little entropy to the order
					time.Sleep(time.Duration(rand.Intn(10)) * time.Millisecond)

					snap, err := cache.AddOrBuild(ctx, ref, rev, builder.build)
					require.NoError(t, err)

					// ensure snap is as expected
					assert.Equal(t, snaps[rev], snap)
				}
				return nil
			})
		}
	}

	require.NoError(t, group.Wait())

	// revisionOne will always be referenced by referenceFixed and so we
	// should expect it to never get built by the build function
	assert.Zero(t, builder.builds[revisionOne])
	// the other revisions should get built at-least once and then fall in
	// and out of the cache due to evictions
	assert.GreaterOrEqual(t, builder.builds[revisionTwo], 1)
	assert.GreaterOrEqual(t, builder.builds[revisionThree], 1)
}

type snapshotBuiler struct {
	mu     sync.Mutex
	snaps  map[string]*Snapshot
	builds map[string]int
}

func newSnapshotBuilder(snaps map[string]*Snapshot) *snapshotBuiler {
	return &snapshotBuiler{builds: map[string]int{}, snaps: snaps}
}

func (b *snapshotBuiler) build(_ context.Context, hash string) (*Snapshot, error) {
	b.mu.Lock()
	defer b.mu.Unlock()

	snap, ok := b.snaps[hash]
	if !ok {
		return nil, errors.New("unexpected hash")
	}

	b.builds[hash] += 1

	return snap, nil
}
