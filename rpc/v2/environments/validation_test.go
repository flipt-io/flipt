package environments

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestBulkApplyResourcesRequestValidate(t *testing.T) {
	t.Run("allows request within target pair limit", func(t *testing.T) {
		req := &BulkApplyResourcesRequest{
			EnvironmentKeys: []string{"dev", "prod"},
			NamespaceKeys:   make([]string, 50),
		}

		require.NoError(t, req.Validate())
	})

	t.Run("rejects request above target pair limit", func(t *testing.T) {
		req := &BulkApplyResourcesRequest{
			EnvironmentKeys: []string{"dev", "prod"},
			NamespaceKeys:   make([]string, 51),
		}

		err := req.Validate()
		require.Error(t, err)
		require.True(t, strings.Contains(err.Error(), "bulk apply exceeds max target pairs"))
	})

	t.Run("uses singular environment_key when environment_keys is empty", func(t *testing.T) {
		req := &BulkApplyResourcesRequest{
			EnvironmentKey: "default",
			NamespaceKeys:  make([]string, maxBulkApplyTargetPairs),
		}

		require.NoError(t, req.Validate())
	})
}
