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

func TestBoolean_DefaultRule_NoRollouts(t *testing.T) {
	var (
		flagKey      = "test-flag"
		namespaceKey = "test-namespace"
		store        = &evaluationStoreMock{}
		logger       = zaptest.NewLogger(t)
		s            = New(logger, store)
	)

	store.On("GetEvaluationRollouts", mock.Anything, flagKey, namespaceKey).Return([]*storage.EvaluationRollout{}, nil)

	store.On("GetFlag", mock.Anything, mock.Anything, mock.Anything).Return(&flipt.Flag{
		NamespaceKey: "test-namespace",
		Key:          "test-flag",
		Enabled:      true,
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

func TestBoolean_DefaultRuleFallthrough_WithPercentageRollout(t *testing.T) {
	var (
		flagKey      = "test-flag"
		namespaceKey = "test-namespace"
		store        = &evaluationStoreMock{}
		logger       = zaptest.NewLogger(t)
		s            = New(logger, store)
	)

	store.On("GetEvaluationRollouts", mock.Anything, flagKey, namespaceKey).Return([]*storage.EvaluationRollout{
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

	store.On("GetFlag", mock.Anything, mock.Anything, mock.Anything).Return(&flipt.Flag{
		NamespaceKey: "test-namespace",
		Key:          "test-flag",
		Enabled:      true,
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

	defer store.AssertNotCalled(t, "GetFlag", flagKey, namespaceKey)

	store.On("GetEvaluationRollouts", mock.Anything, flagKey, namespaceKey).Return([]*storage.EvaluationRollout{
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

	defer store.AssertNotCalled(t, "GetFlag", flagKey, namespaceKey)

	store.On("GetEvaluationRollouts", mock.Anything, flagKey, namespaceKey).Return([]*storage.EvaluationRollout{
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

	defer store.AssertNotCalled(t, "GetFlag", flagKey, namespaceKey)

	store.On("GetEvaluationRollouts", mock.Anything, flagKey, namespaceKey).Return([]*storage.EvaluationRollout{
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

	defer store.AssertNotCalled(t, "GetFlag", flagKey, namespaceKey)

	store.On("GetEvaluationRollouts", mock.Anything, flagKey, namespaceKey).Return([]*storage.EvaluationRollout{
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

func TestBatch_UnknownFlagType(t *testing.T) {
	var (
		flagKey      = "test-flag"
		namespaceKey = "test-namespace"
		store        = &evaluationStoreMock{}
		logger       = zaptest.NewLogger(t)
		s            = New(logger, store)
	)

	defer store.AssertNotCalled(t, "GetEvaluationRollouts", mock.Anything, flagKey, namespaceKey)

	store.On("GetFlag", mock.Anything, namespaceKey, flagKey).Return(&flipt.Flag{
		Key:         flagKey,
		Enabled:     true,
		Description: "test-flag",
		Type:        3,
	}, nil)

	_, err := s.Batch(context.TODO(), &rpcEvaluation.BatchEvaluationRequest{
		Requests: []*rpcEvaluation.EvaluationRequest{
			{
				FlagKey:      flagKey,
				EntityId:     "test-entity",
				NamespaceKey: namespaceKey,
				Context: map[string]string{
					"hello": "world",
				},
			},
		},
		ExcludeNotFound: false,
	})

	assert.Error(t, err)
	assert.EqualError(t, err, "unknown flag type: 3")
}

func TestBatch_NotFoundError(t *testing.T) {
	var (
		flagKey      = "test-flag"
		namespaceKey = "test-namespace"
		store        = &evaluationStoreMock{}
		logger       = zaptest.NewLogger(t)
		s            = New(logger, store)
	)

	defer store.AssertNotCalled(t, "GetEvaluationRollouts", mock.Anything, flagKey, namespaceKey)

	store.On("GetFlag", mock.Anything, namespaceKey, flagKey).Return(&flipt.Flag{}, errs.ErrNotFound("test-flag not found"))

	_, err := s.Batch(context.TODO(), &rpcEvaluation.BatchEvaluationRequest{
		Requests: []*rpcEvaluation.EvaluationRequest{
			{
				FlagKey:      flagKey,
				EntityId:     "test-entity",
				NamespaceKey: namespaceKey,
				Context: map[string]string{
					"hello": "world",
				},
			},
		},
		ExcludeNotFound: false,
	})

	assert.Error(t, err)
	assert.EqualError(t, err, "test-flag not found not found")
}

func TestBatch_EvaluationsExludingNotFound(t *testing.T) {
	var (
		flagKey        = "test-flag"
		anotherFlagKey = "another-test-flag"
		namespaceKey   = "test-namespace"
		store          = &evaluationStoreMock{}
		logger         = zaptest.NewLogger(t)
		s              = New(logger, store)
	)

	store.On("GetFlag", mock.Anything, mock.Anything, mock.Anything).Return(&flipt.Flag{
		NamespaceKey: "test-namespace",
		Key:          "test-flag",
		Enabled:      true,
		Type:         flipt.FlagType_BOOLEAN_FLAG_TYPE,
	}, nil).Times(2)

	store.On("GetFlag", mock.Anything, namespaceKey, anotherFlagKey).Return(&flipt.Flag{}, errs.ErrNotFound("another-test-flag"))

	store.On("GetEvaluationRollouts", mock.Anything, flagKey, namespaceKey).Return([]*storage.EvaluationRollout{
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

	res, err := s.Batch(context.TODO(), &rpcEvaluation.BatchEvaluationRequest{
		Requests: []*rpcEvaluation.EvaluationRequest{
			{
				FlagKey:      flagKey,
				EntityId:     "test-entity",
				NamespaceKey: namespaceKey,
				Context: map[string]string{
					"hello": "world",
				},
			},
			{
				FlagKey:      anotherFlagKey,
				EntityId:     "test-entity",
				NamespaceKey: namespaceKey,
				Context: map[string]string{
					"hello": "world",
				},
			},
		},
		ExcludeNotFound: true,
	})

	require.NoError(t, err)

	assert.Len(t, res.Responses, 1)

	b, ok := res.Responses[0].Response.(*rpcEvaluation.EvaluationResponse_BooleanResponse)
	assert.True(t, ok, "response should be a boolean evaluation response")
	assert.Equal(t, true, b.BooleanResponse.Value)
}
