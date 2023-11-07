package evaluation

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	errs "go.flipt.io/flipt/errors"
	"go.flipt.io/flipt/internal/storage"
	"go.flipt.io/flipt/rpc/flipt"
	"go.flipt.io/flipt/rpc/flipt/evaluation"
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

	res, err := s.Variant(context.TODO(), &evaluation.EvaluationRequest{
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

	res, err := s.Variant(context.TODO(), &evaluation.EvaluationRequest{
		FlagKey:      flagKey,
		EntityId:     "test-entity",
		NamespaceKey: namespaceKey,
		Context: map[string]string{
			"hello": "world",
		},
	})

	require.Nil(t, res)

	assert.EqualError(t, err, "flag type BOOLEAN_FLAG_TYPE invalid")
}

func TestVariant_FlagDisabled(t *testing.T) {
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
		Enabled:      false,
		Type:         flipt.FlagType_VARIANT_FLAG_TYPE,
	}, nil)

	res, err := s.Variant(context.TODO(), &evaluation.EvaluationRequest{
		FlagKey:      flagKey,
		EntityId:     "test-entity",
		NamespaceKey: namespaceKey,
		Context: map[string]string{
			"hello": "world",
		},
	})

	require.NoError(t, err)

	assert.False(t, res.Match)
	assert.Equal(t, evaluation.EvaluationReason_FLAG_DISABLED_EVALUATION_REASON, res.Reason)
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

	res, err := s.Variant(context.TODO(), &evaluation.EvaluationRequest{
		FlagKey:      flagKey,
		EntityId:     "test-entity",
		NamespaceKey: namespaceKey,
		Context: map[string]string{
			"hello": "world",
		},
	})

	require.Nil(t, res)
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
				Id:   "1",
				Rank: 0,
				Segments: []*storage.EvaluationSegment{
					{
						Key:       "bar",
						MatchType: evaluation.EvaluationSegmentMatchType_ALL_SEGMENT_MATCH_TYPE,
						Constraints: []*storage.EvaluationConstraint{
							{
								Id:       "2",
								Type:     evaluation.EvaluationConstraintComparisonType_STRING_CONSTRAINT_COMPARISON_TYPE,
								Property: "hello",
								Operator: flipt.OpEQ,
								Value:    "world",
							},
						},
					},
				},
			},
		}, nil)

	store.On("GetEvaluationDistributions", mock.Anything, "1").Return([]*storage.EvaluationDistribution{}, nil)

	res, err := s.Variant(context.TODO(), &evaluation.EvaluationRequest{
		FlagKey:      flagKey,
		EntityId:     "test-entity",
		NamespaceKey: namespaceKey,
		Context: map[string]string{
			"hello": "world",
		},
	})

	require.NoError(t, err)

	assert.Equal(t, true, res.Match)
	assert.Contains(t, res.SegmentKeys, "bar")
	assert.Equal(t, evaluation.EvaluationReason_MATCH_EVALUATION_REASON, res.Reason)
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

	res, err := s.Boolean(context.TODO(), &evaluation.EvaluationRequest{
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

	res, err := s.Boolean(context.TODO(), &evaluation.EvaluationRequest{
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

	res, err := s.Boolean(context.TODO(), &evaluation.EvaluationRequest{
		FlagKey:      flagKey,
		EntityId:     "test-entity",
		NamespaceKey: namespaceKey,
		Context: map[string]string{
			"hello": "world",
		},
	})

	require.NoError(t, err)

	assert.Equal(t, true, res.Enabled)
	assert.Equal(t, evaluation.EvaluationReason_DEFAULT_EVALUATION_REASON, res.Reason)
	assert.Equal(t, flagKey, res.FlagKey)
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
			Rank: 1,
			Type: evaluation.EvaluationRolloutType_THRESHOLD_ROLLOUT_TYPE,
			Rule: &evaluation.EvaluationRollout_Threshold{
				Threshold: &storage.RolloutThreshold{
					Percentage: 5,
					Value:      false,
				},
			},
		},
	}, nil)

	res, err := s.Boolean(context.TODO(), &evaluation.EvaluationRequest{
		FlagKey:      flagKey,
		EntityId:     "test-entity",
		NamespaceKey: namespaceKey,
		Context: map[string]string{
			"hello": "world",
		},
	})

	require.NoError(t, err)

	assert.Equal(t, true, res.Enabled)
	assert.Equal(t, evaluation.EvaluationReason_DEFAULT_EVALUATION_REASON, res.Reason)
	assert.Equal(t, flagKey, res.FlagKey)
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
			Rank: 1,
			Type: evaluation.EvaluationRolloutType_THRESHOLD_ROLLOUT_TYPE,
			Rule: &evaluation.EvaluationRollout_Threshold{
				Threshold: &storage.RolloutThreshold{
					Percentage: 70,
					Value:      false,
				},
			},
		},
	}, nil)

	res, err := s.Boolean(context.TODO(), &evaluation.EvaluationRequest{
		FlagKey:      flagKey,
		EntityId:     "test-entity",
		NamespaceKey: namespaceKey,
		Context: map[string]string{
			"hello": "world",
		},
	})

	require.NoError(t, err)

	assert.Equal(t, false, res.Enabled)
	assert.Equal(t, evaluation.EvaluationReason_MATCH_EVALUATION_REASON, res.Reason)
	assert.Equal(t, flagKey, res.FlagKey)
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
			Rank: 1,
			Type: evaluation.EvaluationRolloutType_THRESHOLD_ROLLOUT_TYPE,
			Rule: &evaluation.EvaluationRollout_Threshold{
				Threshold: &storage.RolloutThreshold{
					Percentage: 5,
					Value:      false,
				},
			},
		},
		{
			Type: evaluation.EvaluationRolloutType_SEGMENT_ROLLOUT_TYPE,
			Rank: 2,
			Rule: &evaluation.EvaluationRollout_Segment{
				Segment: &storage.RolloutSegment{
					Value:           true,
					SegmentOperator: evaluation.EvaluationSegmentOperator_OR_SEGMENT_OPERATOR,
					Segments: []*storage.EvaluationSegment{
						{
							Key:       "test-segment",
							MatchType: evaluation.EvaluationSegmentMatchType_ANY_SEGMENT_MATCH_TYPE,
							Constraints: []*storage.EvaluationConstraint{
								{
									Type:     evaluation.EvaluationConstraintComparisonType_STRING_CONSTRAINT_COMPARISON_TYPE,
									Property: "hello",
									Operator: flipt.OpEQ,
									Value:    "world",
								},
							},
						},
					},
				},
			},
		},
	}, nil)

	res, err := s.Boolean(context.TODO(), &evaluation.EvaluationRequest{
		FlagKey:      flagKey,
		EntityId:     "test-entity",
		NamespaceKey: namespaceKey,
		Context: map[string]string{
			"hello": "world",
		},
	})

	require.NoError(t, err)

	assert.Equal(t, true, res.Enabled)
	assert.Equal(t, evaluation.EvaluationReason_MATCH_EVALUATION_REASON, res.Reason)
	assert.Equal(t, flagKey, res.FlagKey)
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
			Type: evaluation.EvaluationRolloutType_SEGMENT_ROLLOUT_TYPE,
			Rank: 1,
			Rule: &evaluation.EvaluationRollout_Segment{
				Segment: &storage.RolloutSegment{
					Value:           true,
					SegmentOperator: evaluation.EvaluationSegmentOperator_OR_SEGMENT_OPERATOR,
					Segments: []*storage.EvaluationSegment{
						{
							Key:       "test-segment",
							MatchType: evaluation.EvaluationSegmentMatchType_ANY_SEGMENT_MATCH_TYPE,
							Constraints: []*storage.EvaluationConstraint{
								{
									Type:     evaluation.EvaluationConstraintComparisonType_NUMBER_CONSTRAINT_COMPARISON_TYPE,
									Property: "pitimes100",
									Operator: flipt.OpEQ,
									Value:    "314",
								},
								{
									Type:     evaluation.EvaluationConstraintComparisonType_STRING_CONSTRAINT_COMPARISON_TYPE,
									Property: "hello",
									Operator: flipt.OpEQ,
									Value:    "world",
								},
							},
						},
					},
				},
			},
		},
	}, nil)

	res, err := s.Boolean(context.TODO(), &evaluation.EvaluationRequest{
		FlagKey:      flagKey,
		EntityId:     "test-entity",
		NamespaceKey: namespaceKey,
		Context: map[string]string{
			"hello": "world",
		},
	})

	require.NoError(t, err)

	assert.Equal(t, true, res.Enabled)
	assert.Equal(t, evaluation.EvaluationReason_MATCH_EVALUATION_REASON, res.Reason)
	assert.Equal(t, flagKey, res.FlagKey)
}

