package server

import (
	"context"
	"testing"

	"github.com/gofrs/uuid"
	"github.com/markphelps/flipt/errors"
	flipt "github.com/markphelps/flipt/rpc"
	"github.com/markphelps/flipt/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var _ storage.EvaluationStore = &evaluationStoreMock{}

type evaluationStoreMock struct {
	getEvaluationRulesFn         func(context.Context, string) ([]*storage.EvaluationRule, error)
	getEvaluationDistributionsFn func(context.Context, string) ([]*storage.EvaluationDistribution, error)
}

func (m *evaluationStoreMock) GetEvaluationRules(ctx context.Context, flagKey string) ([]*storage.EvaluationRule, error) {
	if m.getEvaluationRulesFn == nil {
		return []*storage.EvaluationRule{}, nil
	}
	return m.getEvaluationRulesFn(ctx, flagKey)
}

func (m *evaluationStoreMock) GetEvaluationDistributions(ctx context.Context, ruleID string) ([]*storage.EvaluationDistribution, error) {
	if m.getEvaluationDistributionsFn == nil {
		return []*storage.EvaluationDistribution{}, nil
	}
	return m.getEvaluationDistributionsFn(ctx, ruleID)
}

func TestEvaluate_FlagNotFound(t *testing.T) {
	s := &Server{
		logger: logger,
		FlagStore: &flagStoreMock{
			getFlagFn: func(ctx context.Context, r *flipt.GetFlagRequest) (*flipt.Flag, error) {
				return nil, errors.ErrNotFoundf("flag %q", r.Key)
			},
		},
	}

	resp, err := s.Evaluate(context.TODO(), &flipt.EvaluationRequest{
		FlagKey: "foo",
		Context: map[string]string{
			"bar": "boz",
		},
	})

	require.Error(t, err)
	assert.EqualError(t, err, "flag \"foo\" not found")
	assert.False(t, resp.Match)
}

func TestEvaluate_FlagDisabled(t *testing.T) {
	s := &Server{
		logger: logger,
		FlagStore: &flagStoreMock{
			getFlagFn: func(ctx context.Context, r *flipt.GetFlagRequest) (*flipt.Flag, error) {
				return &flipt.Flag{
					Key:     "foo",
					Enabled: false,
				}, nil
			},
		},
	}

	resp, err := s.Evaluate(context.TODO(), &flipt.EvaluationRequest{
		EntityId: "1",
		FlagKey:  "foo",
		Context: map[string]string{
			"bar": "boz",
		},
	})

	require.Error(t, err)
	assert.EqualError(t, err, "flag \"foo\" is disabled")
	assert.False(t, resp.Match)
}

func TestEvaluate_FlagNoRules(t *testing.T) {
	s := &Server{
		logger: logger,
		FlagStore: &flagStoreMock{
			getFlagFn: func(ctx context.Context, r *flipt.GetFlagRequest) (*flipt.Flag, error) {
				return &flipt.Flag{
					Key:     "foo",
					Enabled: true,
				}, nil
			},
		},
		EvaluationStore: &evaluationStoreMock{},
	}

	resp, err := s.Evaluate(context.TODO(), &flipt.EvaluationRequest{
		EntityId: "1",
		FlagKey:  "foo",
		Context: map[string]string{
			"bar": "boz",
		},
	})

	require.NoError(t, err)
	assert.False(t, resp.Match)
}

func TestEvaluate_RulesOutOfOrder(t *testing.T) {
	s := &Server{
		logger: logger,
		FlagStore: &flagStoreMock{
			getFlagFn: func(ctx context.Context, r *flipt.GetFlagRequest) (*flipt.Flag, error) {
				return &flipt.Flag{
					Key:     "foo",
					Enabled: true,
				}, nil
			},
		},
		EvaluationStore: &evaluationStoreMock{
			getEvaluationRulesFn: func(context.Context, string) ([]*storage.EvaluationRule, error) {
				return []*storage.EvaluationRule{
					{
						ID:               "1",
						FlagKey:          "foo",
						SegmentKey:       "bar",
						SegmentMatchType: flipt.MatchType_ALL_MATCH_TYPE,
						Rank:             1,
						Constraints: []storage.EvaluationConstraint{
							{
								ID:       "2",
								Type:     flipt.ComparisonType_STRING_COMPARISON_TYPE,
								Property: "bar",
								Operator: flipt.OpEQ,
								Value:    "baz",
							},
						},
					},
					{
						ID:               "2",
						FlagKey:          "foo",
						SegmentKey:       "bar",
						SegmentMatchType: flipt.MatchType_ALL_MATCH_TYPE,
						Rank:             0,
						Constraints: []storage.EvaluationConstraint{
							{
								ID:       "2",
								Type:     flipt.ComparisonType_STRING_COMPARISON_TYPE,
								Property: "bar",
								Operator: flipt.OpEQ,
								Value:    "baz",
							},
						},
					},
				}, nil
			},
		},
	}

	resp, err := s.Evaluate(context.TODO(), &flipt.EvaluationRequest{
		EntityId: "1",
		FlagKey:  "foo",
		Context: map[string]string{
			"bar": "boz",
		},
	})

	require.Error(t, err)
	assert.EqualError(t, err, "rule rank: 0 detected out of order")
	assert.False(t, resp.Match)
}

// Match ALL constraints
func TestEvaluate_MatchAll_NoVariants_NoDistributions(t *testing.T) {
	s := &Server{
		logger: logger,
		FlagStore: &flagStoreMock{
			getFlagFn: func(ctx context.Context, r *flipt.GetFlagRequest) (*flipt.Flag, error) {
				return &flipt.Flag{
					Key:     "foo",
					Enabled: true,
				}, nil
			},
		},
		EvaluationStore: &evaluationStoreMock{
			getEvaluationRulesFn: func(context.Context, string) ([]*storage.EvaluationRule, error) {
				return []*storage.EvaluationRule{
					{
						ID:               "1",
						FlagKey:          "foo",
						SegmentKey:       "bar",
						SegmentMatchType: flipt.MatchType_ALL_MATCH_TYPE,
						Rank:             0,
						Constraints: []storage.EvaluationConstraint{
							{
								ID:       "2",
								Type:     flipt.ComparisonType_STRING_COMPARISON_TYPE,
								Property: "bar",
								Operator: flipt.OpEQ,
								Value:    "baz",
							},
						},
					},
				}, nil
			},
		},
	}

	tests := []struct {
		name      string
		req       *flipt.EvaluationRequest
		wantMatch bool
	}{
		{
			name: "match string value",
			req: &flipt.EvaluationRequest{
				FlagKey:  "foo",
				EntityId: "1",
				Context: map[string]string{
					"bar": "baz",
				},
			},
			wantMatch: true,
		},
		{
			name: "no match string value",
			req: &flipt.EvaluationRequest{
				FlagKey:  "foo",
				EntityId: "1",
				Context: map[string]string{
					"bar": "boz",
				},
			},
		},
	}

	for _, tt := range tests {
		var (
			req       = tt.req
			wantMatch = tt.wantMatch
		)

		t.Run(tt.name, func(t *testing.T) {
			resp, err := s.Evaluate(context.TODO(), req)
			require.NoError(t, err)
			assert.NotNil(t, resp)
			assert.Equal(t, "foo", resp.FlagKey)
			assert.Equal(t, req.Context, resp.RequestContext)

			if !wantMatch {
				assert.False(t, resp.Match)
				assert.Empty(t, resp.SegmentKey)
				return
			}

			assert.True(t, resp.Match)
			assert.Equal(t, "bar", resp.SegmentKey)
			assert.Empty(t, resp.Value)
		})
	}
}

func TestEvaluate_MatchAll_SingleVariantDistribution(t *testing.T) {
	s := &Server{
		logger: logger,
		FlagStore: &flagStoreMock{
			getFlagFn: func(ctx context.Context, r *flipt.GetFlagRequest) (*flipt.Flag, error) {
				return &flipt.Flag{
					Key:     "foo",
					Enabled: true,
				}, nil
			},
		},
		EvaluationStore: &evaluationStoreMock{
			getEvaluationRulesFn: func(context.Context, string) ([]*storage.EvaluationRule, error) {
				return []*storage.EvaluationRule{
					{
						ID:               "1",
						FlagKey:          "foo",
						SegmentKey:       "bar",
						SegmentMatchType: flipt.MatchType_ALL_MATCH_TYPE,
						Rank:             0,
						Constraints: []storage.EvaluationConstraint{
							// constraint: bar (string) == baz
							{
								ID:       "2",
								Type:     flipt.ComparisonType_STRING_COMPARISON_TYPE,
								Property: "bar",
								Operator: flipt.OpEQ,
								Value:    "baz",
							},
							// constraint: admin (bool) == true
							{
								ID:       "3",
								Type:     flipt.ComparisonType_BOOLEAN_COMPARISON_TYPE,
								Property: "admin",
								Operator: flipt.OpTrue,
							},
						},
					},
				}, nil
			},
			getEvaluationDistributionsFn: func(context.Context, string) ([]*storage.EvaluationDistribution, error) {
				return []*storage.EvaluationDistribution{
					{
						ID:         "4",
						RuleID:     "1",
						VariantID:  "5",
						Rollout:    100,
						VariantKey: "boz",
					},
				}, nil
			},
		},
	}

	tests := []struct {
		name      string
		req       *flipt.EvaluationRequest
		wantMatch bool
	}{
		{
			name: "matches all",
			req: &flipt.EvaluationRequest{
				FlagKey:  "foo",
				EntityId: "1",
				Context: map[string]string{
					"bar":   "baz",
					"admin": "true",
				},
			},
			wantMatch: true,
		},
		{
			name: "no match all",
			req: &flipt.EvaluationRequest{
				FlagKey:  "foo",
				EntityId: "1",
				Context: map[string]string{
					"bar":   "boz",
					"admin": "true",
				},
			},
		},
		{
			name: "no match just bool value",
			req: &flipt.EvaluationRequest{
				FlagKey:  "foo",
				EntityId: "1",
				Context: map[string]string{
					"admin": "true",
				},
			},
		},
		{
			name: "no match just string value",
			req: &flipt.EvaluationRequest{
				FlagKey:  "foo",
				EntityId: "1",
				Context: map[string]string{
					"bar": "baz",
				},
			},
		},
	}

	for _, tt := range tests {
		var (
			req       = tt.req
			wantMatch = tt.wantMatch
		)

		t.Run(tt.name, func(t *testing.T) {
			resp, err := s.Evaluate(context.TODO(), req)
			require.NoError(t, err)
			assert.NotNil(t, resp)
			assert.Equal(t, "foo", resp.FlagKey)
			assert.Equal(t, req.Context, resp.RequestContext)

			if !wantMatch {
				assert.False(t, resp.Match)
				assert.Empty(t, resp.SegmentKey)
				return
			}

			assert.True(t, resp.Match)
			assert.Equal(t, "bar", resp.SegmentKey)
			assert.Equal(t, "boz", resp.Value)
		})
	}
}

func TestEvaluate_MatchAll_RolloutDistribution(t *testing.T) {
	s := &Server{
		logger: logger,
		FlagStore: &flagStoreMock{
			getFlagFn: func(ctx context.Context, r *flipt.GetFlagRequest) (*flipt.Flag, error) {
				return &flipt.Flag{
					Key:     "foo",
					Enabled: true,
				}, nil
			},
		},
		EvaluationStore: &evaluationStoreMock{
			getEvaluationRulesFn: func(context.Context, string) ([]*storage.EvaluationRule, error) {
				return []*storage.EvaluationRule{
					{
						ID:               "1",
						FlagKey:          "foo",
						SegmentKey:       "bar",
						SegmentMatchType: flipt.MatchType_ALL_MATCH_TYPE,
						Rank:             0,
						Constraints: []storage.EvaluationConstraint{
							// constraint: bar (string) == baz
							{
								ID:       "2",
								Type:     flipt.ComparisonType_STRING_COMPARISON_TYPE,
								Property: "bar",
								Operator: flipt.OpEQ,
								Value:    "baz",
							},
						},
					},
				}, nil
			},
			getEvaluationDistributionsFn: func(context.Context, string) ([]*storage.EvaluationDistribution, error) {
				return []*storage.EvaluationDistribution{
					{
						ID:         "4",
						RuleID:     "1",
						VariantID:  "5",
						Rollout:    50,
						VariantKey: "boz",
					},
					{
						ID:         "6",
						RuleID:     "1",
						VariantID:  "7",
						Rollout:    50,
						VariantKey: "booz",
					},
				}, nil
			},
		},
	}

	tests := []struct {
		name              string
		req               *flipt.EvaluationRequest
		matchesVariantKey string
		wantMatch         bool
	}{
		{
			name: "match string value - variant 1",
			req: &flipt.EvaluationRequest{
				FlagKey:  "foo",
				EntityId: "1",
				Context: map[string]string{
					"bar": "baz",
				},
			},
			matchesVariantKey: "boz",
			wantMatch:         true,
		},
		{
			name: "match string value - variant 2",
			req: &flipt.EvaluationRequest{
				FlagKey:  "foo",
				EntityId: "2",
				Context: map[string]string{
					"bar": "baz",
				},
			},
			matchesVariantKey: "booz",
			wantMatch:         true,
		},
		{
			name: "no match string value",
			req: &flipt.EvaluationRequest{
				FlagKey:  "foo",
				EntityId: "1",
				Context: map[string]string{
					"bar": "boz",
				},
			},
		},
	}

	for _, tt := range tests {
		var (
			req               = tt.req
			matchesVariantKey = tt.matchesVariantKey
			wantMatch         = tt.wantMatch
		)

		t.Run(tt.name, func(t *testing.T) {
			resp, err := s.Evaluate(context.TODO(), req)
			require.NoError(t, err)
			assert.NotNil(t, resp)
			assert.Equal(t, "foo", resp.FlagKey)
			assert.Equal(t, req.Context, resp.RequestContext)

			if !wantMatch {
				assert.False(t, resp.Match)
				assert.Empty(t, resp.SegmentKey)
				return
			}

			assert.True(t, resp.Match)
			assert.Equal(t, "bar", resp.SegmentKey)
			assert.Equal(t, matchesVariantKey, resp.Value)
		})
	}
}

func TestEvaluate_MatchAll_RolloutDistribution_MultiRule(t *testing.T) {
	s := &Server{
		logger: logger,
		FlagStore: &flagStoreMock{
			getFlagFn: func(ctx context.Context, r *flipt.GetFlagRequest) (*flipt.Flag, error) {
				return &flipt.Flag{
					Key:     "foo",
					Enabled: true,
				}, nil
			},
		},
		EvaluationStore: &evaluationStoreMock{
			getEvaluationRulesFn: func(context.Context, string) ([]*storage.EvaluationRule, error) {
				return []*storage.EvaluationRule{
					{
						ID:               "1",
						FlagKey:          "foo",
						SegmentKey:       "subscribers",
						SegmentMatchType: flipt.MatchType_ALL_MATCH_TYPE,
						Rank:             0,
						Constraints: []storage.EvaluationConstraint{
							// constraint: premium_user (bool) == true
							{
								ID:       "2",
								Type:     flipt.ComparisonType_BOOLEAN_COMPARISON_TYPE,
								Property: "premium_user",
								Operator: flipt.OpTrue,
							},
						},
					},
					{
						ID:               "2",
						FlagKey:          "foo",
						SegmentKey:       "all_users",
						SegmentMatchType: flipt.MatchType_ALL_MATCH_TYPE,
						Rank:             1,
					},
				}, nil
			},
			getEvaluationDistributionsFn: func(context.Context, string) ([]*storage.EvaluationDistribution, error) {
				return []*storage.EvaluationDistribution{
					{
						ID:         "4",
						RuleID:     "1",
						VariantID:  "5",
						Rollout:    50,
						VariantKey: "released",
					},
					{
						ID:         "6",
						RuleID:     "1",
						VariantID:  "7",
						Rollout:    50,
						VariantKey: "unreleased",
					},
					{
						ID:         "9",
						RuleID:     "2",
						VariantID:  "10",
						Rollout:    100,
						VariantKey: "unreleased",
					},
				}, nil
			},
		},
	}

	resp, err := s.Evaluate(context.TODO(), &flipt.EvaluationRequest{
		FlagKey:  "foo",
		EntityId: uuid.Must(uuid.NewV4()).String(),
		Context: map[string]string{
			"premium_user": "true",
		},
	})

	require.NoError(t, err)

	assert.NotNil(t, resp)
	assert.True(t, resp.Match)
	assert.Equal(t, "subscribers", resp.SegmentKey)
	assert.Equal(t, "foo", resp.FlagKey)
	assert.NotEmpty(t, resp.Value)
}

func TestEvaluate_MatchAll_NoConstraints(t *testing.T) {
	s := &Server{
		logger: logger,
		FlagStore: &flagStoreMock{
			getFlagFn: func(ctx context.Context, r *flipt.GetFlagRequest) (*flipt.Flag, error) {
				return &flipt.Flag{
					Key:     "foo",
					Enabled: true,
				}, nil
			},
		},
		EvaluationStore: &evaluationStoreMock{
			getEvaluationRulesFn: func(context.Context, string) ([]*storage.EvaluationRule, error) {
				return []*storage.EvaluationRule{
					{
						ID:               "1",
						FlagKey:          "foo",
						SegmentKey:       "bar",
						SegmentMatchType: flipt.MatchType_ALL_MATCH_TYPE,
						Rank:             0,
					},
				}, nil
			},
			getEvaluationDistributionsFn: func(context.Context, string) ([]*storage.EvaluationDistribution, error) {
				return []*storage.EvaluationDistribution{
					{
						ID:         "4",
						RuleID:     "1",
						VariantID:  "5",
						Rollout:    50,
						VariantKey: "boz",
					},
					{
						ID:         "6",
						RuleID:     "1",
						VariantID:  "7",
						Rollout:    50,
						VariantKey: "moz",
					},
				}, nil
			},
		},
	}

	tests := []struct {
		name              string
		req               *flipt.EvaluationRequest
		matchesVariantKey string
		wantMatch         bool
	}{
		{
			name: "match no value - variant 1",
			req: &flipt.EvaluationRequest{
				FlagKey:  "foo",
				EntityId: "10",
				Context:  map[string]string{},
			},
			matchesVariantKey: "boz",
			wantMatch:         true,
		},
		{
			name: "match no value - variant 2",
			req: &flipt.EvaluationRequest{
				FlagKey:  "foo",
				EntityId: "01",
				Context:  map[string]string{},
			},
			matchesVariantKey: "moz",
			wantMatch:         true,
		},
		{
			name: "match string value - variant 2",
			req: &flipt.EvaluationRequest{
				FlagKey:  "foo",
				EntityId: "01",
				Context: map[string]string{
					"bar": "boz",
				},
			},
			matchesVariantKey: "moz",
			wantMatch:         true,
		},
	}

	for _, tt := range tests {
		var (
			req               = tt.req
			matchesVariantKey = tt.matchesVariantKey
			wantMatch         = tt.wantMatch
		)

		t.Run(tt.name, func(t *testing.T) {
			resp, err := s.Evaluate(context.TODO(), req)
			require.NoError(t, err)
			assert.NotNil(t, resp)
			assert.Equal(t, "foo", resp.FlagKey)
			assert.Equal(t, req.Context, resp.RequestContext)

			if !wantMatch {
				assert.False(t, resp.Match)
				assert.Empty(t, resp.SegmentKey)
				return
			}

			assert.True(t, resp.Match)
			assert.Equal(t, "bar", resp.SegmentKey)
			assert.Equal(t, matchesVariantKey, resp.Value)
		})
	}
}

// Match ANY constraints

func TestEvaluate_MatchAny_NoVariants_NoDistributions(t *testing.T) {
	s := &Server{
		logger: logger,
		FlagStore: &flagStoreMock{
			getFlagFn: func(ctx context.Context, r *flipt.GetFlagRequest) (*flipt.Flag, error) {
				return &flipt.Flag{
					Key:     "foo",
					Enabled: true,
				}, nil
			},
		},
		EvaluationStore: &evaluationStoreMock{
			getEvaluationRulesFn: func(context.Context, string) ([]*storage.EvaluationRule, error) {
				return []*storage.EvaluationRule{
					{
						ID:               "1",
						FlagKey:          "foo",
						SegmentKey:       "bar",
						SegmentMatchType: flipt.MatchType_ANY_MATCH_TYPE,
						Rank:             0,
						Constraints: []storage.EvaluationConstraint{
							{
								ID:       "2",
								Type:     flipt.ComparisonType_STRING_COMPARISON_TYPE,
								Property: "bar",
								Operator: flipt.OpEQ,
								Value:    "baz",
							},
						},
					},
				}, nil
			},
		},
	}

	tests := []struct {
		name      string
		req       *flipt.EvaluationRequest
		wantMatch bool
	}{
		{
			name: "match string value",
			req: &flipt.EvaluationRequest{
				FlagKey:  "foo",
				EntityId: "1",
				Context: map[string]string{
					"bar": "baz",
				},
			},
			wantMatch: true,
		},
		{
			name: "no match string value",
			req: &flipt.EvaluationRequest{
				FlagKey:  "foo",
				EntityId: "1",
				Context: map[string]string{
					"bar": "boz",
				},
			},
		},
	}

	for _, tt := range tests {
		var (
			req       = tt.req
			wantMatch = tt.wantMatch
		)

		t.Run(tt.name, func(t *testing.T) {
			resp, err := s.Evaluate(context.TODO(), req)
			require.NoError(t, err)
			assert.NotNil(t, resp)
			assert.Equal(t, "foo", resp.FlagKey)
			assert.Equal(t, req.Context, resp.RequestContext)

			if !wantMatch {
				assert.False(t, resp.Match)
				assert.Empty(t, resp.SegmentKey)
				return
			}

			assert.True(t, resp.Match)
			assert.Equal(t, "bar", resp.SegmentKey)
			assert.Empty(t, resp.Value)
		})
	}
}

func TestEvaluate_MatchAny_SingleVariantDistribution(t *testing.T) {
	s := &Server{
		logger: logger,
		FlagStore: &flagStoreMock{
			getFlagFn: func(ctx context.Context, r *flipt.GetFlagRequest) (*flipt.Flag, error) {
				return &flipt.Flag{
					Key:     "foo",
					Enabled: true,
				}, nil
			},
		},
		EvaluationStore: &evaluationStoreMock{
			getEvaluationRulesFn: func(context.Context, string) ([]*storage.EvaluationRule, error) {
				return []*storage.EvaluationRule{
					{
						ID:               "1",
						FlagKey:          "foo",
						SegmentKey:       "bar",
						SegmentMatchType: flipt.MatchType_ANY_MATCH_TYPE,
						Rank:             0,
						Constraints: []storage.EvaluationConstraint{
							// constraint: bar (string) == baz
							{
								ID:       "2",
								Type:     flipt.ComparisonType_STRING_COMPARISON_TYPE,
								Property: "bar",
								Operator: flipt.OpEQ,
								Value:    "baz",
							},
							// constraint: admin (bool) == true
							{
								ID:       "3",
								Type:     flipt.ComparisonType_BOOLEAN_COMPARISON_TYPE,
								Property: "admin",
								Operator: flipt.OpTrue,
							},
						},
					},
				}, nil
			},
			getEvaluationDistributionsFn: func(context.Context, string) ([]*storage.EvaluationDistribution, error) {
				return []*storage.EvaluationDistribution{
					{
						ID:         "4",
						RuleID:     "1",
						VariantID:  "5",
						Rollout:    100,
						VariantKey: "boz",
					},
				}, nil
			},
		},
	}

	tests := []struct {
		name      string
		req       *flipt.EvaluationRequest
		wantMatch bool
	}{
		{
			name: "matches all",
			req: &flipt.EvaluationRequest{
				FlagKey:  "foo",
				EntityId: "1",
				Context: map[string]string{
					"bar":   "baz",
					"admin": "true",
				},
			},
			wantMatch: true,
		},
		{
			name: "matches one",
			req: &flipt.EvaluationRequest{
				FlagKey:  "foo",
				EntityId: "1",
				Context: map[string]string{
					"bar":   "boz",
					"admin": "true",
				},
			},
			wantMatch: true,
		},
		{
			name: "matches just bool value",
			req: &flipt.EvaluationRequest{
				FlagKey:  "foo",
				EntityId: "1",
				Context: map[string]string{
					"admin": "true",
				},
			},
			wantMatch: true,
		},
		{
			name: "matches just string value",
			req: &flipt.EvaluationRequest{
				FlagKey:  "foo",
				EntityId: "1",
				Context: map[string]string{
					"bar": "baz",
				},
			},
			wantMatch: true,
		},
		{
			name: "no matches wrong bool value",
			req: &flipt.EvaluationRequest{
				FlagKey:  "foo",
				EntityId: "1",
				Context: map[string]string{
					"admin": "false",
				},
			},
		},
		{
			name: "no matches wrong string value",
			req: &flipt.EvaluationRequest{
				FlagKey:  "foo",
				EntityId: "1",
				Context: map[string]string{
					"bar": "boz",
				},
			},
		},
		{
			name: "no match none",
			req: &flipt.EvaluationRequest{
				FlagKey:  "foo",
				EntityId: "1",
				Context: map[string]string{
					"boss": "boz",
					"user": "true",
				},
			},
		},
	}

	for _, tt := range tests {
		var (
			req       = tt.req
			wantMatch = tt.wantMatch
		)

		t.Run(tt.name, func(t *testing.T) {
			resp, err := s.Evaluate(context.TODO(), req)
			require.NoError(t, err)
			assert.NotNil(t, resp)
			assert.Equal(t, "foo", resp.FlagKey)
			assert.Equal(t, req.Context, resp.RequestContext)

			if !wantMatch {
				assert.False(t, resp.Match)
				assert.Empty(t, resp.SegmentKey)
				return
			}

			assert.True(t, resp.Match)
			assert.Equal(t, "bar", resp.SegmentKey)
			assert.Equal(t, "boz", resp.Value)
		})
	}
}

func TestEvaluate_MatchAny_RolloutDistribution(t *testing.T) {
	s := &Server{
		logger: logger,
		FlagStore: &flagStoreMock{
			getFlagFn: func(ctx context.Context, r *flipt.GetFlagRequest) (*flipt.Flag, error) {
				return &flipt.Flag{
					Key:     "foo",
					Enabled: true,
				}, nil
			},
		},
		EvaluationStore: &evaluationStoreMock{
			getEvaluationRulesFn: func(context.Context, string) ([]*storage.EvaluationRule, error) {
				return []*storage.EvaluationRule{
					{
						ID:               "1",
						FlagKey:          "foo",
						SegmentKey:       "bar",
						SegmentMatchType: flipt.MatchType_ANY_MATCH_TYPE,
						Rank:             0,
						Constraints: []storage.EvaluationConstraint{
							// constraint: bar (string) == baz
							{
								ID:       "2",
								Type:     flipt.ComparisonType_STRING_COMPARISON_TYPE,
								Property: "bar",
								Operator: flipt.OpEQ,
								Value:    "baz",
							},
						},
					},
				}, nil
			},
			getEvaluationDistributionsFn: func(context.Context, string) ([]*storage.EvaluationDistribution, error) {
				return []*storage.EvaluationDistribution{
					{
						ID:         "4",
						RuleID:     "1",
						VariantID:  "5",
						Rollout:    50,
						VariantKey: "boz",
					},
					{
						ID:         "6",
						RuleID:     "1",
						VariantID:  "7",
						Rollout:    50,
						VariantKey: "booz",
					},
				}, nil
			},
		},
	}

	tests := []struct {
		name              string
		req               *flipt.EvaluationRequest
		matchesVariantKey string
		wantMatch         bool
	}{
		{
			name: "match string value - variant 1",
			req: &flipt.EvaluationRequest{
				FlagKey:  "foo",
				EntityId: "1",
				Context: map[string]string{
					"bar": "baz",
				},
			},
			matchesVariantKey: "boz",
			wantMatch:         true,
		},
		{
			name: "match string value - variant 2",
			req: &flipt.EvaluationRequest{
				FlagKey:  "foo",
				EntityId: "2",
				Context: map[string]string{
					"bar": "baz",
				},
			},
			matchesVariantKey: "booz",
			wantMatch:         true,
		},
		{
			name: "no match string value",
			req: &flipt.EvaluationRequest{
				FlagKey:  "foo",
				EntityId: "1",
				Context: map[string]string{
					"bar": "boz",
				},
			},
		},
	}

	for _, tt := range tests {
		var (
			req               = tt.req
			matchesVariantKey = tt.matchesVariantKey
			wantMatch         = tt.wantMatch
		)

		t.Run(tt.name, func(t *testing.T) {
			resp, err := s.Evaluate(context.TODO(), req)
			require.NoError(t, err)
			assert.NotNil(t, resp)
			assert.Equal(t, "foo", resp.FlagKey)
			assert.Equal(t, req.Context, resp.RequestContext)

			if !wantMatch {
				assert.False(t, resp.Match)
				assert.Empty(t, resp.SegmentKey)
				return
			}

			assert.True(t, resp.Match)
			assert.Equal(t, "bar", resp.SegmentKey)
			assert.Equal(t, matchesVariantKey, resp.Value)
		})
	}
}

func TestEvaluate_MatchAny_RolloutDistribution_MultiRule(t *testing.T) {
	s := &Server{
		logger: logger,
		FlagStore: &flagStoreMock{
			getFlagFn: func(ctx context.Context, r *flipt.GetFlagRequest) (*flipt.Flag, error) {
				return &flipt.Flag{
					Key:     "foo",
					Enabled: true,
				}, nil
			},
		},
		EvaluationStore: &evaluationStoreMock{
			getEvaluationRulesFn: func(context.Context, string) ([]*storage.EvaluationRule, error) {
				return []*storage.EvaluationRule{
					{
						ID:               "1",
						FlagKey:          "foo",
						SegmentKey:       "subscribers",
						SegmentMatchType: flipt.MatchType_ANY_MATCH_TYPE,
						Rank:             0,
						Constraints: []storage.EvaluationConstraint{
							// constraint: premium_user (bool) == true
							{
								ID:       "2",
								Type:     flipt.ComparisonType_BOOLEAN_COMPARISON_TYPE,
								Property: "premium_user",
								Operator: flipt.OpTrue,
							},
						},
					},
					{
						ID:               "2",
						FlagKey:          "foo",
						SegmentKey:       "all_users",
						SegmentMatchType: flipt.MatchType_ALL_MATCH_TYPE,
						Rank:             1,
					},
				}, nil
			},
			getEvaluationDistributionsFn: func(context.Context, string) ([]*storage.EvaluationDistribution, error) {
				return []*storage.EvaluationDistribution{
					{
						ID:         "4",
						RuleID:     "1",
						VariantID:  "5",
						Rollout:    50,
						VariantKey: "released",
					},
					{
						ID:         "6",
						RuleID:     "1",
						VariantID:  "7",
						Rollout:    50,
						VariantKey: "unreleased",
					},
					{
						ID:         "9",
						RuleID:     "2",
						VariantID:  "10",
						Rollout:    100,
						VariantKey: "unreleased",
					},
				}, nil
			},
		},
	}

	resp, err := s.Evaluate(context.TODO(), &flipt.EvaluationRequest{
		FlagKey:  "foo",
		EntityId: uuid.Must(uuid.NewV4()).String(),
		Context: map[string]string{
			"premium_user": "true",
		},
	})

	require.NoError(t, err)

	assert.NotNil(t, resp)
	assert.True(t, resp.Match)
	assert.Equal(t, "subscribers", resp.SegmentKey)
	assert.Equal(t, "foo", resp.FlagKey)
	assert.NotEmpty(t, resp.Value)
}

func TestEvaluate_MatchAny_NoConstraints(t *testing.T) {
	s := &Server{
		logger: logger,
		FlagStore: &flagStoreMock{
			getFlagFn: func(ctx context.Context, r *flipt.GetFlagRequest) (*flipt.Flag, error) {
				return &flipt.Flag{
					Key:     "foo",
					Enabled: true,
				}, nil
			},
		},
		EvaluationStore: &evaluationStoreMock{
			getEvaluationRulesFn: func(context.Context, string) ([]*storage.EvaluationRule, error) {
				return []*storage.EvaluationRule{
					{
						ID:               "1",
						FlagKey:          "foo",
						SegmentKey:       "bar",
						SegmentMatchType: flipt.MatchType_ANY_MATCH_TYPE,
						Rank:             0,
					},
				}, nil
			},
			getEvaluationDistributionsFn: func(context.Context, string) ([]*storage.EvaluationDistribution, error) {
				return []*storage.EvaluationDistribution{
					{
						ID:         "4",
						RuleID:     "1",
						VariantID:  "5",
						Rollout:    50,
						VariantKey: "boz",
					},
					{
						ID:         "6",
						RuleID:     "1",
						VariantID:  "7",
						Rollout:    50,
						VariantKey: "moz",
					},
				}, nil
			},
		},
	}

	tests := []struct {
		name              string
		req               *flipt.EvaluationRequest
		matchesVariantKey string
		wantMatch         bool
	}{
		{
			name: "match no value - variant 1",
			req: &flipt.EvaluationRequest{
				FlagKey:  "foo",
				EntityId: "10",
				Context:  map[string]string{},
			},
			matchesVariantKey: "boz",
			wantMatch:         true,
		},
		{
			name: "match no value - variant 2",
			req: &flipt.EvaluationRequest{
				FlagKey:  "foo",
				EntityId: "01",
				Context:  map[string]string{},
			},
			matchesVariantKey: "moz",
			wantMatch:         true,
		},
		{
			name: "match string value - variant 2",
			req: &flipt.EvaluationRequest{
				FlagKey:  "foo",
				EntityId: "01",
				Context: map[string]string{
					"bar": "boz",
				},
			},
			matchesVariantKey: "moz",
			wantMatch:         true,
		},
	}

	for _, tt := range tests {
		var (
			req               = tt.req
			matchesVariantKey = tt.matchesVariantKey
			wantMatch         = tt.wantMatch
		)

		t.Run(tt.name, func(t *testing.T) {
			resp, err := s.Evaluate(context.TODO(), req)
			require.NoError(t, err)
			assert.NotNil(t, resp)
			assert.Equal(t, "foo", resp.FlagKey)
			assert.Equal(t, req.Context, resp.RequestContext)

			if !wantMatch {
				assert.False(t, resp.Match)
				assert.Empty(t, resp.SegmentKey)
				return
			}

			assert.True(t, resp.Match)
			assert.Equal(t, "bar", resp.SegmentKey)
			assert.Equal(t, matchesVariantKey, resp.Value)
		})
	}
}

func Test_matchesString(t *testing.T) {
	tests := []struct {
		name       string
		constraint storage.EvaluationConstraint
		value      string
		wantMatch  bool
	}{
		{
			name: "eq",
			constraint: storage.EvaluationConstraint{
				Property: "foo",
				Operator: "eq",
				Value:    "bar",
			},
			value:     "bar",
			wantMatch: true,
		},
		{
			name: "negative eq",
			constraint: storage.EvaluationConstraint{
				Property: "foo",
				Operator: "eq",
				Value:    "bar",
			},
			value: "baz",
		},
		{
			name: "neq",
			constraint: storage.EvaluationConstraint{
				Property: "foo",
				Operator: "neq",
				Value:    "bar",
			},
			value:     "baz",
			wantMatch: true,
		},
		{
			name: "negative neq",
			constraint: storage.EvaluationConstraint{
				Property: "foo",
				Operator: "neq",
				Value:    "bar",
			},
			value: "bar",
		},
		{
			name: "empty",
			constraint: storage.EvaluationConstraint{
				Property: "foo",
				Operator: "empty",
			},
			value:     " ",
			wantMatch: true,
		},
		{
			name: "negative empty",
			constraint: storage.EvaluationConstraint{
				Property: "foo",
				Operator: "empty",
			},
			value: "bar",
		},
		{
			name: "not empty",
			constraint: storage.EvaluationConstraint{
				Property: "foo",
				Operator: "notempty",
			},
			value:     "bar",
			wantMatch: true,
		},
		{
			name: "negative not empty",
			constraint: storage.EvaluationConstraint{
				Property: "foo",
				Operator: "notempty",
			},
			value: "",
		},
		{
			name: "unknown operator",
			constraint: storage.EvaluationConstraint{
				Property: "foo",
				Operator: "foo",
				Value:    "bar",
			},
			value: "bar",
		},
		{
			name: "prefix",
			constraint: storage.EvaluationConstraint{
				Property: "foo",
				Operator: "prefix",
				Value:    "ba",
			},
			value:     "bar",
			wantMatch: true,
		},
		{
			name: "negative prefix",
			constraint: storage.EvaluationConstraint{
				Property: "foo",
				Operator: "prefix",
				Value:    "bar",
			},
			value: "nope",
		},
		{
			name: "suffix",
			constraint: storage.EvaluationConstraint{
				Property: "foo",
				Operator: "suffix",
				Value:    "ar",
			},
			value:     "bar",
			wantMatch: true,
		},
		{
			name: "negative suffix",
			constraint: storage.EvaluationConstraint{
				Property: "foo",
				Operator: "suffix",
				Value:    "bar",
			},
			value: "nope",
		},
	}
	for _, tt := range tests {
		var (
			constraint = tt.constraint
			value      = tt.value
			wantMatch  = tt.wantMatch
		)

		t.Run(tt.name, func(t *testing.T) {
			match := matchesString(constraint, value)
			assert.Equal(t, wantMatch, match)
		})
	}
}

func Test_matchesNumber(t *testing.T) {
	tests := []struct {
		name       string
		constraint storage.EvaluationConstraint
		value      string
		wantMatch  bool
		wantErr    bool
	}{
		{
			name: "present",
			constraint: storage.EvaluationConstraint{
				Property: "foo",
				Operator: "present",
			},
			value:     "1",
			wantMatch: true,
		},
		{
			name: "negative present",
			constraint: storage.EvaluationConstraint{
				Property: "foo",
				Operator: "present",
			},
		},
		{
			name: "not present",
			constraint: storage.EvaluationConstraint{
				Property: "foo",
				Operator: "notpresent",
			},
			wantMatch: true,
		},
		{
			name: "negative notpresent",
			constraint: storage.EvaluationConstraint{
				Property: "foo",
				Operator: "notpresent",
			},
			value: "1",
		},
		{
			name: "NAN constraint value",
			constraint: storage.EvaluationConstraint{
				Property: "foo",
				Operator: "eq",
				Value:    "bar",
			},
			value:   "5",
			wantErr: true,
		},
		{
			name: "NAN context value",
			constraint: storage.EvaluationConstraint{
				Property: "foo",
				Operator: "eq",
				Value:    "5",
			},
			value:   "foo",
			wantErr: true,
		},
		{
			name: "eq",
			constraint: storage.EvaluationConstraint{
				Property: "foo",
				Operator: "eq",
				Value:    "42.0",
			},
			value:     "42.0",
			wantMatch: true,
		},
		{
			name: "negative eq",
			constraint: storage.EvaluationConstraint{
				Property: "foo",
				Operator: "eq",
				Value:    "42.0",
			},
			value: "50",
		},
		{
			name: "neq",
			constraint: storage.EvaluationConstraint{
				Property: "foo",
				Operator: "neq",
				Value:    "42.0",
			},
			value:     "50",
			wantMatch: true,
		},
		{
			name: "negative neq",
			constraint: storage.EvaluationConstraint{
				Property: "foo",
				Operator: "neq",
				Value:    "42.0",
			},
			value: "42.0",
		},
		{
			name: "lt",
			constraint: storage.EvaluationConstraint{
				Property: "foo",
				Operator: "lt",
				Value:    "42.0",
			},
			value:     "8",
			wantMatch: true,
		},
		{
			name: "negative lt",
			constraint: storage.EvaluationConstraint{
				Property: "foo",
				Operator: "lt",
				Value:    "42.0",
			},
			value: "50",
		},
		{
			name: "lte",
			constraint: storage.EvaluationConstraint{
				Property: "foo",
				Operator: "lte",
				Value:    "42.0",
			},
			value:     "42.0",
			wantMatch: true,
		},
		{
			name: "negative lte",
			constraint: storage.EvaluationConstraint{
				Property: "foo",
				Operator: "lte",
				Value:    "42.0",
			},
			value: "102.0",
		},
		{
			name: "gt",
			constraint: storage.EvaluationConstraint{
				Property: "foo",
				Operator: "gt",
				Value:    "10.11",
			},
			value:     "10.12",
			wantMatch: true,
		},
		{
			name: "negative gt",
			constraint: storage.EvaluationConstraint{
				Property: "foo",
				Operator: "gt",
				Value:    "10.11",
			},
			value: "1",
		},
		{
			name: "gte",
			constraint: storage.EvaluationConstraint{
				Property: "foo",
				Operator: "gte",
				Value:    "10.11",
			},
			value:     "10.11",
			wantMatch: true,
		},
		{
			name: "negative gte",
			constraint: storage.EvaluationConstraint{
				Property: "foo",
				Operator: "gte",
				Value:    "10.11",
			},
			value: "0.11",
		},
		{
			name: "empty value",
			constraint: storage.EvaluationConstraint{
				Property: "foo",
				Operator: "eq",
				Value:    "0.11",
			},
		},
		{
			name: "unknown operator",
			constraint: storage.EvaluationConstraint{
				Property: "foo",
				Operator: "foo",
				Value:    "0.11",
			},
			value: "0.11",
		},
		{
			name: "negative suffix empty value",
			constraint: storage.EvaluationConstraint{
				Property: "foo",
				Operator: "suffix",
				Value:    "bar",
			},
		},
	}

	for _, tt := range tests {
		var (
			constraint = tt.constraint
			value      = tt.value
			wantMatch  = tt.wantMatch
			wantErr    = tt.wantErr
		)

		t.Run(tt.name, func(t *testing.T) {
			match, err := matchesNumber(constraint, value)

			if wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, wantMatch, match)
		})
	}
}

