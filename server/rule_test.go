package server

import (
	"context"
	"errors"
	"testing"

	"github.com/golang/protobuf/ptypes/empty"
	flipt "github.com/markphelps/flipt/proto"
	"github.com/markphelps/flipt/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var _ storage.RuleRepository = &ruleRepositoryMock{}

type ruleRepositoryMock struct {
	ruleFn               func(context.Context, *flipt.GetRuleRequest) (*flipt.Rule, error)
	rulesFn              func(context.Context, *flipt.ListRuleRequest) ([]*flipt.Rule, error)
	createRuleFn         func(context.Context, *flipt.CreateRuleRequest) (*flipt.Rule, error)
	updateRuleFn         func(context.Context, *flipt.UpdateRuleRequest) (*flipt.Rule, error)
	deleteRuleFn         func(context.Context, *flipt.DeleteRuleRequest) error
	orderRuleFn          func(context.Context, *flipt.OrderRulesRequest) error
	createDistributionFn func(context.Context, *flipt.CreateDistributionRequest) (*flipt.Distribution, error)
	updateDistributionFn func(context.Context, *flipt.UpdateDistributionRequest) (*flipt.Distribution, error)
	deleteDistributionFn func(context.Context, *flipt.DeleteDistributionRequest) error
	evaluateFn           func(context.Context, *flipt.EvaluationRequest) (*flipt.EvaluationResponse, error)
}

func (m *ruleRepositoryMock) Rule(ctx context.Context, r *flipt.GetRuleRequest) (*flipt.Rule, error) {
	return m.ruleFn(ctx, r)
}

func (m *ruleRepositoryMock) Rules(ctx context.Context, r *flipt.ListRuleRequest) ([]*flipt.Rule, error) {
	return m.rulesFn(ctx, r)
}

func (m *ruleRepositoryMock) CreateRule(ctx context.Context, r *flipt.CreateRuleRequest) (*flipt.Rule, error) {
	return m.createRuleFn(ctx, r)
}

func (m *ruleRepositoryMock) UpdateRule(ctx context.Context, r *flipt.UpdateRuleRequest) (*flipt.Rule, error) {
	return m.updateRuleFn(ctx, r)
}

func (m *ruleRepositoryMock) DeleteRule(ctx context.Context, r *flipt.DeleteRuleRequest) error {
	return m.deleteRuleFn(ctx, r)
}

func (m *ruleRepositoryMock) OrderRules(ctx context.Context, r *flipt.OrderRulesRequest) error {
	return m.orderRuleFn(ctx, r)
}

func (m *ruleRepositoryMock) CreateDistribution(ctx context.Context, r *flipt.CreateDistributionRequest) (*flipt.Distribution, error) {
	return m.createDistributionFn(ctx, r)
}

func (m *ruleRepositoryMock) UpdateDistribution(ctx context.Context, r *flipt.UpdateDistributionRequest) (*flipt.Distribution, error) {
	return m.updateDistributionFn(ctx, r)
}

func (m *ruleRepositoryMock) DeleteDistribution(ctx context.Context, r *flipt.DeleteDistributionRequest) error {
	return m.deleteDistributionFn(ctx, r)
}

func (m *ruleRepositoryMock) Evaluate(ctx context.Context, r *flipt.EvaluationRequest) (*flipt.EvaluationResponse, error) {
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
				RuleRepository: &ruleRepositoryMock{
					ruleFn: tt.f,
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
		name  string
		req   *flipt.ListRuleRequest
		f     func(context.Context, *flipt.ListRuleRequest) ([]*flipt.Rule, error)
		rules *flipt.RuleList
		e     error
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
			rules: &flipt.RuleList{
				Rules: []*flipt.Rule{
					&flipt.Rule{
						FlagKey: "flagKey",
					},
				},
			},
			e: nil,
		},
		{
			name: "emptyFlagKey",
			req:  &flipt.ListRuleRequest{FlagKey: ""},
			f: func(ctx context.Context, r *flipt.ListRuleRequest) ([]*flipt.Rule, error) {
				assert.NotNil(t, r)
				assert.Equal(t, "", r.FlagKey)

				return []*flipt.Rule{
					{FlagKey: ""},
				}, nil
			},
			rules: nil,
			e:     EmptyFieldError("flagKey"),
		},
		{
			name: "error test",
			req:  &flipt.ListRuleRequest{FlagKey: "flagKey"},
			f: func(ctx context.Context, r *flipt.ListRuleRequest) ([]*flipt.Rule, error) {
				assert.NotNil(t, r)
				assert.Equal(t, "flagKey", r.FlagKey)

				return nil, errors.New("error test")
			},
			rules: nil,
			e:     errors.New("error test"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Server{
				RuleRepository: &ruleRepositoryMock{
					rulesFn: tt.f,
				},
			}

			rules, err := s.ListRules(context.TODO(), tt.req)
			assert.Equal(t, tt.e, err)
			assert.Equal(t, tt.rules, rules)
		})
	}
}

