package server

import (
	"context"
	"testing"

	"github.com/golang/protobuf/ptypes/empty"
	flipt "github.com/markphelps/flipt/proto"
	"github.com/markphelps/flipt/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var _ storage.RuleStore = &ruleStoreMock{}

type ruleStoreMock struct {
	getRuleFn            func(context.Context, *flipt.GetRuleRequest) (*flipt.Rule, error)
	listRulesFn          func(context.Context, *flipt.ListRuleRequest) ([]*flipt.Rule, error)
	createRuleFn         func(context.Context, *flipt.CreateRuleRequest) (*flipt.Rule, error)
	updateRuleFn         func(context.Context, *flipt.UpdateRuleRequest) (*flipt.Rule, error)
	deleteRuleFn         func(context.Context, *flipt.DeleteRuleRequest) error
	orderRuleFn          func(context.Context, *flipt.OrderRulesRequest) error
	createDistributionFn func(context.Context, *flipt.CreateDistributionRequest) (*flipt.Distribution, error)
	updateDistributionFn func(context.Context, *flipt.UpdateDistributionRequest) (*flipt.Distribution, error)
	deleteDistributionFn func(context.Context, *flipt.DeleteDistributionRequest) error
	evaluateFn           func(context.Context, *flipt.EvaluationRequest) (*flipt.EvaluationResponse, error)
}

func (m *ruleStoreMock) GetRule(ctx context.Context, r *flipt.GetRuleRequest) (*flipt.Rule, error) {
	return m.getRuleFn(ctx, r)
}

func (m *ruleStoreMock) ListRules(ctx context.Context, r *flipt.ListRuleRequest) ([]*flipt.Rule, error) {
	return m.listRulesFn(ctx, r)
}

func (m *ruleStoreMock) CreateRule(ctx context.Context, r *flipt.CreateRuleRequest) (*flipt.Rule, error) {
	return m.createRuleFn(ctx, r)
}

func (m *ruleStoreMock) UpdateRule(ctx context.Context, r *flipt.UpdateRuleRequest) (*flipt.Rule, error) {
	return m.updateRuleFn(ctx, r)
}

func (m *ruleStoreMock) DeleteRule(ctx context.Context, r *flipt.DeleteRuleRequest) error {
	return m.deleteRuleFn(ctx, r)
}

func (m *ruleStoreMock) OrderRules(ctx context.Context, r *flipt.OrderRulesRequest) error {
	return m.orderRuleFn(ctx, r)
}

func (m *ruleStoreMock) CreateDistribution(ctx context.Context, r *flipt.CreateDistributionRequest) (*flipt.Distribution, error) {
	return m.createDistributionFn(ctx, r)
}

func (m *ruleStoreMock) UpdateDistribution(ctx context.Context, r *flipt.UpdateDistributionRequest) (*flipt.Distribution, error) {
	return m.updateDistributionFn(ctx, r)
}

func (m *ruleStoreMock) DeleteDistribution(ctx context.Context, r *flipt.DeleteDistributionRequest) error {
	return m.deleteDistributionFn(ctx, r)
}

func (m *ruleStoreMock) Evaluate(ctx context.Context, r *flipt.EvaluationRequest) (*flipt.EvaluationResponse, error) {
	return m.evaluateFn(ctx, r)
}

func TestGetRule(t *testing.T) {
	tests := []struct {
		name string
		req  *flipt.GetRuleRequest
		f    func(context.Context, *flipt.GetRuleRequest) (*flipt.Rule, error)
	}{
		{
			name: "ok",
			req:  &flipt.GetRuleRequest{Id: "id", FlagKey: "flagKey"},
			f: func(_ context.Context, r *flipt.GetRuleRequest) (*flipt.Rule, error) {
				assert.NotNil(t, r)
				assert.Equal(t, "id", r.Id)
				assert.Equal(t, "flagKey", r.FlagKey)

				return &flipt.Rule{
					Id:      r.Id,
					FlagKey: r.FlagKey,
				}, nil
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Server{
				RuleStore: &ruleStoreMock{
					getRuleFn: tt.f,
				},
			}

			flag, err := s.GetRule(context.TODO(), tt.req)
			require.NoError(t, err)
			assert.NotNil(t, flag)
		})
	}
}

func TestListRules(t *testing.T) {
	tests := []struct {
		name string
		req  *flipt.ListRuleRequest
		f    func(context.Context, *flipt.ListRuleRequest) ([]*flipt.Rule, error)
	}{
		{
			name: "ok",
			req:  &flipt.ListRuleRequest{FlagKey: "flagKey"},
			f: func(ctx context.Context, r *flipt.ListRuleRequest) ([]*flipt.Rule, error) {
				assert.NotNil(t, r)
				assert.Equal(t, "flagKey", r.FlagKey)

				return []*flipt.Rule{
					{FlagKey: r.FlagKey},
				}, nil
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Server{
				RuleStore: &ruleStoreMock{
					listRulesFn: tt.f,
				},
			}

			rules, err := s.ListRules(context.TODO(), tt.req)
			require.NoError(t, err)
			assert.NotEmpty(t, rules)
		})
	}
}

func TestCreateRule(t *testing.T) {
	tests := []struct {
		name string
		req  *flipt.CreateRuleRequest
		f    func(context.Context, *flipt.CreateRuleRequest) (*flipt.Rule, error)
	}{
		{
			name: "ok",
			req: &flipt.CreateRuleRequest{
				FlagKey:    "flagKey",
				SegmentKey: "segmentKey",
				Rank:       1,
			},
			f: func(_ context.Context, r *flipt.CreateRuleRequest) (*flipt.Rule, error) {
				assert.NotNil(t, r)
				assert.Equal(t, "flagKey", r.FlagKey)
				assert.Equal(t, "segmentKey", r.SegmentKey)
				assert.Equal(t, int32(1), r.Rank)

				return &flipt.Rule{
					FlagKey:    r.FlagKey,
					SegmentKey: r.SegmentKey,
					Rank:       r.Rank,
				}, nil
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Server{
				RuleStore: &ruleStoreMock{
					createRuleFn: tt.f,
				},
			}

			rule, err := s.CreateRule(context.TODO(), tt.req)
			require.NoError(t, err)
			assert.NotNil(t, rule)
		})
	}
}

func TestUpdateRule(t *testing.T) {
	tests := []struct {
		name string
		req  *flipt.UpdateRuleRequest
		f    func(context.Context, *flipt.UpdateRuleRequest) (*flipt.Rule, error)
	}{
		{
			name: "ok",
			req: &flipt.UpdateRuleRequest{
				Id:         "id",
				FlagKey:    "flagKey",
				SegmentKey: "segmentKey",
			},
			f: func(_ context.Context, r *flipt.UpdateRuleRequest) (*flipt.Rule, error) {
				assert.NotNil(t, r)
				assert.Equal(t, "id", r.Id)
				assert.Equal(t, "flagKey", r.FlagKey)
				assert.Equal(t, "segmentKey", r.SegmentKey)

				return &flipt.Rule{
					Id:         r.Id,
					FlagKey:    r.FlagKey,
					SegmentKey: r.SegmentKey,
				}, nil
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Server{
				RuleStore: &ruleStoreMock{
					updateRuleFn: tt.f,
				},
			}

			rule, err := s.UpdateRule(context.TODO(), tt.req)
			require.NoError(t, err)
			assert.NotNil(t, rule)
		})
	}
}

func TestDeleteRule(t *testing.T) {
	tests := []struct {
		name string
		req  *flipt.DeleteRuleRequest
		f    func(context.Context, *flipt.DeleteRuleRequest) error
	}{
		{
			name: "ok",
			req:  &flipt.DeleteRuleRequest{Id: "id", FlagKey: "flagKey"},
			f: func(_ context.Context, r *flipt.DeleteRuleRequest) error {
				assert.NotNil(t, r)
				assert.Equal(t, "id", r.Id)
				assert.Equal(t, "flagKey", r.FlagKey)
				return nil
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Server{
				RuleStore: &ruleStoreMock{
					deleteRuleFn: tt.f,
				},
			}

			resp, err := s.DeleteRule(context.TODO(), tt.req)
			require.NoError(t, err)
			assert.Equal(t, &empty.Empty{}, resp)
		})
	}
}

func TestOrderRules(t *testing.T) {
	tests := []struct {
		name string
		req  *flipt.OrderRulesRequest
		f    func(context.Context, *flipt.OrderRulesRequest) error
	}{
		{
			name: "ok",
			req:  &flipt.OrderRulesRequest{FlagKey: "flagKey", RuleIds: []string{"1", "2"}},
			f: func(_ context.Context, r *flipt.OrderRulesRequest) error {
				assert.NotNil(t, r)
				assert.Equal(t, "flagKey", r.FlagKey)
				assert.Equal(t, []string{"1", "2"}, r.RuleIds)
				return nil
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Server{
				RuleStore: &ruleStoreMock{
					orderRuleFn: tt.f,
				},
			}

			resp, err := s.OrderRules(context.TODO(), tt.req)
			require.NoError(t, err)
			assert.Equal(t, &empty.Empty{}, resp)
		})
	}
}

func TestCreateDistribution(t *testing.T) {
	tests := []struct {
		name string
		req  *flipt.CreateDistributionRequest
		f    func(context.Context, *flipt.CreateDistributionRequest) (*flipt.Distribution, error)
	}{
		{
			name: "ok",
			req:  &flipt.CreateDistributionRequest{FlagKey: "flagKey", RuleId: "ruleID", VariantId: "variantID"},
			f: func(ctx context.Context, r *flipt.CreateDistributionRequest) (*flipt.Distribution, error) {
				assert.NotNil(t, r)
				assert.Equal(t, "flagKey", r.FlagKey)
				assert.Equal(t, "ruleID", r.RuleId)
				assert.Equal(t, "variantID", r.VariantId)

				return &flipt.Distribution{
					RuleId:    r.RuleId,
					VariantId: r.VariantId,
				}, nil
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Server{
				RuleStore: &ruleStoreMock{
					createDistributionFn: tt.f,
				},
			}

			distribution, err := s.CreateDistribution(context.TODO(), tt.req)
			require.NoError(t, err)
			assert.NotNil(t, distribution)
		})
	}
}

func TestUpdateDistribution(t *testing.T) {
	tests := []struct {
		name string
		req  *flipt.UpdateDistributionRequest
		f    func(context.Context, *flipt.UpdateDistributionRequest) (*flipt.Distribution, error)
	}{
		{
			name: "ok",
			req:  &flipt.UpdateDistributionRequest{Id: "id", FlagKey: "flagKey", RuleId: "ruleID", VariantId: "variantID"},
			f: func(_ context.Context, r *flipt.UpdateDistributionRequest) (*flipt.Distribution, error) {
				assert.NotNil(t, r)
				assert.Equal(t, "flagKey", r.FlagKey)
				assert.Equal(t, "id", r.Id)
				assert.Equal(t, "ruleID", r.RuleId)
				assert.Equal(t, "variantID", r.VariantId)

				return &flipt.Distribution{
					Id:        r.Id,
					RuleId:    r.RuleId,
					VariantId: r.VariantId,
				}, nil
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Server{
				RuleStore: &ruleStoreMock{
					updateDistributionFn: tt.f,
				},
			}

			distribution, err := s.UpdateDistribution(context.TODO(), tt.req)
			require.NoError(t, err)
			assert.NotNil(t, distribution)
		})
	}
}

func TestDeleteDistribution(t *testing.T) {
	tests := []struct {
		name string
		req  *flipt.DeleteDistributionRequest
		f    func(context.Context, *flipt.DeleteDistributionRequest) error
	}{
		{
			name: "ok",
			req:  &flipt.DeleteDistributionRequest{Id: "id", FlagKey: "flagKey", RuleId: "ruleID", VariantId: "variantID"},
			f: func(_ context.Context, r *flipt.DeleteDistributionRequest) error {
				assert.NotNil(t, r)
				assert.Equal(t, "id", r.Id)
				assert.Equal(t, "flagKey", r.FlagKey)
				assert.Equal(t, "ruleID", r.RuleId)
				assert.Equal(t, "variantID", r.VariantId)

				return nil
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Server{
				RuleStore: &ruleStoreMock{
					deleteDistributionFn: tt.f,
				},
			}

			resp, err := s.DeleteDistribution(context.TODO(), tt.req)
			require.NoError(t, err)
			assert.Equal(t, &empty.Empty{}, resp)
		})
	}
}

func TestEvaluate(t *testing.T) {
	tests := []struct {
		name string
		req  *flipt.EvaluationRequest
		f    func(context.Context, *flipt.EvaluationRequest) (*flipt.EvaluationResponse, error)
	}{
		{
			name: "ok",
			req:  &flipt.EvaluationRequest{FlagKey: "flagKey", EntityId: "entityID"},
			f: func(_ context.Context, r *flipt.EvaluationRequest) (*flipt.EvaluationResponse, error) {
				assert.NotNil(t, r)
				assert.Equal(t, "flagKey", r.FlagKey)
				assert.Equal(t, "entityID", r.EntityId)

				return &flipt.EvaluationResponse{
					FlagKey:  r.FlagKey,
					EntityId: r.EntityId,
				}, nil
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Server{
				RuleStore: &ruleStoreMock{
					evaluateFn: tt.f,
				},
			}
			resp, err := s.Evaluate(context.TODO(), tt.req)
			require.NoError(t, err)
			assert.NotNil(t, resp)
			assert.NotZero(t, resp.RequestDurationMillis)
		})
	}
}
