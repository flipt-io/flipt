package ofrep_test

import (
	"context"
	"testing"

	"github.com/open-feature/go-sdk-contrib/providers/ofrep"
	"github.com/open-feature/go-sdk/openfeature"
	"github.com/stretchr/testify/require"
	"go.flipt.io/build/testing/integration"
)

// OFREP tests the OFREP API.
func TestOFREP(t *testing.T) {
	integration.Harness(t, func(t *testing.T, opts integration.TestOpts) {
		var (
			ctx      = context.Background()
			protocol = opts.Protocol()
		)

		if protocol == integration.ProtocolGRPC {
			t.Skip("REST tests are not applicable for gRPC")
		}

		t.Run("OFREP", func(t *testing.T) {
			t.Logf("Boolean evaluation.")

			provider := ofrep.NewProvider(opts.URL.String(), ofrep.WithBearerToken(opts.Token), ofrep.WithHeaderProvider(oubound.HeaderCallback(func() (string, string) {
				return "X-Flipt-Namespace", "default"
			})))

			resp := provider.BooleanEvaluation(ctx, "flag_boolean", false, openfeature.FlattenedContext{
				"in_segment": "segment_001",
			})

			require.NotNil(t, resp)
		})
	})
}