func TestCreateRule(t *testing.T) {
	tests := []struct {
		name string
		req  *flipt.CreateRuleRequest
		f    func(context.Context, *flipt.CreateRuleRequest) (*flipt.Rule, error)
		rule *flipt.Rule
		e    error
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
			rule: &flipt.Rule{
				FlagKey:    "flagKey",
				SegmentKey: "segmentKey",
				Rank:       int32(1),
			},
			e: nil,
		},
		{
			name: "emptyFlagKey",
			req: &flipt.CreateRuleRequest{
				FlagKey:    "",
				SegmentKey: "segmentKey",
				Rank:       1,
			},
			f: func(_ context.Context, r *flipt.CreateRuleRequest) (*flipt.Rule, error) {
				assert.NotNil(t, r)
				assert.Equal(t, "", r.FlagKey)
				assert.Equal(t, "segmentKey", r.SegmentKey)
				assert.Equal(t, int32(1), r.Rank)

				return &flipt.Rule{
					FlagKey:    "",
					SegmentKey: r.SegmentKey,
					Rank:       r.Rank,
				}, nil
			},
			rule: nil,
			e:    EmptyFieldError("flagKey"),
		},
		{
			name: "emptySegmentKey",
			req: &flipt.CreateRuleRequest{
				FlagKey:    "flagKey",
				SegmentKey: "",
				Rank:       1,
			},
			f: func(_ context.Context, r *flipt.CreateRuleRequest) (*flipt.Rule, error) {
				assert.NotNil(t, r)
				assert.Equal(t, "flagKey", r.FlagKey)
				assert.Equal(t, "", r.SegmentKey)
				assert.Equal(t, int32(1), r.Rank)

				return &flipt.Rule{
					FlagKey:    r.FlagKey,
					SegmentKey: "",
					Rank:       r.Rank,
				}, nil
			},
			rule: nil,
			e:    EmptyFieldError("segmentKey"),
		},
		{
			name: "rank_lesser_than_0",
			req: &flipt.CreateRuleRequest{
				FlagKey:    "flagKey",
				SegmentKey: "segmentKey",
				Rank:       -1,
			},
			f: func(_ context.Context, r *flipt.CreateRuleRequest) (*flipt.Rule, error) {
				assert.NotNil(t, r)
				assert.Equal(t, "flagKey", r.FlagKey)
				assert.Equal(t, "segmentKey", r.SegmentKey)
				assert.Equal(t, int32(-1), r.Rank)

				return &flipt.Rule{
					FlagKey:    r.FlagKey,
					SegmentKey: r.SegmentKey,
					Rank:       r.Rank,
				}, nil
			},
			rule: nil,
			e:    InvalidFieldError("rank", "must be greater than 0"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Server{
				RuleRepository: &ruleRepositoryMock{
					createRuleFn: tt.f,
				},
			}

			rule, err := s.CreateRule(context.TODO(), tt.req)
			assert.Equal(t, tt.e, err)
			assert.Equal(t, tt.rule, rule)
		})
	}
}

