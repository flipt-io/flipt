package evaluation

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	errs "go.flipt.io/flipt/errors"
	"go.flipt.io/flipt/internal/storage"
	"go.flipt.io/flipt/rpc/flipt"
	rpcEvaluation "go.flipt.io/flipt/rpc/flipt/evaluation"
	"go.uber.org/zap/zaptest"
)

func TestBoolean_FlagNotFoundError(t *testing.T) {
	var (
		flagKey      = "test-flag"
		namespaceKey = "test-namespace"
		store        = &evaluationStoreMock{}
		logger       = zaptest.NewLogger(t)
		s            = New(logger, store)
	)
	defer store.AssertNotCalled(t, "GetEvaluationRollouts", mock.Anything, namespaceKey, flagKey)

	store.On("GetFlag", mock.Anything, mock.Anything, mock.Anything).Return(&flipt.Flag{}, errs.ErrNotFound("test-flag"))

	res, err := s.Boolean(context.TODO(), &rpcEvaluation.EvaluationRequest{
		FlagKey:      flagKey,
		EntityId:     "test-entity",
		NamespaceKey: namespaceKey,
		Context: map[string]string{
			"hello": "world",
		},
	})

	assert.EqualError(t, err, "test-flag not found")
	assert.Equal(t, rpcEvaluation.EvaluationReason_FLAG_NOT_FOUND_EVALUATION_REASON, res.Reason)
}

func TestBoolean_NonBooleanFlagError(t *testing.T) {
	var (
		flagKey      = "test-flag"
		namespaceKey = "test-namespace"
		store        = &evaluationStoreMock{}
		logger       = zaptest.NewLogger(t)
		s            = New(logger, store)
	)
	defer store.AssertNotCalled(t, "GetEvaluationRollouts", mock.Anything, namespaceKey, flagKey)

	store.On("GetFlag", mock.Anything, mock.Anything, mock.Anything).Return(&flipt.Flag{
		NamespaceKey: "test-namespace",
		Key:          "test-flag",
		Enabled:      true,
		Type:         flipt.FlagType_VARIANT_FLAG_TYPE,
	}, nil)

	res, err := s.Boolean(context.TODO(), &rpcEvaluation.EvaluationRequest{
		FlagKey:      flagKey,
		EntityId:     "test-entity",
		NamespaceKey: namespaceKey,
		Context: map[string]string{
			"hello": "world",
		},
	})

	assert.EqualError(t, err, "flag type VARIANT_FLAG_TYPE invalid")
	assert.Equal(t, rpcEvaluation.EvaluationReason_ERROR_EVALUATION_REASON, res.Reason)
}

func TestBoolean_DefaultRule_NoRollouts(t *testing.T) {
	var (
		flagKey      = "test-flag"
		namespaceKey = "test-namespace"
		store        = &evaluationStoreMock{}
		logger       = zaptest.NewLogger(t)
		s            = New(logger, store)
	)

	store.On("GetFlag", mock.Anything, mock.Anything, mock.Anything).Return(&flipt.Flag{
		NamespaceKey: "test-namespace",
		Key:          "test-flag",
		Enabled:      true,
		Type:         flipt.FlagType_BOOLEAN_FLAG_TYPE,
	}, nil)

	store.On("GetEvaluationRollouts", mock.Anything, namespaceKey, flagKey).Return([]*storage.EvaluationRollout{}, nil)

	res, err := s.Boolean(context.TODO(), &rpcEvaluation.EvaluationRequest{
		FlagKey:      flagKey,
		EntityId:     "test-entity",
		NamespaceKey: namespaceKey,
		Context: map[string]string{
			"hello": "world",
		},
	})

	require.NoError(t, err)

	assert.Equal(t, true, res.Value)
	assert.Equal(t, rpcEvaluation.EvaluationReason_DEFAULT_EVALUATION_REASON, res.Reason)
}

