package testing

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.flipt.io/flipt/internal/storage/oplock"
	"golang.org/x/sync/errgroup"
)

// Harness is a test harness for all implementations of oplock.Service.
// The test consists of firing multiple goroutines which attempt to acquire
// a lock over a single operation "test". If acquired they increment a counter.
// After attempting to acquire the lock each goroutine sleeps until the lock
// can be attempted again.
func Harness(t *testing.T, s oplock.Service) {
	var (
		count       int64
		op          = oplock.Operation("test")
		ctx, cancel = context.WithCancel(context.Background())
	)

	errgroup, ctx := errgroup.WithContext(ctx)

	for i := 0; i < 5; i++ {
		errgroup.Go(func() error {
			for {
				select {
				case <-ctx.Done():
					return nil
				default:
				}

				acquired, entry, err := s.TryAcquire(ctx, op, 1*time.Second)
				require.NoError(t, err)

				if acquired {
					atomic.AddInt64(&count, 1)
				}

				time.Sleep(time.Until(entry.AcquiredUntil))
			}
		})
	}

	<-time.After(5 * time.Second)

	cancel()

	require.NoError(t, errgroup.Wait())

	// ensure counter only incremented 5 times in 5 seconds
	assert.Equal(t, int64(5), count)
}
