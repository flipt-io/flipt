package db

import (
	"context"
	"testing"

	flipt "github.com/markphelps/flipt/rpc"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetEvaluationRules(t *testing.T) {
	flag, err := flagStore.CreateFlag(context.TODO(), &flipt.CreateFlagRequest{
		Key:         t.Name(),
		Name:        "foo",
		Description: "bar",
		Enabled:     true,
	})

	require.NoError(t, err)

	segment, err := segmentStore.CreateSegment(context.TODO(), &flipt.CreateSegmentRequest{
		Key:         t.Name(),
		Name:        "foo",
		Description: "bar",
		MatchType:   flipt.MatchType_ANY_MATCH_TYPE,
	})

	require.NoError(t, err)

	// constraint 1
	_, err = segmentStore.CreateConstraint(context.TODO(), &flipt.CreateConstraintRequest{
		SegmentKey: segment.Key,
		Type:       flipt.ComparisonType_STRING_COMPARISON_TYPE,
		Property:   "foo",
		Operator:   "EQ",
		Value:      "bar",
	})

	require.NoError(t, err)

	// constraint 2
	_, err = segmentStore.CreateConstraint(context.TODO(), &flipt.CreateConstraintRequest{
		SegmentKey: segment.Key,
		Type:       flipt.ComparisonType_STRING_COMPARISON_TYPE,
		Property:   "foz",
		Operator:   "EQ",
		Value:      "baz",
	})

	require.NoError(t, err)

	// rule rank 1
	rule1, err := ruleStore.CreateRule(context.TODO(), &flipt.CreateRuleRequest{
		FlagKey:    flag.Key,
		SegmentKey: segment.Key,
		Rank:       1,
	})

	require.NoError(t, err)

	// rule rank 2
	rule2, err := ruleStore.CreateRule(context.TODO(), &flipt.CreateRuleRequest{
		FlagKey:    flag.Key,
		SegmentKey: segment.Key,
		Rank:       2,
	})

	require.NoError(t, err)

	evaluationRules, err := evaluationStore.GetEvaluationRules(context.TODO(), flag.Key)
	require.NoError(t, err)

	assert.NotEmpty(t, evaluationRules)
	assert.Equal(t, 2, len(evaluationRules))

	assert.Equal(t, rule1.Id, evaluationRules[0].ID)
	assert.Equal(t, rule1.FlagKey, evaluationRules[0].FlagKey)
	assert.Equal(t, rule1.SegmentKey, evaluationRules[0].SegmentKey)
	assert.Equal(t, segment.MatchType, evaluationRules[0].SegmentMatchType)
	assert.Equal(t, rule1.Rank, evaluationRules[0].Rank)
	assert.Equal(t, 2, len(evaluationRules[0].Constraints))

	assert.Equal(t, rule2.Id, evaluationRules[1].ID)
	assert.Equal(t, rule2.FlagKey, evaluationRules[1].FlagKey)
	assert.Equal(t, rule2.SegmentKey, evaluationRules[1].SegmentKey)
	assert.Equal(t, segment.MatchType, evaluationRules[1].SegmentMatchType)
	assert.Equal(t, rule2.Rank, evaluationRules[1].Rank)
	assert.Equal(t, 2, len(evaluationRules[1].Constraints))
}

func TestGetEvaluationDistributions(t *testing.T) {
	flag, err := flagStore.CreateFlag(context.TODO(), &flipt.CreateFlagRequest{
		Key:         t.Name(),
		Name:        "foo",
		Description: "bar",
		Enabled:     true,
	})

	require.NoError(t, err)

	// variant 1
	variant1, err := flagStore.CreateVariant(context.TODO(), &flipt.CreateVariantRequest{
		FlagKey: flag.Key,
		Key:     "foo",
	})

	require.NoError(t, err)

	// variant 2
	variant2, err := flagStore.CreateVariant(context.TODO(), &flipt.CreateVariantRequest{
		FlagKey: flag.Key,
		Key:     "bar",
	})

	require.NoError(t, err)

	segment, err := segmentStore.CreateSegment(context.TODO(), &flipt.CreateSegmentRequest{
		Key:         t.Name(),
		Name:        "foo",
		Description: "bar",
		MatchType:   flipt.MatchType_ANY_MATCH_TYPE,
	})

	require.NoError(t, err)

	_, err = segmentStore.CreateConstraint(context.TODO(), &flipt.CreateConstraintRequest{
		SegmentKey: segment.Key,
		Type:       flipt.ComparisonType_STRING_COMPARISON_TYPE,
		Property:   "foo",
		Operator:   "EQ",
		Value:      "bar",
	})

	require.NoError(t, err)

	rule, err := ruleStore.CreateRule(context.TODO(), &flipt.CreateRuleRequest{
		FlagKey:    flag.Key,
		SegmentKey: segment.Key,
		Rank:       1,
	})

	require.NoError(t, err)

	// 50/50 distribution
	_, err = ruleStore.CreateDistribution(context.TODO(), &flipt.CreateDistributionRequest{
		FlagKey:   flag.Key,
		RuleId:    rule.Id,
		VariantId: variant1.Id,
		Rollout:   50.00,
	})

	require.NoError(t, err)

	_, err = ruleStore.CreateDistribution(context.TODO(), &flipt.CreateDistributionRequest{
		FlagKey:   flag.Key,
		RuleId:    rule.Id,
		VariantId: variant2.Id,
		Rollout:   50.00,
	})

	require.NoError(t, err)

	evaluationDistributions, err := evaluationStore.GetEvaluationDistributions(context.TODO(), rule.Id)
	require.NoError(t, err)

	assert.Equal(t, 2, len(evaluationDistributions))

	assert.NotEmpty(t, evaluationDistributions[0].ID)
	assert.Equal(t, rule.Id, evaluationDistributions[0].RuleID)
	assert.Equal(t, variant1.Id, evaluationDistributions[0].VariantID)
	assert.Equal(t, variant1.Key, evaluationDistributions[0].VariantKey)
	assert.Equal(t, float32(50.00), evaluationDistributions[0].Rollout)

	assert.NotEmpty(t, evaluationDistributions[1].ID)
	assert.Equal(t, rule.Id, evaluationDistributions[1].RuleID)
	assert.Equal(t, variant2.Id, evaluationDistributions[1].VariantID)
	assert.Equal(t, variant2.Key, evaluationDistributions[1].VariantKey)
	assert.Equal(t, float32(50.00), evaluationDistributions[1].Rollout)
}
