package git

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCategorizeError(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected string
	}{
		{
			name:     "unknown error",
			err:      assert.AnError,
			expected: "unknown",
		},
		{
			name:     "nil error",
			err:      nil,
			expected: "unknown",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := categorizeError(tt.err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestRepoMetrics(t *testing.T) {
	metrics := newRepoMetrics(withRemote("test-remote"))
	ctx := context.Background()

	t.Run("recordSyncStart with success", func(t *testing.T) {
		finished := metrics.recordSyncStart(ctx, "main")
		// Should not panic when called with nil error (success)
		finished(nil)
	})

	t.Run("recordSyncStart with error", func(t *testing.T) {
		finished := metrics.recordSyncStart(ctx, "main")
		// Should not panic when called with error
		finished(assert.AnError)
	})

	t.Run("recordSyncError", func(t *testing.T) {
		// Should not panic
		metrics.recordSyncError(ctx, "main", "network")
	})

	t.Run("recordFilesChanged", func(t *testing.T) {
		// Should not panic
		metrics.recordFilesChanged(ctx, "main", "added", 5)
		metrics.recordFilesChanged(ctx, "main", "modified", 3)
		metrics.recordFilesChanged(ctx, "main", "deleted", 1)
	})
}

func TestRepoMetricsAttributes(t *testing.T) {
	metrics := newRepoMetrics(withRemote("github.com/test/repo"))

	// Test that metrics can be created with remote attribute
	assert.NotNil(t, metrics.syncTotal)
	assert.NotNil(t, metrics.syncErrorsTotal)
	assert.NotNil(t, metrics.syncDuration)
	assert.NotNil(t, metrics.filesChanged)
}
