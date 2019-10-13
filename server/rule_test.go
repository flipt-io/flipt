package server

import (
	"context"
	"errors"
	"testing"

	"github.com/golang/protobuf/ptypes/empty"
	flipt "github.com/markphelps/flipt/rpc"
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
		var (
			f   = tt.f
			req = tt.req
		)

		t.Run(tt.name, func(t *testing.T) {
			s := &Server{
				RuleStore: &ruleStoreMock{
					getRuleFn: f,
				},
			}

			flag, err := s.GetRule(context.TODO(), req)
			require.NoError(t, err)
			assert.NotNil(t, flag)
		})
	}
}

func TestListRules(t *testing.T) {
	tests := []struct {
		name    string
		req     *flipt.ListRuleRequest
		f       func(context.Context, *flipt.ListRuleRequest) ([]*flipt.Rule, error)
		rules   *flipt.RuleList
		wantErr error
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
					{
						FlagKey: "flagKey",
					},
				},
			},
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
			wantErr: emptyFieldError("flagKey"),
		},
		{
			name: "error test",
			req:  &flipt.ListRuleRequest{FlagKey: "flagKey"},
			f: func(ctx context.Context, r *flipt.ListRuleRequest) ([]*flipt.Rule, error) {
				assert.NotNil(t, r)
				assert.Equal(t, "flagKey", r.FlagKey)

				return nil, errors.New("error test")
			},
			wantErr: errors.New("error test"),
		},
	}

	for _, tt := range tests {
		var (
			f       = tt.f
			req     = tt.req
			rules   = tt.rules
			wantErr = tt.wantErr
		)

		t.Run(tt.name, func(t *testing.T) {
			s := &Server{
				RuleStore: &ruleStoreMock{
					listRulesFn: f,
				},
			}

			got, err := s.ListRules(context.TODO(), req)
			assert.Equal(t, wantErr, err)
			assert.Equal(t, rules, got)
		})
	}
}

func TestCreateRule(t *testing.T) {
	tests := []struct {
		name    string
		req     *flipt.CreateRuleRequest
		f       func(context.Context, *flipt.CreateRuleRequest) (*flipt.Rule, error)
		rule    *flipt.Rule
		wantErr error
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
			wantErr: emptyFieldError("flagKey"),
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
			wantErr: emptyFieldError("segmentKey"),
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
			wantErr: invalidFieldError("rank", "must be greater than 0"),
		},
	}

	for _, tt := range tests {
		var (
			f       = tt.f
			req     = tt.req
			rule    = tt.rule
			wantErr = tt.wantErr
		)

		t.Run(tt.name, func(t *testing.T) {
			s := &Server{
				RuleStore: &ruleStoreMock{
					createRuleFn: f,
				},
			}

			got, err := s.CreateRule(context.TODO(), req)
			assert.Equal(t, wantErr, err)
			assert.Equal(t, rule, got)
		})
	}
}

func TestUpdateRule(t *testing.T) {
	tests := []struct {
		name    string
		req     *flipt.UpdateRuleRequest
		f       func(context.Context, *flipt.UpdateRuleRequest) (*flipt.Rule, error)
		rule    *flipt.Rule
		wantErr error
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
			wantErr: emptyFieldError("id"),
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
			wantErr: emptyFieldError("flagKey"),
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
			wantErr: emptyFieldError("segmentKey"),
		},
	}

	for _, tt := range tests {
		var (
			f       = tt.f
			req     = tt.req
			rule    = tt.rule
			wantErr = tt.wantErr
		)

		t.Run(tt.name, func(t *testing.T) {
			s := &Server{
				RuleStore: &ruleStoreMock{
					updateRuleFn: f,
				},
			}

			got, err := s.UpdateRule(context.TODO(), req)
			assert.Equal(t, wantErr, err)
			assert.Equal(t, rule, got)
		})
	}
}

func TestDeleteRule(t *testing.T) {
	tests := []struct {
		name    string
		req     *flipt.DeleteRuleRequest
		f       func(context.Context, *flipt.DeleteRuleRequest) error
		empty   *empty.Empty
		wantErr error
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
			wantErr: emptyFieldError("id"),
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
			wantErr: emptyFieldError("flagKey"),
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
			wantErr: errors.New("error test"),
		},
	}

	for _, tt := range tests {
		var (
			f       = tt.f
			req     = tt.req
			empty   = tt.empty
			wantErr = tt.wantErr
		)

		t.Run(tt.name, func(t *testing.T) {
			s := &Server{
				RuleStore: &ruleStoreMock{
					deleteRuleFn: f,
				},
			}

			got, err := s.DeleteRule(context.TODO(), req)
			assert.Equal(t, wantErr, err)
			assert.Equal(t, empty, got)
		})
	}
}

