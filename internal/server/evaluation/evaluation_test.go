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
	rpcevaluation "go.flipt.io/flipt/rpc/flipt/evaluation"
	"go.uber.org/zap/zaptest"
)

func TestVariant_FlagNotFound(t *testing.T) {
	var (
		flagKey      = "test-flag"
		namespaceKey = "test-namespace"
		store        = &evaluationStoreMock{}
		logger       = zaptest.NewLogger(t)
		s            = New(logger, store)
	)

	store.On("GetFlag", mock.Anything, namespaceKey, flagKey).Return(&flipt.Flag{}, errs.ErrNotFound("test-flag"))

	v, err := s.Variant(context.TODO(), &rpcevaluation.EvaluationRequest{
		FlagKey:      flagKey,
		EntityId:     "test-entity",
		NamespaceKey: namespaceKey,
		Context: map[string]string{
			"hello": "world",
		},
	})

	require.Nil(t, v)

	assert.EqualError(t, err, "test-flag not found")
}

func TestVariant_NonVariantFlag(t *testing.T) {
	var (
		flagKey      = "test-flag"
		namespaceKey = "test-namespace"
		store        = &evaluationStoreMock{}
		logger       = zaptest.NewLogger(t)
		s            = New(logger, store)
	)

	store.On("GetFlag", mock.Anything, namespaceKey, flagKey).Return(&flipt.Flag{
		NamespaceKey: namespaceKey,
		Key:          flagKey,
		Enabled:      true,
		Type:         flipt.FlagType_BOOLEAN_FLAG_TYPE,
	}, nil)

	v, err := s.Variant(context.TODO(), &rpcevaluation.EvaluationRequest{
		FlagKey:      flagKey,
		EntityId:     "test-entity",
		NamespaceKey: namespaceKey,
		Context: map[string]string{
			"hello": "world",
		},
	})

	require.Nil(t, v)

	assert.EqualError(t, err, "flag type BOOLEAN_FLAG_TYPE invalid")
}

func TestVariant_EvaluateFailure_OnGetEvaluationRules(t *testing.T) {
	var (
		flagKey      = "test-flag"
		namespaceKey = "test-namespace"
		store        = &evaluationStoreMock{}
		logger       = zaptest.NewLogger(t)
		s            = New(logger, store)
		flag         = &flipt.Flag{
			NamespaceKey: namespaceKey,
			Key:          flagKey,
			Enabled:      true,
			Type:         flipt.FlagType_VARIANT_FLAG_TYPE,
		}
	)

	store.On("GetFlag", mock.Anything, namespaceKey, flagKey).Return(flag, nil)

	store.On("GetEvaluationRules", mock.Anything, namespaceKey, flagKey).Return([]*storage.EvaluationRule{}, errs.ErrInvalid("some invalid error"))

	v, err := s.Variant(context.TODO(), &rpcevaluation.EvaluationRequest{
		FlagKey:      flagKey,
		EntityId:     "test-entity",
		NamespaceKey: namespaceKey,
		Context: map[string]string{
			"hello": "world",
		},
	})

	require.Nil(t, v)

	assert.EqualError(t, err, "some invalid error")
}