func TestUpdateRule(t *testing.T) {
	tests := []struct {
		name string
		req  *flipt.UpdateRuleRequest
		f    func(context.Context, *flipt.UpdateRuleRequest) (*flipt.Rule, error)
		rule *flipt.Rule
		e    error
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
			rule: &flipt.Rule{
				Id:         "id",
				FlagKey:    "flagKey",
				SegmentKey: "segmentKey",
			},
			e: nil,
		},
		{
			name: "emptyID",
			req: &flipt.UpdateRuleRequest{
				Id:         "",
				FlagKey:    "flagKey",
				SegmentKey: "segmentKey",
			},
			f: func(_ context.Context, r *flipt.UpdateRuleRequest) (*flipt.Rule, error) {
				assert.NotNil(t, r)
				assert.Equal(t, "", r.Id)
				assert.Equal(t, "flagKey", r.FlagKey)
				assert.Equal(t, "segmentKey", r.SegmentKey)

				return &flipt.Rule{
					Id:         "",
					FlagKey:    r.FlagKey,
					SegmentKey: r.SegmentKey,
				}, nil
			},
			rule: nil,
			e:    EmptyFieldError("id"),
		},
		{
			name: "emptyFlagKey",
			req: &flipt.UpdateRuleRequest{
				Id:         "id",
				FlagKey:    "",
				SegmentKey: "segmentKey",
			},
			f: func(_ context.Context, r *flipt.UpdateRuleRequest) (*flipt.Rule, error) {
				assert.NotNil(t, r)
				assert.Equal(t, "id", r.Id)
				assert.Equal(t, "", r.FlagKey)
				assert.Equal(t, "segmentKey", r.SegmentKey)

				return &flipt.Rule{
					Id:         r.Id,
					FlagKey:    "",
					SegmentKey: r.SegmentKey,
				}, nil
			},
			rule: nil,
			e:    EmptyFieldError("flagKey"),
		},
		{
			name: "emptySegmentKey",
			req: &flipt.UpdateRuleRequest{
				Id:         "id",
				FlagKey:    "flagKey",
				SegmentKey: "",
			},
			f: func(_ context.Context, r *flipt.UpdateRuleRequest) (*flipt.Rule, error) {
				assert.NotNil(t, r)
				assert.Equal(t, "id", r.Id)
				assert.Equal(t, "flagKey", r.FlagKey)
				assert.Equal(t, "", r.SegmentKey)

				return &flipt.Rule{
					Id:         r.Id,
					FlagKey:    r.FlagKey,
					SegmentKey: "",
				}, nil
			},
			rule: nil,
			e:    EmptyFieldError("segmentKey"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Server{
				RuleRepository: &ruleRepositoryMock{
					updateRuleFn: tt.f,
				},
			}

			rule, err := s.UpdateRule(context.TODO(), tt.req)
			assert.Equal(t, tt.e, err)
			assert.Equal(t, tt.rule, rule)
		})
	}
}

func TestDeleteRule(t *testing.T) {
	tests := []struct {
		name  string
		req   *flipt.DeleteRuleRequest
		f     func(context.Context, *flipt.DeleteRuleRequest) error
		empty *empty.Empty
		e     error
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
			empty: &empty.Empty{},
			e:     nil,
		},
		{
			name: "emptyID",
			req:  &flipt.DeleteRuleRequest{Id: "", FlagKey: "flagKey"},
			f: func(_ context.Context, r *flipt.DeleteRuleRequest) error {
				assert.NotNil(t, r)
				assert.Equal(t, "", r.Id)
				assert.Equal(t, "flagKey", r.FlagKey)
				return nil
			},
			empty: nil,
			e:     EmptyFieldError("id"),
		},
		{
			name: "emptyFlagKey",
			req:  &flipt.DeleteRuleRequest{Id: "id", FlagKey: ""},
			f: func(_ context.Context, r *flipt.DeleteRuleRequest) error {
				assert.NotNil(t, r)
				assert.Equal(t, "id", r.Id)
				assert.Equal(t, "", r.FlagKey)
				return nil
			},
			empty: nil,
			e:     EmptyFieldError("flagKey"),
		},
		{
			name: "error test",
			req:  &flipt.DeleteRuleRequest{Id: "id", FlagKey: "flagKey"},
			f: func(_ context.Context, r *flipt.DeleteRuleRequest) error {
				assert.NotNil(t, r)
				assert.Equal(t, "id", r.Id)
				assert.Equal(t, "flagKey", r.FlagKey)
				return errors.New("error test")
			},
			empty: nil,
			e:     errors.New("error test"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Server{
				RuleRepository: &ruleRepositoryMock{
					deleteRuleFn: tt.f,
				},
			}

			resp, err := s.DeleteRule(context.TODO(), tt.req)
			assert.Equal(t, tt.e, err)
			assert.Equal(t, tt.empty, resp)
		})
	}
}

