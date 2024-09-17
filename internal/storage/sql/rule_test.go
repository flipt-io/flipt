package sql_test

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"sort"
	"time"

	"github.com/gofrs/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.flipt.io/flipt/internal/storage"
	fliptsql "go.flipt.io/flipt/internal/storage/sql"
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

	got, err := s.store.GetRule(context.TODO(), storage.NewNamespace(storage.DefaultNamespace), rule.Id)

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

func (s *DBTestSuite) TestGetRule_MultipleSegments() {
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

	firstSegment, err := s.store.CreateSegment(context.TODO(), &flipt.CreateSegmentRequest{
		Key:         t.Name(),
		Name:        "foo",
		Description: "bar",
	})

	require.NoError(t, err)
	assert.NotNil(t, firstSegment)

	secondSegment, err := s.store.CreateSegment(context.TODO(), &flipt.CreateSegmentRequest{
		Key:         "another_segment_1",
		Name:        "bar",
		Description: "foo",
	})

	require.NoError(t, err)
	assert.NotNil(t, secondSegment)

	rule, err := s.store.CreateRule(context.TODO(), &flipt.CreateRuleRequest{
		FlagKey:     flag.Key,
		SegmentKeys: []string{firstSegment.Key, secondSegment.Key},
		Rank:        1,
	})

	require.NoError(t, err)
	assert.NotNil(t, rule)

	got, err := s.store.GetRule(context.TODO(), storage.NewNamespace(storage.DefaultNamespace), rule.Id)

	require.NoError(t, err)
	assert.NotNil(t, got)

	assert.Equal(t, rule.Id, got.Id)
	assert.Equal(t, storage.DefaultNamespace, got.NamespaceKey)
	assert.Equal(t, rule.FlagKey, got.FlagKey)

	assert.Len(t, rule.SegmentKeys, 2)
	assert.Equal(t, firstSegment.Key, rule.SegmentKeys[0])
	assert.Equal(t, secondSegment.Key, rule.SegmentKeys[1])
	assert.Equal(t, rule.Rank, got.Rank)
	assert.NotZero(t, got.CreatedAt)
	assert.NotZero(t, got.UpdatedAt)
}

func (s *DBTestSuite) TestGetRuleNamespace() {
	t := s.T()

	flag, err := s.store.CreateFlag(context.TODO(), &flipt.CreateFlagRequest{
		NamespaceKey: s.namespace,
		Key:          t.Name(),
		Name:         "foo",
		Description:  "bar",
		Enabled:      true,
	})

	require.NoError(t, err)
	assert.NotNil(t, flag)

	variant, err := s.store.CreateVariant(context.TODO(), &flipt.CreateVariantRequest{
		NamespaceKey: s.namespace,
		FlagKey:      flag.Key,
		Key:          t.Name(),
		Name:         "foo",
		Description:  "bar",
	})

	require.NoError(t, err)
	assert.NotNil(t, variant)

	segment, err := s.store.CreateSegment(context.TODO(), &flipt.CreateSegmentRequest{
		NamespaceKey: s.namespace,
		Key:          t.Name(),
		Name:         "foo",
		Description:  "bar",
	})

	require.NoError(t, err)
	assert.NotNil(t, segment)

	rule, err := s.store.CreateRule(context.TODO(), &flipt.CreateRuleRequest{
		NamespaceKey: s.namespace,
		FlagKey:      flag.Key,
		SegmentKey:   segment.Key,
		Rank:         1,
	})

	require.NoError(t, err)
	assert.NotNil(t, rule)

	got, err := s.store.GetRule(context.TODO(), storage.NewNamespace(s.namespace), rule.Id)

	require.NoError(t, err)
	assert.NotNil(t, got)

	assert.Equal(t, rule.Id, got.Id)
	assert.Equal(t, s.namespace, got.NamespaceKey)
	assert.Equal(t, rule.FlagKey, got.FlagKey)
	assert.Equal(t, rule.SegmentKey, got.SegmentKey)
	assert.Equal(t, rule.Rank, got.Rank)
	assert.NotZero(t, got.CreatedAt)
	assert.NotZero(t, got.UpdatedAt)
}

func (s *DBTestSuite) TestGetRule_NotFound() {
	t := s.T()

	_, err := s.store.GetRule(context.TODO(), storage.NewNamespace(storage.DefaultNamespace), "0")
	assert.EqualError(t, err, "rule \"default/0\" not found")
}