func TestBoolean_SegmentMatch_MultipleSegments_WithAnd(t *testing.T) {
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
			Type: evaluation.EvaluationRolloutType_SEGMENT_ROLLOUT_TYPE,
			Rank: 1,
			Rule: &evaluation.EvaluationRollout_Segment{
				Segment: &storage.RolloutSegment{
					Value:           true,
					SegmentOperator: evaluation.EvaluationSegmentOperator_AND_SEGMENT_OPERATOR,
					Segments: []*storage.EvaluationSegment{
						{
							Key:       "test-segment",
							MatchType: evaluation.EvaluationSegmentMatchType_ANY_SEGMENT_MATCH_TYPE,
							Constraints: []*storage.EvaluationConstraint{
								{
									Type:     evaluation.EvaluationConstraintComparisonType_NUMBER_CONSTRAINT_COMPARISON_TYPE,
									Property: "pitimes100",
									Operator: flipt.OpEQ,
									Value:    "314",
								},
							},
						},
						{
							Key:       "another-segment",
							MatchType: evaluation.EvaluationSegmentMatchType_ANY_SEGMENT_MATCH_TYPE,
							Constraints: []*storage.EvaluationConstraint{
								{
									Type:     evaluation.EvaluationConstraintComparisonType_STRING_CONSTRAINT_COMPARISON_TYPE,
									Property: "hello",
									Operator: flipt.OpEQ,
									Value:    "world",
								},
							},
						},
					},
				},
			},
		},
	}, nil)

	res, err := s.Boolean(context.TODO(), &evaluation.EvaluationRequest{
		FlagKey:      flagKey,
		EntityId:     "test-entity",
		NamespaceKey: namespaceKey,
		Context: map[string]string{
			"hello":      "world",
			"pitimes100": "314",
		},
	})

	require.NoError(t, err)

	assert.Equal(t, true, res.Enabled)
	assert.Equal(t, evaluation.EvaluationReason_MATCH_EVALUATION_REASON, res.Reason)
	assert.Equal(t, flagKey, res.FlagKey)
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
			Rank: 1,
			Type: evaluation.EvaluationRolloutType_THRESHOLD_ROLLOUT_TYPE,
			Rule: &evaluation.EvaluationRollout_Threshold{
				Threshold: &storage.RolloutThreshold{
					Percentage: 5,
					Value:      false,
				},
			},
		},
		{
			Type: evaluation.EvaluationRolloutType_SEGMENT_ROLLOUT_TYPE,
			Rank: 0,
			Rule: &evaluation.EvaluationRollout_Segment{
				Segment: &storage.RolloutSegment{
					Value:           true,
					SegmentOperator: evaluation.EvaluationSegmentOperator_OR_SEGMENT_OPERATOR,
					Segments: []*storage.EvaluationSegment{
						{
							Key:       "test-segment",
							MatchType: evaluation.EvaluationSegmentMatchType_ANY_SEGMENT_MATCH_TYPE,
							Constraints: []*storage.EvaluationConstraint{
								{
									Type:     evaluation.EvaluationConstraintComparisonType_STRING_CONSTRAINT_COMPARISON_TYPE,
									Property: "hello",
									Operator: flipt.OpEQ,
									Value:    "world",
								},
							},
						},
					},
				},
			},
		},
	}, nil)

	res, err := s.Boolean(context.TODO(), &evaluation.EvaluationRequest{
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

	_, err := s.Batch(context.TODO(), &evaluation.BatchEvaluationRequest{
		Requests: []*evaluation.EvaluationRequest{
			{
				FlagKey:      flagKey,
				EntityId:     "test-entity",
				NamespaceKey: namespaceKey,
				Context: map[string]string{
					"hello": "world",
				},
			},
		},
	})

	assert.Error(t, err)
	assert.EqualError(t, err, "unknown flag type: 3")
}

func TestBatch_InternalError_GetFlag(t *testing.T) {
	var (
		flagKey      = "test-flag"
		namespaceKey = "test-namespace"
		store        = &evaluationStoreMock{}
		logger       = zaptest.NewLogger(t)
		s            = New(logger, store)
	)

	defer store.AssertNotCalled(t, "GetEvaluationRollouts", mock.Anything, flagKey, namespaceKey)

	store.On("GetFlag", mock.Anything, namespaceKey, flagKey).Return(&flipt.Flag{}, errors.New("internal error"))

	_, err := s.Batch(context.TODO(), &evaluation.BatchEvaluationRequest{
		Requests: []*evaluation.EvaluationRequest{
			{
				FlagKey:      flagKey,
				EntityId:     "test-entity",
				NamespaceKey: namespaceKey,
				Context: map[string]string{
					"hello": "world",
				},
			},
		},
	})

	assert.Error(t, err)
	assert.EqualError(t, err, "internal error")
}

func TestBatch_Success(t *testing.T) {
	var (
		flagKey        = "test-flag"
		anotherFlagKey = "another-test-flag"
		variantFlagKey = "variant-test-flag"
		namespaceKey   = "test-namespace"
		store          = &evaluationStoreMock{}
		logger         = zaptest.NewLogger(t)
		s              = New(logger, store)
	)

	store.On("GetFlag", mock.Anything, namespaceKey, flagKey).Return(&flipt.Flag{
		NamespaceKey: "test-namespace",
		Key:          "test-flag",
		Enabled:      true,
		Type:         flipt.FlagType_BOOLEAN_FLAG_TYPE,
	}, nil)

	store.On("GetFlag", mock.Anything, namespaceKey, anotherFlagKey).Return(&flipt.Flag{}, errs.ErrNotFound("another-test-flag"))

	store.On("GetFlag", mock.Anything, namespaceKey, variantFlagKey).Return(&flipt.Flag{
		NamespaceKey: "test-namespace",
		Key:          "variant-test-flag",
		Enabled:      true,
		Type:         flipt.FlagType_VARIANT_FLAG_TYPE,
	}, nil)

	store.On("GetEvaluationRollouts", mock.Anything, namespaceKey, flagKey).Return([]*storage.EvaluationRollout{
		{
			Rank: 1,
			Type: evaluation.EvaluationRolloutType_THRESHOLD_ROLLOUT_TYPE,
			Rule: &evaluation.EvaluationRollout_Threshold{
				Threshold: &storage.RolloutThreshold{
					Percentage: 80,
					Value:      true,
				},
			},
		},
	}, nil)

	store.On("GetEvaluationRules", mock.Anything, namespaceKey, variantFlagKey).Return(
		[]*storage.EvaluationRule{
			{
				Id:   "1",
				Rank: 0,
				Segments: []*storage.EvaluationSegment{
					{
						Key:       "bar",
						MatchType: evaluation.EvaluationSegmentMatchType_ALL_SEGMENT_MATCH_TYPE,
						Constraints: []*storage.EvaluationConstraint{
							{
								Id:       "2",
								Type:     evaluation.EvaluationConstraintComparisonType_STRING_CONSTRAINT_COMPARISON_TYPE,
								Property: "hello",
								Operator: flipt.OpEQ,
								Value:    "world",
							},
						},
					},
				},
			},
		}, nil)

	store.On("GetEvaluationDistributions", mock.Anything, "1").Return([]*storage.EvaluationDistribution{}, nil)

	res, err := s.Batch(context.TODO(), &evaluation.BatchEvaluationRequest{
		Requests: []*evaluation.EvaluationRequest{
			{
				RequestId:    "1",
				FlagKey:      flagKey,
				EntityId:     "test-entity",
				NamespaceKey: namespaceKey,
				Context: map[string]string{
					"hello": "world",
				},
			},
			{
				RequestId:    "2",
				FlagKey:      anotherFlagKey,
				EntityId:     "test-entity",
				NamespaceKey: namespaceKey,
				Context: map[string]string{
					"hello": "world",
				},
			},
			{
				RequestId:    "3",
				FlagKey:      variantFlagKey,
				EntityId:     "test-entity",
				NamespaceKey: namespaceKey,
				Context: map[string]string{
					"hello": "world",
				},
			},
		},
	})

	require.NoError(t, err)

	assert.Len(t, res.Responses, 3)

	b, ok := res.Responses[0].Response.(*evaluation.EvaluationResponse_BooleanResponse)
	assert.True(t, ok, "response should be a boolean evaluation response")
	assert.True(t, b.BooleanResponse.Enabled, "value should be true from match")
	assert.Equal(t, evaluation.EvaluationReason_MATCH_EVALUATION_REASON, b.BooleanResponse.Reason)
	assert.Equal(t, evaluation.EvaluationResponseType_BOOLEAN_EVALUATION_RESPONSE_TYPE, res.Responses[0].Type)
	assert.Equal(t, "1", b.BooleanResponse.RequestId)
	assert.Equal(t, flagKey, b.BooleanResponse.FlagKey)

	e, ok := res.Responses[1].Response.(*evaluation.EvaluationResponse_ErrorResponse)
	assert.True(t, ok, "response should be a error evaluation response")
	assert.Equal(t, anotherFlagKey, e.ErrorResponse.FlagKey)
	assert.Equal(t, namespaceKey, e.ErrorResponse.NamespaceKey)
	assert.Equal(t, evaluation.ErrorEvaluationReason_NOT_FOUND_ERROR_EVALUATION_REASON, e.ErrorResponse.Reason)
	assert.Equal(t, evaluation.EvaluationResponseType_ERROR_EVALUATION_RESPONSE_TYPE, res.Responses[1].Type)

	v, ok := res.Responses[2].Response.(*evaluation.EvaluationResponse_VariantResponse)
	assert.True(t, ok, "response should be a variant evaluation response")
	assert.True(t, v.VariantResponse.Match, "variant response should have matched")
	assert.Contains(t, v.VariantResponse.SegmentKeys, "bar")
	assert.Equal(t, evaluation.EvaluationReason_MATCH_EVALUATION_REASON, v.VariantResponse.Reason)
	assert.Equal(t, evaluation.EvaluationResponseType_VARIANT_EVALUATION_RESPONSE_TYPE, res.Responses[2].Type)
	assert.Equal(t, "3", v.VariantResponse.RequestId)
	assert.Equal(t, variantFlagKey, v.VariantResponse.FlagKey)
}
