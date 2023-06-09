package readonly_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.flipt.io/flipt/build/testing/integration"
	"go.flipt.io/flipt/rpc/flipt"
	sdk "go.flipt.io/flipt/sdk/go"
)

// TestReadOnly is a suite of tests which presumes all the data found in the local testdata
// folder has been loaded into the target instance being tested.
// It then exercises a bunch of read operations via the provided SDK in the target namespace.
func TestReadOnly(t *testing.T) {
	integration.Harness(t, func(t *testing.T, sdk sdk.SDK, namespace string, authenticated bool) {
		ctx := context.Background()
		ns, err := sdk.Flipt().GetNamespace(ctx, &flipt.GetNamespaceRequest{
			Key: namespace,
		})
		require.NoError(t, err)

		assert.Equal(t, namespace, ns.Key)

		expected := "Default"
		if namespace != "" && namespace != "default" {
			expected = namespace
		}
		assert.Equal(t, expected, ns.Name)
		assert.NotZero(t, ns.CreatedAt)
		assert.NotZero(t, ns.UpdatedAt)

		t.Run("ListNamespaces", func(t *testing.T) {
			// NOTE: different cases load different amount of Namespaces
			// so we're just interested in the shape of the response as
			// opposed to the specific contents of namespaces
			ns, err := sdk.Flipt().ListNamespaces(ctx, &flipt.ListNamespaceRequest{})
			require.NoError(t, err)

			// we at-least want to ensure that the namespace used in
			// the context of this test invocation is present in the
			// list response
			var foundNamespaceInContext bool

			assert.NotZero(t, ns.Namespaces)
			for _, n := range ns.Namespaces {
				assert.NotZero(t, n.CreatedAt)
				assert.NotZero(t, n.UpdatedAt)

				foundNamespaceInContext = foundNamespaceInContext || n.Key == namespace
			}

			require.True(t, foundNamespaceInContext, "%q was not found in list response", namespace)
		})

		t.Run("GetFlag", func(t *testing.T) {
			_, err := sdk.Flipt().GetFlag(ctx, &flipt.GetFlagRequest{
				NamespaceKey: namespace,
				Key:          "flag_999",
			})
			require.Error(t, err, "not found")

			flag, err := sdk.Flipt().GetFlag(ctx, &flipt.GetFlagRequest{
				NamespaceKey: namespace,
				Key:          "flag_013",
			})
			require.NoError(t, err)

			assert.Equal(t, namespace, flag.NamespaceKey)
			assert.Equal(t, "flag_013", flag.Key)
			assert.Equal(t, "FLAG_013", flag.Name)
			assert.Equal(t, "Some Description", flag.Description)
			assert.True(t, flag.Enabled)

			require.Len(t, flag.Variants, 2)
			assert.Equal(t, "variant_001", flag.Variants[0].Key)
			assert.Equal(t, "variant_002", flag.Variants[1].Key)
		})

		t.Run("ListFlags", func(t *testing.T) {
			flags, err := sdk.Flipt().ListFlags(ctx, &flipt.ListFlagRequest{
				NamespaceKey: namespace,
			})
			require.NoError(t, err)
			require.Len(t, flags.Flags, 50)

			flag := flags.Flags[0]
			assert.Equal(t, namespace, flag.NamespaceKey)
			assert.Equal(t, "flag_001", flag.Key)
			assert.Equal(t, "FLAG_001", flag.Name)
			assert.Equal(t, "Some Description", flag.Description)
			assert.True(t, flag.Enabled)

			require.Len(t, flag.Variants, 2)
			assert.Equal(t, "variant_001", flag.Variants[0].Key)
			assert.Equal(t, "variant_002", flag.Variants[1].Key)

			t.Run("Paginated (page size 10)", func(t *testing.T) {
				var (
					found    []*flipt.Flag
					nextPage string
				)
				for {
					flags, err := sdk.Flipt().ListFlags(ctx, &flipt.ListFlagRequest{
						NamespaceKey: namespace,
						Limit:        10,
						PageToken:    nextPage,
					})
					require.NoError(t, err)

					found = append(found, flags.Flags...)

					if flags.NextPageToken == "" {
						break
					}

					nextPage = flags.NextPageToken
				}

				require.Len(t, found, 50)
			})
		})

		t.Run("GetSegment", func(t *testing.T) {
			_, err := sdk.Flipt().GetSegment(ctx, &flipt.GetSegmentRequest{
				NamespaceKey: namespace,
				Key:          "segment_999",
			})
			require.Error(t, err, "not found")

			segment, err := sdk.Flipt().GetSegment(ctx, &flipt.GetSegmentRequest{
				NamespaceKey: namespace,
				Key:          "segment_013",
			})
			require.NoError(t, err)

			assert.Equal(t, namespace, segment.NamespaceKey)
			assert.Equal(t, "segment_013", segment.Key)
			assert.Equal(t, "SEGMENT_013", segment.Name)
			assert.Equal(t, "Some Segment Description", segment.Description)

			assert.Len(t, segment.Constraints, 2)
			for _, constraint := range segment.Constraints {
				assert.NotEmpty(t, constraint.Id)
				assert.Equal(t, flipt.ComparisonType_STRING_COMPARISON_TYPE, constraint.Type)
				assert.Equal(t, "in_segment", constraint.Property)
				assert.Equal(t, "eq", constraint.Operator)
				assert.Equal(t, "segment_013", constraint.Value)
			}
		})

		t.Run("ListSegments", func(t *testing.T) {
			segments, err := sdk.Flipt().ListSegments(ctx, &flipt.ListSegmentRequest{
				NamespaceKey: namespace,
			})

			require.NoError(t, err)
			require.Len(t, segments.Segments, 50)

			t.Run("Paginated (page size 10)", func(t *testing.T) {
				var (
					found    []*flipt.Segment
					nextPage string
				)
				for {
					segments, err := sdk.Flipt().ListSegments(ctx, &flipt.ListSegmentRequest{
						NamespaceKey: namespace,
						Limit:        10,
						PageToken:    nextPage,
					})
					require.NoError(t, err)

					found = append(found, segments.Segments...)

					if segments.NextPageToken == "" {
						break
					}

					nextPage = segments.NextPageToken
				}

				require.Len(t, found, 50)
			})
		})

		t.Run("GetRule", func(t *testing.T) {
			_, err := sdk.Flipt().GetRule(ctx, &flipt.GetRuleRequest{
				NamespaceKey: namespace,
				Id:           "notfound",
			})
			require.Error(t, err, "not found")

			rules, err := sdk.Flipt().ListRules(ctx, &flipt.ListRuleRequest{
				NamespaceKey: namespace,
				FlagKey:      "flag_001",
				Limit:        1,
			})
			require.NoError(t, err)

			rule, err := sdk.Flipt().GetRule(ctx, &flipt.GetRuleRequest{
				NamespaceKey: namespace,
				FlagKey:      "flag_001",
				Id:           rules.Rules[0].Id,
			})
			require.NoError(t, err)

			assert.Equal(t, namespace, rule.NamespaceKey)
			assert.Equal(t, "flag_001", rule.FlagKey)
			assert.Equal(t, "segment_001", rule.SegmentKey)
			assert.NotEmpty(t, rule.Id)
			assert.Equal(t, int32(1), rule.Rank)

			require.Len(t, rule.Distributions, 2)

			assert.NotEmpty(t, rule.Distributions[0].Id)
			assert.Equal(t, float32(50.0), rule.Distributions[0].Rollout)
			assert.NotEmpty(t, rule.Distributions[1].Id)
			assert.Equal(t, float32(50.0), rule.Distributions[1].Rollout)
		})

		t.Run("ListRules", func(t *testing.T) {
			rules, err := sdk.Flipt().ListRules(ctx, &flipt.ListRuleRequest{
				NamespaceKey: namespace,
				FlagKey:      "flag_001",
			})
			require.NoError(t, err)
			require.Len(t, rules.Rules, 50, "unexpected number of rules returned")

			rule := rules.Rules[0]
			assert.Equal(t, namespace, rule.NamespaceKey)
			assert.Equal(t, "flag_001", rule.FlagKey)
			assert.Equal(t, "segment_001", rule.SegmentKey)
			assert.NotEmpty(t, rule.Id)
			assert.Equal(t, int32(1), rule.Rank)

			require.Len(t, rule.Distributions, 2)

			assert.NotEmpty(t, rule.Distributions[0].Id)
			assert.Equal(t, float32(50.0), rule.Distributions[0].Rollout)
			assert.NotEmpty(t, rule.Distributions[1].Id)
			assert.Equal(t, float32(50.0), rule.Distributions[1].Rollout)

			t.Run("Paginated (page size 10)", func(t *testing.T) {
				var (
					found    []*flipt.Rule
					nextPage string
				)
				for {
					rules, err := sdk.Flipt().ListRules(ctx, &flipt.ListRuleRequest{
						NamespaceKey: namespace,
						FlagKey:      "flag_001",
						Limit:        10,
						PageToken:    nextPage,
					})
					require.NoError(t, err)

					found = append(found, rules.Rules...)

					if rules.NextPageToken == "" {
						break
					}

					nextPage = rules.NextPageToken
				}

				require.Len(t, found, 50)
				assert.Equal(t, rules.Rules, found)
			})
		})

		t.Run("Evaluate", func(t *testing.T) {
			response, err := sdk.Flipt().Evaluate(ctx, &flipt.EvaluationRequest{
				NamespaceKey: namespace,
				FlagKey:      "flag_001",
				EntityId:     "some-fixed-entity-id",
				Context: map[string]string{
					"in_segment": "segment_005",
				},
			})
			require.NoError(t, err)

			assert.Equal(t, true, response.Match)
			assert.Equal(t, "variant_002", response.Value)
			assert.Equal(t, flipt.EvaluationReason_MATCH_EVALUATION_REASON, response.Reason)
		})

		t.Run("BatchEvaluate", func(t *testing.T) {
			response, err := sdk.Flipt().BatchEvaluate(ctx, &flipt.BatchEvaluationRequest{
				NamespaceKey: namespace,
				Requests: []*flipt.EvaluationRequest{
					{
						FlagKey:  "flag_001",
						EntityId: "some-fixed-entity-id",
						Context: map[string]string{
							"in_segment": "segment_005",
						},
					},
					{
						FlagKey:  "flag_002",
						EntityId: "some-fixed-entity-id",
						Context: map[string]string{
							"in_segment": "segment_006",
						},
					},
				},
			})
			require.NoError(t, err)
			require.Len(t, response.Responses, 2)

			assert.Equal(t, true, response.Responses[0].Match)
			assert.Equal(t, "variant_002", response.Responses[0].Value)
			assert.Equal(t, flipt.EvaluationReason_MATCH_EVALUATION_REASON, response.Responses[0].Reason)

			assert.Equal(t, true, response.Responses[1].Match)
			assert.Equal(t, "variant_001", response.Responses[1].Value)
			assert.Equal(t, flipt.EvaluationReason_MATCH_EVALUATION_REASON, response.Responses[1].Reason)
		})

		t.Run("Auth", func(t *testing.T) {
			t.Run("Self", func(t *testing.T) {
				_, err := sdk.Auth().AuthenticationService().GetAuthenticationSelf(ctx)
				if authenticated {
					assert.NoError(t, err)
				} else {
					assert.EqualError(t, err, "rpc error: code = Unauthenticated desc = request was not authenticated")
				}
			})
			t.Run("Public", func(t *testing.T) {
				_, err := sdk.Auth().PublicAuthenticationService().ListAuthenticationMethods(ctx)
				require.NoError(t, err)
			})
		})
	})
}
