package ofrep_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/open-feature/go-sdk-contrib/providers/ofrep"
	"github.com/open-feature/go-sdk/openfeature"
	"github.com/stretchr/testify/assert"
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

		t.Run("OFREP ", func(t *testing.T) {
			for _, namespace := range integration.Namespaces {
				t.Run(fmt.Sprintf("namespace %q", namespace.Expected), func(t *testing.T) {

					t.Logf("Boolean evaluation.")

					provider := ofrep.NewProvider(opts.URL.String(), ofrep.WithBearerToken(opts.Token), ofrep.WithHeaderProvider(func() (string, string) {
						return "X-Flipt-Namespace", namespace.Key
					}))

					respBool := provider.BooleanEvaluation(ctx, "flag_boolean", false, openfeature.FlattenedContext{
						"in_segment": "segment_001",
					})

					require.NotNil(t, respBool)
					assert.True(t, respBool.Value)
					assert.Empty(t, respBool.Error())
					assert.Equal(t, openfeature.TargetingMatchReason, respBool.Reason)

					t.Logf("Boolean evaluation, flag not found.")

					respBool = provider.BooleanEvaluation(ctx, "idontexist", false, openfeature.FlattenedContext{
						"in_segment": "segment_001",
					})

					require.NotNil(t, respBool)
					assert.False(t, respBool.Value)
					assert.Equal(t, "FLAG_NOT_FOUND: flag for key 'idontexist' does not exist", respBool.Error().Error())

					t.Logf("String evaluation.")

					respStr := provider.StringEvaluation(ctx, "flag_001", "default", openfeature.FlattenedContext{
						"in_segment":   "segment_005",
						"targetingKey": "some-fixed-entity-id",
					})

					require.NotNil(t, respStr)
					assert.Equal(t, "variant_002", respStr.Value)
					assert.Equal(t, "variant_002", respStr.Variant)
					assert.Empty(t, respStr.Error())
					assert.Equal(t, openfeature.TargetingMatchReason, respStr.Reason)

					t.Logf("String evaluation, flag not found.")

					respStr = provider.StringEvaluation(ctx, "idontexist", "default", openfeature.FlattenedContext{
						"in_segment":   "segment_005",
						"targetingKey": "some-fixed-entity-id",
					})

					require.NotNil(t, respStr)
					assert.Equal(t, "default", respStr.Value)
					assert.Equal(t, "FLAG_NOT_FOUND: flag for key 'idontexist' does not exist", respBool.Error().Error())
				})
			}
		})
	})
}