func TestOrderRules(t *testing.T) {
	tests := []struct {
		name  string
		req   *flipt.OrderRulesRequest
		f     func(context.Context, *flipt.OrderRulesRequest) error
		empty *empty.Empty
		e     error
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
			empty: &empty.Empty{},
			e:     nil,
		},
		{
			name: "emptyFlagKey",
			req:  &flipt.OrderRulesRequest{FlagKey: "", RuleIds: []string{"1", "2"}},
			f: func(_ context.Context, r *flipt.OrderRulesRequest) error {
				assert.NotNil(t, r)
				assert.Equal(t, "", r.FlagKey)
				assert.Equal(t, []string{"1", "2"}, r.RuleIds)
				return nil
			},
			empty: nil,
			e:     EmptyFieldError("flagKey"),
		},
		{
			name: "ruleIds length lesser than 2",
			req:  &flipt.OrderRulesRequest{FlagKey: "flagKey", RuleIds: []string{"1"}},
			f: func(_ context.Context, r *flipt.OrderRulesRequest) error {
				assert.NotNil(t, r)
				assert.Equal(t, "", r.FlagKey)
				assert.Equal(t, []string{"1"}, r.RuleIds)
				assert.Equal(t, 1, len(r.RuleIds))
				return nil
			},
			empty: nil,
			e:     InvalidFieldError("ruleIds", "must contain atleast 2 elements"),
		},
		{
			name: "error test",
			req:  &flipt.OrderRulesRequest{FlagKey: "flagKey", RuleIds: []string{"1", "2"}},
			f: func(_ context.Context, r *flipt.OrderRulesRequest) error {
				assert.NotNil(t, r)
				assert.Equal(t, "flagKey", r.FlagKey)
				assert.Equal(t, []string{"1", "2"}, r.RuleIds)
				return errors.New("error test")
			},
			empty: nil,
			e:     errors.New("error test"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Server{
				RuleRepository: &ruleRepositoryMock{
					orderRuleFn: tt.f,
				},
			}

			resp, err := s.OrderRules(context.TODO(), tt.req)
			assert.Equal(t, tt.e, err)
			assert.Equal(t, tt.empty, resp)
		})
	}
}

func TestCreateDistribution(t *testing.T) {
	tests := []struct {
		name string
		req  *flipt.CreateDistributionRequest
		f    func(context.Context, *flipt.CreateDistributionRequest) (*flipt.Distribution, error)
		dist *flipt.Distribution
		e    error
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
			dist: &flipt.Distribution{
				RuleId:    "ruleID",
				VariantId: "variantID",
			},
			e: nil,
		},
		{
			name: "emptyFlagKey",
			req:  &flipt.CreateDistributionRequest{FlagKey: "", RuleId: "ruleID", VariantId: "variantID"},
			f: func(ctx context.Context, r *flipt.CreateDistributionRequest) (*flipt.Distribution, error) {
				assert.NotNil(t, r)
				assert.Equal(t, "", r.FlagKey)
				assert.Equal(t, "ruleID", r.RuleId)
				assert.Equal(t, "variantID", r.VariantId)

				return &flipt.Distribution{
					RuleId:    "ruleID",
					VariantId: "variantID",
				}, nil
			},
			dist: nil,
			e:    EmptyFieldError("flagKey"),
		},
		{
			name: "emptyRuleID",
			req:  &flipt.CreateDistributionRequest{FlagKey: "flagKey", RuleId: "", VariantId: "variantID"},
			f: func(ctx context.Context, r *flipt.CreateDistributionRequest) (*flipt.Distribution, error) {
				assert.NotNil(t, r)
				assert.Equal(t, "flagKey", r.FlagKey)
				assert.Equal(t, "", r.RuleId)
				assert.Equal(t, "variantID", r.VariantId)

				return &flipt.Distribution{
					RuleId:    "",
					VariantId: "variantID",
				}, nil
			},
			dist: nil,
			e:    EmptyFieldError("ruleId"),
		},
		{
			name: "emptyVariantID",
			req:  &flipt.CreateDistributionRequest{FlagKey: "flagKey", RuleId: "ruleID", VariantId: ""},
			f: func(ctx context.Context, r *flipt.CreateDistributionRequest) (*flipt.Distribution, error) {
				assert.NotNil(t, r)
				assert.Equal(t, "flagKey", r.FlagKey)
				assert.Equal(t, "ruleID", r.RuleId)
				assert.Equal(t, "", r.VariantId)

				return &flipt.Distribution{
					RuleId:    "ruleID",
					VariantId: "",
				}, nil
			},
			dist: nil,
			e:    EmptyFieldError("variantId"),
		},
		{
			name: "rollout is less than 0",
			req:  &flipt.CreateDistributionRequest{FlagKey: "flagKey", RuleId: "ruleID", VariantId: "variantID", Rollout: -1},
			f: func(ctx context.Context, r *flipt.CreateDistributionRequest) (*flipt.Distribution, error) {
				assert.NotNil(t, r)
				assert.Equal(t, "flagKey", r.FlagKey)
				assert.Equal(t, "ruleID", r.RuleId)
				assert.Equal(t, "variantID", r.VariantId)

				return &flipt.Distribution{
					RuleId:    "ruleID",
					VariantId: "variantID",
				}, nil
			},
			dist: nil,
			e:    InvalidFieldError("rollout", "must be greater than or equal to '0'"),
		},
		{
			name: "rollout is more than 100",
			req:  &flipt.CreateDistributionRequest{FlagKey: "flagKey", RuleId: "ruleID", VariantId: "variantID", Rollout: 101},
			f: func(ctx context.Context, r *flipt.CreateDistributionRequest) (*flipt.Distribution, error) {
				assert.NotNil(t, r)
				assert.Equal(t, "flagKey", r.FlagKey)
				assert.Equal(t, "ruleID", r.RuleId)
				assert.Equal(t, "variantID", r.VariantId)

				return &flipt.Distribution{
					RuleId:    "ruleID",
					VariantId: "variantID",
				}, nil
			},
			dist: nil,
			e:    InvalidFieldError("rollout", "must be less than or equal to '100'"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Server{
				RuleRepository: &ruleRepositoryMock{
					createDistributionFn: tt.f,
				},
			}

			distribution, err := s.CreateDistribution(context.TODO(), tt.req)
			assert.Equal(t, tt.e, err)
			assert.Equal(t, tt.dist, distribution)
		})
	}
}