func (s *DBTestSuite) TestGetRuleNamespace_NotFound() {
	t := s.T()

	_, err := s.store.GetRule(context.TODO(), storage.NewNamespace(s.namespace), "0")
	assert.EqualError(t, err, fmt.Sprintf("rule \"%s/0\" not found", s.namespace))
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

	_, err = s.store.ListRules(context.TODO(),
		storage.ListWithOptions(
			storage.NewResource(storage.DefaultNamespace, flag.Key),
			storage.ListWithQueryParamOptions[storage.ResourceRequest](storage.WithPageToken("Hello World")),
		),
	)
	require.EqualError(t, err, "pageToken is not valid: \"Hello World\"")

	res, err := s.store.ListRules(context.TODO(), storage.ListWithOptions(storage.NewResource(storage.DefaultNamespace, flag.Key)))
	require.NoError(t, err)

	got := res.Results
	assert.NotEmpty(t, got)

	for _, rule := range got {
		assert.Equal(t, storage.DefaultNamespace, rule.NamespaceKey)
		assert.NotZero(t, rule.CreatedAt)
		assert.NotZero(t, rule.UpdatedAt)
	}
}

func (s *DBTestSuite) TestListRules_MultipleSegments() {
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

	firstSegment, err := s.store.CreateSegment(context.TODO(), &flipt.CreateSegmentRequest{
		Key:         t.Name(),
		Name:        "foo",
		Description: "bar",
	})

	require.NoError(t, err)
	assert.NotNil(t, firstSegment)

	secondSegment, err := s.store.CreateSegment(context.TODO(), &flipt.CreateSegmentRequest{
		Key:         "another_segment_2",
		Name:        "foo",
		Description: "bar",
	})

	require.NoError(t, err)
	assert.NotNil(t, secondSegment)

	reqs := []*flipt.CreateRuleRequest{
		{
			FlagKey:     flag.Key,
			SegmentKeys: []string{firstSegment.Key, secondSegment.Key},
			Rank:        1,
		},
		{
			FlagKey:     flag.Key,
			SegmentKeys: []string{firstSegment.Key, secondSegment.Key},
			Rank:        2,
		},
	}

	for _, req := range reqs {
		_, err := s.store.CreateRule(context.TODO(), req)
		require.NoError(t, err)
	}

	_, err = s.store.ListRules(context.TODO(),
		storage.ListWithOptions(
			storage.NewResource(storage.DefaultNamespace, flag.Key),
			storage.ListWithQueryParamOptions[storage.ResourceRequest](storage.WithPageToken("Hello World"))))
	require.EqualError(t, err, "pageToken is not valid: \"Hello World\"")

	res, err := s.store.ListRules(context.TODO(), storage.ListWithOptions(storage.NewResource(storage.DefaultNamespace, flag.Key)))
	require.NoError(t, err)

	got := res.Results
	assert.NotEmpty(t, got)

	for _, rule := range got {
		assert.Equal(t, storage.DefaultNamespace, rule.NamespaceKey)
		assert.Len(t, rule.SegmentKeys, 2)
		assert.Contains(t, rule.SegmentKeys, firstSegment.Key)
		assert.Contains(t, rule.SegmentKeys, secondSegment.Key)
		assert.NotZero(t, rule.CreatedAt)
		assert.NotZero(t, rule.UpdatedAt)
	}
}

