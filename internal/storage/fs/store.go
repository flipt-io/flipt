package fs

import (
	"context"
	"fmt"
	"path"

	"go.uber.org/zap"
)

// SnapshotSource produces instances of the storage snapshot.
// A single snapshot can be produced via Get or a channel
// may be provided to Subscribe in order to received
// new instances when new state becomes available.
type SnapshotSource interface {
	fmt.Stringer

	// Get builds a single instance of a *SnapshotSource
	Get(context.Context) (*StoreSnapshot, error)

	// Subscribe feeds instances of *SnapshotSource onto the provided channel.
	// It should block until the provided context is cancelled (it will be called in a goroutine).
	// It should close the provided channel before it returns.
	Subscribe(context.Context, chan<- *StoreSnapshot)
}

// Store is an implementation of storage.Store backed by an SnapshotSource.
// The store subscribes to the source for instances of *SnapshotSource with new contents.
// When a new fs is received the contents is fetched and built into a snapshot
// of Flipt feature flag state.
type Store struct {
	*syncedStore

	logger *zap.Logger
	source SnapshotSource

	// notify is used for test purposes
	// it is invoked if defined when a snapshot update finishes
	notify func()

	cancel context.CancelFunc
	done   chan struct{}
}

func (l *Store) updateSnapshot(storeSnapshot *StoreSnapshot) {
	l.mu.Lock()
	l.Store = storeSnapshot
	l.mu.Unlock()

	// NOTE: this is really just a trick for unit tests
	// It is used to signal that an update occurred
	// so we dont have to e.g. sleep to know when
	// to check state.
	if l.notify != nil {
		l.notify()
	}
}

// NewStore constructs and configure a Store.
// The store creates a background goroutine which feeds a channel of *SnapshotSource.
func NewStore(logger *zap.Logger, source SnapshotSource) (*Store, error) {
	store := &Store{
		syncedStore: &syncedStore{},
		logger:      logger,
		source:      source,
		done:        make(chan struct{}),
	}

	// get an initial snapshot from source.
	f, err := source.Get(context.Background())
	if err != nil {
		return nil, err
	}

	store.updateSnapshot(f)

	var ctx context.Context
	ctx, store.cancel = context.WithCancel(context.Background())

	ch := make(chan *StoreSnapshot)
	go source.Subscribe(ctx, ch)

	go func() {
		defer close(store.done)
		for snap := range ch {
			logger.Debug("received new snapshot")
			store.updateSnapshot(snap)
			logger.Debug("updated latest snapshot")
		}

		logger.Info("source subscription closed")
	}()

	return store, nil
}

// Close cancels the polling routine and waits for the routine to return.
func (l *Store) Close() error {
	l.cancel()

	<-l.done

	return nil
}

// String returns an identifier string for the store type.
func (l *Store) String() string {
	return path.Join("filesystem", l.source.String())
}
