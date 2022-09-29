//nolint:goconst
package server

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	flipt "go.flipt.io/flipt/rpc/flipt"
	"go.flipt.io/flipt/storage"
	"go.uber.org/zap/zaptest"
)

func TestGetRule(t *testing.T) {
	var (
		store  = &storeMock{}
		logger = zaptest.NewLogger(t)
		s      = &Server{
			logger: logger,
			store:  store,
		}
		req = &flipt.GetRuleRequest{Id: "id", FlagKey: "flagKey"}
	)

	store.On("GetRule", mock.Anything, "id").Return(&flipt.Rule{
		Id: "1",
	}, nil)

	got, err := s.GetRule(context.TODO(), req)
	require.NoError(t, err)

	assert.NotNil(t, got)
}

func TestListRules_PaginationDefaultLimits(t *testing.T) {
	var (
		store  = &storeMock{}
		logger = zaptest.NewLogger(t)
		s      = &Server{
			logger: logger,
			store:  store,
		}
	)

	defer store.AssertExpectations(t)

	params := storage.QueryParams{}
	store.On("ListRules", mock.Anything, "flagKey", mock.MatchedBy(func(opts []storage.QueryOption) bool {
		for _, opt := range opts {
			opt(&params)
		}

		// assert defaults are applied
		return params.Limit == defaultListLimit && params.Offset == 0
	})).Return(
		storage.ResultSet[*flipt.Rule]{
			Results: []*flipt.Rule{
				{
					FlagKey: "flagKey",
				},
			},
		}, nil)

	store.On("CountRules", mock.Anything).Return(uint64(1), nil)

	got, err := s.ListRules(context.TODO(), &flipt.ListRuleRequest{FlagKey: "flagKey",
		Limit:  0,
		Offset: -1,
	})

	require.NoError(t, err)

	assert.NotEmpty(t, got.Rules)
	assert.Empty(t, got.NextPageToken)
	assert.Equal(t, int32(1), got.TotalCount)
}

func TestListRules_PaginationMaxLimits(t *testing.T) {
	var (
		store  = &storeMock{}
		logger = zaptest.NewLogger(t)
		s      = &Server{
			logger: logger,
			store:  store,
		}
	)

	defer store.AssertExpectations(t)

	params := storage.QueryParams{}
	store.On("ListRules", mock.Anything, "flagKey", mock.MatchedBy(func(opts []storage.QueryOption) bool {
		for _, opt := range opts {
			opt(&params)
		}

		// assert max is applied
		return params.Limit == maxListLimit && params.Offset == 0
	})).Return(
		storage.ResultSet[*flipt.Rule]{
			Results: []*flipt.Rule{
				{
					FlagKey: "flagKey",
				},
			},
		}, nil)

	store.On("CountRules", mock.Anything).Return(uint64(1), nil)

	got, err := s.ListRules(context.TODO(), &flipt.ListRuleRequest{FlagKey: "flagKey",
		Limit:  200,
		Offset: -1,
	})

	require.NoError(t, err)

	assert.NotEmpty(t, got.Rules)
	assert.Empty(t, got.NextPageToken)
	assert.Equal(t, int32(1), got.TotalCount)
}

func TestListRules_PaginationNextPageToken(t *testing.T) {
	var (
		store  = &storeMock{}
		logger = zaptest.NewLogger(t)
		s      = &Server{
			logger: logger,
			store:  store,
		}
	)

	defer store.AssertExpectations(t)

	params := storage.QueryParams{}
	store.On("ListRules", mock.Anything, "flagKey", mock.MatchedBy(func(opts []storage.QueryOption) bool {
		for _, opt := range opts {
			opt(&params)
		}

		// assert page token is preferred over offset
		return params.PageToken == "foo" && params.Offset == 0
	})).Return(
		storage.ResultSet[*flipt.Rule]{
			Results: []*flipt.Rule{
				{
					FlagKey: "flagKey",
				},
			},
			NextPageToken: "bar",
		}, nil)

	store.On("CountRules", mock.Anything).Return(uint64(1), nil)

	got, err := s.ListRules(context.TODO(), &flipt.ListRuleRequest{FlagKey: "flagKey",
		PageToken: "Zm9v",
		Offset:    10,
	})

	require.NoError(t, err)

	assert.NotEmpty(t, got.Rules)
	assert.Equal(t, "YmFy", got.NextPageToken)
	assert.Equal(t, int32(1), got.TotalCount)
}