func (s *DBTestSuite) TestListRulesNamespace() {
	t := s.T()

	flag, err := s.store.CreateFlag(context.TODO(), &flipt.CreateFlagRequest{
		NamespaceKey: s.namespace,
		Key:          t.Name(),
		Name:         "foo",
		Description:  "bar",
		Enabled:      true,
	})

	require.NoError(t, err)
	assert.NotNil(t, flag)

	variant, err := s.store.CreateVariant(context.TODO(), &flipt.CreateVariantRequest{
		NamespaceKey: s.namespace,
		FlagKey:      flag.Key,
		Key:          t.Name(),
		Name:         "foo",
		Description:  "bar",
	})

	require.NoError(t, err)
	assert.NotNil(t, variant)

	segment, err := s.store.CreateSegment(context.TODO(), &flipt.CreateSegmentRequest{
		NamespaceKey: s.namespace,
		Key:          t.Name(),
		Name:         "foo",
		Description:  "bar",
	})

	require.NoError(t, err)
	assert.NotNil(t, segment)

	reqs := []*flipt.CreateRuleRequest{
		{
			NamespaceKey: s.namespace,
			FlagKey:      flag.Key,
			SegmentKey:   segment.Key,
			Rank:         1,
		},
		{
			NamespaceKey: s.namespace,
			FlagKey:      flag.Key,
			SegmentKey:   segment.Key,
			Rank:         2,
		},
	}

	for _, req := range reqs {
		_, err := s.store.CreateRule(context.TODO(), req)
		require.NoError(t, err)
	}

	res, err := s.store.ListRules(context.TODO(), storage.ListWithOptions(storage.NewResource(s.namespace, flag.Key)))
	require.NoError(t, err)

	got := res.Results
	assert.NotEmpty(t, got)

	for _, rule := range got {
		assert.Equal(t, s.namespace, rule.NamespaceKey)
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

	res, err := s.store.ListRules(context.TODO(),
		storage.ListWithOptions(
			storage.NewResource(storage.DefaultNamespace, flag.Key),
			storage.ListWithQueryParamOptions[storage.ResourceRequest](storage.WithLimit(1), storage.WithOffset(1)),
		))
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

	res, err := s.store.ListRules(context.TODO(),
		storage.ListWithOptions(
			storage.NewResource(storage.DefaultNamespace, flag.Key),
			storage.ListWithQueryParamOptions[storage.ResourceRequest](opts...),
		))
	require.NoError(t, err)

	got := res.Results
	assert.Len(t, got, 1)
	assert.Equal(t, reqs[1].Rank, got[0].Rank)
	assert.NotEmpty(t, res.NextPageToken)

	pTokenB, err := base64.StdEncoding.DecodeString(res.NextPageToken)
	require.NoError(t, err)

	pageToken := &common.PageToken{}
	err = json.Unmarshal(pTokenB, pageToken)
	require.NoError(t, err)
	assert.NotEmpty(t, pageToken.Key)
	assert.Equal(t, uint64(1), pageToken.Offset)

	opts = append(opts, storage.WithPageToken(res.NextPageToken))

	res, err = s.store.ListRules(context.TODO(),
		storage.ListWithOptions(
			storage.NewResource(storage.DefaultNamespace, flag.Key),
			storage.ListWithQueryParamOptions[storage.ResourceRequest](opts...),
		))
	require.NoError(t, err)

	got = res.Results
	assert.Len(t, got, 1)
	assert.Equal(t, reqs[0].Rank, got[0].Rank)
}

func (s *DBTestSuite) TestListRulesPagination_FullWalk() {
	t := s.T()

	namespace := uuid.Must(uuid.NewV4()).String()

	ctx := context.Background()
	_, err := s.store.CreateNamespace(ctx, &flipt.CreateNamespaceRequest{
		Key: namespace,
	})
	require.NoError(t, err)

	flag, err := s.store.CreateFlag(ctx, &flipt.CreateFlagRequest{
		NamespaceKey: namespace,
		Key:          "flag-list-rules-full-walk",
		Name:         "flag-list-rules-full-walk",
	})
	require.NoError(t, err)

	variant, err := s.store.CreateVariant(ctx, &flipt.CreateVariantRequest{
		NamespaceKey: namespace,
		FlagKey:      flag.Key,
		Key:          "variant-list-rules-full-walk",
	})
	require.NoError(t, err)

	segment, err := s.store.CreateSegment(ctx, &flipt.CreateSegmentRequest{
		NamespaceKey: namespace,
		Key:          "segment-list-rules-full-walk",
		Name:         "segment-list-rules-full-walk",
	})
	require.NoError(t, err)

	var (
		totalRules = 9
		pageSize   = uint64(3)
	)

	for i := 0; i < totalRules; i++ {
		req := flipt.CreateRuleRequest{
			NamespaceKey: namespace,
			FlagKey:      flag.Key,
			SegmentKey:   segment.Key,
			Rank:         int32(i + 1),
		}

		rule, err := s.store.CreateRule(ctx, &req)
		require.NoError(t, err)

		for i := 0; i < 2; i++ {
			if i > 0 && s.db.Driver == fliptsql.MySQL {
				// required for MySQL since it only s.stores timestamps to the second and not millisecond granularity
				time.Sleep(time.Second)
			}

			_, err := s.store.CreateDistribution(ctx, &flipt.CreateDistributionRequest{
				NamespaceKey: namespace,
				FlagKey:      flag.Key,
				VariantId:    variant.Id,
				RuleId:       rule.Id,
				Rollout:      100.0,
			})
			require.NoError(t, err)
		}
	}

	resp, err := s.store.ListRules(ctx,
		storage.ListWithOptions(
			storage.NewResource(namespace, flag.Key),
			storage.ListWithQueryParamOptions[storage.ResourceRequest](
				storage.WithLimit(pageSize),
			),
		))
	require.NoError(t, err)

	found := resp.Results
	for token := resp.NextPageToken; token != ""; token = resp.NextPageToken {
		resp, err = s.store.ListRules(ctx,
			storage.ListWithOptions(
				storage.NewResource(namespace, flag.Key),
				storage.ListWithQueryParamOptions[storage.ResourceRequest](
					storage.WithLimit(pageSize),
					storage.WithPageToken(token),
				),
			),
		)
		require.NoError(t, err)

		found = append(found, resp.Results...)
	}

	require.Len(t, found, totalRules)

	for i := 0; i < totalRules; i++ {
		assert.Equal(t, namespace, found[i].NamespaceKey)
		assert.Equal(t, flag.Key, found[i].FlagKey)
		assert.Equal(t, segment.Key, found[i].SegmentKey)
		assert.Equal(t, int32(i+1), found[i].Rank)

		require.Len(t, found[i].Distributions, 2)
		assert.Equal(t, found[i].Id, found[i].Distributions[0].RuleId)
		assert.Equal(t, variant.Id, found[i].Distributions[0].VariantId)
		assert.InDelta(t, 100.0, found[i].Distributions[0].Rollout, 0)

		assert.Equal(t, found[i].Id, found[i].Distributions[1].RuleId)
		assert.Equal(t, variant.Id, found[i].Distributions[1].VariantId)
		assert.InDelta(t, 100.0, found[i].Distributions[1].Rollout, 0)
	}
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
		FlagKey:   flag.Key,
		RuleId:    rule.Id,
		VariantId: variant.Id,
		Rollout:   100,
	})

	require.NoError(t, err)
	assert.NotZero(t, distribution.Id)
	assert.Equal(t, rule.Id, distribution.RuleId)
	assert.Equal(t, variant.Id, distribution.VariantId)
	assert.InDelta(t, 100, distribution.Rollout, 0)
	assert.NotZero(t, distribution.CreatedAt)
	assert.Equal(t, distribution.CreatedAt.Seconds, distribution.UpdatedAt.Seconds)
}