func TestUpdateDistribution(t *testing.T) {
	tests := []struct {
		name string
		req  *flipt.UpdateDistributionRequest
		f    func(context.Context, *flipt.UpdateDistributionRequest) (*flipt.Distribution, error)
		dist *flipt.Distribution
		e    error
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
			dist: &flipt.Distribution{
				Id:        "id",
				RuleId:    "ruleID",
				VariantId: "variantID",
			},
			e: nil,
		},
		{
			name: "emptyID",
			req:  &flipt.UpdateDistributionRequest{Id: "", FlagKey: "flagKey", RuleId: "ruleID", VariantId: "variantID"},
			f: func(_ context.Context, r *flipt.UpdateDistributionRequest) (*flipt.Distribution, error) {
				assert.NotNil(t, r)
				assert.Equal(t, "flagKey", r.FlagKey)
				assert.Equal(t, "", r.Id)
				assert.Equal(t, "ruleID", r.RuleId)
				assert.Equal(t, "variantID", r.VariantId)

				return &flipt.Distribution{
					Id:        "",
					RuleId:    r.RuleId,
					VariantId: r.VariantId,
				}, nil
			},
			dist: nil,
			e:    EmptyFieldError("id"),
		},
		{
			name: "emptyFlagKey",
			req:  &flipt.UpdateDistributionRequest{Id: "id", FlagKey: "", RuleId: "ruleID", VariantId: "variantID"},
			f: func(_ context.Context, r *flipt.UpdateDistributionRequest) (*flipt.Distribution, error) {
				assert.NotNil(t, r)
				assert.Equal(t, "", r.FlagKey)
				assert.Equal(t, "id", r.Id)
				assert.Equal(t, "ruleID", r.RuleId)
				assert.Equal(t, "variantID", r.VariantId)

				return &flipt.Distribution{
					Id:        r.Id,
					RuleId:    r.RuleId,
					VariantId: r.VariantId,
				}, nil
			},
			dist: nil,
			e:    EmptyFieldError("flagKey"),
		},
		{
			name: "emptyRuleID",
			req:  &flipt.UpdateDistributionRequest{Id: "id", FlagKey: "flagKey", RuleId: "", VariantId: "variantID"},
			f: func(_ context.Context, r *flipt.UpdateDistributionRequest) (*flipt.Distribution, error) {
				assert.NotNil(t, r)
				assert.Equal(t, "flagKey", r.FlagKey)
				assert.Equal(t, "id", r.Id)
				assert.Equal(t, "", r.RuleId)
				assert.Equal(t, "variantID", r.VariantId)

				return &flipt.Distribution{
					Id:        r.Id,
					RuleId:    "",
					VariantId: r.VariantId,
				}, nil
			},
			dist: nil,
			e:    EmptyFieldError("ruleId"),
		},
		{
			name: "emptyVariantID",
			req:  &flipt.UpdateDistributionRequest{Id: "id", FlagKey: "flagKey", RuleId: "ruleID", VariantId: ""},
			f: func(_ context.Context, r *flipt.UpdateDistributionRequest) (*flipt.Distribution, error) {
				assert.NotNil(t, r)
				assert.Equal(t, "flagKey", r.FlagKey)
				assert.Equal(t, "id", r.Id)
				assert.Equal(t, "ruleID", r.RuleId)
				assert.Equal(t, "", r.VariantId)

				return &flipt.Distribution{
					Id:        r.Id,
					RuleId:    r.RuleId,
					VariantId: "",
				}, nil
			},
			dist: nil,
			e:    EmptyFieldError("variantId"),
		},
		{
			name: "rollout is lesser than 0",
			req:  &flipt.UpdateDistributionRequest{Id: "id", FlagKey: "flagKey", RuleId: "ruleID", VariantId: "variantID", Rollout: -1},
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
			dist: nil,
			e:    InvalidFieldError("rollout", "must be greater than or equal to '0'"),
		},
		{
			name: "rollout is greater than 100",
			req:  &flipt.UpdateDistributionRequest{Id: "id", FlagKey: "flagKey", RuleId: "ruleID", VariantId: "variantID", Rollout: 101},
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
			dist: nil,
			e:    InvalidFieldError("rollout", "must be less than or equal to '100'"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Server{
				RuleRepository: &ruleRepositoryMock{
					updateDistributionFn: tt.f,
				},
			}

			distribution, err := s.UpdateDistribution(context.TODO(), tt.req)
			assert.Equal(t, tt.e, err)
			assert.Equal(t, tt.dist, distribution)
		})
	}
}