func Test_matchesBool(t *testing.T) {
	tests := []struct {
		name       string
		constraint storage.EvaluationConstraint
		value      string
		wantMatch  bool
		wantErr    bool
	}{
		{
			name: "present",
			constraint: storage.EvaluationConstraint{
				Property: "foo",
				Operator: "present",
			},
			value:     "true",
			wantMatch: true,
		},
		{
			name: "negative present",
			constraint: storage.EvaluationConstraint{
				Property: "foo",
				Operator: "present",
			},
		},
		{
			name: "not present",
			constraint: storage.EvaluationConstraint{
				Property: "foo",
				Operator: "notpresent",
			},
			wantMatch: true,
		},
		{
			name: "negative notpresent",
			constraint: storage.EvaluationConstraint{
				Property: "foo",
				Operator: "notpresent",
			},
			value: "true",
		},
		{
			name: "not a bool",
			constraint: storage.EvaluationConstraint{
				Property: "foo",
				Operator: "true",
			},
			value:   "foo",
			wantErr: true,
		},
		{
			name: "is true",
			constraint: storage.EvaluationConstraint{
				Property: "foo",
				Operator: "true",
			},
			value:     "true",
			wantMatch: true,
		},
		{
			name: "negative is true",
			constraint: storage.EvaluationConstraint{
				Property: "foo",
				Operator: "true",
			},
			value: "false",
		},
		{
			name: "is false",
			constraint: storage.EvaluationConstraint{
				Property: "foo",
				Operator: "false",
			},
			value:     "false",
			wantMatch: true,
		},
		{
			name: "negative is false",
			constraint: storage.EvaluationConstraint{
				Property: "foo",
				Operator: "false",
			},
			value: "true",
		},
		{
			name: "empty value",
			constraint: storage.EvaluationConstraint{
				Property: "foo",
				Operator: "false",
			},
		},
		{
			name: "unknown operator",
			constraint: storage.EvaluationConstraint{
				Property: "foo",
				Operator: "foo",
			},
			value: "true",
		},
	}

	for _, tt := range tests {
		var (
			constraint = tt.constraint
			value      = tt.value
			wantMatch  = tt.wantMatch
			wantErr    = tt.wantErr
		)

		t.Run(tt.name, func(t *testing.T) {
			match, err := matchesBool(constraint, value)

			if wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, wantMatch, match)
		})
	}
}