func (s *DBTestSuite) TestCreateRuleAndDistributionNamespace() {
	t := s.T()

	flag, err := s.store.CreateFlag(context.TODO(), &flipt.CreateFlagRequest{
		NamespaceKey: s.namespace,
		Key:          t.Name(),
		Name:         "foo",
		Description:  "bar",
		Enabled:      true,
	})

	require.NoError(t, err)
	assert.NotNil(t, flag)

	variant, err := s.store.CreateVariant(context.TODO(), &flipt.CreateVariantRequest{
		NamespaceKey: s.namespace,
		FlagKey:      flag.Key,
		Key:          t.Name(),
		Name:         "foo",
		Description:  "bar",
	})

	require.NoError(t, err)
	assert.NotNil(t, variant)

	segment, err := s.store.CreateSegment(context.TODO(), &flipt.CreateSegmentRequest{
		NamespaceKey: s.namespace,
		Key:          t.Name(),
		Name:         "foo",
		Description:  "bar",
	})

	require.NoError(t, err)
	assert.NotNil(t, segment)

	rule, err := s.store.CreateRule(context.TODO(), &flipt.CreateRuleRequest{
		NamespaceKey:    s.namespace,
		FlagKey:         flag.Key,
		SegmentKey:      segment.Key,
		SegmentOperator: flipt.SegmentOperator_AND_SEGMENT_OPERATOR,
		Rank:            1,
	})

	require.NoError(t, err)
	assert.NotNil(t, rule)

	assert.Equal(t, s.namespace, rule.NamespaceKey)
	assert.NotZero(t, rule.Id)
	assert.Equal(t, flag.Key, rule.FlagKey)
	assert.Equal(t, segment.Key, rule.SegmentKey)
	assert.Equal(t, int32(1), rule.Rank)
	assert.NotZero(t, rule.CreatedAt)
	assert.Equal(t, rule.CreatedAt.Seconds, rule.UpdatedAt.Seconds)
	assert.Equal(t, flipt.SegmentOperator_OR_SEGMENT_OPERATOR, rule.SegmentOperator)

	distribution, err := s.store.CreateDistribution(context.TODO(), &flipt.CreateDistributionRequest{
		NamespaceKey: s.namespace,
		FlagKey:      flag.Key,
		RuleId:       rule.Id,
		VariantId:    variant.Id,
		Rollout:      100,
	})

	require.NoError(t, err)
	assert.NotZero(t, distribution.Id)
	assert.Equal(t, rule.Id, distribution.RuleId)
	assert.Equal(t, variant.Id, distribution.VariantId)
	assert.InDelta(t, 100, distribution.Rollout, 0)
	assert.NotZero(t, distribution.CreatedAt)
	assert.Equal(t, distribution.CreatedAt.Seconds, distribution.UpdatedAt.Seconds)
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
		FlagKey:   flag.Key,
		RuleId:    "foo",
		VariantId: variant.Id,
		Rollout:   100,
	})

	assert.EqualError(t, err, fmt.Sprintf("variant %q, rule %q, flag %q in namespace %q not found", variant.Id, "foo", flag.Key, "default"))
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

