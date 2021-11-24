package server

import (
	"context"
	"testing"

	flipt "github.com/markphelps/flipt/rpc/flipt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestGetRule(t *testing.T) {
	var (
		store = &storeMock{}
		s     = &Server{
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

func TestListRules(t *testing.T) {
	var (
		store = &storeMock{}
		s     = &Server{
			logger: logger,
			store:  store,
		}
		req = &flipt.ListRuleRequest{FlagKey: "flagKey"}
	)

	store.On("ListRules", mock.Anything, "flagKey", mock.Anything).Return(
		[]*flipt.Rule{
			{
				FlagKey: req.FlagKey,
			},
		}, nil)

	got, err := s.ListRules(context.TODO(), req)
	require.NoError(t, err)

	assert.NotEmpty(t, got.Rules)
}

func TestCreateRule(t *testing.T) {
	var (
		store = &storeMock{}
		s     = &Server{
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
		store = &storeMock{}
		s     = &Server{
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
		store = &storeMock{}
		s     = &Server{
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
		store = &storeMock{}
		s     = &Server{
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
		store = &storeMock{}
		s     = &Server{
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
		store = &storeMock{}
		s     = &Server{
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
		store = &storeMock{}
		s     = &Server{
			logger: logger,
			store:  store,
		}
		req = &flipt.DeleteRuleRequest{
			Id: "foo",
		}
	)

	store.On("DeleteRule", mock.Anything, req).Return(nil)

	got, err := s.DeleteRule(context.TODO(), req)
	require.NoError(t, err)

	assert.NotNil(t, got)
}
