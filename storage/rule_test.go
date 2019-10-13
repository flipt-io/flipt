package storage

import (
	"context"
	"fmt"
	"sort"
	"testing"

	flipt "github.com/markphelps/flipt/rpc"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetRule(t *testing.T) {
	flag, err := flagStore.CreateFlag(context.TODO(), &flipt.CreateFlagRequest{
		Key:         t.Name(),
		Name:        "foo",
		Description: "bar",
		Enabled:     true,
	})

	require.NoError(t, err)
	assert.NotNil(t, flag)

	variant, err := flagStore.CreateVariant(context.TODO(), &flipt.CreateVariantRequest{
		FlagKey:     flag.Key,
		Key:         t.Name(),
		Name:        "foo",
		Description: "bar",
	})

	require.NoError(t, err)
	assert.NotNil(t, variant)

	segment, err := segmentStore.CreateSegment(context.TODO(), &flipt.CreateSegmentRequest{
		Key:         t.Name(),
		Name:        "foo",
		Description: "bar",
	})

	require.NoError(t, err)
	assert.NotNil(t, segment)

	rule, err := ruleStore.CreateRule(context.TODO(), &flipt.CreateRuleRequest{
		FlagKey:    flag.Key,
		SegmentKey: segment.Key,
		Rank:       1,
	})

	require.NoError(t, err)
	assert.NotNil(t, rule)

	got, err := ruleStore.GetRule(context.TODO(), &flipt.GetRuleRequest{
		Id:      rule.Id,
		FlagKey: flag.Key,
	})

	require.NoError(t, err)
	assert.NotNil(t, got)

	assert.Equal(t, rule.Id, got.Id)
	assert.Equal(t, rule.FlagKey, got.FlagKey)
	assert.Equal(t, rule.SegmentKey, got.SegmentKey)
	assert.Equal(t, rule.Rank, got.Rank)
	assert.NotZero(t, got.CreatedAt)
	assert.NotZero(t, got.UpdatedAt)

	_, err = ruleStore.GetRule(context.TODO(), &flipt.GetRuleRequest{
		Id:      "0",
		FlagKey: flag.Key,
	})

	require.Error(t, err)
}

func TestListRules(t *testing.T) {
	flag, err := flagStore.CreateFlag(context.TODO(), &flipt.CreateFlagRequest{
		Key:         t.Name(),
		Name:        "foo",
		Description: "bar",
		Enabled:     true,
	})

	require.NoError(t, err)
	assert.NotNil(t, flag)

	variant, err := flagStore.CreateVariant(context.TODO(), &flipt.CreateVariantRequest{
		FlagKey:     flag.Key,
		Key:         t.Name(),
		Name:        "foo",
		Description: "bar",
	})

	require.NoError(t, err)
	assert.NotNil(t, variant)

	segment, err := segmentStore.CreateSegment(context.TODO(), &flipt.CreateSegmentRequest{
		Key:         t.Name(),
		Name:        "foo",
		Description: "bar",
	})

	require.NoError(t, err)
	assert.NotNil(t, segment)

	reqs := []*flipt.CreateRuleRequest{
		{
			FlagKey:    flag.Key,
			SegmentKey: segment.Key,
			Rank:       1,
		},
		{
			FlagKey:    flag.Key,
			SegmentKey: segment.Key,
			Rank:       2,
		},
	}

	for _, req := range reqs {
		_, err := ruleStore.CreateRule(context.TODO(), req)
		require.NoError(t, err)
	}

	got, err := ruleStore.ListRules(context.TODO(), &flipt.ListRuleRequest{
		FlagKey: flag.Key,
	})

	require.NoError(t, err)
	assert.NotZero(t, len(got))
}

func TestListRulesPagination(t *testing.T) {
	flag, err := flagStore.CreateFlag(context.TODO(), &flipt.CreateFlagRequest{
		Key:         t.Name(),
		Name:        "foo",
		Description: "bar",
		Enabled:     true,
	})

	require.NoError(t, err)
	assert.NotNil(t, flag)

	variant, err := flagStore.CreateVariant(context.TODO(), &flipt.CreateVariantRequest{
		FlagKey:     flag.Key,
		Key:         t.Name(),
		Name:        "foo",
		Description: "bar",
	})

	require.NoError(t, err)
	assert.NotNil(t, variant)

	segment, err := segmentStore.CreateSegment(context.TODO(), &flipt.CreateSegmentRequest{
		Key:         t.Name(),
		Name:        "foo",
		Description: "bar",
	})

	require.NoError(t, err)
	assert.NotNil(t, segment)

	reqs := []*flipt.CreateRuleRequest{
		{
			FlagKey:    flag.Key,
			SegmentKey: segment.Key,
			Rank:       1,
		},
		{
			FlagKey:    flag.Key,
			SegmentKey: segment.Key,
			Rank:       2,
		},
	}

	for _, req := range reqs {
		_, err := ruleStore.CreateRule(context.TODO(), req)
		require.NoError(t, err)
	}

	got, err := ruleStore.ListRules(context.TODO(), &flipt.ListRuleRequest{
		FlagKey: flag.Key,
		Limit:   1,
		Offset:  1,
	})

	require.NoError(t, err)
	assert.Len(t, got, 1)
}

func TestCreateRuleAndDistribution(t *testing.T) {
	flag, err := flagStore.CreateFlag(context.TODO(), &flipt.CreateFlagRequest{
		Key:         t.Name(),
		Name:        "foo",
		Description: "bar",
		Enabled:     true,
	})

	require.NoError(t, err)
	assert.NotNil(t, flag)

	variant, err := flagStore.CreateVariant(context.TODO(), &flipt.CreateVariantRequest{
		FlagKey:     flag.Key,
		Key:         t.Name(),
		Name:        "foo",
		Description: "bar",
	})

	require.NoError(t, err)
	assert.NotNil(t, variant)

	segment, err := segmentStore.CreateSegment(context.TODO(), &flipt.CreateSegmentRequest{
		Key:         t.Name(),
		Name:        "foo",
		Description: "bar",
	})

	require.NoError(t, err)
	assert.NotNil(t, segment)

	rule, err := ruleStore.CreateRule(context.TODO(), &flipt.CreateRuleRequest{
		FlagKey:    flag.Key,
		SegmentKey: segment.Key,
		Rank:       1,
	})

	require.NoError(t, err)
	assert.NotNil(t, rule)

	assert.NotZero(t, rule.Id)
	assert.Equal(t, flag.Key, rule.FlagKey)
	assert.Equal(t, segment.Key, rule.SegmentKey)
	assert.Equal(t, int32(1), rule.Rank)
	assert.NotZero(t, rule.CreatedAt)
	assert.Equal(t, rule.CreatedAt.Seconds, rule.UpdatedAt.Seconds)

	distribution, err := ruleStore.CreateDistribution(context.TODO(), &flipt.CreateDistributionRequest{
		RuleId:    rule.Id,
		VariantId: variant.Id,
		Rollout:   100,
	})

	require.NoError(t, err)
	assert.NotZero(t, distribution.Id)
	assert.Equal(t, rule.Id, distribution.RuleId)
	assert.Equal(t, variant.Id, distribution.VariantId)
	assert.Equal(t, float32(100), distribution.Rollout)
	assert.NotZero(t, distribution.CreatedAt)
	assert.Equal(t, distribution.CreatedAt.Seconds, distribution.UpdatedAt.Seconds)
}

func TestCreateDistribution_NoRule(t *testing.T) {
	flag, err := flagStore.CreateFlag(context.TODO(), &flipt.CreateFlagRequest{
		Key:         t.Name(),
		Name:        "foo",
		Description: "bar",
		Enabled:     true,
	})

	require.NoError(t, err)
	assert.NotNil(t, flag)

	variant, err := flagStore.CreateVariant(context.TODO(), &flipt.CreateVariantRequest{
		FlagKey:     flag.Key,
		Key:         t.Name(),
		Name:        "foo",
		Description: "bar",
	})

	require.NoError(t, err)
	assert.NotNil(t, variant)

	_, err = ruleStore.CreateDistribution(context.TODO(), &flipt.CreateDistributionRequest{
		RuleId:    "foo",
		VariantId: variant.Id,
		Rollout:   100,
	})

	assert.EqualError(t, err, "rule \"foo\" not found")
}

func TestCreateRule_FlagNotFound(t *testing.T) {
	_, err := ruleStore.CreateRule(context.TODO(), &flipt.CreateRuleRequest{
		FlagKey:    "foo",
		SegmentKey: "bar",
		Rank:       1,
	})

	assert.EqualError(t, err, "flag \"foo\" or segment \"bar\" not found")
}

func TestCreateRule_SegmentNotFound(t *testing.T) {
	flag, err := flagStore.CreateFlag(context.TODO(), &flipt.CreateFlagRequest{
		Key:         t.Name(),
		Name:        "foo",
		Description: "bar",
		Enabled:     true,
	})

	require.NoError(t, err)
	assert.NotNil(t, flag)

	_, err = ruleStore.CreateRule(context.TODO(), &flipt.CreateRuleRequest{
		FlagKey:    flag.Key,
		SegmentKey: "foo",
		Rank:       1,
	})

	assert.EqualError(t, err, "flag \"TestCreateRule_SegmentNotFound\" or segment \"foo\" not found")
}

func TestUpdateRuleAndDistribution(t *testing.T) {
	flag, err := flagStore.CreateFlag(context.TODO(), &flipt.CreateFlagRequest{
		Key:         t.Name(),
		Name:        "foo",
		Description: "bar",
		Enabled:     true,
	})

	require.NoError(t, err)
	assert.NotNil(t, flag)

	variant, err := flagStore.CreateVariant(context.TODO(), &flipt.CreateVariantRequest{
		FlagKey:     flag.Key,
		Key:         t.Name(),
		Name:        "foo",
		Description: "bar",
	})

	require.NoError(t, err)
	assert.NotNil(t, variant)

	segmentOne, err := segmentStore.CreateSegment(context.TODO(), &flipt.CreateSegmentRequest{
		Key:         fmt.Sprintf("%s_one", t.Name()),
		Name:        "foo",
		Description: "bar",
	})

	require.NoError(t, err)
	assert.NotNil(t, segmentOne)

	rule, err := ruleStore.CreateRule(context.TODO(), &flipt.CreateRuleRequest{
		FlagKey:    flag.Key,
		SegmentKey: segmentOne.Key,
		Rank:       1,
	})

	require.NoError(t, err)
	assert.NotNil(t, rule)

	assert.NotZero(t, rule.Id)
	assert.Equal(t, flag.Key, rule.FlagKey)
	assert.Equal(t, segmentOne.Key, rule.SegmentKey)
	assert.Equal(t, int32(1), rule.Rank)
	assert.NotZero(t, rule.CreatedAt)
	assert.Equal(t, rule.CreatedAt.Seconds, rule.UpdatedAt.Seconds)

	distribution, err := ruleStore.CreateDistribution(context.TODO(), &flipt.CreateDistributionRequest{
		RuleId:    rule.Id,
		VariantId: variant.Id,
		Rollout:   100,
	})

	require.NoError(t, err)
	assert.NotZero(t, distribution.Id)
	assert.Equal(t, rule.Id, distribution.RuleId)
	assert.Equal(t, variant.Id, distribution.VariantId)
	assert.Equal(t, float32(100), distribution.Rollout)
	assert.NotZero(t, distribution.CreatedAt)
	assert.Equal(t, distribution.CreatedAt.Seconds, distribution.UpdatedAt.Seconds)

	segmentTwo, err := segmentStore.CreateSegment(context.TODO(), &flipt.CreateSegmentRequest{
		Key:         fmt.Sprintf("%s_two", t.Name()),
		Name:        "foo",
		Description: "bar",
	})

	require.NoError(t, err)
	assert.NotNil(t, segmentTwo)

	updatedRule, err := ruleStore.UpdateRule(context.TODO(), &flipt.UpdateRuleRequest{
		Id:         rule.Id,
		FlagKey:    flag.Key,
		SegmentKey: segmentTwo.Key,
	})

	require.NoError(t, err)
	assert.NotNil(t, updatedRule)

	assert.Equal(t, rule.Id, updatedRule.Id)
	assert.Equal(t, rule.FlagKey, updatedRule.FlagKey)
	assert.Equal(t, segmentTwo.Key, updatedRule.SegmentKey)
	assert.Equal(t, int32(1), updatedRule.Rank)
	assert.Equal(t, rule.CreatedAt.Seconds, updatedRule.CreatedAt.Seconds)
	assert.NotEqual(t, rule.CreatedAt, updatedRule.UpdatedAt)

	updatedDistribution, err := ruleStore.UpdateDistribution(context.TODO(), &flipt.UpdateDistributionRequest{
		Id:        distribution.Id,
		RuleId:    rule.Id,
		VariantId: variant.Id,
		Rollout:   10,
	})

	require.NoError(t, err)
	assert.Equal(t, distribution.Id, updatedDistribution.Id)
	assert.Equal(t, rule.Id, updatedDistribution.RuleId)
	assert.Equal(t, variant.Id, updatedDistribution.VariantId)
	assert.Equal(t, float32(10), updatedDistribution.Rollout)
	assert.Equal(t, distribution.CreatedAt.Seconds, updatedDistribution.CreatedAt.Seconds)
	assert.NotEqual(t, distribution.CreatedAt, updatedDistribution.UpdatedAt)

	err = ruleStore.DeleteDistribution(context.TODO(), &flipt.DeleteDistributionRequest{
		Id:        distribution.Id,
		RuleId:    rule.Id,
		VariantId: variant.Id,
	})
	require.NoError(t, err)
}

func TestUpdateRule_NotFound(t *testing.T) {
	flag, err := flagStore.CreateFlag(context.TODO(), &flipt.CreateFlagRequest{
		Key:         t.Name(),
		Name:        "foo",
		Description: "bar",
		Enabled:     true,
	})

	require.NoError(t, err)
	assert.NotNil(t, flag)

	segment, err := segmentStore.CreateSegment(context.TODO(), &flipt.CreateSegmentRequest{
		Key:         t.Name(),
		Name:        "foo",
		Description: "bar",
	})

	require.NoError(t, err)
	assert.NotNil(t, segment)

	_, err = ruleStore.UpdateRule(context.TODO(), &flipt.UpdateRuleRequest{
		Id:         "foo",
		FlagKey:    flag.Key,
		SegmentKey: segment.Key,
	})

	assert.EqualError(t, err, "rule \"foo\" not found")
}

func TestDeleteRule(t *testing.T) {
	flag, err := flagStore.CreateFlag(context.TODO(), &flipt.CreateFlagRequest{
		Key:         t.Name(),
		Name:        "foo",
		Description: "bar",
		Enabled:     true,
	})

	require.NoError(t, err)
	assert.NotNil(t, flag)

	variant, err := flagStore.CreateVariant(context.TODO(), &flipt.CreateVariantRequest{
		FlagKey:     flag.Key,
		Key:         t.Name(),
		Name:        "foo",
		Description: "bar",
	})

	require.NoError(t, err)
	assert.NotNil(t, variant)

	segment, err := segmentStore.CreateSegment(context.TODO(), &flipt.CreateSegmentRequest{
		Key:         t.Name(),
		Name:        "foo",
		Description: "bar",
	})

	require.NoError(t, err)
	assert.NotNil(t, segment)

	var rules []*flipt.Rule

	// create 3 rules
	for i := 0; i < 3; i++ {
		rule, err := ruleStore.CreateRule(context.TODO(), &flipt.CreateRuleRequest{
			FlagKey:    flag.Key,
			SegmentKey: segment.Key,
			Rank:       int32(i + 1),
		})

		require.NoError(t, err)
		assert.NotNil(t, rule)
		rules = append(rules, rule)
	}

	// delete second rule
	err = ruleStore.DeleteRule(context.TODO(), &flipt.DeleteRuleRequest{
		FlagKey: flag.Key,
		Id:      rules[1].Id,
	})

	require.NoError(t, err)

	got, err := ruleStore.ListRules(context.TODO(), &flipt.ListRuleRequest{
		FlagKey: flag.Key,
	})

	// ensure rules are in correct order
	require.NoError(t, err)
	assert.NotNil(t, got)
	assert.Equal(t, 2, len(got))
	assert.Equal(t, rules[0].Id, got[0].Id)
	assert.Equal(t, int32(1), got[0].Rank)
	assert.Equal(t, rules[2].Id, got[1].Id)
	assert.Equal(t, int32(2), got[1].Rank)
}

func TestDeleteRule_NotFound(t *testing.T) {
	flag, err := flagStore.CreateFlag(context.TODO(), &flipt.CreateFlagRequest{
		Key:         t.Name(),
		Name:        "foo",
		Description: "bar",
		Enabled:     true,
	})

	require.NoError(t, err)
	assert.NotNil(t, flag)

	err = ruleStore.DeleteRule(context.TODO(), &flipt.DeleteRuleRequest{
		Id:      "foo",
		FlagKey: flag.Key,
	})

	require.NoError(t, err)
}

func TestOrderRules(t *testing.T) {
	flag, err := flagStore.CreateFlag(context.TODO(), &flipt.CreateFlagRequest{
		Key:         t.Name(),
		Name:        "foo",
		Description: "bar",
		Enabled:     true,
	})

	require.NoError(t, err)
	assert.NotNil(t, flag)

	variant, err := flagStore.CreateVariant(context.TODO(), &flipt.CreateVariantRequest{
		FlagKey:     flag.Key,
		Key:         t.Name(),
		Name:        "foo",
		Description: "bar",
	})

	require.NoError(t, err)
	assert.NotNil(t, variant)

	segment, err := segmentStore.CreateSegment(context.TODO(), &flipt.CreateSegmentRequest{
		Key:         t.Name(),
		Name:        "foo",
		Description: "bar",
	})

	require.NoError(t, err)
	assert.NotNil(t, segment)

	var rules []*flipt.Rule

	// create 3 rules
	for i := 0; i < 3; i++ {
		rule, err := ruleStore.CreateRule(context.TODO(), &flipt.CreateRuleRequest{
			FlagKey:    flag.Key,
			SegmentKey: segment.Key,
			Rank:       int32(i + 1),
		})

		require.NoError(t, err)
		assert.NotNil(t, rule)
		rules = append(rules, rule)
	}

	// order rules in reverse order
	sort.Slice(rules, func(i, j int) bool { return rules[i].Rank > rules[j].Rank })

	var ruleIds []string
	for _, rule := range rules {
		ruleIds = append(ruleIds, rule.Id)
	}

	// re-order rules
	err = ruleStore.OrderRules(context.TODO(), &flipt.OrderRulesRequest{
		FlagKey: flag.Key,
		RuleIds: ruleIds,
	})

	require.NoError(t, err)

	got, err := ruleStore.ListRules(context.TODO(), &flipt.ListRuleRequest{
		FlagKey: flag.Key,
	})

	// ensure rules are in correct order
	require.NoError(t, err)
	assert.NotNil(t, got)
	assert.Equal(t, 3, len(got))
	assert.Equal(t, rules[0].Id, got[0].Id)
	assert.Equal(t, int32(1), got[0].Rank)
	assert.Equal(t, rules[1].Id, got[1].Id)
	assert.Equal(t, int32(2), got[1].Rank)
	assert.Equal(t, rules[2].Id, got[2].Id)
	assert.Equal(t, int32(3), got[2].Rank)
}