func TestVariant_Success(t *testing.T) {
	var (
		flagKey      = "test-flag"
		namespaceKey = "test-namespace"
		store        = &evaluationStoreMock{}
		logger       = zaptest.NewLogger(t)
		s            = New(logger, store)
		flag         = &flipt.Flag{
			NamespaceKey: namespaceKey,
			Key:          flagKey,
			Enabled:      true,
			Type:         flipt.FlagType_VARIANT_FLAG_TYPE,
		}
	)

	store.On("GetFlag", mock.Anything, namespaceKey, flagKey).Return(flag, nil)

	store.On("GetEvaluationRules", mock.Anything, namespaceKey, flagKey).Return(
		[]*storage.EvaluationRule{
			{
				ID:               "1",
				FlagKey:          flagKey,
				SegmentKey:       "bar",
				SegmentMatchType: flipt.MatchType_ALL_MATCH_TYPE,
				Rank:             0,
				Constraints: []storage.EvaluationConstraint{
					{
						ID:       "2",
						Type:     flipt.ComparisonType_STRING_COMPARISON_TYPE,
						Property: "hello",
						Operator: flipt.OpEQ,
						Value:    "world",
					},
				},
			},
		}, nil)

	store.On("GetEvaluationDistributions", mock.Anything, "1").Return([]*storage.EvaluationDistribution{}, nil)

	v, err := s.Variant(context.TODO(), &rpcevaluation.EvaluationRequest{
		FlagKey:      flagKey,
		EntityId:     "test-entity",
		NamespaceKey: namespaceKey,
		Context: map[string]string{
			"hello": "world",
		},
	})

	require.NoError(t, err)

	assert.Equal(t, true, v.Match)
	assert.Equal(t, "bar", v.SegmentKey)
	assert.Equal(t, rpcevaluation.EvaluationReason_MATCH_EVALUATION_REASON, v.Reason)
}

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

	res, err := s.Boolean(context.TODO(), &rpcevaluation.EvaluationRequest{
		FlagKey:      flagKey,
		EntityId:     "test-entity",
		NamespaceKey: namespaceKey,
		Context: map[string]string{
			"hello": "world",
		},
	})

	require.Nil(t, res)

	assert.EqualError(t, err, "test-flag not found")
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

	res, err := s.Boolean(context.TODO(), &rpcevaluation.EvaluationRequest{
		FlagKey:      flagKey,
		EntityId:     "test-entity",
		NamespaceKey: namespaceKey,
		Context: map[string]string{
			"hello": "world",
		},
	})

	require.Nil(t, res)

	assert.EqualError(t, err, "flag type VARIANT_FLAG_TYPE invalid")
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

	res, err := s.Boolean(context.TODO(), &rpcevaluation.EvaluationRequest{
		FlagKey:      flagKey,
		EntityId:     "test-entity",
		NamespaceKey: namespaceKey,
		Context: map[string]string{
			"hello": "world",
		},
	})

	require.NoError(t, err)

	assert.Equal(t, true, res.Value)
	assert.Equal(t, rpcevaluation.EvaluationReason_DEFAULT_EVALUATION_REASON, res.Reason)
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
			RolloutType:  flipt.RolloutType_THRESHOLD_ROLLOUT_TYPE,
			Threshold: &storage.RolloutThreshold{
				Percentage: 5,
				Value:      false,
			},
		},
	}, nil)

	res, err := s.Boolean(context.TODO(), &rpcevaluation.EvaluationRequest{
		FlagKey:      flagKey,
		EntityId:     "test-entity",
		NamespaceKey: namespaceKey,
		Context: map[string]string{
			"hello": "world",
		},
	})

	require.NoError(t, err)

	assert.Equal(t, true, res.Value)
	assert.Equal(t, rpcevaluation.EvaluationReason_DEFAULT_EVALUATION_REASON, res.Reason)
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
			RolloutType:  flipt.RolloutType_THRESHOLD_ROLLOUT_TYPE,
			Threshold: &storage.RolloutThreshold{
				Percentage: 70,
				Value:      false,
			},
		},
	}, nil)

	res, err := s.Boolean(context.TODO(), &rpcevaluation.EvaluationRequest{
		FlagKey:      flagKey,
		EntityId:     "test-entity",
		NamespaceKey: namespaceKey,
		Context: map[string]string{
			"hello": "world",
		},
	})

	require.NoError(t, err)

	assert.Equal(t, false, res.Value)
	assert.Equal(t, rpcevaluation.EvaluationReason_MATCH_EVALUATION_REASON, res.Reason)
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
			RolloutType:  flipt.RolloutType_THRESHOLD_ROLLOUT_TYPE,
			Threshold: &storage.RolloutThreshold{
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

	res, err := s.Boolean(context.TODO(), &rpcevaluation.EvaluationRequest{
		FlagKey:      flagKey,
		EntityId:     "test-entity",
		NamespaceKey: namespaceKey,
		Context: map[string]string{
			"hello": "world",
		},
	})

	require.NoError(t, err)

	assert.Equal(t, true, res.Value)
	assert.Equal(t, rpcevaluation.EvaluationReason_MATCH_EVALUATION_REASON, res.Reason)
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

	res, err := s.Boolean(context.TODO(), &rpcevaluation.EvaluationRequest{
		FlagKey:      flagKey,
		EntityId:     "test-entity",
		NamespaceKey: namespaceKey,
		Context: map[string]string{
			"hello": "world",
		},
	})

	require.NoError(t, err)

	assert.Equal(t, true, res.Value)
	assert.Equal(t, rpcevaluation.EvaluationReason_MATCH_EVALUATION_REASON, res.Reason)
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
			RolloutType:  flipt.RolloutType_THRESHOLD_ROLLOUT_TYPE,
			Threshold: &storage.RolloutThreshold{
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

	res, err := s.Boolean(context.TODO(), &rpcevaluation.EvaluationRequest{
		FlagKey:      flagKey,
		EntityId:     "test-entity",
		NamespaceKey: namespaceKey,
		Context: map[string]string{
			"hello": "world",
		},
	})

	require.Nil(t, res)

	assert.EqualError(t, err, "rollout rank: 0 detected out of order")
}