// Since we skip rollout buckets that have 0% distribution, ensure that things still work
// when a 0% distribution is the first available one.
// Fixes a previously existing bug in that specific situation
func TestEvaluate_FirstRolloutRuleIsZero(t *testing.T) {
	s := &Server{
		logger: logger,
		FlagStore: &flagStoreMock{
			getFlagFn: func(ctx context.Context, r *flipt.GetFlagRequest) (*flipt.Flag, error) {
				return &flipt.Flag{
					Key:     "foo",
					Enabled: true,
				}, nil
			},
		},
		EvaluationStore: &evaluationStoreMock{
			getEvaluationRulesFn: func(context.Context, string) ([]*storage.EvaluationRule, error) {
				return []*storage.EvaluationRule{
					{
						ID:               "1",
						FlagKey:          "foo",
						SegmentKey:       "bar",
						SegmentMatchType: flipt.MatchType_ALL_MATCH_TYPE,
						Rank:             0,
						Constraints: []storage.EvaluationConstraint{
							// constraint: bar (string) == baz
							{
								ID:       "2",
								Type:     flipt.ComparisonType_STRING_COMPARISON_TYPE,
								Property: "bar",
								Operator: flipt.OpEQ,
								Value:    "baz",
							},
						},
					},
				}, nil
			},
			getEvaluationDistributionsFn: func(context.Context, string) ([]*storage.EvaluationDistribution, error) {
				return []*storage.EvaluationDistribution{
					{
						ID:         "4",
						RuleID:     "1",
						VariantID:  "5",
						Rollout:    0,
						VariantKey: "boz",
					},
					{
						ID:         "6",
						RuleID:     "1",
						VariantID:  "7",
						Rollout:    100,
						VariantKey: "booz",
					},
				}, nil
			},
		},
	}

	tests := []struct {
		name              string
		req               *flipt.EvaluationRequest
		matchesVariantKey string
		wantMatch         bool
	}{
		{
			name: "match string value - variant 1",
			req: &flipt.EvaluationRequest{
				FlagKey:  "foo",
				EntityId: "1",
				Context: map[string]string{
					"bar": "baz",
				},
			},
			matchesVariantKey: "booz",
			wantMatch:         true,
		},
	}

	for _, tt := range tests {
		var (
			req               = tt.req
			matchesVariantKey = tt.matchesVariantKey
			wantMatch         = tt.wantMatch
		)

		t.Run(tt.name, func(t *testing.T) {
			resp, err := s.Evaluate(context.TODO(), req)
			require.NoError(t, err)
			assert.NotNil(t, resp)
			assert.Equal(t, "foo", resp.FlagKey)
			assert.Equal(t, req.Context, resp.RequestContext)

			if !wantMatch {
				assert.False(t, resp.Match)
				assert.Empty(t, resp.SegmentKey)
				return
			}

			assert.True(t, resp.Match)
			assert.Equal(t, "bar", resp.SegmentKey)
			assert.Equal(t, matchesVariantKey, resp.Value)
		})
	}
}

