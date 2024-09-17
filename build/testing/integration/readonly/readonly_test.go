package readonly_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.flipt.io/build/testing/integration"
	"go.flipt.io/flipt/rpc/flipt"
	"go.flipt.io/flipt/rpc/flipt/evaluation"
)

// TestReadOnly is a suite of tests which presumes all the data found in the local testdata
// folder has been loaded into the target instance being tested.
// It then exercises a bunch of read operations via the provided SDK in the target namespace.
func TestReadOnly(t *testing.T) {
	integration.Harness(t, func(t *testing.T, opts integration.TestOpts) {
		var (
			ctx = context.Background()
			sdk = opts.TokenClient(t)
		)

		ns, err := sdk.Flipt().GetNamespace(ctx, &flipt.GetNamespaceRequest{
			Key: integration.DefaultNamespace,
		})
		require.NoError(t, err)

		assert.Equal(t, integration.DefaultNamespace, ns.Key)

		assert.Equal(t, "Default", ns.Name)
		assert.NotZero(t, ns.CreatedAt)
		assert.NotZero(t, ns.UpdatedAt)

		t.Run("ListNamespaces", func(t *testing.T) {
			// Ensure that invalid page token returns an appropriate response.
			_, err := sdk.Flipt().ListNamespaces(ctx, &flipt.ListNamespaceRequest{
				Limit:     10,
				PageToken: "Hello World",
			})
			require.EqualError(t, err, "rpc error: code = InvalidArgument desc = pageToken is not valid: \"Hello World\"")

			ns, err := sdk.Flipt().ListNamespaces(ctx, &flipt.ListNamespaceRequest{})
			require.NoError(t, err)

			assert.Len(t, ns.Namespaces, 2)
		})

		for _, namespace := range integration.Namespaces {
			t.Run(fmt.Sprintf("namespace %q", namespace.Key), func(t *testing.T) {
				t.Run("GetFlag", func(t *testing.T) {
					_, err := sdk.Flipt().GetFlag(ctx, &flipt.GetFlagRequest{
						NamespaceKey: namespace.Key,
						Key:          "flag_999",
					})
					require.Error(t, err, "not found")

					flag, err := sdk.Flipt().GetFlag(ctx, &flipt.GetFlagRequest{
						NamespaceKey: namespace.Key,
						Key:          "flag_013",
					})
					require.NoError(t, err)

					assert.Equal(t, namespace.Expected, flag.NamespaceKey)
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
						NamespaceKey: namespace.Key,
						Limit:        10,
						PageToken:    "Hello World",
					})
					require.EqualError(t, err, "rpc error: code = InvalidArgument desc = pageToken is not valid: \"Hello World\"")

					flags, err := sdk.Flipt().ListFlags(ctx, &flipt.ListFlagRequest{
						NamespaceKey: namespace.Key,
					})
					require.NoError(t, err)
					require.Len(t, flags.Flags, 57)

					flag := flags.Flags[0]
					assert.Equal(t, namespace.Expected, flag.NamespaceKey)
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
								NamespaceKey: namespace.Key,
								Limit:        10,
								PageToken:    nextPage,
							})
							require.NoError(t, err)

							if flags.NextPageToken == "" {
								// ensure last page contains 3 entries (boolean and disabled)
								assert.Len(t, flags.Flags, 7)

								found = append(found, flags.Flags...)

								break
							}

							// ensure each full page is of length 10
							assert.Len(t, flags.Flags, 10)

							found = append(found, flags.Flags...)

							nextPage = flags.NextPageToken
						}

						require.Len(t, found, 57)
					})
				})

				t.Run("GetSegment", func(t *testing.T) {
					_, err := sdk.Flipt().GetSegment(ctx, &flipt.GetSegmentRequest{
						NamespaceKey: namespace.Key,
						Key:          "segment_999",
					})
					require.Error(t, err, "not found")

					segment, err := sdk.Flipt().GetSegment(ctx, &flipt.GetSegmentRequest{
						NamespaceKey: namespace.Key,
						Key:          "segment_013",
					})
					require.NoError(t, err)

					assert.Equal(t, namespace.Expected, segment.NamespaceKey)
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
						NamespaceKey: namespace.Key,
						Limit:        10,
						PageToken:    "Hello World",
					})
					require.EqualError(t, err, "rpc error: code = InvalidArgument desc = pageToken is not valid: \"Hello World\"")

					segments, err := sdk.Flipt().ListSegments(ctx, &flipt.ListSegmentRequest{
						NamespaceKey: namespace.Key,
					})

					require.NoError(t, err)
					require.Len(t, segments.Segments, 53)

					t.Run("Paginated (page size 10)", func(t *testing.T) {
						var (
							found    []*flipt.Segment
							nextPage string
						)
						for {
							segments, err := sdk.Flipt().ListSegments(ctx, &flipt.ListSegmentRequest{
								NamespaceKey: namespace.Key,
								Limit:        10,
								PageToken:    nextPage,
							})
							require.NoError(t, err)

							// ensure each page is of length 10

							found = append(found, segments.Segments...)

							if segments.NextPageToken == "" {
								assert.Len(t, segments.Segments, 3)
								break
							}

							assert.Len(t, segments.Segments, 10)

							nextPage = segments.NextPageToken
						}

						require.Len(t, found, 53)
					})
				})

				t.Run("GetRule", func(t *testing.T) {
					_, err := sdk.Flipt().GetRule(ctx, &flipt.GetRuleRequest{
						NamespaceKey: namespace.Key,
						Id:           "notfound",
					})
					require.Error(t, err, "not found")

					rules, err := sdk.Flipt().ListRules(ctx, &flipt.ListRuleRequest{
						NamespaceKey: namespace.Key,
						FlagKey:      "flag_001",
						Limit:        1,
					})
					require.NoError(t, err)

					rule, err := sdk.Flipt().GetRule(ctx, &flipt.GetRuleRequest{
						NamespaceKey: namespace.Key,
						FlagKey:      "flag_001",
						Id:           rules.Rules[0].Id,
					})
					require.NoError(t, err)

					assert.Equal(t, namespace.Expected, rule.NamespaceKey)
					assert.Equal(t, "flag_001", rule.FlagKey)
					assert.Equal(t, "segment_001", rule.SegmentKey)
					assert.NotEmpty(t, rule.Id)
					assert.Equal(t, int32(1), rule.Rank)

					require.Len(t, rule.Distributions, 2)

					assert.NotEmpty(t, rule.Distributions[0].Id)
					assert.InDelta(t, 50.0, rule.Distributions[0].Rollout, 0)
					assert.NotEmpty(t, rule.Distributions[1].Id)
					assert.InDelta(t, 50.0, rule.Distributions[1].Rollout, 0)
				})

				t.Run("ListRules", func(t *testing.T) {
					_, err := sdk.Flipt().ListRules(ctx, &flipt.ListRuleRequest{
						NamespaceKey: namespace.Key,
						FlagKey:      "flag_001",
						Limit:        10,
						PageToken:    "Hello World",
					})
					require.EqualError(t, err, "rpc error: code = InvalidArgument desc = pageToken is not valid: \"Hello World\"")

					rules, err := sdk.Flipt().ListRules(ctx, &flipt.ListRuleRequest{
						NamespaceKey: namespace.Key,
						FlagKey:      "flag_001",
					})
					require.NoError(t, err)
					require.Len(t, rules.Rules, 50, "unexpected number of rules returned")

					rule := rules.Rules[0]
					assert.Equal(t, namespace.Expected, rule.NamespaceKey)
					assert.Equal(t, "flag_001", rule.FlagKey)
					assert.Equal(t, "segment_001", rule.SegmentKey)
					assert.NotEmpty(t, rule.Id)
					assert.Equal(t, int32(1), rule.Rank)

					require.Len(t, rule.Distributions, 2)

					assert.NotEmpty(t, rule.Distributions[0].Id)
					assert.InDelta(t, 50.0, rule.Distributions[0].Rollout, 0)
					assert.NotEmpty(t, rule.Distributions[1].Id)
					assert.InDelta(t, 50.0, rule.Distributions[1].Rollout, 0)

					t.Run("Paginated (page size 10)", func(t *testing.T) {
						var (
							found    []*flipt.Rule
							nextPage string
						)
						for {
							rules, err := sdk.Flipt().ListRules(ctx, &flipt.ListRuleRequest{
								NamespaceKey: namespace.Key,
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
						NamespaceKey: namespace.Key,
						FlagKey:      "flag_boolean",
						PageToken:    "Hello World",
					})
					require.EqualError(t, err, "rpc error: code = InvalidArgument desc = pageToken is not valid: \"Hello World\"")

					rules, err := sdk.Flipt().ListRollouts(ctx, &flipt.ListRolloutRequest{
						NamespaceKey: namespace.Key,
						FlagKey:      "flag_boolean",
					})
					require.NoError(t, err)
					require.Len(t, rules.Rules, 5, "unexpected number of rules returned")

					rule := rules.Rules[0]
					assert.Equal(t, namespace.Expected, rule.NamespaceKey)
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
								NamespaceKey: namespace.Key,
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
							NamespaceKey: namespace.Key,
							FlagKey:      "flag_001",
							EntityId:     "some-fixed-entity-id",
							Context: map[string]string{
								"in_segment": "segment_005",
							},
						})
						require.NoError(t, err)

						assert.True(t, response.Match)
						assert.Equal(t, "variant_002", response.Value)
						assert.Equal(t, flipt.EvaluationReason_MATCH_EVALUATION_REASON, response.Reason)
					})

					t.Run("BatchEvaluate", func(t *testing.T) {
						response, err := sdk.Flipt().BatchEvaluate(ctx, &flipt.BatchEvaluationRequest{
							NamespaceKey: namespace.Key,
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

						assert.True(t, response.Responses[0].Match)
						assert.Equal(t, "variant_002", response.Responses[0].Value)
						assert.Equal(t, flipt.EvaluationReason_MATCH_EVALUATION_REASON, response.Responses[0].Reason)

						assert.True(t, response.Responses[1].Match)
						assert.Equal(t, "variant_001", response.Responses[1].Value)
						assert.Equal(t, flipt.EvaluationReason_MATCH_EVALUATION_REASON, response.Responses[1].Reason)
					})
				})

				t.Run("Evaluation", func(t *testing.T) {
					t.Run("Variant", func(t *testing.T) {
						t.Run("match", func(t *testing.T) {
							response, err := sdk.Evaluation().Variant(ctx, &evaluation.EvaluationRequest{
								NamespaceKey: namespace.Key,
								FlagKey:      "flag_001",
								EntityId:     "some-fixed-entity-id",
								Context: map[string]string{
									"in_segment": "segment_005",
								},
							})
							require.NoError(t, err)

							assert.True(t, response.Match)
							assert.Equal(t, "flag_001", response.FlagKey)
							assert.Equal(t, "variant_002", response.VariantKey)
							assert.Equal(t, evaluation.EvaluationReason_MATCH_EVALUATION_REASON, response.Reason)
							assert.Contains(t, response.SegmentKeys, "segment_005")
						})

						t.Run("match segment ANDing", func(t *testing.T) {
							response, err := sdk.Evaluation().Variant(ctx, &evaluation.EvaluationRequest{
								NamespaceKey: namespace.Key,
								FlagKey:      "flag_variant_and_segments",
								EntityId:     "some-fixed-entity-id",
								Context: map[string]string{
									"in_segment": "segment_001",
									"anding":     "segment",
								},
							})
							require.NoError(t, err)

							assert.True(t, response.Match)
							assert.Equal(t, "variant_002", response.VariantKey)
							assert.Equal(t, "flag_variant_and_segments", response.FlagKey)
							assert.Equal(t, evaluation.EvaluationReason_MATCH_EVALUATION_REASON, response.Reason)
							assert.Contains(t, response.SegmentKeys, "segment_001")
							assert.Contains(t, response.SegmentKeys, "segment_anding")
						})

						t.Run("no match", func(t *testing.T) {
							response, err := sdk.Evaluation().Variant(ctx, &evaluation.EvaluationRequest{
								NamespaceKey: namespace.Key,
								FlagKey:      "flag_001",
								EntityId:     "some-fixed-entity-id",
								Context: map[string]string{
									"in_segment": "unknown",
								},
							})
							require.NoError(t, err)

							assert.False(t, response.Match)
							assert.Empty(t, response.VariantKey)
							assert.Equal(t, evaluation.EvaluationReason_UNKNOWN_EVALUATION_REASON, response.Reason)
						})

						t.Run("entity id matching", func(t *testing.T) {
							response, err := sdk.Evaluation().Variant(ctx, &evaluation.EvaluationRequest{
								NamespaceKey: namespace.Key,
								FlagKey:      "flag_using_entity_id",
								EntityId:     "user@flipt.io",
								Context:      map[string]string{},
							})
							require.NoError(t, err)

							assert.True(t, response.Match)
							assert.Equal(t, "variant_001", response.VariantKey)
							assert.Equal(t, "flag_using_entity_id", response.FlagKey)
							assert.Equal(t, evaluation.EvaluationReason_MATCH_EVALUATION_REASON, response.Reason)
							assert.Contains(t, response.SegmentKeys, "segment_entity_id")
						})

						t.Run("flag disabled", func(t *testing.T) {
							result, err := sdk.Evaluation().Variant(ctx, &evaluation.EvaluationRequest{
								NamespaceKey: namespace.Key,
								FlagKey:      "flag_disabled",
								EntityId:     "some-fixed-entity-id",
								Context:      map[string]string{},
							})
							require.NoError(t, err)

							assert.False(t, result.Match, "Evaluation should not have matched.")
							assert.Equal(t, evaluation.EvaluationReason_FLAG_DISABLED_EVALUATION_REASON, result.Reason)
							assert.Empty(t, result.VariantKey)
						})

						t.Run("flag not found", func(t *testing.T) {
							result, err := sdk.Evaluation().Variant(ctx, &evaluation.EvaluationRequest{
								NamespaceKey: namespace.Key,
								FlagKey:      "unknown_flag",
								EntityId:     "some-fixed-entity-id",
								Context:      map[string]string{},
							})
							msg := fmt.Sprintf("rpc error: code = NotFound desc = flag \"%s/unknown_flag\" not found", namespace.Expected)
							require.EqualError(t, err, msg)

							require.Nil(t, result)
						})

						t.Run("match no distributions", func(t *testing.T) {
							response, err := sdk.Evaluation().Variant(ctx, &evaluation.EvaluationRequest{
								NamespaceKey: namespace.Key,
								FlagKey:      "flag_no_distributions",
								EntityId:     "some-fixed-entity-id",
								Context: map[string]string{
									"in_segment": "segment_001",
								},
							})
							require.NoError(t, err)

							assert.True(t, response.Match)
							assert.Equal(t, "flag_no_distributions", response.FlagKey)
							assert.Equal(t, evaluation.EvaluationReason_MATCH_EVALUATION_REASON, response.Reason)
							assert.Contains(t, response.SegmentKeys, "segment_001")
						})
					})

					t.Run("Boolean", func(t *testing.T) {
						t.Run("default flag value", func(t *testing.T) {
							result, err := sdk.Evaluation().Boolean(ctx, &evaluation.EvaluationRequest{
								NamespaceKey: namespace.Key,
								FlagKey:      "flag_boolean",
								EntityId:     "hello",
								Context:      map[string]string{},
							})

							require.NoError(t, err)

							assert.Equal(t, evaluation.EvaluationReason_DEFAULT_EVALUATION_REASON, result.Reason)
							assert.Equal(t, "flag_boolean", result.FlagKey)
							assert.False(t, result.Enabled, "default flag value should be false")
						})

						t.Run("percentage match", func(t *testing.T) {
							result, err := sdk.Evaluation().Boolean(ctx, &evaluation.EvaluationRequest{
								NamespaceKey: namespace.Key,
								FlagKey:      "flag_boolean",
								EntityId:     "fixed",
								Context:      map[string]string{},
							})

							require.NoError(t, err)

							assert.Equal(t, evaluation.EvaluationReason_MATCH_EVALUATION_REASON, result.Reason)
							assert.Equal(t, "flag_boolean", result.FlagKey)
							assert.True(t, result.Enabled, "boolean evaluation value should be true")
						})

						t.Run("segment match", func(t *testing.T) {
							result, err := sdk.Evaluation().Boolean(ctx, &evaluation.EvaluationRequest{
								NamespaceKey: namespace.Key,
								FlagKey:      "flag_boolean",
								EntityId:     "fixed",
								Context: map[string]string{
									"in_segment": "segment_001",
								},
							})

							require.NoError(t, err)

							assert.Equal(t, evaluation.EvaluationReason_MATCH_EVALUATION_REASON, result.Reason)
							assert.Equal(t, "flag_boolean", result.FlagKey)
							assert.True(t, result.Enabled, "segment evaluation value should be true")
						})

						t.Run("segment with no constraints", func(t *testing.T) {
							result, err := sdk.Evaluation().Boolean(ctx, &evaluation.EvaluationRequest{
								NamespaceKey: namespace.Key,
								FlagKey:      "flag_boolean_no_constraints",
							})

							require.NoError(t, err)

							assert.Equal(t, evaluation.EvaluationReason_MATCH_EVALUATION_REASON, result.Reason)
							assert.Equal(t, "flag_boolean_no_constraints", result.FlagKey)
							assert.True(t, result.Enabled, "segment evaluation value should be true")
						})

						t.Run("segment with ANDing", func(t *testing.T) {
							result, err := sdk.Evaluation().Boolean(ctx, &evaluation.EvaluationRequest{
								NamespaceKey: namespace.Key,
								FlagKey:      "flag_boolean_and_segments",
								Context: map[string]string{
									"in_segment": "segment_001",
									"anding":     "segment",
								},
							})
							require.NoError(t, err)

							assert.Equal(t, evaluation.EvaluationReason_MATCH_EVALUATION_REASON, result.Reason)
							assert.Equal(t, "flag_boolean_and_segments", result.FlagKey)
							assert.True(t, result.Enabled, "segment evaluation value should be true")
						})
					})

					t.Run("Batch", func(t *testing.T) {
						t.Run("batch evaluations (with not found errors)", func(t *testing.T) {
							result, err := sdk.Evaluation().Batch(ctx, &evaluation.BatchEvaluationRequest{
								Requests: []*evaluation.EvaluationRequest{
									{
										NamespaceKey: namespace.Key,
										FlagKey:      "flag_boolean",
										EntityId:     "fixed",
										Context:      map[string]string{},
									},
									{
										NamespaceKey: namespace.Key,
										FlagKey:      "foobarnotfound",
									},
									{
										NamespaceKey: namespace.Key,
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
							assert.Equal(t, "flag_boolean", b.BooleanResponse.FlagKey)
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
							assert.Equal(t, "flag_001", v.VariantResponse.FlagKey)
							assert.Contains(t, v.VariantResponse.SegmentKeys, "segment_005")
							assert.Equal(t, evaluation.EvaluationResponseType_VARIANT_EVALUATION_RESPONSE_TYPE, result.Responses[2].Type)
						})
					})
				})

				t.Run("Custom Reference", func(t *testing.T) {
					if !opts.References {
						t.Skip("backend does not support references")
						return
					}

					reference := "alternate"
					t.Run("ListFlags", func(t *testing.T) {
						flags, err := sdk.Flipt().ListFlags(ctx, &flipt.ListFlagRequest{
							NamespaceKey: namespace.Key,
							Reference:    reference,
						})

						require.NoError(t, err)

						assert.Equal(t, int32(2), flags.TotalCount)
						assert.Len(t, flags.Flags, 2)
					})

					t.Run("GetFlag", func(t *testing.T) {
						variant, err := sdk.Flipt().GetFlag(ctx, &flipt.GetFlagRequest{
							NamespaceKey: namespace.Key,
							Key:          "flag_variant",
							Reference:    reference,
						})
						require.NoError(t, err)

						assert.Equal(t, namespace.Expected, variant.NamespaceKey)
						assert.Equal(t, "flag_variant", variant.Key)
						assert.Equal(t, flipt.FlagType_VARIANT_FLAG_TYPE, variant.Type)
						assert.Equal(t, "Alternate Flag Type Variant", variant.Name)
						assert.Equal(t, "Variant Flag Description", variant.Description)
						assert.False(t, variant.Enabled)

						boolean, err := sdk.Flipt().GetFlag(ctx, &flipt.GetFlagRequest{
							NamespaceKey: namespace.Key,
							Key:          "flag_boolean",
							Reference:    reference,
						})
						require.NoError(t, err)

						assert.Equal(t, namespace.Expected, boolean.NamespaceKey)
						assert.Equal(t, "flag_boolean", boolean.Key)
						assert.Equal(t, flipt.FlagType_BOOLEAN_FLAG_TYPE, boolean.Type)
						assert.Equal(t, "Alternate Flag Type Boolean", boolean.Name)
						assert.Equal(t, "Boolean Flag Description", boolean.Description)
						assert.True(t, boolean.Enabled)
					})
				})
			})
		}

		t.Run("Auth", func(t *testing.T) {
			t.Run("Self", func(t *testing.T) {
				_, err := sdk.Auth().AuthenticationService().GetAuthenticationSelf(ctx)
				assert.NoError(t, err)
			})

			t.Run("Public", func(t *testing.T) {
				_, err := sdk.Auth().PublicAuthenticationService().ListAuthenticationMethods(ctx)
				require.NoError(t, err)
			})
		})

	})
}