func TestBoolean_DefaultRuleFallthrough_WithPercentageRollout(t *testing.T) {
	var (
		flagKey      = "test-flag"
		namespaceKey = "test-namespace"
		store        = &evaluationStoreMock{}
		logger       = zaptest.NewLogger(t)
		s            = New(logger, store)
	)

	store.On("GetFlag", mock.Anything, mock.Anything, mock.Anything).Return(&flipt.Flag{
		NamespaceKey: "test-namespace",
		Key:          "test-flag",
		Enabled:      true,
		Type:         flipt.FlagType_BOOLEAN_FLAG_TYPE,
	}, nil)

	store.On("GetEvaluationRollouts", mock.Anything, namespaceKey, flagKey).Return([]*storage.EvaluationRollout{
		{
			NamespaceKey: namespaceKey,
			Rank:         1,
			RolloutType:  flipt.RolloutType_PERCENTAGE_ROLLOUT_TYPE,
			Percentage: &storage.RolloutPercentage{
				Percentage: 5,
				Value:      false,
			},
		},
	}, nil)

	res, err := s.Boolean(context.TODO(), &rpcEvaluation.EvaluationRequest{
		FlagKey:      flagKey,
		EntityId:     "test-entity",
		NamespaceKey: namespaceKey,
		Context: map[string]string{
			"hello": "world",
		},
	})

	require.NoError(t, err)

	assert.Equal(t, true, res.Value)
	assert.Equal(t, rpcEvaluation.EvaluationReason_DEFAULT_EVALUATION_REASON, res.Reason)
}

func TestBoolean_PercentageRuleMatch(t *testing.T) {
	var (
		flagKey      = "test-flag"
		namespaceKey = "test-namespace"
		store        = &evaluationStoreMock{}
		logger       = zaptest.NewLogger(t)
		s            = New(logger, store)
	)

	store.On("GetFlag", mock.Anything, namespaceKey, flagKey).Return(&flipt.Flag{
		NamespaceKey: "test-namespace",
		Key:          "test-flag",
		Enabled:      true,
		Type:         flipt.FlagType_BOOLEAN_FLAG_TYPE,
	}, nil)

	store.On("GetEvaluationRollouts", mock.Anything, namespaceKey, flagKey).Return([]*storage.EvaluationRollout{
		{
			NamespaceKey: namespaceKey,
			Rank:         1,
			RolloutType:  flipt.RolloutType_PERCENTAGE_ROLLOUT_TYPE,
			Percentage: &storage.RolloutPercentage{
				Percentage: 70,
				Value:      false,
			},
		},
	}, nil)

	res, err := s.Boolean(context.TODO(), &rpcEvaluation.EvaluationRequest{
		FlagKey:      flagKey,
		EntityId:     "test-entity",
		NamespaceKey: namespaceKey,
		Context: map[string]string{
			"hello": "world",
		},
	})

	require.NoError(t, err)

	assert.Equal(t, false, res.Value)
	assert.Equal(t, rpcEvaluation.EvaluationReason_MATCH_EVALUATION_REASON, res.Reason)
}

func TestBoolean_PercentageRuleFallthrough_SegmentMatch(t *testing.T) {
	var (
		flagKey      = "test-flag"
		namespaceKey = "test-namespace"
		store        = &evaluationStoreMock{}
		logger       = zaptest.NewLogger(t)
		s            = New(logger, store)
	)

	store.On("GetFlag", mock.Anything, namespaceKey, flagKey).Return(&flipt.Flag{
		NamespaceKey: "test-namespace",
		Key:          "test-flag",
		Enabled:      true,
		Type:         flipt.FlagType_BOOLEAN_FLAG_TYPE,
	}, nil)

	store.On("GetEvaluationRollouts", mock.Anything, namespaceKey, flagKey).Return([]*storage.EvaluationRollout{
		{
			NamespaceKey: namespaceKey,
			Rank:         1,
			RolloutType:  flipt.RolloutType_PERCENTAGE_ROLLOUT_TYPE,
			Percentage: &storage.RolloutPercentage{
				Percentage: 5,
				Value:      false,
			},
		},
		{
			NamespaceKey: namespaceKey,
			RolloutType:  flipt.RolloutType_SEGMENT_ROLLOUT_TYPE,
			Rank:         2,
			Segment: &storage.RolloutSegment{
				SegmentKey:       "test-segment",
				SegmentMatchType: flipt.MatchType_ANY_MATCH_TYPE,
				Value:            true,
				Constraints: []storage.EvaluationConstraint{
					{
						Type:     flipt.ComparisonType_STRING_COMPARISON_TYPE,
						Property: "hello",
						Operator: flipt.OpEQ,
						Value:    "world",
					},
				},
			},
		},
	}, nil)

	res, err := s.Boolean(context.TODO(), &rpcEvaluation.EvaluationRequest{
		FlagKey:      flagKey,
		EntityId:     "test-entity",
		NamespaceKey: namespaceKey,
		Context: map[string]string{
			"hello": "world",
		},
	})

	require.NoError(t, err)

	assert.Equal(t, true, res.Value)
	assert.Equal(t, rpcEvaluation.EvaluationReason_MATCH_EVALUATION_REASON, res.Reason)
}

