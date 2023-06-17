package fs

import (
	"context"
	"fmt"
	"path"

	"go.flipt.io/flipt/internal/storage"
	"go.flipt.io/flipt/rpc/flipt"
	"go.uber.org/zap"
)

// Source produces Snapshot types.
// A single Snapshot can be produced via Get or a channel
// may be provided to Subscribe in order to received
// new instances when new state becomes available.
type Source interface {
	fmt.Stringer

	// Get builds a single instance of an Snapshot
	Get() (*Snapshot, error)

	// Subscribe feeds implementations of Snapshot onto the provided channel.
	// It should block until the provided context is cancelled (it will be called in a goroutine).
	// It should close the provided channel before it returns.
	Subscribe(context.Context, chan<- *Snapshot)
}

// Store is an implementation of storage.Store backed by an Source.
// The store subscribes to the source for instances of Snapshot with new contents.
// When a new fs is received the contents is fetched and built into a snapshot
// of Flipt feature flag state.
type Store struct {
	*syncedStore

	logger *zap.Logger
	source Source

	// notify is used for test purposes
	// it is invoked if defined when a snapshot update finishes
	notify func()

	cancel context.CancelFunc
	done   chan struct{}
}

func (l *Store) updateSnapshot(s *Snapshot) error {
	l.mu.Lock()
	l.Snapshot = s
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
// The store creates a background goroutine which feeds a channel of *Snapshot.
func NewStore(logger *zap.Logger, source Source) (*Store, error) {
	store := &Store{
		syncedStore: &syncedStore{},
		logger:      logger,
		source:      source,
		done:        make(chan struct{}),
	}

	// get an initial FS from source.
	s, err := source.Get()
	if err != nil {
		return nil, err
	}

	if err := store.updateSnapshot(s); err != nil {
		return nil, err
	}

	var ctx context.Context
	ctx, store.cancel = context.WithCancel(context.Background())

	ch := make(chan *Snapshot)
	go source.Subscribe(ctx, ch)

	go func() {
		defer close(store.done)
		for s := range ch {
			logger.Debug("received new fs")

			if err = store.updateSnapshot(s); err != nil {
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

type ProposableSource interface {
	Source

	storage.ProposalStore
}

func (s *Store) Propose(ctx context.Context, r *flipt.ProposeRequest) (*flipt.Proposal, error) {
	if ps, ok := s.source.(ProposableSource); ok {
		return ps.Propose(ctx, r)
	}

	return nil, ErrNotImplemented
}