func TestDeleteDistribution(t *testing.T) {
	tests := []struct {
		name  string
		req   *flipt.DeleteDistributionRequest
		f     func(context.Context, *flipt.DeleteDistributionRequest) error
		empty *empty.Empty
		e     error
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
			empty: &empty.Empty{},
			e:     nil,
		},
		{
			name: "emptyID",
			req:  &flipt.DeleteDistributionRequest{Id: "", FlagKey: "flagKey", RuleId: "ruleID", VariantId: "variantID"},
			f: func(_ context.Context, r *flipt.DeleteDistributionRequest) error {
				assert.NotNil(t, r)
				assert.Equal(t, "", r.Id)
				assert.Equal(t, "flagKey", r.FlagKey)
				assert.Equal(t, "ruleID", r.RuleId)
				assert.Equal(t, "variantID", r.VariantId)

				return nil
			},
			empty: nil,
			e:     EmptyFieldError("id"),
		},
		{
			name: "emptyFlagKey",
			req:  &flipt.DeleteDistributionRequest{Id: "id", FlagKey: "", RuleId: "ruleID", VariantId: "variantID"},
			f: func(_ context.Context, r *flipt.DeleteDistributionRequest) error {
				assert.NotNil(t, r)
				assert.Equal(t, "id", r.Id)
				assert.Equal(t, "", r.FlagKey)
				assert.Equal(t, "ruleID", r.RuleId)
				assert.Equal(t, "variantID", r.VariantId)

				return nil
			},
			empty: nil,
			e:     EmptyFieldError("flagKey"),
		},
		{
			name: "emptyRuleID",
			req:  &flipt.DeleteDistributionRequest{Id: "id", FlagKey: "flagKey", RuleId: "", VariantId: "variantID"},
			f: func(_ context.Context, r *flipt.DeleteDistributionRequest) error {
				assert.NotNil(t, r)
				assert.Equal(t, "id", r.Id)
				assert.Equal(t, "flagKey", r.FlagKey)
				assert.Equal(t, "", r.RuleId)
				assert.Equal(t, "variantID", r.VariantId)

				return nil
			},
			empty: nil,
			e:     EmptyFieldError("ruleId"),
		},
		{
			name: "emptyVariantID",
			req:  &flipt.DeleteDistributionRequest{Id: "id", FlagKey: "flagKey", RuleId: "ruleID", VariantId: ""},
			f: func(_ context.Context, r *flipt.DeleteDistributionRequest) error {
				assert.NotNil(t, r)
				assert.Equal(t, "id", r.Id)
				assert.Equal(t, "flagKey", r.FlagKey)
				assert.Equal(t, "ruleID", r.RuleId)
				assert.Equal(t, "", r.VariantId)

				return nil
			},
			empty: nil,
			e:     EmptyFieldError("variantId"),
		},
		{
			name: "error test",
			req:  &flipt.DeleteDistributionRequest{Id: "id", FlagKey: "flagKey", RuleId: "ruleID", VariantId: "variantID"},
			f: func(_ context.Context, r *flipt.DeleteDistributionRequest) error {
				assert.NotNil(t, r)
				assert.Equal(t, "id", r.Id)
				assert.Equal(t, "flagKey", r.FlagKey)
				assert.Equal(t, "ruleID", r.RuleId)
				assert.Equal(t, "variantID", r.VariantId)

				return errors.New("error test")
			},
			empty: nil,
			e:     errors.New("error test"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Server{
				RuleRepository: &ruleRepositoryMock{
					deleteDistributionFn: tt.f,
				},
			}

			resp, err := s.DeleteDistribution(context.TODO(), tt.req)
			assert.Equal(t, tt.e, err)
			assert.Equal(t, tt.empty, resp)
		})
	}
}