func TestBoolean_SegmentMatch_MultipleConstraints(t *testing.T) {
	var (
		flagKey      = "test-flag"
		namespaceKey = "test-namespace"
		store        = &evaluationStoreMock{}
		logger       = zaptest.NewLogger(t)
		s            = New(logger, store)
	)

	store.On("GetFlag", mock.Anything, namespaceKey, flagKey).Return(
		&flipt.Flag{
			NamespaceKey: "test-namespace",
			Key:          "test-flag",
			Enabled:      true,
			Type:         flipt.FlagType_BOOLEAN_FLAG_TYPE,
		}, nil)

	store.On("GetEvaluationRollouts", mock.Anything, namespaceKey, flagKey).Return([]*storage.EvaluationRollout{
		{
			NamespaceKey: namespaceKey,
			RolloutType:  flipt.RolloutType_SEGMENT_ROLLOUT_TYPE,
			Rank:         1,
			Segment: &storage.RolloutSegment{
				SegmentKey:       "test-segment",
				SegmentMatchType: flipt.MatchType_ANY_MATCH_TYPE,
				Value:            true,
				Constraints: []storage.EvaluationConstraint{
					{
						Type:     flipt.ComparisonType_NUMBER_COMPARISON_TYPE,
						Property: "pitimes100",
						Operator: flipt.OpEQ,
						Value:    "314",
					},
					{
						Type:     flipt.ComparisonType_STRING_COMPARISON_TYPE,
						Property: "hello",
						Operator: flipt.OpEQ,
						Value:    "world",
					},
				},
			},
		},
	}, nil)

	res, err := s.Boolean(context.TODO(), &rpcEvaluation.EvaluationRequest{
		FlagKey:      flagKey,
		EntityId:     "test-entity",
		NamespaceKey: namespaceKey,
		Context: map[string]string{
			"hello": "world",
		},
	})

	require.NoError(t, err)

	assert.Equal(t, true, res.Value)
	assert.Equal(t, rpcEvaluation.EvaluationReason_MATCH_EVALUATION_REASON, res.Reason)
}

func TestBoolean_RulesOutOfOrder(t *testing.T) {
	var (
		flagKey      = "test-flag"
		namespaceKey = "test-namespace"
		store        = &evaluationStoreMock{}
		logger       = zaptest.NewLogger(t)
		s            = New(logger, store)
	)

	store.On("GetFlag", mock.Anything, namespaceKey, flagKey).Return(
		&flipt.Flag{
			NamespaceKey: "test-namespace",
			Key:          "test-flag",
			Enabled:      true,
			Type:         flipt.FlagType_BOOLEAN_FLAG_TYPE,
		}, nil)

	store.On("GetEvaluationRollouts", mock.Anything, namespaceKey, flagKey).Return([]*storage.EvaluationRollout{
		{
			NamespaceKey: namespaceKey,
			Rank:         1,
			RolloutType:  flipt.RolloutType_PERCENTAGE_ROLLOUT_TYPE,
			Percentage: &storage.RolloutPercentage{
				Percentage: 5,
				Value:      false,
			},
		},
		{
			NamespaceKey: namespaceKey,
			RolloutType:  flipt.RolloutType_SEGMENT_ROLLOUT_TYPE,
			Rank:         0,
			Segment: &storage.RolloutSegment{
				SegmentKey:       "test-segment",
				SegmentMatchType: flipt.MatchType_ANY_MATCH_TYPE,
				Value:            true,
				Constraints: []storage.EvaluationConstraint{
					{
						Type:     flipt.ComparisonType_STRING_COMPARISON_TYPE,
						Property: "hello",
						Operator: flipt.OpEQ,
						Value:    "world",
					},
				},
			},
		},
	}, nil)

	res, err := s.Boolean(context.TODO(), &rpcEvaluation.EvaluationRequest{
		FlagKey:      flagKey,
		EntityId:     "test-entity",
		NamespaceKey: namespaceKey,
		Context: map[string]string{
			"hello": "world",
		},
	})

	assert.Error(t, err)
	assert.EqualError(t, err, "rollout rank: 0 detected out of order")
	assert.Equal(t, rpcEvaluation.EvaluationReason_ERROR_EVALUATION_REASON, res.Reason)
}