func TestCreateRule(t *testing.T) {
	var (
		store  = &storeMock{}
		logger = zaptest.NewLogger(t)
		s      = &Server{
			logger: logger,
			store:  store,
		}
		req = &flipt.CreateRuleRequest{
			FlagKey:    "flagKey",
			SegmentKey: "segmentKey",
			Rank:       1,
		}
	)

	store.On("CreateRule", mock.Anything, req).Return(&flipt.Rule{
		Id:         "1",
		FlagKey:    req.FlagKey,
		SegmentKey: req.SegmentKey,
		Rank:       req.Rank,
	}, nil)

	got, err := s.CreateRule(context.TODO(), req)
	require.NoError(t, err)

	assert.NotNil(t, got)
}

func TestUpdateRule(t *testing.T) {
	var (
		store  = &storeMock{}
		logger = zaptest.NewLogger(t)
		s      = &Server{
			logger: logger,
			store:  store,
		}
		req = &flipt.UpdateRuleRequest{
			Id:         "1",
			FlagKey:    "flagKey",
			SegmentKey: "segmentKey",
		}
	)

	store.On("UpdateRule", mock.Anything, req).Return(&flipt.Rule{
		Id:         "1",
		FlagKey:    req.FlagKey,
		SegmentKey: req.SegmentKey,
	}, nil)

	got, err := s.UpdateRule(context.TODO(), req)
	require.NoError(t, err)

	assert.NotNil(t, got)
}

func TestDeleteRule(t *testing.T) {
	var (
		store  = &storeMock{}
		logger = zaptest.NewLogger(t)
		s      = &Server{
			logger: logger,
			store:  store,
		}
		req = &flipt.DeleteRuleRequest{
			Id: "id",
		}
	)

	store.On("DeleteRule", mock.Anything, req).Return(nil)

	got, err := s.DeleteRule(context.TODO(), req)
	require.NoError(t, err)

	assert.NotNil(t, got)
}

func TestOrderRules(t *testing.T) {
	var (
		store  = &storeMock{}
		logger = zaptest.NewLogger(t)
		s      = &Server{
			logger: logger,
			store:  store,
		}
		req = &flipt.OrderRulesRequest{FlagKey: "flagKey", RuleIds: []string{"1", "2"}}
	)

	store.On("OrderRules", mock.Anything, req).Return(nil)

	got, err := s.OrderRules(context.TODO(), req)
	require.NoError(t, err)

	assert.NotNil(t, got)
}

func TestCreateDistribution(t *testing.T) {
	var (
		store  = &storeMock{}
		logger = zaptest.NewLogger(t)
		s      = &Server{
			logger: logger,
			store:  store,
		}
		req = &flipt.CreateDistributionRequest{FlagKey: "flagKey", RuleId: "ruleID", VariantId: "variantID"}
	)

	store.On("CreateDistribution", mock.Anything, req).Return(&flipt.Distribution{
		Id:        "1",
		RuleId:    req.RuleId,
		VariantId: req.VariantId,
	}, nil)

	got, err := s.CreateDistribution(context.TODO(), req)
	require.NoError(t, err)

	assert.NotNil(t, got)
}

func TestUpdateDistribution(t *testing.T) {
	var (
		store  = &storeMock{}
		logger = zaptest.NewLogger(t)
		s      = &Server{
			logger: logger,
			store:  store,
		}
		req = &flipt.UpdateDistributionRequest{Id: "1", FlagKey: "flagKey", RuleId: "ruleID", VariantId: "variantID"}
	)

	store.On("UpdateDistribution", mock.Anything, req).Return(&flipt.Distribution{
		Id:        req.Id,
		RuleId:    req.RuleId,
		VariantId: req.VariantId,
	}, nil)

	got, err := s.UpdateDistribution(context.TODO(), req)
	require.NoError(t, err)

	assert.NotNil(t, got)
}

func TestDeleteDistribution(t *testing.T) {
	var (
		store  = &storeMock{}
		logger = zaptest.NewLogger(t)
		s      = &Server{
			logger: logger,
			store:  store,
		}
		req = &flipt.DeleteDistributionRequest{Id: "1", FlagKey: "flagKey", RuleId: "ruleID", VariantId: "variantID"}
	)

	store.On("DeleteDistribution", mock.Anything, req).Return(nil)

	got, err := s.DeleteDistribution(context.TODO(), req)
	require.NoError(t, err)

	assert.NotNil(t, got)
}