func (s *DBTestSuite) TestCreateRuleNamespace_FlagNotFound() {
	t := s.T()

	_, err := s.store.CreateRule(context.TODO(), &flipt.CreateRuleRequest{
		NamespaceKey: s.namespace,
		FlagKey:      "foo",
		SegmentKey:   "bar",
		Rank:         1,
	})

	assert.EqualError(t, err, fmt.Sprintf("flag \"%s/foo\" or segment \"%s/bar\" not found", s.namespace, s.namespace))
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

func (s *DBTestSuite) TestCreateRuleNamespace_SegmentNotFound() {
	t := s.T()

	flag, err := s.store.CreateFlag(context.TODO(), &flipt.CreateFlagRequest{
		NamespaceKey: s.namespace,
		Key:          t.Name(),
		Name:         "foo",
		Description:  "bar",
		Enabled:      true,
	})

	require.NoError(t, err)
	assert.NotNil(t, flag)

	_, err = s.store.CreateRule(context.TODO(), &flipt.CreateRuleRequest{
		NamespaceKey: s.namespace,
		FlagKey:      flag.Key,
		SegmentKey:   "foo",
		Rank:         1,
	})

	assert.EqualError(t, err, fmt.Sprintf("flag \"%s/%s\" or segment \"%s/foo\" not found", s.namespace, t.Name(), s.namespace))
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

	variantOne, err := s.store.CreateVariant(context.TODO(), &flipt.CreateVariantRequest{
		FlagKey:     flag.Key,
		Key:         t.Name() + "foo",
		Name:        "foo",
		Description: "bar",
	})

	require.NoError(t, err)
	assert.NotNil(t, variantOne)

	variantTwo, err := s.store.CreateVariant(context.TODO(), &flipt.CreateVariantRequest{
		FlagKey:     flag.Key,
		Key:         t.Name() + "bar",
		Name:        "bar",
		Description: "baz",
	})

	require.NoError(t, err)
	assert.NotNil(t, variantOne)

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
		FlagKey:   flag.Key,
		RuleId:    rule.Id,
		VariantId: variantOne.Id,
		Rollout:   100,
	})

	require.NoError(t, err)
	assert.NotZero(t, distribution.Id)
	assert.Equal(t, rule.Id, distribution.RuleId)
	assert.Equal(t, variantOne.Id, distribution.VariantId)
	assert.InDelta(t, 100, distribution.Rollout, 0)
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
		Id:              rule.Id,
		FlagKey:         flag.Key,
		SegmentKey:      segmentTwo.Key,
		SegmentOperator: flipt.SegmentOperator_AND_SEGMENT_OPERATOR,
	})

	require.NoError(t, err)
	assert.NotNil(t, updatedRule)

	assert.Equal(t, rule.Id, updatedRule.Id)
	assert.Equal(t, rule.FlagKey, updatedRule.FlagKey)
	assert.Equal(t, segmentTwo.Key, updatedRule.SegmentKey)
	assert.Equal(t, int32(1), updatedRule.Rank)
	assert.Equal(t, flipt.SegmentOperator_OR_SEGMENT_OPERATOR, updatedRule.SegmentOperator)
	// assert.Equal(t, rule.CreatedAt.Seconds, updatedRule.CreatedAt.Seconds)
	assert.NotZero(t, rule.UpdatedAt)

	t.Log("Update rule to references two segments.")

	updatedRule, err = s.store.UpdateRule(context.TODO(), &flipt.UpdateRuleRequest{
		Id:              rule.Id,
		FlagKey:         flag.Key,
		SegmentKeys:     []string{segmentOne.Key, segmentTwo.Key},
		SegmentOperator: flipt.SegmentOperator_AND_SEGMENT_OPERATOR,
	})

	require.NoError(t, err)
	assert.NotNil(t, updatedRule)

	assert.Equal(t, rule.Id, updatedRule.Id)
	assert.Equal(t, rule.FlagKey, updatedRule.FlagKey)
	assert.Contains(t, updatedRule.SegmentKeys, segmentOne.Key)
	assert.Contains(t, updatedRule.SegmentKeys, segmentTwo.Key)
	assert.Equal(t, flipt.SegmentOperator_AND_SEGMENT_OPERATOR, updatedRule.SegmentOperator)
	assert.Equal(t, int32(1), updatedRule.Rank)
	assert.NotZero(t, rule.UpdatedAt)

	// update distribution rollout
	updatedDistribution, err := s.store.UpdateDistribution(context.TODO(), &flipt.UpdateDistributionRequest{
		FlagKey:   flag.Key,
		Id:        distribution.Id,
		RuleId:    rule.Id,
		VariantId: variantOne.Id,
		Rollout:   10,
	})

	require.NoError(t, err)
	assert.Equal(t, distribution.Id, updatedDistribution.Id)
	assert.Equal(t, rule.Id, updatedDistribution.RuleId)
	assert.Equal(t, variantOne.Id, updatedDistribution.VariantId)
	assert.InDelta(t, 10, updatedDistribution.Rollout, 0)
	assert.NotZero(t, rule.UpdatedAt)

	// update distribution variant
	updatedDistribution, err = s.store.UpdateDistribution(context.TODO(), &flipt.UpdateDistributionRequest{
		FlagKey:   flag.Key,
		Id:        distribution.Id,
		RuleId:    rule.Id,
		VariantId: variantTwo.Id,
		Rollout:   10,
	})

	require.NoError(t, err)
	assert.Equal(t, distribution.Id, updatedDistribution.Id)
	assert.Equal(t, rule.Id, updatedDistribution.RuleId)
	assert.Equal(t, variantTwo.Id, updatedDistribution.VariantId)
	assert.InDelta(t, 10, updatedDistribution.Rollout, 0)
	assert.NotZero(t, rule.UpdatedAt)

	err = s.store.DeleteDistribution(context.TODO(), &flipt.DeleteDistributionRequest{
		Id:        distribution.Id,
		RuleId:    rule.Id,
		VariantId: variantOne.Id,
	})
	require.NoError(t, err)
}

