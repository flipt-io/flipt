package sql_test

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.flipt.io/flipt/internal/storage"
	"go.flipt.io/flipt/internal/storage/sql/common"
	flipt "go.flipt.io/flipt/rpc/flipt"
)

func (s *DBTestSuite) TestGetRule() {
	t := s.T()

	flag, err := s.store.CreateFlag(context.TODO(), &flipt.CreateFlagRequest{
		Key:         t.Name(),
		Name:        "foo",
		Description: "bar",
		Enabled:     true,
	})

	require.NoError(t, err)
	assert.NotNil(t, flag)

	variant, err := s.store.CreateVariant(context.TODO(), &flipt.CreateVariantRequest{
		FlagKey:     flag.Key,
		Key:         t.Name(),
		Name:        "foo",
		Description: "bar",
	})

	require.NoError(t, err)
	assert.NotNil(t, variant)

	segment, err := s.store.CreateSegment(context.TODO(), &flipt.CreateSegmentRequest{
		Key:         t.Name(),
		Name:        "foo",
		Description: "bar",
	})

	require.NoError(t, err)
	assert.NotNil(t, segment)

	rule, err := s.store.CreateRule(context.TODO(), &flipt.CreateRuleRequest{
		FlagKey:    flag.Key,
		SegmentKey: segment.Key,
		Rank:       1,
	})

	require.NoError(t, err)
	assert.NotNil(t, rule)

	got, err := s.store.GetRule(context.TODO(), storage.DefaultNamespace, rule.Id)

	require.NoError(t, err)
	assert.NotNil(t, got)

	assert.Equal(t, rule.Id, got.Id)
	assert.Equal(t, storage.DefaultNamespace, got.NamespaceKey)
	assert.Equal(t, rule.FlagKey, got.FlagKey)
	assert.Equal(t, rule.SegmentKey, got.SegmentKey)
	assert.Equal(t, rule.Rank, got.Rank)
	assert.NotZero(t, got.CreatedAt)
	assert.NotZero(t, got.UpdatedAt)
}

func (s *DBTestSuite) TestGetRule_NotFound() {
	t := s.T()

	_, err := s.store.GetRule(context.TODO(), storage.DefaultNamespace, "0")
	assert.EqualError(t, err, "rule \"default/0\" not found")
}

