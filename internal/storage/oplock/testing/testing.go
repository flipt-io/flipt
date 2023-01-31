package testing

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	errs "go.flipt.io/flipt/errors"
	"go.flipt.io/flipt/internal/storage/oplock"
	"golang.org/x/sync/errgroup"
)

// Harness is a test harness for all implementations of oplock.Service.
// The test consists of firing multiple goroutines which attempt to acquire
// a lock over a single operation "test".
// Each acquisitions timestamp is pushed down a channel.
// When five lock acquisitions have occurred the test ensures that it took
// at-least a specified duration to do so (interval * (iterations - 1)).
// Also that acquisitions occurred in ascending timestamp order with a delta
// between each tick of at-least the configured interval.
func Harness(t *testing.T, s oplock.Service) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}

	var (
		acquiredAt  = make(chan time.Time, 1)
		interval    = 2 * time.Second
		op          = oplock.Operation("test")
		ctx, cancel = context.WithCancel(context.Background())
	)

	errgroup, ctx := errgroup.WithContext(ctx)

	for i := 0; i < 5; i++ {
		var acquiredUntil = time.Now().UTC()

		errgroup.Go(func() error {
			for {
				select {
				case <-ctx.Done():
					return nil
				case <-time.After(time.Until(acquiredUntil)):
				}

				acquired, entry, err := s.TryAcquire(ctx, op, interval)
				if err != nil {
					return err
				}

				if acquired {
					acquiredAt <- entry.LastAcquired
				}

				acquiredUntil = entry.AcquiredUntil
			}
		})
	}

	now := time.Now().UTC()
	var acquisitions []time.Time
	for tick := range acquiredAt {
		acquisitions = append(acquisitions, tick)

		if len(acquisitions) == 5 {
			break
		}
	}

	since := time.Since(now)

	// ensure it took at-least 8s second to acquire 5 locks
	require.Greater(t, since, 8*time.Second)

	t.Logf("It took %s to consume the lock 5 times with an interval of %s\n", since, interval)

	cancel()

	if err := errgroup.Wait(); err != nil {
		// there are a few acceptable errors here (e.g. context.Canceled / "operation was canceled")
		// stdlib net package can adapt context.Canceled into an unexported errCanceled
		// https://github.com/golang/go/blob/6b45863e47ad1a27ba3051ce0407f0bdc7b46113/src/net/net.go#L428-L439
		// Postgres reports this as a "query_canceled" error code name.
		_, isCanceled := errs.As[errs.ErrCanceled](err)
		switch {
		case isCanceled:
		case errors.Is(err, context.Canceled):
		case strings.Contains(err.Error(), "operation was canceled"):
		default:
			require.FailNowf(t, "error not as expected", "%v", err)
		}
	}

	close(acquiredAt)

	// ensure ticks were acquired sequentially
	assert.IsIncreasing(t, acquisitions)

	for i, tick := range acquisitions {
		if len(acquisitions) == i+1 {
			break
		}

		// tick at T(n+1) occurs at-least <interval> after T(n)
		assert.Greater(t, acquisitions[i+1].Sub(tick), interval)
	}
}
