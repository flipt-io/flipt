package integration

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.flipt.io/flipt/rpc/flipt"
	"go.flipt.io/flipt/sdk"
)

func Harness(t *testing.T, fn func(t *testing.T) sdk.SDK) {
	client := fn(t)

	ctx := context.Background()

	t.Run("Flags and Variants", func(t *testing.T) {
		t.Log("Create a new flag with key \"test\".")

		created, err := client.Flipt().CreateFlag(ctx, &flipt.CreateFlagRequest{
			Key:         "test",
			Name:        "Test",
			Description: "This is a test flag",
			Enabled:     true,
		})
		require.NoError(t, err)

		assert.Equal(t, "test", created.Key)
		assert.Equal(t, "Test", created.Name)
		assert.Equal(t, "This is a test flag", created.Description)
		assert.True(t, created.Enabled, "Flag should be enabled")

		t.Log("Retrieve flag with key \"test\".")

		flag, err := client.Flipt().GetFlag(ctx, &flipt.GetFlagRequest{Key: "test"})
		require.NoError(t, err)

		assert.Equal(t, created, flag)

		t.Log("Update flag with key \"test\".")

		updated, err := client.Flipt().UpdateFlag(ctx, &flipt.UpdateFlagRequest{
			Key:  "test",
			Name: "Test 2",
		})
		require.NoError(t, err)

		assert.Equal(t, "Test 2", updated.Name)

		t.Log("List all flags.")

		flags, err := client.Flipt().ListFlags(ctx, &flipt.ListFlagRequest{})
		require.NoError(t, err)

		assert.Len(t, flags.Flags, 1)
		assert.Equal(t, updated.Key, flags.Flags[0].Key)
		assert.Equal(t, updated.Name, flags.Flags[0].Name)
		assert.Equal(t, updated.Description, flags.Flags[0].Description)

		for _, key := range []string{"one", "two"} {
			t.Logf("Create variant with key %q.", key)

			createdVariant, err := client.Flipt().CreateVariant(ctx, &flipt.CreateVariantRequest{
				Key:     key,
				Name:    key,
				FlagKey: "test",
			})
			require.NoError(t, err)

			assert.Equal(t, key, createdVariant.Key)
			assert.Equal(t, key, createdVariant.Name)
		}

		t.Log("Get flag \"test\" and check variants.")

		flag, err = client.Flipt().GetFlag(ctx, &flipt.GetFlagRequest{Key: "test"})
		require.NoError(t, err)

		assert.Len(t, flag.Variants, 2)

		t.Log(`Update variant "one" (rename from "one" to "One").`)

		updatedVariant, err := client.Flipt().UpdateVariant(ctx, &flipt.UpdateVariantRequest{
			FlagKey: "test",
			Id:      flag.Variants[0].Id,
			Key:     "one",
			Name:    "One",
		})
		require.NoError(t, err)

		assert.Equal(t, "one", updatedVariant.Key)
		assert.Equal(t, "One", updatedVariant.Name)
	})

	t.Run("Segments and Constraints", func(t *testing.T) {
		t.Log(`Create segment "everyone".`)

		createdSegment, err := client.Flipt().CreateSegment(ctx, &flipt.CreateSegmentRequest{
			Key:       "everyone",
			Name:      "Everyone",
			MatchType: flipt.MatchType_ALL_MATCH_TYPE,
		})
		require.NoError(t, err)

		assert.Equal(t, "everyone", createdSegment.Key)
		assert.Equal(t, "Everyone", createdSegment.Name)
		assert.Equal(t, flipt.MatchType_ALL_MATCH_TYPE, createdSegment.MatchType)

		t.Log(`Get segment "everyone".`)

		retrievedSegment, err := client.Flipt().GetSegment(ctx, &flipt.GetSegmentRequest{
			Key: "everyone",
		})
		require.NoError(t, err)

		assert.Equal(t, createdSegment.Key, retrievedSegment.Key)
		assert.Equal(t, createdSegment.Name, retrievedSegment.Name)
		assert.Equal(t, createdSegment.Description, retrievedSegment.Description)
		assert.Equal(t, createdSegment.MatchType, createdSegment.MatchType)

		t.Log(`Update segment "everyone" (rename "Everyone" to "All the peeps").`)

		updatedSegment, err := client.Flipt().UpdateSegment(ctx, &flipt.UpdateSegmentRequest{
			Key:  "everyone",
			Name: "All the peeps",
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
				SegmentKey: "everyone",
				Type:       flipt.ComparisonType_STRING_COMPARISON_TYPE,
				Property:   constraint.property,
				Operator:   constraint.operator,
				Value:      constraint.value,
			})
			require.NoError(t, err)

			assert.Equal(t, constraint.property, createdConstraint.Property)
			assert.Equal(t, constraint.operator, createdConstraint.Operator)
			assert.Equal(t, constraint.value, createdConstraint.Value)
		}

		t.Log(`Get segment "everyone" with constraints.`)

		retrievedSegment, err = client.Flipt().GetSegment(ctx, &flipt.GetSegmentRequest{
			Key: "everyone",
		})
		require.NoError(t, err)

		assert.Len(t, retrievedSegment.Constraints, 2)
		assert.Equal(t, "foo", retrievedSegment.Constraints[0].Property)
		assert.Equal(t, "fizz", retrievedSegment.Constraints[1].Property)

		t.Log(`Update constraint value (from "bar" to "baz").`)
		updatedConstraint, err := client.Flipt().UpdateConstraint(ctx, &flipt.UpdateConstraintRequest{
			SegmentKey: "everyone",
			Type:       retrievedSegment.Constraints[0].Type,
			Id:         retrievedSegment.Constraints[0].Id,
			Property:   retrievedSegment.Constraints[0].Property,
			Operator:   retrievedSegment.Constraints[0].Operator,
			Value:      "baz",
		})
		require.NoError(t, err)

		assert.Equal(t, flipt.ComparisonType_STRING_COMPARISON_TYPE, updatedConstraint.Type)
		assert.Equal(t, "foo", updatedConstraint.Property)
		assert.Equal(t, "eq", updatedConstraint.Operator)
		assert.Equal(t, "baz", updatedConstraint.Value)
	})

	t.Run("Rules and Distributions", func(*testing.T) {
		t.Log(`Create rule "rank 1".`)

		ruleOne, err := client.Flipt().CreateRule(ctx, &flipt.CreateRuleRequest{
			FlagKey:    "test",
			SegmentKey: "everyone",
			Rank:       1,
		})

		require.NoError(t, err)

		assert.Equal(t, "test", ruleOne.FlagKey)
		assert.Equal(t, "everyone", ruleOne.SegmentKey)
		assert.Equal(t, int32(1), ruleOne.Rank)

		t.Log(`List rules.`)

		allRules, err := client.Flipt().ListRules(ctx, &flipt.ListRuleRequest{
			FlagKey: "test",
		})
		require.NoError(t, err)

		assert.Len(t, allRules.Rules, 1)

		t.Log(`Get rules "rank 1".`)

		retrievedRule, err := client.Flipt().GetRule(ctx, &flipt.GetRuleRequest{
			FlagKey: "test",
			Id:      ruleOne.Id,
		})
		require.NoError(t, err)

		assert.Equal(t, "test", retrievedRule.FlagKey)
		assert.Equal(t, "everyone", retrievedRule.SegmentKey)
		assert.Equal(t, int32(1), retrievedRule.Rank)
	})
}
