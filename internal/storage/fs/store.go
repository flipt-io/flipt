package fs

import (
	"context"
	"fmt"
	"io/fs"
	"path"

	"go.uber.org/zap"
)

// FSSource produces implementations of fs.FS.
// A single FS can be produced via Get or a channel
// may be provided to Subscribe in order to received
// new instances when new state becomes available.
type FSSource interface {
	fmt.Stringer

	// Get builds a single instance of an fs.FS
	Get() (fs.FS, error)

	// Subscribe feeds implementations of fs.FS onto the provided channel.
	// It should block until the provided context is cancelled (it will be called in a goroutine).
	// It should close the provided channel before it returns.
	Subscribe(context.Context, chan<- fs.FS)
}

// Store is an implementation of storage.Store backed by an FSSource.
// The store subscribes to the source for instances of fs.FS with new contents.
// When a new fs is received the contents is fetched and built into a snapshot
// of Flipt feature flag state.
type Store struct {
	*syncedStore

	logger *zap.Logger
	source FSSource

	// notify is used for test purposes
	// it is invoked if defined when a snapshot update finishes
	notify func()

	cancel context.CancelFunc
	done   chan struct{}
}

func (l *Store) updateSnapshot(fs fs.FS) error {
	storeSnapshot, err := snapshotFromFS(l.logger, fs)
	if err != nil {
		return err
	}

	l.mu.Lock()
	l.storeSnapshot = storeSnapshot
	l.mu.Unlock()

	// NOTE: this is really just a trick for unit tests
	// It is used to signal that an update occurred
	// so we dont have to e.g. sleep to know when
	// to check state.
	if l.notify != nil {
		l.notify()
	}

	return nil
}

// NewStore constructs and configure a Store.
// The store creates a background goroutine which feeds a channel of fs.FS.
func NewStore(logger *zap.Logger, source FSSource) (*Store, error) {
	store := &Store{
		syncedStore: &syncedStore{},
		logger:      logger,
		source:      source,
		done:        make(chan struct{}),
	}

	// get an initial FS from source.
	f, err := source.Get()
	if err != nil {
		return nil, err
	}

	if err := store.updateSnapshot(f); err != nil {
		return nil, err
	}

	var ctx context.Context
	ctx, store.cancel = context.WithCancel(context.Background())

	ch := make(chan fs.FS)
	go source.Subscribe(ctx, ch)

	go func() {
		defer close(store.done)
		for fs := range ch {
			logger.Debug("received new fs")

			if err = store.updateSnapshot(fs); err != nil {
				logger.Error("failed updating snapshot", zap.Error(err))
				continue
			}

			logger.Debug("updated latest fs")
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