func (s *DBTestSuite) TestUpdateRuleAndDistributionNamespace() {
	t := s.T()

	flag, err := s.store.CreateFlag(context.TODO(), &flipt.CreateFlagRequest{
		NamespaceKey: s.namespace,
		Key:          t.Name(),
		Name:         "foo",
		Description:  "bar",
		Enabled:      true,
	})

	require.NoError(t, err)
	assert.NotNil(t, flag)

	variantOne, err := s.store.CreateVariant(context.TODO(), &flipt.CreateVariantRequest{
		NamespaceKey: s.namespace,
		FlagKey:      flag.Key,
		Key:          t.Name() + "foo",
		Name:         "foo",
		Description:  "bar",
	})

	require.NoError(t, err)
	assert.NotNil(t, variantOne)

	variantTwo, err := s.store.CreateVariant(context.TODO(), &flipt.CreateVariantRequest{
		NamespaceKey: s.namespace,
		FlagKey:      flag.Key,
		Key:          t.Name() + "bar",
		Name:         "bar",
		Description:  "baz",
	})

	require.NoError(t, err)
	assert.NotNil(t, variantOne)

	segmentOne, err := s.store.CreateSegment(context.TODO(), &flipt.CreateSegmentRequest{
		NamespaceKey: s.namespace,
		Key:          fmt.Sprintf("%s_one", t.Name()),
		Name:         "foo",
		Description:  "bar",
	})

	require.NoError(t, err)
	assert.NotNil(t, segmentOne)

	rule, err := s.store.CreateRule(context.TODO(), &flipt.CreateRuleRequest{
		NamespaceKey: s.namespace,
		FlagKey:      flag.Key,
		SegmentKey:   segmentOne.Key,
		Rank:         1,
	})

	require.NoError(t, err)
	assert.NotNil(t, rule)

	assert.Equal(t, s.namespace, rule.NamespaceKey)
	assert.NotZero(t, rule.Id)
	assert.Equal(t, flag.Key, rule.FlagKey)
	assert.Equal(t, segmentOne.Key, rule.SegmentKey)
	assert.Equal(t, int32(1), rule.Rank)
	assert.NotZero(t, rule.CreatedAt)
	assert.Equal(t, rule.CreatedAt.Seconds, rule.UpdatedAt.Seconds)

	distribution, err := s.store.CreateDistribution(context.TODO(), &flipt.CreateDistributionRequest{
		NamespaceKey: s.namespace,
		FlagKey:      flag.Key,
		RuleId:       rule.Id,
		VariantId:    variantOne.Id,
		Rollout:      100,
	})

	require.NoError(t, err)
	assert.NotZero(t, distribution.Id)
	assert.Equal(t, rule.Id, distribution.RuleId)
	assert.Equal(t, variantOne.Id, distribution.VariantId)
	assert.InDelta(t, 100, distribution.Rollout, 0)
	assert.NotZero(t, distribution.CreatedAt)
	assert.Equal(t, distribution.CreatedAt.Seconds, distribution.UpdatedAt.Seconds)

	segmentTwo, err := s.store.CreateSegment(context.TODO(), &flipt.CreateSegmentRequest{
		NamespaceKey: s.namespace,
		Key:          fmt.Sprintf("%s_two", t.Name()),
		Name:         "foo",
		Description:  "bar",
	})

	require.NoError(t, err)
	assert.NotNil(t, segmentTwo)

	updatedRule, err := s.store.UpdateRule(context.TODO(), &flipt.UpdateRuleRequest{
		NamespaceKey: s.namespace,
		Id:           rule.Id,
		FlagKey:      flag.Key,
		SegmentKey:   segmentTwo.Key,
	})

	require.NoError(t, err)
	assert.NotNil(t, updatedRule)

	assert.Equal(t, s.namespace, rule.NamespaceKey)
	assert.Equal(t, rule.Id, updatedRule.Id)
	assert.Equal(t, rule.FlagKey, updatedRule.FlagKey)
	assert.Equal(t, segmentTwo.Key, updatedRule.SegmentKey)
	assert.Equal(t, int32(1), updatedRule.Rank)
	// assert.Equal(t, rule.CreatedAt.Seconds, updatedRule.CreatedAt.Seconds)
	assert.NotZero(t, rule.UpdatedAt)

	// update distribution rollout
	updatedDistribution, err := s.store.UpdateDistribution(context.TODO(), &flipt.UpdateDistributionRequest{
		NamespaceKey: s.namespace,
		FlagKey:      flag.Key,
		Id:           distribution.Id,
		RuleId:       rule.Id,
		VariantId:    variantOne.Id,
		Rollout:      10,
	})

	require.NoError(t, err)
	assert.Equal(t, distribution.Id, updatedDistribution.Id)
	assert.Equal(t, rule.Id, updatedDistribution.RuleId)
	assert.Equal(t, variantOne.Id, updatedDistribution.VariantId)
	assert.InDelta(t, 10, updatedDistribution.Rollout, 0)
	// assert.Equal(t, distribution.CreatedAt.Seconds, updatedDistribution.CreatedAt.Seconds)
	assert.NotZero(t, rule.UpdatedAt)

	// update distribution variant
	updatedDistribution, err = s.store.UpdateDistribution(context.TODO(), &flipt.UpdateDistributionRequest{
		NamespaceKey: s.namespace,
		FlagKey:      flag.Key,
		Id:           distribution.Id,
		RuleId:       rule.Id,
		VariantId:    variantTwo.Id,
		Rollout:      10,
	})

	require.NoError(t, err)
	assert.Equal(t, distribution.Id, updatedDistribution.Id)
	assert.Equal(t, rule.Id, updatedDistribution.RuleId)
	assert.Equal(t, variantTwo.Id, updatedDistribution.VariantId)
	assert.InDelta(t, 10, updatedDistribution.Rollout, 0)
	// assert.Equal(t, distribution.CreatedAt.Seconds, updatedDistribution.CreatedAt.Seconds)
	assert.NotZero(t, rule.UpdatedAt)

	err = s.store.DeleteDistribution(context.TODO(), &flipt.DeleteDistributionRequest{
		Id:        distribution.Id,
		RuleId:    rule.Id,
		VariantId: variantOne.Id,
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

func (s *DBTestSuite) TestUpdateRuleNamespace_NotFound() {
	t := s.T()

	flag, err := s.store.CreateFlag(context.TODO(), &flipt.CreateFlagRequest{
		NamespaceKey: s.namespace,
		Key:          t.Name(),
		Name:         "foo",
		Description:  "bar",
		Enabled:      true,
	})

	require.NoError(t, err)
	assert.NotNil(t, flag)

	segment, err := s.store.CreateSegment(context.TODO(), &flipt.CreateSegmentRequest{
		NamespaceKey: s.namespace,
		Key:          t.Name(),
		Name:         "foo",
		Description:  "bar",
	})

	require.NoError(t, err)
	assert.NotNil(t, segment)

	_, err = s.store.UpdateRule(context.TODO(), &flipt.UpdateRuleRequest{
		NamespaceKey: s.namespace,
		Id:           "foo",
		FlagKey:      flag.Key,
		SegmentKey:   segment.Key,
	})

	assert.EqualError(t, err, fmt.Sprintf("rule \"%s/foo\" not found", s.namespace))
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

	res, err := s.store.ListRules(context.TODO(), storage.ListWithOptions(storage.NewResource(storage.DefaultNamespace, flag.Key)))
	// ensure rules are in correct order
	require.NoError(t, err)

	got := res.Results
	assert.NotNil(t, got)
	assert.Len(t, got, 2)
	assert.Equal(t, rules[0].Id, got[0].Id)
	assert.Equal(t, int32(1), got[0].Rank)
	assert.Equal(t, storage.DefaultNamespace, got[0].NamespaceKey)
	assert.Equal(t, rules[2].Id, got[1].Id)
	assert.Equal(t, int32(2), got[1].Rank)
	assert.Equal(t, storage.DefaultNamespace, got[1].NamespaceKey)
}

func (s *DBTestSuite) TestDeleteRuleNamespace() {
	t := s.T()

	flag, err := s.store.CreateFlag(context.TODO(), &flipt.CreateFlagRequest{
		NamespaceKey: s.namespace,
		Key:          t.Name(),
		Name:         "foo",
		Description:  "bar",
		Enabled:      true,
	})

	require.NoError(t, err)
	assert.NotNil(t, flag)

	variant, err := s.store.CreateVariant(context.TODO(), &flipt.CreateVariantRequest{
		NamespaceKey: s.namespace,
		FlagKey:      flag.Key,
		Key:          t.Name(),
		Name:         "foo",
		Description:  "bar",
	})

	require.NoError(t, err)
	assert.NotNil(t, variant)

	segment, err := s.store.CreateSegment(context.TODO(), &flipt.CreateSegmentRequest{
		NamespaceKey: s.namespace,
		Key:          t.Name(),
		Name:         "foo",
		Description:  "bar",
	})

	require.NoError(t, err)
	assert.NotNil(t, segment)

	var rules []*flipt.Rule

	// create 3 rules
	for i := 0; i < 3; i++ {
		rule, err := s.store.CreateRule(context.TODO(), &flipt.CreateRuleRequest{
			NamespaceKey: s.namespace,
			FlagKey:      flag.Key,
			SegmentKey:   segment.Key,
			Rank:         int32(i + 1),
		})

		require.NoError(t, err)
		assert.NotNil(t, rule)
		rules = append(rules, rule)
	}

	// delete second rule
	err = s.store.DeleteRule(context.TODO(), &flipt.DeleteRuleRequest{
		NamespaceKey: s.namespace,
		FlagKey:      flag.Key,
		Id:           rules[1].Id,
	})

	require.NoError(t, err)

	res, err := s.store.ListRules(context.TODO(), storage.ListWithOptions(storage.NewResource(s.namespace, flag.Key)))
	// ensure rules are in correct order
	require.NoError(t, err)

	got := res.Results
	assert.NotNil(t, got)
	assert.Len(t, got, 2)
	assert.Equal(t, rules[0].Id, got[0].Id)
	assert.Equal(t, int32(1), got[0].Rank)
	assert.Equal(t, s.namespace, got[0].NamespaceKey)
	assert.Equal(t, rules[2].Id, got[1].Id)
	assert.Equal(t, int32(2), got[1].Rank)
	assert.Equal(t, s.namespace, got[1].NamespaceKey)
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

func (s *DBTestSuite) TestDeleteRuleNamespace_NotFound() {
	t := s.T()

	flag, err := s.store.CreateFlag(context.TODO(), &flipt.CreateFlagRequest{
		NamespaceKey: s.namespace,
		Key:          t.Name(),
		Name:         "foo",
		Description:  "bar",
		Enabled:      true,
	})

	require.NoError(t, err)
	assert.NotNil(t, flag)

	err = s.store.DeleteRule(context.TODO(), &flipt.DeleteRuleRequest{
		NamespaceKey: s.namespace,
		Id:           "foo",
		FlagKey:      flag.Key,
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

	res, err := s.store.ListRules(context.TODO(), storage.ListWithOptions(storage.NewResource(storage.DefaultNamespace, flag.Key)))

	// ensure rules are in correct order
	require.NoError(t, err)
	got := res.Results
	assert.NotNil(t, got)
	assert.Len(t, got, 3)

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

func (s *DBTestSuite) TestOrderRulesNamespace() {
	t := s.T()

	flag, err := s.store.CreateFlag(context.TODO(), &flipt.CreateFlagRequest{
		NamespaceKey: s.namespace,
		Key:          t.Name(),
		Name:         "foo",
		Description:  "bar",
		Enabled:      true,
	})

	require.NoError(t, err)
	assert.NotNil(t, flag)

	variant, err := s.store.CreateVariant(context.TODO(), &flipt.CreateVariantRequest{
		NamespaceKey: s.namespace,
		FlagKey:      flag.Key,
		Key:          t.Name(),
		Name:         "foo",
		Description:  "bar",
	})

	require.NoError(t, err)
	assert.NotNil(t, variant)

	segment, err := s.store.CreateSegment(context.TODO(), &flipt.CreateSegmentRequest{
		NamespaceKey: s.namespace,
		Key:          t.Name(),
		Name:         "foo",
		Description:  "bar",
	})

	require.NoError(t, err)
	assert.NotNil(t, segment)

	var rules []*flipt.Rule

	// create 3 rules
	for i := 0; i < 3; i++ {
		rule, err := s.store.CreateRule(context.TODO(), &flipt.CreateRuleRequest{
			NamespaceKey: s.namespace,
			FlagKey:      flag.Key,
			SegmentKey:   segment.Key,
			Rank:         int32(i + 1),
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
		NamespaceKey: s.namespace,
		FlagKey:      flag.Key,
		RuleIds:      ruleIds,
	})

	require.NoError(t, err)

	res, err := s.store.ListRules(context.TODO(), storage.ListWithOptions(storage.NewResource(s.namespace, flag.Key)))

	// ensure rules are in correct order
	require.NoError(t, err)
	got := res.Results
	assert.NotNil(t, got)
	assert.Len(t, got, 3)

	assert.Equal(t, rules[0].Id, got[0].Id)
	assert.Equal(t, int32(1), got[0].Rank)
	assert.Equal(t, s.namespace, got[0].NamespaceKey)

	assert.Equal(t, rules[1].Id, got[1].Id)
	assert.Equal(t, int32(2), got[1].Rank)
	assert.Equal(t, s.namespace, got[1].NamespaceKey)

	assert.Equal(t, rules[2].Id, got[2].Id)
	assert.Equal(t, int32(3), got[2].Rank)
	assert.Equal(t, s.namespace, got[2].NamespaceKey)
}