func (s *DBTestSuite) TestListRules() {
	t := s.T()

	flag, err := s.store.CreateFlag(context.TODO(), &flipt.CreateFlagRequest{
		Key:         t.Name(),
		Name:        "foo",
		Description: "bar",
		Enabled:     true,
	})

	require.NoError(t, err)
	assert.NotNil(t, flag)

	variant, err := s.store.CreateVariant(context.TODO(), &flipt.CreateVariantRequest{
		FlagKey:     flag.Key,
		Key:         t.Name(),
		Name:        "foo",
		Description: "bar",
	})

	require.NoError(t, err)
	assert.NotNil(t, variant)

	segment, err := s.store.CreateSegment(context.TODO(), &flipt.CreateSegmentRequest{
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
		_, err := s.store.CreateRule(context.TODO(), req)
		require.NoError(t, err)
	}

	res, err := s.store.ListRules(context.TODO(), storage.DefaultNamespace, flag.Key)
	require.NoError(t, err)

	got := res.Results
	assert.NotZero(t, len(got))

	for _, rule := range got {
		assert.Equal(t, storage.DefaultNamespace, rule.NamespaceKey)
		assert.NotZero(t, rule.CreatedAt)
		assert.NotZero(t, rule.UpdatedAt)
	}
}

func (s *DBTestSuite) TestListRulesPagination_LimitOffset() {
	t := s.T()

	flag, err := s.store.CreateFlag(context.TODO(), &flipt.CreateFlagRequest{
		Key:         t.Name(),
		Name:        "foo",
		Description: "bar",
		Enabled:     true,
	})

	require.NoError(t, err)
	assert.NotNil(t, flag)

	variant, err := s.store.CreateVariant(context.TODO(), &flipt.CreateVariantRequest{
		FlagKey:     flag.Key,
		Key:         t.Name(),
		Name:        "foo",
		Description: "bar",
	})

	require.NoError(t, err)
	assert.NotNil(t, variant)

	segment, err := s.store.CreateSegment(context.TODO(), &flipt.CreateSegmentRequest{
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
		_, err := s.store.CreateRule(context.TODO(), req)
		require.NoError(t, err)
	}

	res, err := s.store.ListRules(context.TODO(), storage.DefaultNamespace, flag.Key, storage.WithLimit(1), storage.WithOffset(1))
	require.NoError(t, err)

	got := res.Results
	assert.Len(t, got, 1)
}

func (s *DBTestSuite) TestListRulesPagination_LimitWithNextPage() {
	t := s.T()

	flag, err := s.store.CreateFlag(context.TODO(), &flipt.CreateFlagRequest{
		Key:         t.Name(),
		Name:        "foo",
		Description: "bar",
		Enabled:     true,
	})

	require.NoError(t, err)
	assert.NotNil(t, flag)

	variant, err := s.store.CreateVariant(context.TODO(), &flipt.CreateVariantRequest{
		FlagKey:     flag.Key,
		Key:         t.Name(),
		Name:        "foo",
		Description: "bar",
	})

	require.NoError(t, err)
	assert.NotNil(t, variant)

	segment, err := s.store.CreateSegment(context.TODO(), &flipt.CreateSegmentRequest{
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
		_, err := s.store.CreateRule(context.TODO(), req)
		require.NoError(t, err)
	}

	// TODO: the ordering (DESC) is required because the default ordering is ASC and we are not clearing the DB between tests
	opts := []storage.QueryOption{storage.WithOrder(storage.OrderDesc), storage.WithLimit(1)}

	res, err := s.store.ListRules(context.TODO(), storage.DefaultNamespace, flag.Key, opts...)
	require.NoError(t, err)

	got := res.Results
	assert.Len(t, got, 1)
	assert.Equal(t, reqs[1].Rank, got[0].Rank)
	assert.NotEmpty(t, res.NextPageToken)

	pageToken := &common.PageToken{}
	err = json.Unmarshal([]byte(res.NextPageToken), pageToken)
	require.NoError(t, err)
	assert.NotEmpty(t, pageToken.Key)
	assert.Equal(t, uint64(1), pageToken.Offset)

	opts = append(opts, storage.WithPageToken(res.NextPageToken))

	res, err = s.store.ListRules(context.TODO(), storage.DefaultNamespace, flag.Key, opts...)
	require.NoError(t, err)

	got = res.Results
	assert.Len(t, got, 1)
	assert.Equal(t, reqs[0].Rank, got[0].Rank)
}

func (s *DBTestSuite) TestCreateRuleAndDistribution() {
	t := s.T()

	flag, err := s.store.CreateFlag(context.TODO(), &flipt.CreateFlagRequest{
		Key:         t.Name(),
		Name:        "foo",
		Description: "bar",
		Enabled:     true,
	})

	require.NoError(t, err)
	assert.NotNil(t, flag)

	variant, err := s.store.CreateVariant(context.TODO(), &flipt.CreateVariantRequest{
		FlagKey:     flag.Key,
		Key:         t.Name(),
		Name:        "foo",
		Description: "bar",
	})

	require.NoError(t, err)
	assert.NotNil(t, variant)

	segment, err := s.store.CreateSegment(context.TODO(), &flipt.CreateSegmentRequest{
		Key:         t.Name(),
		Name:        "foo",
		Description: "bar",
	})

	require.NoError(t, err)
	assert.NotNil(t, segment)

	rule, err := s.store.CreateRule(context.TODO(), &flipt.CreateRuleRequest{
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

	distribution, err := s.store.CreateDistribution(context.TODO(), &flipt.CreateDistributionRequest{
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

	// TODO: create a rule with a different namespace
	// rule, err := s.store.CreateRule(context.TODO(), &flipt.CreateRuleRequest{
	// 	NamespaceKey: "foo",
	// 	FlagKey:    flag.Key,
	// 	SegmentKey: segment.Key,
	// 	Rank:       1,
	// })

	// require.NoError(t, err)
	// assert.NotNil(t, rule)

	// assert.NotZero(t, rule.Id)
	// assert.Equal(t, "foo", rule.NamespaceKey)
	// assert.Equal(t, flag.Key, rule.FlagKey)
	// assert.Equal(t, segment.Key, rule.SegmentKey)
	// assert.Equal(t, int32(1), rule.Rank)
	// assert.NotZero(t, rule.CreatedAt)
	// assert.Equal(t, rule.CreatedAt.Seconds, rule.UpdatedAt.Seconds)
}

func (s *DBTestSuite) TestCreateDistribution_NoRule() {
	t := s.T()

	flag, err := s.store.CreateFlag(context.TODO(), &flipt.CreateFlagRequest{
		Key:         t.Name(),
		Name:        "foo",
		Description: "bar",
		Enabled:     true,
	})

	require.NoError(t, err)
	assert.NotNil(t, flag)

	variant, err := s.store.CreateVariant(context.TODO(), &flipt.CreateVariantRequest{
		FlagKey:     flag.Key,
		Key:         t.Name(),
		Name:        "foo",
		Description: "bar",
	})

	require.NoError(t, err)
	assert.NotNil(t, variant)

	_, err = s.store.CreateDistribution(context.TODO(), &flipt.CreateDistributionRequest{
		RuleId:    "foo",
		VariantId: variant.Id,
		Rollout:   100,
	})

	assert.EqualError(t, err, "rule \"foo\" not found")
}

func (s *DBTestSuite) TestCreateRule_FlagNotFound() {
	t := s.T()

	_, err := s.store.CreateRule(context.TODO(), &flipt.CreateRuleRequest{
		FlagKey:    "foo",
		SegmentKey: "bar",
		Rank:       1,
	})

	assert.EqualError(t, err, "flag \"default/foo\" or segment \"default/bar\" not found")
}

func (s *DBTestSuite) TestCreateRule_SegmentNotFound() {
	t := s.T()

	flag, err := s.store.CreateFlag(context.TODO(), &flipt.CreateFlagRequest{
		Key:         t.Name(),
		Name:        "foo",
		Description: "bar",
		Enabled:     true,
	})

	require.NoError(t, err)
	assert.NotNil(t, flag)

	_, err = s.store.CreateRule(context.TODO(), &flipt.CreateRuleRequest{
		FlagKey:    flag.Key,
		SegmentKey: "foo",
		Rank:       1,
	})

	assert.EqualError(t, err, "flag \"default/TestDBTestSuite/TestCreateRule_SegmentNotFound\" or segment \"default/foo\" not found")
}

func (s *DBTestSuite) TestUpdateRuleAndDistribution() {
	t := s.T()

	flag, err := s.store.CreateFlag(context.TODO(), &flipt.CreateFlagRequest{
		Key:         t.Name(),
		Name:        "foo",
		Description: "bar",
		Enabled:     true,
	})

	require.NoError(t, err)
	assert.NotNil(t, flag)

	variant, err := s.store.CreateVariant(context.TODO(), &flipt.CreateVariantRequest{
		FlagKey:     flag.Key,
		Key:         t.Name(),
		Name:        "foo",
		Description: "bar",
	})

	require.NoError(t, err)
	assert.NotNil(t, variant)

	segmentOne, err := s.store.CreateSegment(context.TODO(), &flipt.CreateSegmentRequest{
		Key:         fmt.Sprintf("%s_one", t.Name()),
		Name:        "foo",
		Description: "bar",
	})

	require.NoError(t, err)
	assert.NotNil(t, segmentOne)

	rule, err := s.store.CreateRule(context.TODO(), &flipt.CreateRuleRequest{
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

	distribution, err := s.store.CreateDistribution(context.TODO(), &flipt.CreateDistributionRequest{
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

	segmentTwo, err := s.store.CreateSegment(context.TODO(), &flipt.CreateSegmentRequest{
		Key:         fmt.Sprintf("%s_two", t.Name()),
		Name:        "foo",
		Description: "bar",
	})

	require.NoError(t, err)
	assert.NotNil(t, segmentTwo)

	updatedRule, err := s.store.UpdateRule(context.TODO(), &flipt.UpdateRuleRequest{
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
	// assert.Equal(t, rule.CreatedAt.Seconds, updatedRule.CreatedAt.Seconds)
	assert.NotZero(t, rule.UpdatedAt)

	updatedDistribution, err := s.store.UpdateDistribution(context.TODO(), &flipt.UpdateDistributionRequest{
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
	// assert.Equal(t, distribution.CreatedAt.Seconds, updatedDistribution.CreatedAt.Seconds)
	assert.NotZero(t, rule.UpdatedAt)

	err = s.store.DeleteDistribution(context.TODO(), &flipt.DeleteDistributionRequest{
		Id:        distribution.Id,
		RuleId:    rule.Id,
		VariantId: variant.Id,
	})
	require.NoError(t, err)
}

func (s *DBTestSuite) TestUpdateRule_NotFound() {
	t := s.T()

	flag, err := s.store.CreateFlag(context.TODO(), &flipt.CreateFlagRequest{
		Key:         t.Name(),
		Name:        "foo",
		Description: "bar",
		Enabled:     true,
	})

	require.NoError(t, err)
	assert.NotNil(t, flag)

	segment, err := s.store.CreateSegment(context.TODO(), &flipt.CreateSegmentRequest{
		Key:         t.Name(),
		Name:        "foo",
		Description: "bar",
	})

	require.NoError(t, err)
	assert.NotNil(t, segment)

	_, err = s.store.UpdateRule(context.TODO(), &flipt.UpdateRuleRequest{
		Id:         "foo",
		FlagKey:    flag.Key,
		SegmentKey: segment.Key,
	})

	assert.EqualError(t, err, "rule \"default/foo\" not found")
}

func (s *DBTestSuite) TestDeleteRule() {
	t := s.T()

	flag, err := s.store.CreateFlag(context.TODO(), &flipt.CreateFlagRequest{
		Key:         t.Name(),
		Name:        "foo",
		Description: "bar",
		Enabled:     true,
	})

	require.NoError(t, err)
	assert.NotNil(t, flag)

	variant, err := s.store.CreateVariant(context.TODO(), &flipt.CreateVariantRequest{
		FlagKey:     flag.Key,
		Key:         t.Name(),
		Name:        "foo",
		Description: "bar",
	})

	require.NoError(t, err)
	assert.NotNil(t, variant)

	segment, err := s.store.CreateSegment(context.TODO(), &flipt.CreateSegmentRequest{
		Key:         t.Name(),
		Name:        "foo",
		Description: "bar",
	})

	require.NoError(t, err)
	assert.NotNil(t, segment)

	var rules []*flipt.Rule

	// create 3 rules
	for i := 0; i < 3; i++ {
		rule, err := s.store.CreateRule(context.TODO(), &flipt.CreateRuleRequest{
			FlagKey:    flag.Key,
			SegmentKey: segment.Key,
			Rank:       int32(i + 1),
		})

		require.NoError(t, err)
		assert.NotNil(t, rule)
		rules = append(rules, rule)
	}

	// delete second rule
	err = s.store.DeleteRule(context.TODO(), &flipt.DeleteRuleRequest{
		FlagKey: flag.Key,
		Id:      rules[1].Id,
	})

	require.NoError(t, err)

	res, err := s.store.ListRules(context.TODO(), storage.DefaultNamespace, flag.Key)
	// ensure rules are in correct order
	require.NoError(t, err)

	got := res.Results
	assert.NotNil(t, got)
	assert.Equal(t, 2, len(got))
	assert.Equal(t, rules[0].Id, got[0].Id)
	assert.Equal(t, int32(1), got[0].Rank)
	assert.Equal(t, storage.DefaultNamespace, got[0].NamespaceKey)
	assert.Equal(t, rules[2].Id, got[1].Id)
	assert.Equal(t, int32(2), got[1].Rank)
	assert.Equal(t, storage.DefaultNamespace, got[1].NamespaceKey)
}

func (s *DBTestSuite) TestDeleteRule_NotFound() {
	t := s.T()

	flag, err := s.store.CreateFlag(context.TODO(), &flipt.CreateFlagRequest{
		Key:         t.Name(),
		Name:        "foo",
		Description: "bar",
		Enabled:     true,
	})

	require.NoError(t, err)
	assert.NotNil(t, flag)

	err = s.store.DeleteRule(context.TODO(), &flipt.DeleteRuleRequest{
		Id:      "foo",
		FlagKey: flag.Key,
	})

	require.NoError(t, err)
}

func (s *DBTestSuite) TestOrderRules() {
	t := s.T()

	flag, err := s.store.CreateFlag(context.TODO(), &flipt.CreateFlagRequest{
		Key:         t.Name(),
		Name:        "foo",
		Description: "bar",
		Enabled:     true,
	})

	require.NoError(t, err)
	assert.NotNil(t, flag)

	variant, err := s.store.CreateVariant(context.TODO(), &flipt.CreateVariantRequest{
		FlagKey:     flag.Key,
		Key:         t.Name(),
		Name:        "foo",
		Description: "bar",
	})

	require.NoError(t, err)
	assert.NotNil(t, variant)

	segment, err := s.store.CreateSegment(context.TODO(), &flipt.CreateSegmentRequest{
		Key:         t.Name(),
		Name:        "foo",
		Description: "bar",
	})

	require.NoError(t, err)
	assert.NotNil(t, segment)

	var rules []*flipt.Rule

	// create 3 rules
	for i := 0; i < 3; i++ {
		rule, err := s.store.CreateRule(context.TODO(), &flipt.CreateRuleRequest{
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
	err = s.store.OrderRules(context.TODO(), &flipt.OrderRulesRequest{
		FlagKey: flag.Key,
		RuleIds: ruleIds,
	})

	require.NoError(t, err)

	res, err := s.store.ListRules(context.TODO(), storage.DefaultNamespace, flag.Key)

	// ensure rules are in correct order
	require.NoError(t, err)
	got := res.Results
	assert.NotNil(t, got)
	assert.Equal(t, 3, len(got))

	assert.Equal(t, rules[0].Id, got[0].Id)
	assert.Equal(t, int32(1), got[0].Rank)
	assert.Equal(t, storage.DefaultNamespace, got[0].NamespaceKey)

	assert.Equal(t, rules[1].Id, got[1].Id)
	assert.Equal(t, int32(2), got[1].Rank)
	assert.Equal(t, storage.DefaultNamespace, got[1].NamespaceKey)

	assert.Equal(t, rules[2].Id, got[2].Id)
	assert.Equal(t, int32(3), got[2].Rank)
	assert.Equal(t, storage.DefaultNamespace, got[2].NamespaceKey)
}