func TestOrderRules(t *testing.T) {
	tests := []struct {
		name    string
		req     *flipt.OrderRulesRequest
		f       func(context.Context, *flipt.OrderRulesRequest) error
		empty   *empty.Empty
		wantErr error
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
			wantErr: emptyFieldError("flagKey"),
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
			wantErr: invalidFieldError("ruleIds", "must contain atleast 2 elements"),
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
			wantErr: errors.New("error test"),
		},
	}

	for _, tt := range tests {
		var (
			f       = tt.f
			req     = tt.req
			empty   = tt.empty
			wantErr = tt.wantErr
		)

		t.Run(tt.name, func(t *testing.T) {
			s := &Server{
				RuleStore: &ruleStoreMock{
					orderRuleFn: f,
				},
			}

			got, err := s.OrderRules(context.TODO(), req)
			assert.Equal(t, wantErr, err)
			assert.Equal(t, empty, got)
		})
	}
}

func TestCreateDistribution(t *testing.T) {
	tests := []struct {
		name    string
		req     *flipt.CreateDistributionRequest
		f       func(context.Context, *flipt.CreateDistributionRequest) (*flipt.Distribution, error)
		dist    *flipt.Distribution
		wantErr error
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
			wantErr: emptyFieldError("flagKey"),
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
			wantErr: emptyFieldError("ruleId"),
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
			wantErr: emptyFieldError("variantId"),
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
			wantErr: invalidFieldError("rollout", "must be greater than or equal to '0'"),
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
			wantErr: invalidFieldError("rollout", "must be less than or equal to '100'"),
		},
	}

	for _, tt := range tests {
		var (
			f       = tt.f
			req     = tt.req
			dist    = tt.dist
			wantErr = tt.wantErr
		)

		t.Run(tt.name, func(t *testing.T) {
			s := &Server{
				RuleStore: &ruleStoreMock{
					createDistributionFn: f,
				},
			}

			got, err := s.CreateDistribution(context.TODO(), req)
			assert.Equal(t, wantErr, err)
			assert.Equal(t, dist, got)
		})
	}
}

func TestUpdateDistribution(t *testing.T) {
	tests := []struct {
		name    string
		req     *flipt.UpdateDistributionRequest
		f       func(context.Context, *flipt.UpdateDistributionRequest) (*flipt.Distribution, error)
		dist    *flipt.Distribution
		wantErr error
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
			wantErr: emptyFieldError("id"),
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
			wantErr: emptyFieldError("flagKey"),
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
			wantErr: emptyFieldError("ruleId"),
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
			wantErr: emptyFieldError("variantId"),
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
			wantErr: invalidFieldError("rollout", "must be greater than or equal to '0'"),
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
			wantErr: invalidFieldError("rollout", "must be less than or equal to '100'"),
		},
	}

	for _, tt := range tests {
		var (
			f       = tt.f
			req     = tt.req
			dist    = tt.dist
			wantErr = tt.wantErr
		)

		t.Run(tt.name, func(t *testing.T) {
			s := &Server{
				RuleStore: &ruleStoreMock{
					updateDistributionFn: f,
				},
			}

			got, err := s.UpdateDistribution(context.TODO(), req)
			assert.Equal(t, wantErr, err)
			assert.Equal(t, dist, got)
		})
	}
}

func TestDeleteDistribution(t *testing.T) {
	tests := []struct {
		name    string
		req     *flipt.DeleteDistributionRequest
		f       func(context.Context, *flipt.DeleteDistributionRequest) error
		empty   *empty.Empty
		wantErr error
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
			wantErr: emptyFieldError("id"),
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
			wantErr: emptyFieldError("flagKey"),
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
			wantErr: emptyFieldError("ruleId"),
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
			wantErr: emptyFieldError("variantId"),
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
			wantErr: errors.New("error test"),
		},
	}

	for _, tt := range tests {
		var (
			f       = tt.f
			req     = tt.req
			empty   = tt.empty
			wantErr = tt.wantErr
		)

		t.Run(tt.name, func(t *testing.T) {
			s := &Server{
				RuleStore: &ruleStoreMock{
					deleteDistributionFn: f,
				},
			}

			got, err := s.DeleteDistribution(context.TODO(), req)
			assert.Equal(t, wantErr, err)
			assert.Equal(t, empty, got)
		})
	}
}
