//nolint:goconst
package server

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.flipt.io/flipt/internal/common"
	"go.flipt.io/flipt/internal/storage"
	flipt "go.flipt.io/flipt/rpc/flipt"
	"go.uber.org/zap/zaptest"
)

func TestGetRule(t *testing.T) {
	var (
		store  = &common.StoreMock{}
		logger = zaptest.NewLogger(t)
		s      = &Server{
			logger: logger,
			store:  store,
		}
		req = &flipt.GetRuleRequest{Id: "id", FlagKey: "flagKey"}
	)

	store.On("GetRule", mock.Anything, storage.NewNamespace(""), "id").Return(&flipt.Rule{
		Id: "1",
	}, nil)

	got, err := s.GetRule(context.TODO(), req)
	require.NoError(t, err)

	assert.NotNil(t, got)
}

func TestListRules_PaginationOffset(t *testing.T) {
	var (
		store  = &common.StoreMock{}
		logger = zaptest.NewLogger(t)
		s      = &Server{
			logger: logger,
			store:  store,
		}
	)

	defer store.AssertExpectations(t)

	store.On("ListRules", mock.Anything, storage.ListWithOptions(storage.NewResource("", "flagKey"),
		storage.ListWithQueryParamOptions[storage.ResourceRequest](
			storage.WithOffset(10),
		),
	)).Return(
		storage.ResultSet[*flipt.Rule]{
			Results: []*flipt.Rule{
				{
					FlagKey: "flagKey",
				},
			},
			NextPageToken: "YmFy",
		}, nil)

	store.On("CountRules", mock.Anything, storage.NewResource("", "flagKey")).Return(uint64(1), nil)

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
		store  = &common.StoreMock{}
		logger = zaptest.NewLogger(t)
		s      = &Server{
			logger: logger,
			store:  store,
		}
	)

	defer store.AssertExpectations(t)

	store.On("ListRules", mock.Anything, storage.ListWithOptions(storage.NewResource("", "flagKey"),
		storage.ListWithQueryParamOptions[storage.ResourceRequest](
			storage.WithPageToken("Zm9v"),
			storage.WithOffset(10),
		),
	)).Return(
		storage.ResultSet[*flipt.Rule]{
			Results: []*flipt.Rule{
				{
					FlagKey: "flagKey",
				},
			},
			NextPageToken: "YmFy",
		}, nil)

	store.On("CountRules", mock.Anything, storage.NewResource("", "flagKey")).Return(uint64(1), nil)

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
		store  = &common.StoreMock{}
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

func TestCreateRule_MultipleSegments(t *testing.T) {
	var (
		store  = &common.StoreMock{}
		logger = zaptest.NewLogger(t)
		s      = &Server{
			logger: logger,
			store:  store,
		}
		req = &flipt.CreateRuleRequest{
			FlagKey:         "flagKey",
			SegmentKeys:     []string{"segmentKey1", "segmentKey2"},
			SegmentOperator: flipt.SegmentOperator_AND_SEGMENT_OPERATOR,
			Rank:            1,
		}
	)

	store.On("CreateRule", mock.Anything, req).Return(&flipt.Rule{
		Id:              "1",
		FlagKey:         req.FlagKey,
		SegmentKeys:     req.SegmentKeys,
		SegmentOperator: req.SegmentOperator,
		Rank:            req.Rank,
	}, nil)

	got, err := s.CreateRule(context.TODO(), req)
	require.NoError(t, err)

	assert.NotNil(t, got)
}

func TestUpdateRule(t *testing.T) {
	var (
		store  = &common.StoreMock{}
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
		store  = &common.StoreMock{}
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
		store  = &common.StoreMock{}
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
		store  = &common.StoreMock{}
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
		store  = &common.StoreMock{}
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
		store  = &common.StoreMock{}
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