func TestEvaluate(t *testing.T) {
	tests := []struct {
		name string
		req  *flipt.EvaluationRequest
		f    func(context.Context, *flipt.EvaluationRequest) (*flipt.EvaluationResponse, error)
		eval *flipt.EvaluationResponse
		e    error
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
			eval: &flipt.EvaluationResponse{
				FlagKey:  "flagKey",
				EntityId: "entityID",
			},
			e: nil,
		},
		{
			name: "emptyFlagKey",
			req:  &flipt.EvaluationRequest{FlagKey: "", EntityId: "entityID"},
			f: func(_ context.Context, r *flipt.EvaluationRequest) (*flipt.EvaluationResponse, error) {
				assert.NotNil(t, r)
				assert.Equal(t, "", r.FlagKey)
				assert.Equal(t, "entityID", r.EntityId)

				return &flipt.EvaluationResponse{
					FlagKey:  "",
					EntityId: r.EntityId,
				}, nil
			},
			eval: nil,
			e:    EmptyFieldError("flagKey"),
		},
		{
			name: "emptyEntityId",
			req:  &flipt.EvaluationRequest{FlagKey: "flagKey", EntityId: ""},
			f: func(_ context.Context, r *flipt.EvaluationRequest) (*flipt.EvaluationResponse, error) {
				assert.NotNil(t, r)
				assert.Equal(t, "flagKey", r.FlagKey)
				assert.Equal(t, "", r.EntityId)

				return &flipt.EvaluationResponse{
					FlagKey:  r.FlagKey,
					EntityId: "",
				}, nil
			},
			eval: nil,
			e:    EmptyFieldError("entityId"),
		},
		{
			name: "error test",
			req:  &flipt.EvaluationRequest{FlagKey: "flagKey", EntityId: "entityID"},
			f: func(_ context.Context, r *flipt.EvaluationRequest) (*flipt.EvaluationResponse, error) {
				assert.NotNil(t, r)
				assert.Equal(t, "flagKey", r.FlagKey)
				assert.Equal(t, "entityID", r.EntityId)

				return nil, errors.New("error test")
			},
			eval: nil,
			e:    errors.New("error test"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Server{
				RuleRepository: &ruleRepositoryMock{
					evaluateFn: tt.f,
				},
			}
			resp, err := s.Evaluate(context.TODO(), tt.req)
			assert.Equal(t, tt.e, err)
			if resp != nil {
				assert.NotZero(t, resp.RequestDurationMillis)
			} else {
				assert.Equal(t, tt.eval, resp)
			}
		})
	}
}
