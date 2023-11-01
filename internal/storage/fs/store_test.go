package fs

import (
	"context"
	"io/fs"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"go.uber.org/zap/zaptest"
)

func Test_Store(t *testing.T) {
	var (
		logger = zaptest.NewLogger(t)
		notify = make(chan struct{})
		source = source{
			get: mustSub(t, testdata, "fixtures/fswithindex"),
			ch:  make(chan *StoreSnapshot),
		}
	)

	store, err := NewStore(logger, source)
	require.NoError(t, err)

	// register a function to be called when updates have
	// finished
	store.notify = func() {
		notify <- struct{}{}
	}

	assert.Equal(t, "filesystem/test", store.String())

	// run FS with index suite against current store
	suite.Run(t, &FSIndexSuite{store: store})

	// update snapshot by sending fs without index
	source.ch <- mustSub(t, testdata, "fixtures/fswithoutindex")

	// wait for update to apply
	<-notify

	// run FS without index suite against current store
	suite.Run(t, &FSWithoutIndexSuite{store: store})

	// shutdown store
	require.NoError(t, store.Close())
}

type source struct {
	get *StoreSnapshot
	ch  chan *StoreSnapshot
}

func (s source) String() string {
	return "test"
}

// Get builds a single instance of an *StoreSnapshot
func (s source) Get() (*StoreSnapshot, error) {
	return s.get, nil
}

// Subscribe feeds implementations of *StoreSnapshot onto the provided channel.
// It should block until the provided context is cancelled (it will be called in a goroutine).
// It should close the provided channel before it returns.
func (s source) Subscribe(ctx context.Context, ch chan<- *StoreSnapshot) {
	defer close(ch)

	for {
		select {
		case <-ctx.Done():
			return
		case snap := <-s.ch:
			ch <- snap
		}
	}
}

func mustSub(t *testing.T, f fs.FS, dir string) *StoreSnapshot {
	t.Helper()
	var err error
	f, err = fs.Sub(f, dir)
	require.NoError(t, err)

	snap, err := SnapshotFromFS(zaptest.NewLogger(t), f)
	require.NoError(t, err)
	return snap
}
