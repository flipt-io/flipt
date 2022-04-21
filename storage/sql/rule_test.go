package sql

import (
	"context"
	"fmt"
	"sort"
	"testing"
	"time"

	flipt "github.com/markphelps/flipt/rpc/flipt"
	"github.com/markphelps/flipt/storage"
	"github.com/markphelps/flipt/storage/cache"
	"github.com/markphelps/flipt/storage/cache/memory"
	"github.com/sirupsen/logrus"
	"github.com/sirupsen/logrus/hooks/test"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetRule(t *testing.T) {
	flag, err := store.CreateFlag(context.TODO(), &flipt.CreateFlagRequest{
		Key:         t.Name(),
		Name:        "foo",
		Description: "bar",
		Enabled:     true,
	})

	require.NoError(t, err)
	assert.NotNil(t, flag)

	variant, err := store.CreateVariant(context.TODO(), &flipt.CreateVariantRequest{
		FlagKey:     flag.Key,
		Key:         t.Name(),
		Name:        "foo",
		Description: "bar",
	})

	require.NoError(t, err)
	assert.NotNil(t, variant)

	segment, err := store.CreateSegment(context.TODO(), &flipt.CreateSegmentRequest{
		Key:         t.Name(),
		Name:        "foo",
		Description: "bar",
	})

	require.NoError(t, err)
	assert.NotNil(t, segment)

	rule, err := store.CreateRule(context.TODO(), &flipt.CreateRuleRequest{
		FlagKey:    flag.Key,
		SegmentKey: segment.Key,
		Rank:       1,
	})

	require.NoError(t, err)
	assert.NotNil(t, rule)

	got, err := store.GetRule(context.TODO(), rule.Id)

	require.NoError(t, err)
	assert.NotNil(t, got)

	assert.Equal(t, rule.Id, got.Id)
	assert.Equal(t, rule.FlagKey, got.FlagKey)
	assert.Equal(t, rule.SegmentKey, got.SegmentKey)
	assert.Equal(t, rule.Rank, got.Rank)
	assert.NotZero(t, got.CreatedAt)
	assert.NotZero(t, got.UpdatedAt)
}

func TestGetRule_NotFound(t *testing.T) {
	_, err := store.GetRule(context.TODO(), "0")
	assert.EqualError(t, err, "rule \"0\" not found")
}

func TestListRules(t *testing.T) {
	flag, err := store.CreateFlag(context.TODO(), &flipt.CreateFlagRequest{
		Key:         t.Name(),
		Name:        "foo",
		Description: "bar",
		Enabled:     true,
	})

	require.NoError(t, err)
	assert.NotNil(t, flag)

	variant, err := store.CreateVariant(context.TODO(), &flipt.CreateVariantRequest{
		FlagKey:     flag.Key,
		Key:         t.Name(),
		Name:        "foo",
		Description: "bar",
	})

	require.NoError(t, err)
	assert.NotNil(t, variant)

	segment, err := store.CreateSegment(context.TODO(), &flipt.CreateSegmentRequest{
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
		_, err := store.CreateRule(context.TODO(), req)
		require.NoError(t, err)
	}

	got, err := store.ListRules(context.TODO(), flag.Key)

	require.NoError(t, err)
	assert.NotZero(t, len(got))
}

func TestListRulesPagination(t *testing.T) {
	flag, err := store.CreateFlag(context.TODO(), &flipt.CreateFlagRequest{
		Key:         t.Name(),
		Name:        "foo",
		Description: "bar",
		Enabled:     true,
	})

	require.NoError(t, err)
	assert.NotNil(t, flag)

	variant, err := store.CreateVariant(context.TODO(), &flipt.CreateVariantRequest{
		FlagKey:     flag.Key,
		Key:         t.Name(),
		Name:        "foo",
		Description: "bar",
	})

	require.NoError(t, err)
	assert.NotNil(t, variant)

	segment, err := store.CreateSegment(context.TODO(), &flipt.CreateSegmentRequest{
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
		_, err := store.CreateRule(context.TODO(), req)
		require.NoError(t, err)
	}

	got, err := store.ListRules(context.TODO(), flag.Key, storage.WithLimit(1), storage.WithOffset(1))

	require.NoError(t, err)
	assert.Len(t, got, 1)
}

func TestCreateRuleAndDistribution(t *testing.T) {
	flag, err := store.CreateFlag(context.TODO(), &flipt.CreateFlagRequest{
		Key:         t.Name(),
		Name:        "foo",
		Description: "bar",
		Enabled:     true,
	})

	require.NoError(t, err)
	assert.NotNil(t, flag)

	variant, err := store.CreateVariant(context.TODO(), &flipt.CreateVariantRequest{
		FlagKey:     flag.Key,
		Key:         t.Name(),
		Name:        "foo",
		Description: "bar",
	})

	require.NoError(t, err)
	assert.NotNil(t, variant)

	segment, err := store.CreateSegment(context.TODO(), &flipt.CreateSegmentRequest{
		Key:         t.Name(),
		Name:        "foo",
		Description: "bar",
	})

	require.NoError(t, err)
	assert.NotNil(t, segment)

	rule, err := store.CreateRule(context.TODO(), &flipt.CreateRuleRequest{
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

	distribution, err := store.CreateDistribution(context.TODO(), &flipt.CreateDistributionRequest{
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
	flag, err := store.CreateFlag(context.TODO(), &flipt.CreateFlagRequest{
		Key:         t.Name(),
		Name:        "foo",
		Description: "bar",
		Enabled:     true,
	})

	require.NoError(t, err)
	assert.NotNil(t, flag)

	variant, err := store.CreateVariant(context.TODO(), &flipt.CreateVariantRequest{
		FlagKey:     flag.Key,
		Key:         t.Name(),
		Name:        "foo",
		Description: "bar",
	})

	require.NoError(t, err)
	assert.NotNil(t, variant)

	_, err = store.CreateDistribution(context.TODO(), &flipt.CreateDistributionRequest{
		RuleId:    "foo",
		VariantId: variant.Id,
		Rollout:   100,
	})

	assert.EqualError(t, err, "rule \"foo\" not found")
}

func TestCreateRule_FlagNotFound(t *testing.T) {
	_, err := store.CreateRule(context.TODO(), &flipt.CreateRuleRequest{
		FlagKey:    "foo",
		SegmentKey: "bar",
		Rank:       1,
	})

	assert.EqualError(t, err, "flag \"foo\" or segment \"bar\" not found")
}

func TestCreateRule_SegmentNotFound(t *testing.T) {
	flag, err := store.CreateFlag(context.TODO(), &flipt.CreateFlagRequest{
		Key:         t.Name(),
		Name:        "foo",
		Description: "bar",
		Enabled:     true,
	})

	require.NoError(t, err)
	assert.NotNil(t, flag)

	_, err = store.CreateRule(context.TODO(), &flipt.CreateRuleRequest{
		FlagKey:    flag.Key,
		SegmentKey: "foo",
		Rank:       1,
	})

	assert.EqualError(t, err, "flag \"TestCreateRule_SegmentNotFound\" or segment \"foo\" not found")
}

func TestUpdateRuleAndDistribution(t *testing.T) {
	flag, err := store.CreateFlag(context.TODO(), &flipt.CreateFlagRequest{
		Key:         t.Name(),
		Name:        "foo",
		Description: "bar",
		Enabled:     true,
	})

	require.NoError(t, err)
	assert.NotNil(t, flag)

	variant, err := store.CreateVariant(context.TODO(), &flipt.CreateVariantRequest{
		FlagKey:     flag.Key,
		Key:         t.Name(),
		Name:        "foo",
		Description: "bar",
	})

	require.NoError(t, err)
	assert.NotNil(t, variant)

	segmentOne, err := store.CreateSegment(context.TODO(), &flipt.CreateSegmentRequest{
		Key:         fmt.Sprintf("%s_one", t.Name()),
		Name:        "foo",
		Description: "bar",
	})

	require.NoError(t, err)
	assert.NotNil(t, segmentOne)

	rule, err := store.CreateRule(context.TODO(), &flipt.CreateRuleRequest{
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

	distribution, err := store.CreateDistribution(context.TODO(), &flipt.CreateDistributionRequest{
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

	segmentTwo, err := store.CreateSegment(context.TODO(), &flipt.CreateSegmentRequest{
		Key:         fmt.Sprintf("%s_two", t.Name()),
		Name:        "foo",
		Description: "bar",
	})

	require.NoError(t, err)
	assert.NotNil(t, segmentTwo)

	updatedRule, err := store.UpdateRule(context.TODO(), &flipt.UpdateRuleRequest{
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

	updatedDistribution, err := store.UpdateDistribution(context.TODO(), &flipt.UpdateDistributionRequest{
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

	err = store.DeleteDistribution(context.TODO(), &flipt.DeleteDistributionRequest{
		Id:        distribution.Id,
		RuleId:    rule.Id,
		VariantId: variant.Id,
	})
	require.NoError(t, err)
}

func TestUpdateRule_NotFound(t *testing.T) {
	flag, err := store.CreateFlag(context.TODO(), &flipt.CreateFlagRequest{
		Key:         t.Name(),
		Name:        "foo",
		Description: "bar",
		Enabled:     true,
	})

	require.NoError(t, err)
	assert.NotNil(t, flag)

	segment, err := store.CreateSegment(context.TODO(), &flipt.CreateSegmentRequest{
		Key:         t.Name(),
		Name:        "foo",
		Description: "bar",
	})

	require.NoError(t, err)
	assert.NotNil(t, segment)

	_, err = store.UpdateRule(context.TODO(), &flipt.UpdateRuleRequest{
		Id:         "foo",
		FlagKey:    flag.Key,
		SegmentKey: segment.Key,
	})

	assert.EqualError(t, err, "rule \"foo\" not found")
}

func TestDeleteRule(t *testing.T) {
	flag, err := store.CreateFlag(context.TODO(), &flipt.CreateFlagRequest{
		Key:         t.Name(),
		Name:        "foo",
		Description: "bar",
		Enabled:     true,
	})

	require.NoError(t, err)
	assert.NotNil(t, flag)

	variant, err := store.CreateVariant(context.TODO(), &flipt.CreateVariantRequest{
		FlagKey:     flag.Key,
		Key:         t.Name(),
		Name:        "foo",
		Description: "bar",
	})

	require.NoError(t, err)
	assert.NotNil(t, variant)

	segment, err := store.CreateSegment(context.TODO(), &flipt.CreateSegmentRequest{
		Key:         t.Name(),
		Name:        "foo",
		Description: "bar",
	})

	require.NoError(t, err)
	assert.NotNil(t, segment)

	var rules []*flipt.Rule

	// create 3 rules
	for i := 0; i < 3; i++ {
		rule, err := store.CreateRule(context.TODO(), &flipt.CreateRuleRequest{
			FlagKey:    flag.Key,
			SegmentKey: segment.Key,
			Rank:       int32(i + 1),
		})

		require.NoError(t, err)
		assert.NotNil(t, rule)
		rules = append(rules, rule)
	}

	// delete second rule
	err = store.DeleteRule(context.TODO(), &flipt.DeleteRuleRequest{
		FlagKey: flag.Key,
		Id:      rules[1].Id,
	})

	require.NoError(t, err)

	got, err := store.ListRules(context.TODO(), flag.Key)

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
	flag, err := store.CreateFlag(context.TODO(), &flipt.CreateFlagRequest{
		Key:         t.Name(),
		Name:        "foo",
		Description: "bar",
		Enabled:     true,
	})

	require.NoError(t, err)
	assert.NotNil(t, flag)

	err = store.DeleteRule(context.TODO(), &flipt.DeleteRuleRequest{
		Id:      "foo",
		FlagKey: flag.Key,
	})

	require.NoError(t, err)
}

func TestOrderRules(t *testing.T) {
	flag, err := store.CreateFlag(context.TODO(), &flipt.CreateFlagRequest{
		Key:         t.Name(),
		Name:        "foo",
		Description: "bar",
		Enabled:     true,
	})

	require.NoError(t, err)
	assert.NotNil(t, flag)

	variant, err := store.CreateVariant(context.TODO(), &flipt.CreateVariantRequest{
		FlagKey:     flag.Key,
		Key:         t.Name(),
		Name:        "foo",
		Description: "bar",
	})

	require.NoError(t, err)
	assert.NotNil(t, variant)

	segment, err := store.CreateSegment(context.TODO(), &flipt.CreateSegmentRequest{
		Key:         t.Name(),
		Name:        "foo",
		Description: "bar",
	})

	require.NoError(t, err)
	assert.NotNil(t, segment)

	var rules []*flipt.Rule

	// create 3 rules
	for i := 0; i < 3; i++ {
		rule, err := store.CreateRule(context.TODO(), &flipt.CreateRuleRequest{
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
	err = store.OrderRules(context.TODO(), &flipt.OrderRulesRequest{
		FlagKey: flag.Key,
		RuleIds: ruleIds,
	})

	require.NoError(t, err)

	got, err := store.ListRules(context.TODO(), flag.Key)

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

var benchRule *flipt.Rule

func BenchmarkGetRule(b *testing.B) {
	var (
		ctx       = context.Background()
		flag, err = store.CreateFlag(ctx, &flipt.CreateFlagRequest{
			Key:         b.Name(),
			Name:        "foo",
			Description: "bar",
			Enabled:     true,
		})
	)

	if err != nil {
		b.Fatal(err)
	}

	_, err = store.CreateVariant(ctx, &flipt.CreateVariantRequest{
		FlagKey:     flag.Key,
		Key:         b.Name(),
		Name:        "foo",
		Description: "bar",
	})

	if err != nil {
		b.Fatal(err)
	}

	segment, err := store.CreateSegment(ctx, &flipt.CreateSegmentRequest{
		Key:         b.Name(),
		Name:        "foo",
		Description: "bar",
	})

	if err != nil {
		b.Fatal(err)
	}

	rule, err := store.CreateRule(ctx, &flipt.CreateRuleRequest{
		FlagKey:    flag.Key,
		SegmentKey: segment.Key,
		Rank:       1,
	})

	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()

	b.Run("get-rule", func(b *testing.B) {
		var r *flipt.Rule

		for i := 0; i < b.N; i++ {
			r, _ = store.GetRule(ctx, rule.Id)
		}

		benchRule = r
	})
}

func BenchmarkGetRule_CacheMemory(b *testing.B) {
	var (
		l, _       = test.NewNullLogger()
		logger     = logrus.NewEntry(l)
		cacher     = memory.NewCache(5*time.Minute, 10*time.Minute, logger)
		storeCache = cache.NewStore(logger, cacher, store)

		ctx       = context.Background()
		flag, err = store.CreateFlag(ctx, &flipt.CreateFlagRequest{
			Key:         b.Name(),
			Name:        "foo",
			Description: "bar",
			Enabled:     true,
		})
	)

	if err != nil {
		b.Fatal(err)
	}

	_, err = store.CreateVariant(ctx, &flipt.CreateVariantRequest{
		FlagKey:     flag.Key,
		Key:         b.Name(),
		Name:        "foo",
		Description: "bar",
	})

	if err != nil {
		b.Fatal(err)
	}

	segment, err := store.CreateSegment(ctx, &flipt.CreateSegmentRequest{
		Key:         b.Name(),
		Name:        "foo",
		Description: "bar",
	})

	if err != nil {
		b.Fatal(err)
	}

	rule, err := storeCache.CreateRule(ctx, &flipt.CreateRuleRequest{
		FlagKey:    flag.Key,
		SegmentKey: segment.Key,
		Rank:       1,
	})

	if err != nil {
		b.Fatal(err)
	}

	var r *flipt.Rule

	// warm the cache
	r, _ = storeCache.GetRule(ctx, rule.Id)

	b.ResetTimer()

	b.Run("get-rule-cache", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			r, _ = storeCache.GetRule(ctx, rule.Id)
		}

		benchRule = r
	})
}
