package api

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/gofrs/uuid"
	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.flipt.io/flipt/rpc/flipt"
	"go.flipt.io/flipt/rpc/flipt/evaluation"
	sdk "go.flipt.io/flipt/sdk/go"
	"google.golang.org/protobuf/testing/protocmp"
)

func API(t *testing.T, ctx context.Context, client sdk.SDK, namespace string, authenticated bool) {
	t.Run("Namespaces", func(t *testing.T) {
		if !namespaceIsDefault(namespace) {
			t.Log(`Create namespace.`)

			created, err := client.Flipt().CreateNamespace(ctx, &flipt.CreateNamespaceRequest{
				Key:  namespace,
				Name: namespace,
			})
			require.NoError(t, err)

			assert.Equal(t, namespace, created.Key)
			assert.Equal(t, namespace, created.Name)

			t.Log(`Get namespace by key.`)

			retrieved, err := client.Flipt().GetNamespace(ctx, &flipt.GetNamespaceRequest{
				Key: namespace,
			})

			assert.Equal(t, created.Name, retrieved.Name)

			t.Log(`List namespaces error with invalid page token`)

			_, err = client.Flipt().ListNamespaces(ctx, &flipt.ListNamespaceRequest{
				PageToken: "Hello World",
			})
			require.EqualError(t, err, "rpc error: code = InvalidArgument desc = pageToken is not valid: \"Hello World\"")

			t.Log(`List namespaces.`)

			namespaces, err := client.Flipt().ListNamespaces(ctx, &flipt.ListNamespaceRequest{})
			require.NoError(t, err)

			assert.Len(t, namespaces.Namespaces, 2)

			assert.Equal(t, "default", namespaces.Namespaces[0].Key)
			assert.Equal(t, namespace, namespaces.Namespaces[1].Key)

			t.Log(`Update namespace.`)

			updated, err := client.Flipt().UpdateNamespace(ctx, &flipt.UpdateNamespaceRequest{
				Key:         namespace,
				Name:        namespace,
				Description: "Some kind of description",
			})
			require.NoError(t, err)

			assert.Equal(t, "Some kind of description", updated.Description)
		} else {
			t.Log(`Ensure default cannot be created.`)

			_, err := client.Flipt().CreateNamespace(ctx, &flipt.CreateNamespaceRequest{
				Key:  "default",
				Name: "Default",
			})
			require.EqualError(t, err, "rpc error: code = InvalidArgument desc = namespace \"default\" is not unique")

			t.Log(`Ensure default can be updated.`)

			updated, err := client.Flipt().UpdateNamespace(ctx, &flipt.UpdateNamespaceRequest{
				Key:         "default",
				Name:        "Default",
				Description: "Some namespace description.",
			})
			require.NoError(t, err)
			assert.Equal(t, "Default", updated.Name)

			t.Log(`Ensure default cannot be deleted.`)

			err = client.Flipt().DeleteNamespace(ctx, &flipt.DeleteNamespaceRequest{
				Key: "default",
			})
			require.EqualError(t, err, "rpc error: code = InvalidArgument desc = namespace \"default\" is protected")
		}
	})

	t.Run("Flags and Variants", func(t *testing.T) {
		t.Log("Create a new enabled flag with key \"test\".")

		enabled, err := client.Flipt().CreateFlag(ctx, &flipt.CreateFlagRequest{
			NamespaceKey: namespace,
			Key:          "test",
			Name:         "Test",
			Description:  "This is a test flag",
			Enabled:      true,
		})
		require.NoError(t, err)

		assert.Equal(t, "test", enabled.Key)
		assert.Equal(t, "Test", enabled.Name)
		assert.Equal(t, "This is a test flag", enabled.Description)
		assert.True(t, enabled.Enabled, "Flag should be enabled")

		t.Log("Create a new flag in a disabled state.")

		disabled, err := client.Flipt().CreateFlag(ctx, &flipt.CreateFlagRequest{
			NamespaceKey: namespace,
			Key:          "disabled",
			Name:         "Disabled",
			Description:  "This is a disabled test flag",
			Enabled:      false,
		})
		require.NoError(t, err)

		assert.Equal(t, "disabled", disabled.Key)
		assert.Equal(t, "Disabled", disabled.Name)
		assert.Equal(t, "This is a disabled test flag", disabled.Description)
		assert.False(t, disabled.Enabled, "Flag should be disabled")

		t.Log("Create a new enabled boolean flag with key \"boolean_enabled\".")

		booleanEnabled, err := client.Flipt().CreateFlag(ctx, &flipt.CreateFlagRequest{
			Type:         flipt.FlagType_BOOLEAN_FLAG_TYPE,
			NamespaceKey: namespace,
			Key:          "boolean_enabled",
			Name:         "Boolean Enabled",
			Description:  "This is an enabled boolean test flag",
			Enabled:      true,
		})
		require.NoError(t, err)

		assert.Equal(t, "boolean_enabled", booleanEnabled.Key)
		assert.Equal(t, "Boolean Enabled", booleanEnabled.Name)
		assert.Equal(t, "This is an enabled boolean test flag", booleanEnabled.Description)
		assert.True(t, booleanEnabled.Enabled, "Flag should be enabled")

		t.Log("Create a new flag in a disabled state.")

		booleanDisabled, err := client.Flipt().CreateFlag(ctx, &flipt.CreateFlagRequest{
			Type:         flipt.FlagType_BOOLEAN_FLAG_TYPE,
			NamespaceKey: namespace,
			Key:          "boolean_disabled",
			Name:         "Boolean Disabled",
			Description:  "This is a disabled boolean test flag",
			Enabled:      false,
		})
		require.NoError(t, err)

		assert.Equal(t, "boolean_disabled", booleanDisabled.Key)
		assert.Equal(t, "Boolean Disabled", booleanDisabled.Name)
		assert.Equal(t, "This is a disabled boolean test flag", booleanDisabled.Description)
		assert.False(t, booleanDisabled.Enabled, "Flag should be disabled")

		t.Log("Retrieve flag with key \"test\".")

		flag, err := client.Flipt().GetFlag(ctx, &flipt.GetFlagRequest{
			NamespaceKey: namespace,
			Key:          "test",
		})
		require.NoError(t, err)

		assert.Equal(t, enabled, flag)

		t.Log("Update flag with key \"test\".")

		updated, err := client.Flipt().UpdateFlag(ctx, &flipt.UpdateFlagRequest{
			NamespaceKey: namespace,
			Key:          enabled.Key,
			Name:         "Test 2",
			Description:  enabled.Description,
			Enabled:      true,
		})
		require.NoError(t, err)

		assert.Equal(t, "Test 2", updated.Name)

		t.Log("List all Flags error with invalid page token")

		_, err = client.Flipt().ListFlags(ctx, &flipt.ListFlagRequest{
			NamespaceKey: namespace,
			PageToken:    "Hello World",
			Limit:        10,
		})
		require.EqualError(t, err, "rpc error: code = InvalidArgument desc = pageToken is not valid: \"Hello World\"")

		t.Log("List all flags.")

		flags, err := client.Flipt().ListFlags(ctx, &flipt.ListFlagRequest{
			NamespaceKey: namespace,
		})
		require.NoError(t, err)

		assert.Len(t, flags.Flags, 4)
		for i, flag := range []*flipt.Flag{
			updated,
			disabled,
			booleanEnabled,
			booleanDisabled,
		} {
			assert.Equal(t, flag.Key, flags.Flags[i].Key)
			assert.Equal(t, flag.Name, flags.Flags[i].Name)
			assert.Equal(t, flag.Description, flags.Flags[i].Description)
			assert.Equal(t, flag.Enabled, flags.Flags[i].Enabled)
		}

		for _, key := range []string{"one", "two"} {
			t.Logf("Create variant with key %q.", key)

			createdVariant, err := client.Flipt().CreateVariant(ctx, &flipt.CreateVariantRequest{
				NamespaceKey: namespace,
				Key:          key,
				Name:         key,
				FlagKey:      "test",
			})
			require.NoError(t, err)

			assert.Equal(t, key, createdVariant.Key)
			assert.Equal(t, key, createdVariant.Name)
		}

		t.Log("Get flag \"test\" and check variants.")

		flag, err = client.Flipt().GetFlag(ctx, &flipt.GetFlagRequest{
			NamespaceKey: namespace,
			Key:          "test",
		})
		require.NoError(t, err)

		assert.Len(t, flag.Variants, 2)

		t.Log(`Update variant "one" (rename from "one" to "One").`)

		updatedVariant, err := client.Flipt().UpdateVariant(ctx, &flipt.UpdateVariantRequest{
			NamespaceKey: namespace,
			FlagKey:      "test",
			Id:           flag.Variants[0].Id,
			Key:          "one",
			Name:         "One",
		})
		require.NoError(t, err)

		assert.Equal(t, "one", updatedVariant.Key)
		assert.Equal(t, "One", updatedVariant.Name)
	})

	t.Run("Segments and Constraints", func(t *testing.T) {
		t.Log(`Create segment "everyone".`)

		createdSegment, err := client.Flipt().CreateSegment(ctx, &flipt.CreateSegmentRequest{
			NamespaceKey: namespace,
			Key:          "everyone",
			Name:         "Everyone",
			MatchType:    flipt.MatchType_ALL_MATCH_TYPE,
		})
		require.NoError(t, err)

		assert.Equal(t, "everyone", createdSegment.Key)
		assert.Equal(t, "Everyone", createdSegment.Name)
		assert.Equal(t, flipt.MatchType_ALL_MATCH_TYPE, createdSegment.MatchType)

		t.Log(`Get segment "everyone".`)

		retrievedSegment, err := client.Flipt().GetSegment(ctx, &flipt.GetSegmentRequest{
			NamespaceKey: namespace,
			Key:          "everyone",
		})
		require.NoError(t, err)

		assert.Equal(t, createdSegment.Key, retrievedSegment.Key)
		assert.Equal(t, createdSegment.Name, retrievedSegment.Name)
		assert.Equal(t, createdSegment.Description, retrievedSegment.Description)
		assert.Equal(t, createdSegment.MatchType, createdSegment.MatchType)

		t.Log(`Update segment "everyone" (rename "Everyone" to "All the peeps").`)

		updatedSegment, err := client.Flipt().UpdateSegment(ctx, &flipt.UpdateSegmentRequest{
			NamespaceKey: namespace,
			Key:          "everyone",
			Name:         "All the peeps",
		})
		require.NoError(t, err)

		assert.Equal(t, "everyone", updatedSegment.Key)
		assert.Equal(t, "All the peeps", updatedSegment.Name)
		assert.Equal(t, flipt.MatchType_ALL_MATCH_TYPE, updatedSegment.MatchType)

		for _, constraint := range []struct {
			property, operator, value string
		}{
			{"foo", "eq", "bar"},
			{"fizz", "neq", "buzz"},
		} {
			t.Logf(`Create constraint with key %q %s %q.`,
				constraint.property,
				constraint.operator,
				constraint.value,
			)

			createdConstraint, err := client.Flipt().CreateConstraint(ctx, &flipt.CreateConstraintRequest{
				NamespaceKey: namespace,
				SegmentKey:   "everyone",
				Type:         flipt.ComparisonType_STRING_COMPARISON_TYPE,
				Property:     constraint.property,
				Operator:     constraint.operator,
				Value:        constraint.value,
			})
			require.NoError(t, err)

			assert.Equal(t, constraint.property, createdConstraint.Property)
			assert.Equal(t, constraint.operator, createdConstraint.Operator)
			assert.Equal(t, constraint.value, createdConstraint.Value)
		}

		t.Log(`Get segment "everyone" with constraints.`)

		retrievedSegment, err = client.Flipt().GetSegment(ctx, &flipt.GetSegmentRequest{
			NamespaceKey: namespace,
			Key:          "everyone",
		})
		require.NoError(t, err)

		assert.Len(t, retrievedSegment.Constraints, 2)
		assert.Equal(t, "foo", retrievedSegment.Constraints[0].Property)
		assert.Equal(t, "fizz", retrievedSegment.Constraints[1].Property)

		t.Log(`Update constraint value (from "bar" to "baz").`)
		updatedConstraint, err := client.Flipt().UpdateConstraint(ctx, &flipt.UpdateConstraintRequest{
			NamespaceKey: namespace,
			SegmentKey:   "everyone",
			Type:         retrievedSegment.Constraints[0].Type,
			Id:           retrievedSegment.Constraints[0].Id,
			Property:     retrievedSegment.Constraints[0].Property,
			Operator:     retrievedSegment.Constraints[0].Operator,
			Value:        "baz",
			Description:  "newdesc",
		})
		require.NoError(t, err)

		assert.Equal(t, flipt.ComparisonType_STRING_COMPARISON_TYPE, updatedConstraint.Type)
		assert.Equal(t, "foo", updatedConstraint.Property)
		assert.Equal(t, "eq", updatedConstraint.Operator)
		assert.Equal(t, "baz", updatedConstraint.Value)
		assert.Equal(t, "newdesc", updatedConstraint.Description)
	})

	t.Run("Rules and Distributions", func(t *testing.T) {
		t.Log(`Create rule "rank 1".`)

		ruleOne, err := client.Flipt().CreateRule(ctx, &flipt.CreateRuleRequest{
			NamespaceKey: namespace,
			FlagKey:      "test",
			SegmentKey:   "everyone",
			Rank:         1,
		})

		require.NoError(t, err)

		assert.Equal(t, "test", ruleOne.FlagKey)
		assert.Equal(t, "everyone", ruleOne.SegmentKey)
		assert.Equal(t, int32(1), ruleOne.Rank)

		t.Log(`Get rules "rank 1".`)

		retrievedRule, err := client.Flipt().GetRule(ctx, &flipt.GetRuleRequest{
			NamespaceKey: namespace,
			FlagKey:      "test",
			Id:           ruleOne.Id,
		})
		require.NoError(t, err)

		assert.Equal(t, "test", retrievedRule.FlagKey)
		assert.Equal(t, "everyone", retrievedRule.SegmentKey)
		assert.Equal(t, int32(1), retrievedRule.Rank)

		t.Log(`Create rule "rank 2".`)

		ruleTwo, err := client.Flipt().CreateRule(ctx, &flipt.CreateRuleRequest{
			NamespaceKey: namespace,
			FlagKey:      "test",
			SegmentKey:   "everyone",
			Rank:         2,
		})

		require.NoError(t, err)

		assert.Equal(t, "test", ruleTwo.FlagKey)
		assert.Equal(t, "everyone", ruleTwo.SegmentKey)
		assert.Equal(t, int32(2), ruleTwo.Rank)

		// ensure you can not link flags and segments from different namespaces.
		if !namespaceIsDefault(namespace) {
			t.Log(`Ensure that rules can only link entities in the same namespace.`)
			_, err = client.Flipt().CreateRule(ctx, &flipt.CreateRuleRequest{
				FlagKey:    "test",
				SegmentKey: "everyone",
				Rank:       2,
			})

			msg := "rpc error: code = NotFound desc = flag \"default/test\" or segment \"default/everyone\" not found"
			require.EqualError(t, err, msg)
		}
		t.Log(`List rules error with invalid page token.`)

		_, err = client.Flipt().ListRules(ctx, &flipt.ListRuleRequest{
			NamespaceKey: namespace,
			FlagKey:      "test",
			PageToken:    "Hello World",
			Limit:        10,
		})

		require.EqualError(t, err, "rpc error: code = InvalidArgument desc = pageToken is not valid: \"Hello World\"")

		t.Log(`List rules.`)

		allRules, err := client.Flipt().ListRules(ctx, &flipt.ListRuleRequest{
			NamespaceKey: namespace,
			FlagKey:      "test",
		})
		require.NoError(t, err)

		assert.Len(t, allRules.Rules, 2)

		assert.Equal(t, ruleOne.Id, allRules.Rules[0].Id)
		assert.Equal(t, ruleTwo.Id, allRules.Rules[1].Id)

		t.Log(`Re-order rules.`)

		err = client.Flipt().OrderRules(ctx, &flipt.OrderRulesRequest{
			NamespaceKey: namespace,
			FlagKey:      "test",
			RuleIds:      []string{ruleTwo.Id, ruleOne.Id},
		})
		require.NoError(t, err)

		t.Log(`List rules again.`)

		allRules, err = client.Flipt().ListRules(ctx, &flipt.ListRuleRequest{
			NamespaceKey: namespace,
			FlagKey:      "test",
		})
		require.NoError(t, err)

		assert.Len(t, allRules.Rules, 2)

		// ensure the order has switched
		assert.Equal(t, ruleTwo.Id, allRules.Rules[0].Id)
		assert.Equal(t, int32(1), allRules.Rules[0].Rank)
		assert.Equal(t, ruleOne.Id, allRules.Rules[1].Id)
		assert.Equal(t, int32(2), allRules.Rules[1].Rank)

		t.Log(`Create distribution "rollout 100".`)

		flag, err := client.Flipt().GetFlag(ctx, &flipt.GetFlagRequest{
			NamespaceKey: namespace,
			Key:          "test",
		})
		require.NoError(t, err)

		distribution, err := client.Flipt().CreateDistribution(ctx, &flipt.CreateDistributionRequest{
			NamespaceKey: namespace,
			FlagKey:      "test",
			RuleId:       ruleTwo.Id,
			VariantId:    flag.Variants[0].Id,
			Rollout:      100,
		})
		require.NoError(t, err)

		if !namespaceIsDefault(namespace) {
			t.Log(`Ensure that distributions and all entities associated with them are part of same namespace.`)
			var (
				flagKey = "defaultflag"
			)

			defaultFlag, err := client.Flipt().CreateFlag(ctx, &flipt.CreateFlagRequest{
				Key:         flagKey,
				Name:        "DefaultFlag",
				Description: "default flag description",
				Enabled:     true,
			})
			require.NoError(t, err)

			defaultVariant, err := client.Flipt().CreateVariant(ctx, &flipt.CreateVariantRequest{
				Key:         "defaultvariant",
				Name:        "DefaultVariant",
				FlagKey:     flagKey,
				Description: "default variant description",
			})
			require.NoError(t, err)

			_, err = client.Flipt().CreateDistribution(ctx, &flipt.CreateDistributionRequest{
				FlagKey:   "test",
				RuleId:    ruleTwo.Id,
				VariantId: defaultVariant.Id,
				Rollout:   100,
			})

			msg := fmt.Sprintf("rpc error: code = NotFound desc = variant %q, rule %q, flag %q in namespace %q not found", defaultVariant.Id, ruleTwo.Id, "test", "default")
			require.EqualError(t, err, msg)

			_, err = client.Flipt().UpdateDistribution(ctx, &flipt.UpdateDistributionRequest{
				NamespaceKey: namespace,
				Id:           distribution.Id,
				FlagKey:      "test",
				VariantId:    defaultVariant.Id,
				RuleId:       ruleTwo.Id,
				Rollout:      50,
			})

			msg = fmt.Sprintf("rpc error: code = NotFound desc = variant %q, rule %q, flag %q in namespace %q not found", defaultVariant.Id, ruleTwo.Id, "test", namespace)
			require.EqualError(t, err, msg)

			err = client.Flipt().DeleteFlag(ctx, &flipt.DeleteFlagRequest{
				Key: defaultFlag.Key,
			})
			require.NoError(t, err)

			err = client.Flipt().DeleteVariant(ctx, &flipt.DeleteVariantRequest{
				Id:      defaultVariant.Id,
				FlagKey: flagKey,
			})
			require.NoError(t, err)
		}

		assert.Equal(t, ruleTwo.Id, distribution.RuleId)
		assert.Equal(t, float32(100), distribution.Rollout)
	})

	t.Run("Boolean Rollouts", func(t *testing.T) {
		rolloutSegment, err := client.Flipt().CreateRollout(ctx, &flipt.CreateRolloutRequest{
			NamespaceKey: namespace,
			FlagKey:      "boolean_disabled",
			Description:  "matches a segment",
			Rank:         1,
			Rule: &flipt.CreateRolloutRequest_Segment{
				Segment: &flipt.RolloutSegment{
					SegmentKey: "everyone",
					Value:      true,
				},
			},
		})
		require.NoError(t, err)

		assert.Equal(t, namespace, rolloutSegment.NamespaceKey)
		assert.Equal(t, "boolean_disabled", rolloutSegment.FlagKey)
		assert.Equal(t, int32(1), rolloutSegment.Rank)
		assert.Equal(t, "everyone", rolloutSegment.Rule.(*flipt.Rollout_Segment).Segment.SegmentKey)
		assert.Equal(t, true, rolloutSegment.Rule.(*flipt.Rollout_Segment).Segment.Value)

		rolloutThreshold, err := client.Flipt().CreateRollout(ctx, &flipt.CreateRolloutRequest{
			NamespaceKey: namespace,
			FlagKey:      "boolean_disabled",
			Description:  "50% disabled",
			Rank:         2,
			Rule: &flipt.CreateRolloutRequest_Threshold{
				Threshold: &flipt.RolloutThreshold{
					Percentage: 50,
					Value:      false,
				},
			},
		})
		require.NoError(t, err)

		assert.Equal(t, namespace, rolloutThreshold.NamespaceKey)
		assert.Equal(t, "boolean_disabled", rolloutThreshold.FlagKey)
		assert.Equal(t, "50% disabled", rolloutThreshold.Description)
		assert.Equal(t, int32(2), rolloutThreshold.Rank)
		assert.Equal(t, float32(50.0), rolloutThreshold.Rule.(*flipt.Rollout_Threshold).Threshold.Percentage)
		assert.Equal(t, false, rolloutThreshold.Rule.(*flipt.Rollout_Threshold).Threshold.Value)

		rollouts, err := client.Flipt().ListRollouts(ctx, &flipt.ListRolloutRequest{
			NamespaceKey: namespace,
			FlagKey:      "boolean_disabled",
		})
		require.NoError(t, err)

		assert.Empty(t, cmp.Diff([]*flipt.Rollout{
			rolloutSegment,
			rolloutThreshold,
		}, rollouts.Rules, protocmp.Transform()))

		updatedRollout, err := client.Flipt().UpdateRollout(ctx, &flipt.UpdateRolloutRequest{
			NamespaceKey: namespace,
			FlagKey:      "boolean_disabled",
			Id:           rolloutThreshold.Id,
			Description:  "50% enabled",
			Rule: &flipt.UpdateRolloutRequest_Threshold{
				Threshold: &flipt.RolloutThreshold{
					Percentage: 50,
					Value:      true,
				},
			},
		})
		require.NoError(t, err)

		assert.Equal(t, namespace, updatedRollout.NamespaceKey)
		assert.Equal(t, "boolean_disabled", updatedRollout.FlagKey)
		assert.Equal(t, "50% enabled", updatedRollout.Description)
		assert.Equal(t, int32(2), updatedRollout.Rank)
		assert.Equal(t, float32(50.0), updatedRollout.Rule.(*flipt.Rollout_Threshold).Threshold.Percentage)
		assert.Equal(t, true, updatedRollout.Rule.(*flipt.Rollout_Threshold).Threshold.Value)

		t.Run("Cannot change rollout type", func(t *testing.T) {
			_, err := client.Flipt().UpdateRollout(ctx, &flipt.UpdateRolloutRequest{
				NamespaceKey: namespace,
				FlagKey:      "boolean_disabled",
				Id:           rolloutThreshold.Id,
				Description:  "50% enabled",
				Rule: &flipt.UpdateRolloutRequest_Segment{
					Segment: &flipt.RolloutSegment{
						SegmentKey: "everyone",
						Value:      true,
					},
				},
			})

			require.EqualError(t, err, "rpc error: code = InvalidArgument desc = cannot change type of rollout: have \"THRESHOLD_ROLLOUT_TYPE\" attempted \"SEGMENT_ROLLOUT_TYPE\"")
		})
	})

	t.Run("Legacy", func(t *testing.T) {
		t.Run("Evaluation", func(t *testing.T) {
			t.Log(`Successful match.`)

			result, err := client.Flipt().Evaluate(ctx, &flipt.EvaluationRequest{
				NamespaceKey: namespace,
				FlagKey:      "test",
				EntityId:     uuid.Must(uuid.NewV4()).String(),
				Context: map[string]string{
					"foo":  "baz",
					"fizz": "bozz",
				},
			})
			require.NoError(t, err)

			require.True(t, result.Match, "Evaluation should have matched.")
			assert.Equal(t, "everyone", result.SegmentKey)
			assert.Equal(t, "one", result.Value)
			assert.Equal(t, flipt.EvaluationReason_MATCH_EVALUATION_REASON, result.Reason)

			t.Log(`Unsuccessful match.`)

			result, err = client.Flipt().Evaluate(ctx, &flipt.EvaluationRequest{
				NamespaceKey: namespace,
				FlagKey:      "test",
				EntityId:     uuid.Must(uuid.NewV4()).String(),
				Context: map[string]string{
					"fizz": "buzz",
				},
			})
			require.NoError(t, err)

			assert.False(t, result.Match, "Evaluation should not have matched.")
		})

		t.Run("Batch Evaluation", func(t *testing.T) {
			t.Log(`Successful match.`)

			results, err := client.Flipt().BatchEvaluate(ctx, &flipt.BatchEvaluationRequest{
				NamespaceKey: namespace,
				Requests: []*flipt.EvaluationRequest{
					{
						NamespaceKey: namespace,
						FlagKey:      "test",
						EntityId:     uuid.Must(uuid.NewV4()).String(),
						Context: map[string]string{
							"foo":  "baz",
							"fizz": "bozz",
						},
					},
				},
			})
			require.NoError(t, err)

			require.Len(t, results.Responses, 1)
			result := results.Responses[0]

			require.True(t, result.Match, "Evaluation should have matched.")
			assert.Equal(t, "everyone", result.SegmentKey)
			assert.Equal(t, "one", result.Value)
			assert.Equal(t, flipt.EvaluationReason_MATCH_EVALUATION_REASON, result.Reason)

			t.Log(`Unsuccessful match.`)

			results, err = client.Flipt().BatchEvaluate(ctx, &flipt.BatchEvaluationRequest{
				NamespaceKey: namespace,
				Requests: []*flipt.EvaluationRequest{
					{
						NamespaceKey: namespace,
						FlagKey:      "test",
						EntityId:     uuid.Must(uuid.NewV4()).String(),
						Context: map[string]string{
							"fizz": "buzz",
						},
					},
				},
			})
			require.NoError(t, err)

			require.Len(t, results.Responses, 1)
			result = results.Responses[0]

			assert.False(t, result.Match, "Evaluation should not have matched.")
		})
	})

	t.Run("Evaluation", func(t *testing.T) {
		t.Run("Variant", func(t *testing.T) {
			t.Run("successful match", func(t *testing.T) {
				result, err := client.Evaluation().Variant(ctx, &evaluation.EvaluationRequest{
					NamespaceKey: namespace,
					FlagKey:      "test",
					EntityId:     uuid.Must(uuid.NewV4()).String(),
					Context: map[string]string{
						"foo":  "baz",
						"fizz": "bozz",
					},
				})
				require.NoError(t, err)

				require.True(t, result.Match, "Evaluation should have matched.")
				assert.Equal(t, "everyone", result.SegmentKey)
				assert.Equal(t, "one", result.VariantKey)
				assert.Equal(t, evaluation.EvaluationReason_MATCH_EVALUATION_REASON, result.Reason)
			})

			t.Run("no match", func(t *testing.T) {
				result, err := client.Evaluation().Variant(ctx, &evaluation.EvaluationRequest{
					NamespaceKey: namespace,
					FlagKey:      "test",
					EntityId:     uuid.Must(uuid.NewV4()).String(),
					Context: map[string]string{
						"fizz": "buzz",
					},
				})
				require.NoError(t, err)

				assert.False(t, result.Match, "Evaluation should not have matched.")
				assert.Equal(t, evaluation.EvaluationReason_UNKNOWN_EVALUATION_REASON, result.Reason)
				assert.Empty(t, result.SegmentKey)
				assert.Empty(t, result.VariantKey)
			})

			t.Run("flag disabled", func(t *testing.T) {
				result, err := client.Evaluation().Variant(ctx, &evaluation.EvaluationRequest{
					NamespaceKey: namespace,
					FlagKey:      "disabled",
					EntityId:     uuid.Must(uuid.NewV4()).String(),
					Context:      map[string]string{},
				})
				require.NoError(t, err)

				assert.False(t, result.Match, "Evaluation should not have matched.")
				assert.Equal(t, evaluation.EvaluationReason_FLAG_DISABLED_EVALUATION_REASON, result.Reason)
				assert.Empty(t, result.SegmentKey)
				assert.Empty(t, result.VariantKey)
			})

			t.Run("flag not found", func(t *testing.T) {
				result, err := client.Evaluation().Variant(ctx, &evaluation.EvaluationRequest{
					NamespaceKey: namespace,
					FlagKey:      "unknown_flag",
					EntityId:     uuid.Must(uuid.NewV4()).String(),
					Context:      map[string]string{},
				})

				msg := fmt.Sprintf("rpc error: code = NotFound desc = flag \"%s/unknown_flag\" not found", namespace)
				require.EqualError(t, err, msg)
				require.Nil(t, result)
			})
		})
	})

	t.Run("Delete", func(t *testing.T) {
		if !namespaceIsDefault(namespace) {
			t.Log(`Namespace with flags fails.`)
			err := client.Flipt().DeleteNamespace(ctx, &flipt.DeleteNamespaceRequest{
				Key: namespace,
			})

			msg := fmt.Sprintf("rpc error: code = InvalidArgument desc = namespace %q cannot be deleted; flags must be deleted first", namespace)
			require.EqualError(t, err, msg)
		}

		t.Log(`Rules.`)
		rules, err := client.Flipt().ListRules(ctx, &flipt.ListRuleRequest{
			NamespaceKey: namespace,
			FlagKey:      "test",
		})
		require.NoError(t, err)

		for _, rule := range rules.Rules {
			client.Flipt().DeleteRule(ctx, &flipt.DeleteRuleRequest{
				NamespaceKey: namespace,
				FlagKey:      "test",
				Id:           rule.Id,
			})
		}

		t.Log(`Flag.`)

		flags, err := client.Flipt().ListFlags(ctx, &flipt.ListFlagRequest{
			NamespaceKey: namespace,
		})
		for _, flag := range flags.Flags {
			err = client.Flipt().DeleteFlag(ctx, &flipt.DeleteFlagRequest{
				NamespaceKey: namespace,
				Key:          flag.Key,
			})
			require.NoError(t, err)
		}

		t.Log(`Segment.`)
		err = client.Flipt().DeleteSegment(ctx, &flipt.DeleteSegmentRequest{
			NamespaceKey: namespace,
			Key:          "everyone",
		})
		require.NoError(t, err)

		if !namespaceIsDefault(namespace) {
			t.Log(`Namespace.`)
			err = client.Flipt().DeleteNamespace(ctx, &flipt.DeleteNamespaceRequest{
				Key: namespace,
			})
			require.NoError(t, err)
		}
	})

	t.Run("Meta", func(t *testing.T) {
		t.Log(`Returns Flipt service information.`)

		info, err := client.Meta().GetInfo(ctx)
		require.NoError(t, err)

		assert.Equal(t, "application/json", info.ContentType)

		var infoMap map[string]any
		require.NoError(t, json.Unmarshal(info.Data, &infoMap))

		version, ok := infoMap["version"]
		assert.True(t, ok, "Missing version.")
		assert.NotEmpty(t, version)

		goVersion, ok := infoMap["goVersion"]
		assert.True(t, ok, "Missing Go version.")
		assert.NotEmpty(t, goVersion)

		t.Log(`Returns expected configuration.`)

		config, err := client.Meta().GetConfiguration(ctx)
		require.NoError(t, err)

		assert.Equal(t, "application/json", info.ContentType)

		var configMap map[string]any
		require.NoError(t, json.Unmarshal(config.Data, &configMap))

		for _, name := range []string{
			"log",
			"ui",
			"cache",
			"server",
			"db",
		} {
			field, ok := configMap[name]
			assert.True(t, ok, "Missing %s.", name)
			assert.NotEmpty(t, field)
		}
	})

	t.Run("Auth", func(t *testing.T) {
		t.Run("Self", func(t *testing.T) {
			_, err := client.Auth().AuthenticationService().GetAuthenticationSelf(ctx)
			if authenticated {
				assert.NoError(t, err)
			} else {
				assert.EqualError(t, err, "rpc error: code = Unauthenticated desc = request was not authenticated")
			}
		})
		t.Run("Public", func(t *testing.T) {
			_, err := client.Auth().PublicAuthenticationService().ListAuthenticationMethods(ctx)
			require.NoError(t, err)
		})
	})
}

func namespaceIsDefault(ns string) bool {
	return ns == "" || ns == "default"
}
