package evaluation_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.flipt.io/flipt/build/testing/integration"
	"go.flipt.io/flipt/rpc/flipt/evaluation"
	sdk "go.flipt.io/flipt/sdk/go"
)

func TestEvaluations_All_Match_Type(t *testing.T) {
	integration.Harness(t, func(t *testing.T, sdk sdk.SDK, opts integration.TestOpts) {
		var ctx = context.Background()
		t.Run("ALL_MATCH_TYPE", func(t *testing.T) {
			t.Run("no variants no distributions", func(t *testing.T) {
				variant, err := sdk.Evaluation().Variant(ctx, &evaluation.EvaluationRequest{
					NamespaceKey: "default",
					FlagKey:      "no_distributions_all",
					EntityId:     "1",
					Context: map[string]string{
						"segment_all_a": "segment_all_a",
					},
				})
				require.NoError(t, err)

				assert.True(t, variant.Match)
				assert.Contains(t, variant.SegmentKeys, "segment_all_a")
				assert.Equal(t, evaluation.EvaluationReason_MATCH_EVALUATION_REASON, variant.Reason)
			})

			t.Run("multiple segments matched", func(t *testing.T) {
				variant, err := sdk.Evaluation().Variant(ctx, &evaluation.EvaluationRequest{
					NamespaceKey: "default",
					FlagKey:      "multiple_segments_matched_all",
					EntityId:     "1",
					Context: map[string]string{
						"segment_all_a": "segment_all_a",
						"segment_all_b": "segment_all_b",
					},
				})
				require.NoError(t, err)

				assert.True(t, variant.Match)
				assert.Equal(t, evaluation.EvaluationReason_MATCH_EVALUATION_REASON, variant.Reason)
				assert.Contains(t, variant.SegmentKeys, "segment_all_a")
				assert.Contains(t, variant.SegmentKeys, "segment_all_b")
			})
		})
	})
}
