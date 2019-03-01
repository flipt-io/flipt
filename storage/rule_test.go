package storage

import (
	"context"
	"fmt"
	"sort"
	"testing"

	flipt "github.com/markphelps/flipt/proto"
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

	got, err = ruleStore.GetRule(context.TODO(), &flipt.GetRuleRequest{
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
	assert.Equal(t, rule.CreatedAt, rule.UpdatedAt)

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
	assert.Equal(t, distribution.CreatedAt, distribution.UpdatedAt)
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
	assert.Equal(t, rule.CreatedAt, rule.UpdatedAt)

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
	assert.Equal(t, distribution.CreatedAt, distribution.UpdatedAt)

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
	assert.Equal(t, rule.CreatedAt, updatedRule.CreatedAt)
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
	assert.Equal(t, distribution.CreatedAt, updatedDistribution.CreatedAt)
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

func TestEvaluate_flagNotFound(t *testing.T) {
	_, err := ruleStore.Evaluate(context.TODO(), &flipt.EvaluationRequest{
		FlagKey: "foo",
		Context: map[string]string{
			"bar": "boz",
		},
	})
	require.Error(t, err)
	assert.EqualError(t, err, "flag \"foo\" not found")
}

func TestEvaluate_flagDisabled(t *testing.T) {
	flag, err := flagStore.CreateFlag(context.TODO(), &flipt.CreateFlagRequest{
		Key:         t.Name(),
		Name:        "foo",
		Description: "bar",
		Enabled:     false,
	})

	require.NoError(t, err)

	_, err = ruleStore.Evaluate(context.TODO(), &flipt.EvaluationRequest{
		FlagKey:  flag.Key,
		EntityId: "1",
		Context: map[string]string{
			"bar": "boz",
		},
	})

	require.Error(t, err)
	assert.EqualError(t, err, "flag \"TestEvaluate_flagDisabled\" is disabled")
}

func TestEvaluate_flagNoRules(t *testing.T) {
	flag, err := flagStore.CreateFlag(context.TODO(), &flipt.CreateFlagRequest{
		Key:         t.Name(),
		Name:        "foo",
		Description: "bar",
		Enabled:     true,
	})

	require.NoError(t, err)

	_, err = flagStore.CreateVariant(context.TODO(), &flipt.CreateVariantRequest{
		FlagKey:     flag.Key,
		Key:         t.Name(),
		Name:        "foo",
		Description: "bar",
	})

	require.NoError(t, err)

	resp, err := ruleStore.Evaluate(context.TODO(), &flipt.EvaluationRequest{
		FlagKey:  flag.Key,
		EntityId: "1",
		Context: map[string]string{
			"bar": "boz",
		},
	})

	require.NoError(t, err)
	assert.False(t, resp.Match)
}

func TestEvaluate_singleVariantDistribution(t *testing.T) {
	flag, err := flagStore.CreateFlag(context.TODO(), &flipt.CreateFlagRequest{
		Key:         t.Name(),
		Name:        t.Name(),
		Description: "foo",
		Enabled:     true,
	})

	require.NoError(t, err)

	var variants []*flipt.Variant

	for _, req := range []*flipt.CreateVariantRequest{
		{
			FlagKey: flag.Key,
			Key:     fmt.Sprintf("foo_%s", t.Name()),
		},
		{
			FlagKey: flag.Key,
			Key:     fmt.Sprintf("bar_%s", t.Name()),
		},
	} {
		variant, err := flagStore.CreateVariant(context.TODO(), req)
		require.NoError(t, err)
		variants = append(variants, variant)
	}

	segment, err := segmentStore.CreateSegment(context.TODO(), &flipt.CreateSegmentRequest{
		Key:         t.Name(),
		Name:        t.Name(),
		Description: "foo",
	})

	require.NoError(t, err)

	_, err = segmentStore.CreateConstraint(context.TODO(), &flipt.CreateConstraintRequest{
		SegmentKey: segment.Key,
		Type:       flipt.ComparisonType_STRING_COMPARISON_TYPE,
		Property:   "bar",
		Operator:   opEQ,
		Value:      "baz",
	})

	require.NoError(t, err)

	rule, err := ruleStore.CreateRule(context.TODO(), &flipt.CreateRuleRequest{
		FlagKey:    flag.Key,
		SegmentKey: segment.Key,
	})

	require.NoError(t, err)

	_, err = ruleStore.CreateDistribution(context.TODO(), &flipt.CreateDistributionRequest{
		FlagKey:   flag.Key,
		RuleId:    rule.Id,
		VariantId: variants[0].Id,
		Rollout:   100,
	})

	require.NoError(t, err)

	tests := []struct {
		name      string
		req       *flipt.EvaluationRequest
		wantMatch bool
	}{
		{
			name: "match string value",
			req: &flipt.EvaluationRequest{
				FlagKey:  flag.Key,
				EntityId: "1",
				Context: map[string]string{
					"bar": "baz",
				},
			},
			wantMatch: true,
		},
		{
			name: "no match string value",
			req: &flipt.EvaluationRequest{
				FlagKey:  flag.Key,
				EntityId: "1",
				Context: map[string]string{
					"bar": "boz",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := ruleStore.Evaluate(context.TODO(), tt.req)
			require.NoError(t, err)
			assert.NotNil(t, resp)
			assert.Equal(t, flag.Key, resp.FlagKey)
			assert.Equal(t, tt.req.Context, resp.RequestContext)

			if !tt.wantMatch {
				assert.False(t, resp.Match)
				assert.Empty(t, resp.SegmentKey)
				return
			}

			assert.True(t, resp.Match)
			assert.Equal(t, segment.Key, resp.SegmentKey)
			assert.Equal(t, variants[0].Key, resp.Value)
		})
	}
}

func TestEvaluate_rolloutDistribution(t *testing.T) {
	flag, err := flagStore.CreateFlag(context.TODO(), &flipt.CreateFlagRequest{
		Key:         t.Name(),
		Name:        t.Name(),
		Description: "foo",
		Enabled:     true,
	})

	require.NoError(t, err)

	var variants []*flipt.Variant

	for _, req := range []*flipt.CreateVariantRequest{
		{
			FlagKey: flag.Key,
			Key:     fmt.Sprintf("foo_%s", t.Name()),
		},
		{
			FlagKey: flag.Key,
			Key:     fmt.Sprintf("bar_%s", t.Name()),
		},
	} {
		variant, err := flagStore.CreateVariant(context.TODO(), req)
		require.NoError(t, err)
		variants = append(variants, variant)
	}

	segment, err := segmentStore.CreateSegment(context.TODO(), &flipt.CreateSegmentRequest{
		Key:         t.Name(),
		Name:        t.Name(),
		Description: "foo",
	})

	require.NoError(t, err)

	_, err = segmentStore.CreateConstraint(context.TODO(), &flipt.CreateConstraintRequest{
		SegmentKey: segment.Key,
		Type:       flipt.ComparisonType_STRING_COMPARISON_TYPE,
		Property:   "bar",
		Operator:   opEQ,
		Value:      "baz",
	})

	require.NoError(t, err)

	rule, err := ruleStore.CreateRule(context.TODO(), &flipt.CreateRuleRequest{
		FlagKey:    flag.Key,
		SegmentKey: segment.Key,
	})

	require.NoError(t, err)

	for _, req := range []*flipt.CreateDistributionRequest{
		{
			FlagKey:   flag.Key,
			RuleId:    rule.Id,
			VariantId: variants[0].Id,
			Rollout:   50,
		},
		{
			FlagKey:   flag.Key,
			RuleId:    rule.Id,
			VariantId: variants[1].Id,
			Rollout:   50,
		},
	} {
		_, err := ruleStore.CreateDistribution(context.TODO(), req)
		require.NoError(t, err)
	}

	tests := []struct {
		name              string
		req               *flipt.EvaluationRequest
		matchesVariantKey string
		wantMatch         bool
	}{
		{
			name: "match string value - variant 1",
			req: &flipt.EvaluationRequest{
				FlagKey:  flag.Key,
				EntityId: "1",
				Context: map[string]string{
					"bar": "baz",
				},
			},
			matchesVariantKey: variants[0].Key,
			wantMatch:         true,
		},
		{
			name: "match string value - variant 2",
			req: &flipt.EvaluationRequest{
				FlagKey:  flag.Key,
				EntityId: "100",
				Context: map[string]string{
					"bar": "baz",
				},
			},
			matchesVariantKey: variants[1].Key,
			wantMatch:         true,
		},
		{
			name: "no match string value",
			req: &flipt.EvaluationRequest{
				FlagKey:  flag.Key,
				EntityId: "1",
				Context: map[string]string{
					"bar": "boz",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := ruleStore.Evaluate(context.TODO(), tt.req)
			require.NoError(t, err)
			assert.NotNil(t, resp)
			assert.Equal(t, flag.Key, resp.FlagKey)
			assert.Equal(t, tt.req.Context, resp.RequestContext)

			if !tt.wantMatch {
				assert.False(t, resp.Match)
				assert.Empty(t, resp.SegmentKey)
				return
			}

			assert.True(t, resp.Match)
			assert.Equal(t, segment.Key, resp.SegmentKey)
			assert.Equal(t, tt.matchesVariantKey, resp.Value)
		})
	}
}

func Test_validateConstraint(t *testing.T) {
	tests := []struct {
		name       string
		constraint constraint
		wantErr    bool
	}{
		{
			name: "missing property",
			constraint: constraint{
				Operator: "eq",
				Value:    "bar",
			},
			wantErr: true,
		},
		{
			name: "missing operator",
			constraint: constraint{
				Property: "foo",
				Value:    "bar",
			},
			wantErr: true,
		},
		{
			name: "invalid operator",
			constraint: constraint{
				Property: "foo",
				Operator: "?",
				Value:    "bar",
			},
			wantErr: true,
		},
		{
			name: "valid",
			constraint: constraint{
				Property: "foo",
				Operator: "eq",
				Value:    "bar",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validate(tt.constraint)

			if tt.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
		})
	}
}

func Test_matchesString(t *testing.T) {
	tests := []struct {
		name       string
		constraint constraint
		value      string
		wantMatch  bool
		wantErr    bool
	}{
		{
			name: "eq",
			constraint: constraint{
				Property: "foo",
				Operator: "eq",
				Value:    "bar",
			},
			value:     "bar",
			wantMatch: true,
		},
		{
			name: "negative eq",
			constraint: constraint{
				Property: "foo",
				Operator: "eq",
				Value:    "bar",
			},
			value: "baz",
		},
		{
			name: "neq",
			constraint: constraint{
				Property: "foo",
				Operator: "neq",
				Value:    "bar",
			},
			value:     "baz",
			wantMatch: true,
		},
		{
			name: "negative neq",
			constraint: constraint{
				Property: "foo",
				Operator: "neq",
				Value:    "bar",
			},
			value: "bar",
		},
		{
			name: "empty",
			constraint: constraint{
				Property: "foo",
				Operator: "empty",
			},
			value:     " ",
			wantMatch: true,
		},
		{
			name: "negative empty",
			constraint: constraint{
				Property: "foo",
				Operator: "empty",
			},
			value: "bar",
		},
		{
			name: "not empty",
			constraint: constraint{
				Property: "foo",
				Operator: "notempty",
			},
			value:     "bar",
			wantMatch: true,
		},
		{
			name: "negative not empty",
			constraint: constraint{
				Property: "foo",
				Operator: "notempty",
			},
			value: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			match, err := matchesString(tt.constraint, tt.value)

			if tt.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.wantMatch, match)
		})
	}
}

func Test_matchesNumber(t *testing.T) {
	tests := []struct {
		name       string
		constraint constraint
		value      string
		wantMatch  bool
		wantErr    bool
	}{
		{
			name: "present",
			constraint: constraint{
				Property: "foo",
				Operator: "present",
			},
			value:     "1",
			wantMatch: true,
		},
		{
			name: "negative present",
			constraint: constraint{
				Property: "foo",
				Operator: "present",
			},
		},
		{
			name: "not present",
			constraint: constraint{
				Property: "foo",
				Operator: "notpresent",
			},
			wantMatch: true,
		},
		{
			name: "negative notpresent",
			constraint: constraint{
				Property: "foo",
				Operator: "notpresent",
			},
			value: "1",
		},
		{
			name: "NAN constraint value",
			constraint: constraint{
				Property: "foo",
				Operator: "eq",
				Value:    "bar",
			},
			value:   "5",
			wantErr: true,
		},
		{
			name: "NAN context value",
			constraint: constraint{
				Property: "foo",
				Operator: "eq",
				Value:    "5",
			},
			value:   "foo",
			wantErr: true,
		},
		{
			name: "eq",
			constraint: constraint{
				Property: "foo",
				Operator: "eq",
				Value:    "42.0",
			},
			value:     "42.0",
			wantMatch: true,
		},
		{
			name: "negative eq",
			constraint: constraint{
				Property: "foo",
				Operator: "eq",
				Value:    "42.0",
			},
			value: "50",
		},
		{
			name: "neq",
			constraint: constraint{
				Property: "foo",
				Operator: "neq",
				Value:    "42.0",
			},
			value:     "50",
			wantMatch: true,
		},
		{
			name: "negative neq",
			constraint: constraint{
				Property: "foo",
				Operator: "neq",
				Value:    "42.0",
			},
			value: "42.0",
		},
		{
			name: "lt",
			constraint: constraint{
				Property: "foo",
				Operator: "lt",
				Value:    "42.0",
			},
			value:     "8",
			wantMatch: true,
		},
		{
			name: "negative lt",
			constraint: constraint{
				Property: "foo",
				Operator: "lt",
				Value:    "42.0",
			},
			value: "50",
		},
		{
			name: "lte",
			constraint: constraint{
				Property: "foo",
				Operator: "lte",
				Value:    "42.0",
			},
			value:     "42.0",
			wantMatch: true,
		},
		{
			name: "negative lte",
			constraint: constraint{
				Property: "foo",
				Operator: "lte",
				Value:    "42.0",
			},
			value: "102.0",
		},
		{
			name: "gt",
			constraint: constraint{
				Property: "foo",
				Operator: "gt",
				Value:    "10.11",
			},
			value:     "10.12",
			wantMatch: true,
		},
		{
			name: "negative gt",
			constraint: constraint{
				Property: "foo",
				Operator: "gt",
				Value:    "10.11",
			},
			value: "1",
		},
		{
			name: "gte",
			constraint: constraint{
				Property: "foo",
				Operator: "gte",
				Value:    "10.11",
			},
			value:     "10.11",
			wantMatch: true,
		},
		{
			name: "negative gte",
			constraint: constraint{
				Property: "foo",
				Operator: "gte",
				Value:    "10.11",
			},
			value: "0.11",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			match, err := matchesNumber(tt.constraint, tt.value)

			if tt.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.wantMatch, match)
		})
	}
}

func Test_matchesBool(t *testing.T) {
	tests := []struct {
		name       string
		constraint constraint
		value      string
		wantMatch  bool
		wantErr    bool
	}{
		{
			name: "present",
			constraint: constraint{
				Property: "foo",
				Operator: "present",
			},
			value:     "true",
			wantMatch: true,
		},
		{
			name: "negative present",
			constraint: constraint{
				Property: "foo",
				Operator: "present",
			},
		},
		{
			name: "not present",
			constraint: constraint{
				Property: "foo",
				Operator: "notpresent",
			},
			wantMatch: true,
		},
		{
			name: "negative notpresent",
			constraint: constraint{
				Property: "foo",
				Operator: "notpresent",
			},
			value: "true",
		},
		{
			name: "not a bool",
			constraint: constraint{
				Property: "foo",
				Operator: "true",
			},
			value:   "foo",
			wantErr: true,
		},
		{
			name: "is true",
			constraint: constraint{
				Property: "foo",
				Operator: "true",
			},
			value:     "true",
			wantMatch: true,
		},
		{
			name: "negative is true",
			constraint: constraint{
				Property: "foo",
				Operator: "true",
			},
			value: "false",
		},
		{
			name: "is false",
			constraint: constraint{
				Property: "foo",
				Operator: "false",
			},
			value:     "false",
			wantMatch: true,
		},
		{
			name: "negative is false",
			constraint: constraint{
				Property: "foo",
				Operator: "false",
			},
			value: "true",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			match, err := matchesBool(tt.constraint, tt.value)

			if tt.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.wantMatch, match)
		})
	}
}