// Ensure things work properly when many rollout distributions have a 0% value.
// Fixes a previously existing bug in that specific situation
func TestEvaluate_MultipleZeroRolloutDistributions(t *testing.T) {
	s := &Server{
		logger: logger,
		FlagStore: &flagStoreMock{
			getFlagFn: func(ctx context.Context, r *flipt.GetFlagRequest) (*flipt.Flag, error) {
				return &flipt.Flag{
					Key:     "foo",
					Enabled: true,
				}, nil
			},
		},
		EvaluationStore: &evaluationStoreMock{
			getEvaluationRulesFn: func(context.Context, string) ([]*storage.EvaluationRule, error) {
				return []*storage.EvaluationRule{
					{
						ID:               "1",
						FlagKey:          "foo",
						SegmentKey:       "bar",
						SegmentMatchType: flipt.MatchType_ALL_MATCH_TYPE,
						Rank:             0,
						Constraints: []storage.EvaluationConstraint{
							// constraint: bar (string) == baz
							{
								ID:       "2",
								Type:     flipt.ComparisonType_STRING_COMPARISON_TYPE,
								Property: "bar",
								Operator: flipt.OpEQ,
								Value:    "baz",
							},
						},
					},
				}, nil
			},
			getEvaluationDistributionsFn: func(context.Context, string) ([]*storage.EvaluationDistribution, error) {
				return []*storage.EvaluationDistribution{
					{
						ID:         "1",
						RuleID:     "1",
						VariantID:  "1",
						VariantKey: "1",
						Rollout:    0,
					},
					{
						ID:         "2",
						RuleID:     "2",
						VariantID:  "2",
						VariantKey: "2",
						Rollout:    0,
					},
					{
						ID:         "3",
						RuleID:     "3",
						VariantID:  "3",
						VariantKey: "3",
						Rollout:    50,
					},
					{
						ID:         "4",
						RuleID:     "4",
						VariantID:  "4",
						VariantKey: "4",
						Rollout:    0,
					},
					{
						ID:         "5",
						RuleID:     "5",
						VariantID:  "5",
						VariantKey: "5",
						Rollout:    0,
					},
					{
						ID:         "6",
						RuleID:     "6",
						VariantID:  "6",
						VariantKey: "6",
						Rollout:    50,
					},
				}, nil
			},
		},
	}

	tests := []struct {
		name              string
		req               *flipt.EvaluationRequest
		matchesVariantKey string
		wantMatch         bool
	}{
		{
			name: "match string value - variant 1",
			req: &flipt.EvaluationRequest{
				FlagKey:  "foo",
				EntityId: "1",
				Context: map[string]string{
					"bar": "baz",
				},
			},
			matchesVariantKey: "3",
			wantMatch:         true,
		},
	}

	for _, tt := range tests {
		var (
			req               = tt.req
			matchesVariantKey = tt.matchesVariantKey
			wantMatch         = tt.wantMatch
		)

		t.Run(tt.name, func(t *testing.T) {
			resp, err := s.Evaluate(context.TODO(), req)
			require.NoError(t, err)
			assert.NotNil(t, resp)
			assert.Equal(t, "foo", resp.FlagKey)
			assert.Equal(t, req.Context, resp.RequestContext)

			if !wantMatch {
				assert.False(t, resp.Match)
				assert.Empty(t, resp.SegmentKey)
				return
			}

			assert.True(t, resp.Match)
			assert.Equal(t, "bar", resp.SegmentKey)
			assert.Equal(t, matchesVariantKey, resp.Value)
		})
	}
}
