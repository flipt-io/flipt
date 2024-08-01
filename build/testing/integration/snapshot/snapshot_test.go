package snapshot_test

import (
	"context"
	"testing"

	"go.flipt.io/build/testing/integration"
	"go.flipt.io/build/testing/integration/snapshot"
)

func TestSnapshot(t *testing.T) {
	integration.Harness(t, func(t *testing.T, opts integration.TestOpts) {
		ctx := context.Background()

		snapshot.Snapshot(t, ctx, opts)
	})
}
