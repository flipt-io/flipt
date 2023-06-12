//nolint:goconst
package server

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.flipt.io/flipt/internal/storage"
	flipt "go.flipt.io/flipt/rpc/flipt"
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

	store.On("GetRule", mock.Anything, mock.Anything, "id").Return(&flipt.Rule{
		Id: "1",
	}, nil)

	got, err := s.GetRule(context.TODO(), req)
	require.NoError(t, err)

	assert.NotNil(t, got)
}

func TestListRules_PaginationOffset(t *testing.T) {
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
	store.On("ListRules", mock.Anything, mock.Anything, "flagKey", mock.MatchedBy(func(opts []storage.QueryOption) bool {
		for _, opt := range opts {
			opt(&params)
		}

		// assert offset is provided
		return params.PageToken == "" && params.Offset > 0
	})).Return(
		storage.ResultSet[*flipt.Rule]{
			Results: []*flipt.Rule{
				{
					FlagKey: "flagKey",
				},
			},
			NextPageToken: "bar",
		}, nil)

	store.On("CountRules", mock.Anything, mock.Anything, mock.Anything).Return(uint64(1), nil)

	got, err := s.ListRules(context.TODO(), &flipt.ListRuleRequest{FlagKey: "flagKey",
		Offset: 10,
	})

	require.NoError(t, err)

	assert.NotEmpty(t, got.Rules)
	assert.Equal(t, "YmFy", got.NextPageToken)
	assert.Equal(t, int32(1), got.TotalCount)
}

func TestListRules_PaginationPageToken(t *testing.T) {
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
	store.On("ListRules", mock.Anything, mock.Anything, "flagKey", mock.MatchedBy(func(opts []storage.QueryOption) bool {
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

	store.On("CountRules", mock.Anything, mock.Anything, mock.Anything).Return(uint64(1), nil)

	got, err := s.ListRules(context.TODO(), &flipt.ListRuleRequest{FlagKey: "flagKey",
		PageToken: "Zm9v",
		Offset:    10,
	})

	require.NoError(t, err)

	assert.NotEmpty(t, got.Rules)
	assert.Equal(t, "YmFy", got.NextPageToken)
	assert.Equal(t, int32(1), got.TotalCount)
}

func TestListRules_PaginationInvalidPageToken(t *testing.T) {
	var (
		store  = &storeMock{}
		logger = zaptest.NewLogger(t)
		s      = &Server{
			logger: logger,
			store:  store,
		}
	)

	defer store.AssertExpectations(t)

	store.AssertNotCalled(t, "ListRules")

	_, err := s.ListRules(context.TODO(), &flipt.ListRuleRequest{FlagKey: "flagKey",
		PageToken: "Invalid string",
		Offset:    10,
	})

	assert.EqualError(t, err, `pageToken is not valid: "Invalid string"`)
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
