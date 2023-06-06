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
			getFS: mustSub(t, testdata, "fixtures/fswithindex"),
			ch:    make(chan fs.FS),
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
	getFS fs.FS
	ch    chan fs.FS
}

func (s source) String() string {
	return "test"
}

// Get builds a single instance of an fs.FS
func (s source) Get() (fs.FS, error) {
	return s.getFS, nil
}

// Subscribe feeds implementations of fs.FS onto the provided channel.
// It should block until the provided context is cancelled (it will be called in a goroutine).
// It should close the provided channel before it returns.
func (s source) Subscribe(ctx context.Context, ch chan<- fs.FS) {
	defer close(ch)

	for {
		select {
		case <-ctx.Done():
			return
		case fs := <-s.ch:
			ch <- fs
		}
	}
}

func mustSub(t *testing.T, f fs.FS, dir string) fs.FS {
	t.Helper()
	var err error
	f, err = fs.Sub(f, dir)
	require.NoError(t, err)
	return f
}
