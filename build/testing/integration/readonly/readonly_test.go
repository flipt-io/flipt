package readonly_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.flipt.io/flipt/build/testing/integration"
	"go.flipt.io/flipt/rpc/flipt"
	"go.flipt.io/flipt/rpc/flipt/evaluation"
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
			// Ensure that invalid page token returns an appropriate response.
			_, err := sdk.Flipt().ListNamespaces(ctx, &flipt.ListNamespaceRequest{
				Limit:     10,
				PageToken: "Hello World",
			})
			require.EqualError(t, err, "rpc error: code = InvalidArgument desc = pageToken is not valid: \"Hello World\"")

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
			// Ensure that invalid page token returns an appropriate response.
			_, err := sdk.Flipt().ListFlags(ctx, &flipt.ListFlagRequest{
				NamespaceKey: namespace,
				Limit:        10,
				PageToken:    "Hello World",
			})
			require.EqualError(t, err, "rpc error: code = InvalidArgument desc = pageToken is not valid: \"Hello World\"")

			flags, err := sdk.Flipt().ListFlags(ctx, &flipt.ListFlagRequest{
				NamespaceKey: namespace,
			})
			require.NoError(t, err)
			require.Len(t, flags.Flags, 52)

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

					if flags.NextPageToken == "" {
						// ensure last page contains 2 entries (boolean and disabled)
						assert.Len(t, flags.Flags, 2)

						found = append(found, flags.Flags...)

						break
					}

					// ensure each full page is of length 10
					assert.Len(t, flags.Flags, 10)

					found = append(found, flags.Flags...)

					nextPage = flags.NextPageToken
				}

				require.Len(t, found, 52)
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
			_, err := sdk.Flipt().ListSegments(ctx, &flipt.ListSegmentRequest{
				NamespaceKey: namespace,
				Limit:        10,
				PageToken:    "Hello World",
			})
			require.EqualError(t, err, "rpc error: code = InvalidArgument desc = pageToken is not valid: \"Hello World\"")

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

					// ensure each page is of length 10
					assert.Len(t, segments.Segments, 10)

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
			_, err := sdk.Flipt().ListRules(ctx, &flipt.ListRuleRequest{
				NamespaceKey: namespace,
				FlagKey:      "flag_001",
				Limit:        10,
				PageToken:    "Hello World",
			})
			require.EqualError(t, err, "rpc error: code = InvalidArgument desc = pageToken is not valid: \"Hello World\"")

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

					// ensure each page is of length 10
					assert.Len(t, rules.Rules, 10)

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

		t.Run("ListRollouts", func(t *testing.T) {
			_, err := sdk.Flipt().ListRollouts(ctx, &flipt.ListRolloutRequest{
				NamespaceKey: namespace,
				FlagKey:      "flag_boolean",
				PageToken:    "Hello World",
			})
			require.EqualError(t, err, "rpc error: code = InvalidArgument desc = pageToken is not valid: \"Hello World\"")

			rules, err := sdk.Flipt().ListRollouts(ctx, &flipt.ListRolloutRequest{
				NamespaceKey: namespace,
				FlagKey:      "flag_boolean",
			})
			require.NoError(t, err)
			require.Len(t, rules.Rules, 5, "unexpected number of rules returned")

			rule := rules.Rules[0]
			assert.Equal(t, namespace, rule.NamespaceKey)
			assert.Equal(t, "flag_boolean", rule.FlagKey)
			assert.Equal(t, "segment_001", rule.GetSegment().SegmentKey)
			assert.True(t, rule.GetSegment().Value)
			assert.NotEmpty(t, rule.Id)
			assert.Equal(t, int32(1), rule.Rank)

			t.Run("Paginated (page size 2)", func(t *testing.T) {
				var (
					found    []*flipt.Rollout
					nextPage string
				)
				for {
					rules, err := sdk.Flipt().ListRollouts(ctx, &flipt.ListRolloutRequest{
						NamespaceKey: namespace,
						FlagKey:      "flag_boolean",
						Limit:        2,
						PageToken:    nextPage,
					})
					require.NoError(t, err)

					if rules.NextPageToken == "" {
						// ensure each page is of length 1
						assert.Len(t, rules.Rules, 1)
						found = append(found, rules.Rules...)
						break
					}

					// ensure each page is of length 2
					assert.Len(t, rules.Rules, 2)

					found = append(found, rules.Rules...)

					nextPage = rules.NextPageToken
				}

				require.Len(t, found, 5)
				assert.Equal(t, rules.Rules, found)
			})
		})

		t.Run("Legacy", func(t *testing.T) {
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
		})

		t.Run("Evaluation", func(t *testing.T) {
			t.Run("Variant", func(t *testing.T) {
				t.Run("match", func(t *testing.T) {
					response, err := sdk.Evaluation().Variant(ctx, &evaluation.EvaluationRequest{
						NamespaceKey: namespace,
						FlagKey:      "flag_001",
						EntityId:     "some-fixed-entity-id",
						Context: map[string]string{
							"in_segment": "segment_005",
						},
					})
					require.NoError(t, err)

					assert.Equal(t, true, response.Match)
					assert.Equal(t, "variant_002", response.VariantKey)
					assert.Equal(t, evaluation.EvaluationReason_MATCH_EVALUATION_REASON, response.Reason)
				})

				t.Run("no match", func(t *testing.T) {
					response, err := sdk.Evaluation().Variant(ctx, &evaluation.EvaluationRequest{
						NamespaceKey: namespace,
						FlagKey:      "flag_001",
						EntityId:     "some-fixed-entity-id",
						Context: map[string]string{
							"in_segment": "unknown",
						},
					})
					require.NoError(t, err)

					assert.Equal(t, false, response.Match)
					assert.Empty(t, response.VariantKey)
					assert.Equal(t, evaluation.EvaluationReason_UNKNOWN_EVALUATION_REASON, response.Reason)
				})

				t.Run("flag disabled", func(t *testing.T) {
					result, err := sdk.Evaluation().Variant(ctx, &evaluation.EvaluationRequest{
						NamespaceKey: namespace,
						FlagKey:      "flag_disabled",
						EntityId:     "some-fixed-entity-id",
						Context:      map[string]string{},
					})
					require.NoError(t, err)

					assert.False(t, result.Match, "Evaluation should not have matched.")
					assert.Equal(t, evaluation.EvaluationReason_FLAG_DISABLED_EVALUATION_REASON, result.Reason)
					assert.Empty(t, result.SegmentKey)
					assert.Empty(t, result.VariantKey)
				})

				t.Run("flag not found", func(t *testing.T) {
					result, err := sdk.Evaluation().Variant(ctx, &evaluation.EvaluationRequest{
						NamespaceKey: namespace,
						FlagKey:      "unknown_flag",
						EntityId:     "some-fixed-entity-id",
						Context:      map[string]string{},
					})
					msg := fmt.Sprintf("rpc error: code = NotFound desc = flag \"%s/unknown_flag\" not found", namespace)
					require.EqualError(t, err, msg)

					require.Nil(t, result)
				})
			})

			t.Run("Boolean", func(t *testing.T) {
				t.Run("default flag value", func(t *testing.T) {
					result, err := sdk.Evaluation().Boolean(ctx, &evaluation.EvaluationRequest{
						NamespaceKey: namespace,
						FlagKey:      "flag_boolean",
						EntityId:     "hello",
						Context:      map[string]string{},
					})

					require.NoError(t, err)

					assert.Equal(t, evaluation.EvaluationReason_DEFAULT_EVALUATION_REASON, result.Reason)
					assert.False(t, result.Enabled, "default flag value should be false")
				})

				t.Run("percentage match", func(t *testing.T) {
					result, err := sdk.Evaluation().Boolean(ctx, &evaluation.EvaluationRequest{
						NamespaceKey: namespace,
						FlagKey:      "flag_boolean",
						EntityId:     "fixed",
						Context:      map[string]string{},
					})

					require.NoError(t, err)

					assert.Equal(t, evaluation.EvaluationReason_MATCH_EVALUATION_REASON, result.Reason)
					assert.True(t, result.Enabled, "boolean evaluation value should be true")
				})

				t.Run("segment match", func(t *testing.T) {
					result, err := sdk.Evaluation().Boolean(ctx, &evaluation.EvaluationRequest{
						NamespaceKey: namespace,
						FlagKey:      "flag_boolean",
						EntityId:     "fixed",
						Context: map[string]string{
							"in_segment": "segment_001",
						},
					})

					require.NoError(t, err)

					assert.Equal(t, evaluation.EvaluationReason_MATCH_EVALUATION_REASON, result.Reason)
					assert.True(t, result.Enabled, "segment evaluation value should be true")
				})
			})

			t.Run("Batch", func(t *testing.T) {
				t.Run("batch evaluations (with not found errors)", func(t *testing.T) {
					result, err := sdk.Evaluation().Batch(ctx, &evaluation.BatchEvaluationRequest{
						Requests: []*evaluation.EvaluationRequest{
							{
								NamespaceKey: namespace,
								FlagKey:      "flag_boolean",
								EntityId:     "fixed",
								Context:      map[string]string{},
							},
							{
								NamespaceKey: namespace,
								FlagKey:      "foobarnotfound",
							},
							{
								NamespaceKey: namespace,
								FlagKey:      "flag_001",
								EntityId:     "some-fixed-entity-id",
								Context: map[string]string{
									"in_segment": "segment_005",
								},
							},
						},
					})

					require.NoError(t, err)

					assert.Len(t, result.Responses, 3)

					b, ok := result.Responses[0].Response.(*evaluation.EvaluationResponse_BooleanResponse)
					assert.True(t, ok, "response should be boolean evaluation response")
					assert.True(t, b.BooleanResponse.Enabled, "boolean response should have true value")
					assert.Equal(t, evaluation.EvaluationReason_MATCH_EVALUATION_REASON, b.BooleanResponse.Reason)
					assert.Equal(t, evaluation.EvaluationResponseType_BOOLEAN_EVALUATION_RESPONSE_TYPE, result.Responses[0].Type)

					e, ok := result.Responses[1].Response.(*evaluation.EvaluationResponse_ErrorResponse)
					assert.True(t, ok, "response should be error evaluation response")
					assert.Equal(t, "foobarnotfound", e.ErrorResponse.FlagKey)
					assert.Equal(t, evaluation.ErrorEvaluationReason_NOT_FOUND_ERROR_EVALUATION_REASON, e.ErrorResponse.Reason)
					assert.Equal(t, evaluation.EvaluationResponseType_ERROR_EVALUATION_RESPONSE_TYPE, result.Responses[1].Type)

					v, ok := result.Responses[2].Response.(*evaluation.EvaluationResponse_VariantResponse)
					assert.True(t, ok, "response should be boolean evaluation response")
					assert.True(t, v.VariantResponse.Match, "variant response match should have true value")
					assert.Equal(t, evaluation.EvaluationReason_MATCH_EVALUATION_REASON, v.VariantResponse.Reason)
					assert.Equal(t, "variant_002", v.VariantResponse.VariantKey)
					assert.Equal(t, "segment_005", v.VariantResponse.SegmentKey)
					assert.Equal(t, evaluation.EvaluationResponseType_VARIANT_EVALUATION_RESPONSE_TYPE, result.Responses[2].Type)
				})
			})
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
