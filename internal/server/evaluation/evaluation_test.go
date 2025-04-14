package evaluation

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	errs "go.flipt.io/flipt/errors"
	"go.flipt.io/flipt/internal/server/environments"
	"go.flipt.io/flipt/internal/storage"
	"go.flipt.io/flipt/rpc/flipt"
	"go.flipt.io/flipt/rpc/flipt/core"
	rpcevaluation "go.flipt.io/flipt/rpc/flipt/evaluation"
	"go.uber.org/zap/zaptest"
)

func TestVariant_FlagNotFound(t *testing.T) {
	var (
		flagKey      = "test-flag"
		namespaceKey = "test-namespace"
		envStore     = NewMockEnvironmentStore(t)
		environment  = environments.NewMockEnvironment(t)
		store        = storage.NewMockReadOnlyStore(t)
		logger       = zaptest.NewLogger(t)
		s            = New(logger, envStore)
	)

	envStore.On("GetFromContext", mock.Anything).Return(environment)

	environment.On("EvaluationStore").Return(store, nil)

	store.On("GetFlag", mock.Anything, storage.NewResource(namespaceKey, flagKey)).Return(&core.Flag{}, errs.ErrNotFound("test-flag"))

	res, err := s.Variant(context.TODO(), &rpcevaluation.EvaluationRequest{
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
		envStore     = NewMockEnvironmentStore(t)
		environment  = environments.NewMockEnvironment(t)
		store        = storage.NewMockReadOnlyStore(t)
		logger       = zaptest.NewLogger(t)
		s            = New(logger, envStore)
	)

	envStore.On("GetFromContext", mock.Anything).Return(environment)
	environment.On("EvaluationStore").Return(store, nil)

	store.On("GetFlag", mock.Anything, storage.NewResource(namespaceKey, flagKey)).Return(&core.Flag{
		Key:     flagKey,
		Enabled: true,
		Type:    core.FlagType_BOOLEAN_FLAG_TYPE,
	}, nil)

	res, err := s.Variant(context.TODO(), &rpcevaluation.EvaluationRequest{
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
		envStore     = NewMockEnvironmentStore(t)
		environment  = environments.NewMockEnvironment(t)
		store        = storage.NewMockReadOnlyStore(t)
		logger       = zaptest.NewLogger(t)
		s            = New(logger, envStore)
	)

	envStore.On("GetFromContext", mock.Anything).Return(environment)
	environment.On("EvaluationStore").Return(store, nil)

	store.On("GetFlag", mock.Anything, storage.NewResource(namespaceKey, flagKey)).Return(&core.Flag{
		Key:     flagKey,
		Enabled: false,
		Type:    core.FlagType_VARIANT_FLAG_TYPE,
	}, nil)

	res, err := s.Variant(context.TODO(), &rpcevaluation.EvaluationRequest{
		FlagKey:      flagKey,
		EntityId:     "test-entity",
		NamespaceKey: namespaceKey,
		Context: map[string]string{
			"hello": "world",
		},
	})

	require.NoError(t, err)
	assert.False(t, res.Match)
	assert.Equal(t, rpcevaluation.EvaluationReason_FLAG_DISABLED_EVALUATION_REASON, res.Reason)
}

func TestVariant_EvaluateFailure_OnGetEvaluationRules(t *testing.T) {
	var (
		flagKey      = "test-flag"
		namespaceKey = "test-namespace"
		envStore     = NewMockEnvironmentStore(t)
		environment  = environments.NewMockEnvironment(t)
		store        = storage.NewMockReadOnlyStore(t)
		logger       = zaptest.NewLogger(t)
		s            = New(logger, envStore)
		flag         = &core.Flag{
			Key:     flagKey,
			Enabled: true,
			Type:    core.FlagType_VARIANT_FLAG_TYPE,
		}
	)

	envStore.On("GetFromContext", mock.Anything).Return(environment)
	environment.On("EvaluationStore").Return(store, nil)

	store.On("GetFlag", mock.Anything, storage.NewResource(namespaceKey, flagKey)).Return(flag, nil)
	store.On("GetEvaluationRules", mock.Anything, storage.NewResource(namespaceKey, flagKey)).Return([]*storage.EvaluationRule{}, errs.ErrInvalid("some invalid error"))

	res, err := s.Variant(context.TODO(), &rpcevaluation.EvaluationRequest{
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
		envStore     = NewMockEnvironmentStore(t)
		environment  = environments.NewMockEnvironment(t)
		store        = storage.NewMockReadOnlyStore(t)
		logger       = zaptest.NewLogger(t)
		s            = New(logger, envStore)
		flag         = &core.Flag{
			Key:     flagKey,
			Enabled: true,
			Type:    core.FlagType_VARIANT_FLAG_TYPE,
		}
	)

	envStore.On("GetFromContext", mock.Anything).Return(environment)
	environment.On("EvaluationStore").Return(store, nil)

	store.On("GetFlag", mock.Anything, storage.NewResource(namespaceKey, flagKey)).Return(flag, nil)

	store.On("GetEvaluationRules", mock.Anything, storage.NewResource(namespaceKey, flagKey)).Return(
		[]*storage.EvaluationRule{
			{
				ID:      "1",
				FlagKey: flagKey,
				Rank:    0,
				Segments: map[string]*storage.EvaluationSegment{
					"bar": {
						SegmentKey: "bar",
						MatchType:  core.MatchType_ALL_MATCH_TYPE,
						Constraints: []storage.EvaluationConstraint{
							{
								ID:       "2",
								Type:     core.ComparisonType_STRING_COMPARISON_TYPE,
								Property: "hello",
								Operator: flipt.OpEQ,
								Value:    "world",
							},
						},
					},
				},
			},
		}, nil)

	store.On(
		"GetEvaluationDistributions",
		mock.Anything,
		storage.NewResource(namespaceKey, flagKey),
		storage.NewID("1"),
	).Return([]*storage.EvaluationDistribution{}, nil)

	res, err := s.Variant(context.TODO(), &rpcevaluation.EvaluationRequest{
		FlagKey:      flagKey,
		EntityId:     "test-entity",
		NamespaceKey: namespaceKey,
		Context: map[string]string{
			"hello": "world",
		},
	})

	require.NoError(t, err)
	assert.True(t, res.Match)
	assert.Contains(t, res.SegmentKeys, "bar")
	assert.Equal(t, rpcevaluation.EvaluationReason_MATCH_EVALUATION_REASON, res.Reason)
}

func TestBoolean_FlagNotFoundError(t *testing.T) {
	var (
		flagKey      = "test-flag"
		namespaceKey = "test-namespace"
		envStore     = NewMockEnvironmentStore(t)
		environment  = environments.NewMockEnvironment(t)
		store        = storage.NewMockReadOnlyStore(t)
		logger       = zaptest.NewLogger(t)
		s            = New(logger, envStore)
	)

	envStore.On("GetFromContext", mock.Anything).Return(environment)
	environment.On("EvaluationStore").Return(store, nil)

	store.On("GetFlag", mock.Anything, storage.NewResource(namespaceKey, flagKey)).Return(&core.Flag{}, errs.ErrNotFound("test-flag"))

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
		envStore     = NewMockEnvironmentStore(t)
		environment  = environments.NewMockEnvironment(t)
		store        = storage.NewMockReadOnlyStore(t)
		logger       = zaptest.NewLogger(t)
		s            = New(logger, envStore)
	)

	envStore.On("GetFromContext", mock.Anything).Return(environment)
	environment.On("EvaluationStore").Return(store, nil)

	store.On("GetFlag", mock.Anything, storage.NewResource(namespaceKey, flagKey)).Return(&core.Flag{
		Key:     flagKey,
		Enabled: true,
		Type:    core.FlagType_VARIANT_FLAG_TYPE,
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
		envStore     = NewMockEnvironmentStore(t)
		environment  = environments.NewMockEnvironment(t)
		store        = storage.NewMockReadOnlyStore(t)
		logger       = zaptest.NewLogger(t)
		s            = New(logger, envStore)
	)

	envStore.On("GetFromContext", mock.Anything).Return(environment)
	environment.On("EvaluationStore").Return(store, nil)

	store.On("GetFlag", mock.Anything, storage.NewResource(namespaceKey, flagKey)).Return(&core.Flag{
		Key:     flagKey,
		Enabled: true,
		Type:    core.FlagType_BOOLEAN_FLAG_TYPE,
	}, nil)

	store.On("GetEvaluationRollouts", mock.Anything, storage.NewResource(namespaceKey, flagKey)).Return([]*storage.EvaluationRollout{}, nil)

	res, err := s.Boolean(context.TODO(), &rpcevaluation.EvaluationRequest{
		FlagKey:      flagKey,
		EntityId:     "test-entity",
		NamespaceKey: namespaceKey,
		Context: map[string]string{
			"hello": "world",
		},
	})

	require.NoError(t, err)
	assert.True(t, res.Enabled)
	assert.Equal(t, rpcevaluation.EvaluationReason_DEFAULT_EVALUATION_REASON, res.Reason)
	assert.Equal(t, flagKey, res.FlagKey)
}

func TestBoolean_DefaultRuleFallthrough_WithPercentageRollout(t *testing.T) {
	var (
		flagKey      = "test-flag"
		namespaceKey = "test-namespace"
		envStore     = NewMockEnvironmentStore(t)
		environment  = environments.NewMockEnvironment(t)
		store        = storage.NewMockReadOnlyStore(t)
		logger       = zaptest.NewLogger(t)
		s            = New(logger, envStore)
	)

	envStore.On("GetFromContext", mock.Anything).Return(environment)
	environment.On("EvaluationStore").Return(store, nil)

	store.On("GetFlag", mock.Anything, storage.NewResource(namespaceKey, flagKey)).Return(&core.Flag{
		Key:     flagKey,
		Enabled: true,
		Type:    core.FlagType_BOOLEAN_FLAG_TYPE,
	}, nil)

	store.On("GetEvaluationRollouts", mock.Anything, storage.NewResource(namespaceKey, flagKey)).Return([]*storage.EvaluationRollout{
		{
			NamespaceKey: namespaceKey,
			Rank:         1,
			RolloutType:  core.RolloutType_THRESHOLD_ROLLOUT_TYPE,
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
	assert.True(t, res.Enabled)
	assert.Equal(t, rpcevaluation.EvaluationReason_DEFAULT_EVALUATION_REASON, res.Reason)
	assert.Equal(t, flagKey, res.FlagKey)
}

func TestBoolean_PercentageRuleMatch(t *testing.T) {
	var (
		flagKey      = "test-flag"
		namespaceKey = "test-namespace"
		envStore     = NewMockEnvironmentStore(t)
		environment  = environments.NewMockEnvironment(t)
		store        = storage.NewMockReadOnlyStore(t)
		logger       = zaptest.NewLogger(t)
		s            = New(logger, envStore)
	)

	envStore.On("GetFromContext", mock.Anything).Return(environment)
	environment.On("EvaluationStore").Return(store, nil)

	store.On("GetFlag", mock.Anything, storage.NewResource(namespaceKey, flagKey)).Return(&core.Flag{
		Key:     flagKey,
		Enabled: true,
		Type:    core.FlagType_BOOLEAN_FLAG_TYPE,
	}, nil)

	store.On("GetEvaluationRollouts", mock.Anything, storage.NewResource(namespaceKey, flagKey)).Return([]*storage.EvaluationRollout{
		{
			NamespaceKey: namespaceKey,
			Rank:         1,
			RolloutType:  core.RolloutType_THRESHOLD_ROLLOUT_TYPE,
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
	assert.False(t, res.Enabled)
	assert.Equal(t, rpcevaluation.EvaluationReason_MATCH_EVALUATION_REASON, res.Reason)
	assert.Equal(t, flagKey, res.FlagKey)
}

func TestBoolean_PercentageRuleFallthrough_SegmentMatch(t *testing.T) {
	var (
		flagKey      = "test-flag"
		namespaceKey = "test-namespace"
		envStore     = NewMockEnvironmentStore(t)
		environment  = environments.NewMockEnvironment(t)
		store        = storage.NewMockReadOnlyStore(t)
		logger       = zaptest.NewLogger(t)
		s            = New(logger, envStore)
	)

	envStore.On("GetFromContext", mock.Anything).Return(environment)
	environment.On("EvaluationStore").Return(store, nil)

	store.On("GetFlag", mock.Anything, storage.NewResource(namespaceKey, flagKey)).Return(&core.Flag{
		Key:     flagKey,
		Enabled: true,
		Type:    core.FlagType_BOOLEAN_FLAG_TYPE,
	}, nil)

	store.On("GetEvaluationRollouts", mock.Anything, storage.NewResource(namespaceKey, flagKey)).Return([]*storage.EvaluationRollout{
		{
			NamespaceKey: namespaceKey,
			Rank:         1,
			RolloutType:  core.RolloutType_THRESHOLD_ROLLOUT_TYPE,
			Threshold: &storage.RolloutThreshold{
				Percentage: 5,
				Value:      false,
			},
		},
		{
			NamespaceKey: namespaceKey,
			RolloutType:  core.RolloutType_SEGMENT_ROLLOUT_TYPE,
			Rank:         2,
			Segment: &storage.RolloutSegment{
				Value:           true,
				SegmentOperator: core.SegmentOperator_OR_SEGMENT_OPERATOR,
				Segments: map[string]*storage.EvaluationSegment{
					"test-segment": {
						SegmentKey: "test-segment",
						MatchType:  core.MatchType_ANY_MATCH_TYPE,
						Constraints: []storage.EvaluationConstraint{
							{
								Type:     core.ComparisonType_STRING_COMPARISON_TYPE,
								Property: "hello",
								Operator: flipt.OpEQ,
								Value:    "world",
							},
						},
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
	assert.True(t, res.Enabled)
	assert.Equal(t, rpcevaluation.EvaluationReason_MATCH_EVALUATION_REASON, res.Reason)
	assert.Equal(t, flagKey, res.FlagKey)
}

func TestBoolean_SegmentMatch_MultipleConstraints(t *testing.T) {
	var (
		flagKey      = "test-flag"
		namespaceKey = "test-namespace"
		envStore     = NewMockEnvironmentStore(t)
		environment  = environments.NewMockEnvironment(t)
		store        = storage.NewMockReadOnlyStore(t)
		logger       = zaptest.NewLogger(t)
		s            = New(logger, envStore)
	)

	envStore.On("GetFromContext", mock.Anything).Return(environment)
	environment.On("EvaluationStore").Return(store, nil)

	store.On("GetFlag", mock.Anything, storage.NewResource(namespaceKey, flagKey)).Return(&core.Flag{
		Key:     flagKey,
		Enabled: true,
		Type:    core.FlagType_BOOLEAN_FLAG_TYPE,
	}, nil)

	store.On("GetEvaluationRollouts", mock.Anything, storage.NewResource(namespaceKey, flagKey)).Return([]*storage.EvaluationRollout{
		{
			NamespaceKey: namespaceKey,
			RolloutType:  core.RolloutType_SEGMENT_ROLLOUT_TYPE,
			Rank:         1,
			Segment: &storage.RolloutSegment{
				Value:           true,
				SegmentOperator: core.SegmentOperator_OR_SEGMENT_OPERATOR,
				Segments: map[string]*storage.EvaluationSegment{
					"test-segment": {
						SegmentKey: "test-segment",
						MatchType:  core.MatchType_ANY_MATCH_TYPE,
						Constraints: []storage.EvaluationConstraint{
							{
								Type:     core.ComparisonType_NUMBER_COMPARISON_TYPE,
								Property: "pitimes100",
								Operator: flipt.OpEQ,
								Value:    "314",
							},
							{
								Type:     core.ComparisonType_STRING_COMPARISON_TYPE,
								Property: "hello",
								Operator: flipt.OpEQ,
								Value:    "world",
							},
						},
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
	assert.True(t, res.Enabled)
	assert.Equal(t, rpcevaluation.EvaluationReason_MATCH_EVALUATION_REASON, res.Reason)
	assert.Equal(t, flagKey, res.FlagKey)
}

func TestBoolean_SegmentMatch_Constraint_EntityId(t *testing.T) {
	var (
		flagKey      = "test-flag"
		namespaceKey = "test-namespace"
		envStore     = NewMockEnvironmentStore(t)
		environment  = environments.NewMockEnvironment(t)
		store        = storage.NewMockReadOnlyStore(t)
		logger       = zaptest.NewLogger(t)
		s            = New(logger, envStore)
	)

	envStore.On("GetFromContext", mock.Anything).Return(environment)
	environment.On("EvaluationStore").Return(store, nil)

	store.On("GetFlag", mock.Anything, storage.NewResource(namespaceKey, flagKey)).Return(&core.Flag{
		Key:     flagKey,
		Enabled: true,
		Type:    core.FlagType_BOOLEAN_FLAG_TYPE,
	}, nil)

	store.On("GetEvaluationRollouts", mock.Anything, storage.NewResource(namespaceKey, flagKey)).Return([]*storage.EvaluationRollout{
		{
			NamespaceKey: namespaceKey,
			RolloutType:  core.RolloutType_SEGMENT_ROLLOUT_TYPE,
			Rank:         1,
			Segment: &storage.RolloutSegment{
				Value:           true,
				SegmentOperator: core.SegmentOperator_OR_SEGMENT_OPERATOR,
				Segments: map[string]*storage.EvaluationSegment{
					"test-segment": {
						SegmentKey: "test-segment",
						MatchType:  core.MatchType_ANY_MATCH_TYPE,
						Constraints: []storage.EvaluationConstraint{
							{
								Type:     core.ComparisonType_ENTITY_ID_COMPARISON_TYPE,
								Property: "entity",
								Operator: flipt.OpEQ,
								Value:    "user@core.io",
							},
						},
					},
				},
			},
		},
	}, nil)

	res, err := s.Boolean(context.TODO(), &rpcevaluation.EvaluationRequest{
		FlagKey:      flagKey,
		EntityId:     "user@core.io",
		NamespaceKey: namespaceKey,
		Context:      map[string]string{},
	})

	require.NoError(t, err)
	assert.True(t, res.Enabled)
	assert.Equal(t, rpcevaluation.EvaluationReason_MATCH_EVALUATION_REASON, res.Reason)
	assert.Equal(t, flagKey, res.FlagKey)
}

func TestBoolean_SegmentMatch_MultipleSegments_WithAnd(t *testing.T) {
	var (
		flagKey      = "test-flag"
		namespaceKey = "test-namespace"
		envStore     = NewMockEnvironmentStore(t)
		environment  = environments.NewMockEnvironment(t)
		store        = storage.NewMockReadOnlyStore(t)
		logger       = zaptest.NewLogger(t)
		s            = New(logger, envStore)
	)

	envStore.On("GetFromContext", mock.Anything).Return(environment)
	environment.On("EvaluationStore").Return(store, nil)

	store.On("GetFlag", mock.Anything, storage.NewResource(namespaceKey, flagKey)).Return(&core.Flag{
		Key:     flagKey,
		Enabled: true,
		Type:    core.FlagType_BOOLEAN_FLAG_TYPE,
	}, nil)

	store.On("GetEvaluationRollouts", mock.Anything, storage.NewResource(namespaceKey, flagKey)).Return([]*storage.EvaluationRollout{
		{
			NamespaceKey: namespaceKey,
			RolloutType:  core.RolloutType_SEGMENT_ROLLOUT_TYPE,
			Rank:         1,
			Segment: &storage.RolloutSegment{
				Value:           true,
				SegmentOperator: core.SegmentOperator_AND_SEGMENT_OPERATOR,
				Segments: map[string]*storage.EvaluationSegment{
					"test-segment": {
						SegmentKey: "test-segment",
						MatchType:  core.MatchType_ANY_MATCH_TYPE,
						Constraints: []storage.EvaluationConstraint{
							{
								Type:     core.ComparisonType_NUMBER_COMPARISON_TYPE,
								Property: "pitimes100",
								Operator: flipt.OpEQ,
								Value:    "314",
							},
						},
					},
					"another-segment": {
						SegmentKey: "another-segment",
						MatchType:  core.MatchType_ANY_MATCH_TYPE,
						Constraints: []storage.EvaluationConstraint{
							{
								Type:     core.ComparisonType_STRING_COMPARISON_TYPE,
								Property: "hello",
								Operator: flipt.OpEQ,
								Value:    "world",
							},
						},
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
			"hello":      "world",
			"pitimes100": "314",
		},
	})

	require.NoError(t, err)
	assert.True(t, res.Enabled)
	assert.Equal(t, rpcevaluation.EvaluationReason_MATCH_EVALUATION_REASON, res.Reason)
	assert.Equal(t, flagKey, res.FlagKey)
}

func TestBoolean_RulesOutOfOrder(t *testing.T) {
	var (
		flagKey      = "test-flag"
		namespaceKey = "test-namespace"
		envStore     = NewMockEnvironmentStore(t)
		environment  = environments.NewMockEnvironment(t)
		store        = storage.NewMockReadOnlyStore(t)
		logger       = zaptest.NewLogger(t)
		s            = New(logger, envStore)
	)

	envStore.On("GetFromContext", mock.Anything).Return(environment)
	environment.On("EvaluationStore").Return(store, nil)

	store.On("GetFlag", mock.Anything, storage.NewResource(namespaceKey, flagKey)).Return(&core.Flag{
		Key:     flagKey,
		Enabled: true,
		Type:    core.FlagType_BOOLEAN_FLAG_TYPE,
	}, nil)

	store.On("GetEvaluationRollouts", mock.Anything, storage.NewResource(namespaceKey, flagKey)).Return([]*storage.EvaluationRollout{
		{
			NamespaceKey: namespaceKey,
			Rank:         1,
			RolloutType:  core.RolloutType_THRESHOLD_ROLLOUT_TYPE,
			Threshold: &storage.RolloutThreshold{
				Percentage: 5,
				Value:      false,
			},
		},
		{
			NamespaceKey: namespaceKey,
			RolloutType:  core.RolloutType_SEGMENT_ROLLOUT_TYPE,
			Rank:         0,
			Segment: &storage.RolloutSegment{
				Value:           true,
				SegmentOperator: core.SegmentOperator_OR_SEGMENT_OPERATOR,
				Segments: map[string]*storage.EvaluationSegment{
					"test-segment": {
						SegmentKey: "test-segment",
						MatchType:  core.MatchType_ANY_MATCH_TYPE,
						Constraints: []storage.EvaluationConstraint{
							{
								Type:     core.ComparisonType_STRING_COMPARISON_TYPE,
								Property: "hello",
								Operator: flipt.OpEQ,
								Value:    "world",
							},
						},
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

func TestBatch_UnknownFlagType(t *testing.T) {
	var (
		flagKey      = "test-flag"
		namespaceKey = "test-namespace"
		envStore     = NewMockEnvironmentStore(t)
		environment  = environments.NewMockEnvironment(t)
		store        = storage.NewMockReadOnlyStore(t)
		logger       = zaptest.NewLogger(t)
		s            = New(logger, envStore)
	)

	envStore.On("GetFromContext", mock.Anything).Return(environment)
	environment.On("EvaluationStore").Return(store, nil)

	store.On("GetFlag", mock.Anything, storage.NewResource(namespaceKey, flagKey)).Return(&core.Flag{
		Key:         flagKey,
		Enabled:     true,
		Description: "test-flag",
		Type:        3,
	}, nil)

	_, err := s.Batch(context.TODO(), &rpcevaluation.BatchEvaluationRequest{
		Requests: []*rpcevaluation.EvaluationRequest{
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

	require.Error(t, err)
	assert.EqualError(t, err, "unknown flag type: 3")
}

func TestBatch_InternalError_GetFlag(t *testing.T) {
	var (
		flagKey      = "test-flag"
		namespaceKey = "test-namespace"
		envStore     = NewMockEnvironmentStore(t)
		environment  = environments.NewMockEnvironment(t)
		store        = storage.NewMockReadOnlyStore(t)
		logger       = zaptest.NewLogger(t)
		s            = New(logger, envStore)
	)

	envStore.On("GetFromContext", mock.Anything).Return(environment)
	environment.On("EvaluationStore").Return(store, nil)

	store.On("GetFlag", mock.Anything, storage.NewResource(namespaceKey, flagKey)).Return(&core.Flag{}, errors.New("internal error"))

	_, err := s.Batch(context.TODO(), &rpcevaluation.BatchEvaluationRequest{
		Requests: []*rpcevaluation.EvaluationRequest{
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

	require.Error(t, err)
	assert.EqualError(t, err, "internal error")
}

func TestBatch_Success(t *testing.T) {
	var (
		flagKey        = "test-flag"
		anotherFlagKey = "another-test-flag"
		variantFlagKey = "variant-test-flag"
		namespaceKey   = "test-namespace"
		envStore       = NewMockEnvironmentStore(t)
		environment    = environments.NewMockEnvironment(t)
		store          = storage.NewMockReadOnlyStore(t)
		logger         = zaptest.NewLogger(t)
		s              = New(logger, envStore)
	)

	envStore.On("GetFromContext", mock.Anything).Return(environment)
	environment.On("EvaluationStore").Return(store, nil)

	store.On("GetFlag", mock.Anything, storage.NewResource(namespaceKey, flagKey)).Return(&core.Flag{
		Key:     flagKey,
		Enabled: true,
		Type:    core.FlagType_BOOLEAN_FLAG_TYPE,
	}, nil)

	store.On("GetFlag", mock.Anything, storage.NewResource(namespaceKey, anotherFlagKey)).Return(&core.Flag{}, errs.ErrNotFound("another-test-flag"))

	store.On("GetFlag", mock.Anything, storage.NewResource(namespaceKey, variantFlagKey)).Return(&core.Flag{
		Key:     variantFlagKey,
		Enabled: true,
		Type:    core.FlagType_VARIANT_FLAG_TYPE,
	}, nil)

	store.On("GetEvaluationRollouts", mock.Anything, storage.NewResource(namespaceKey, flagKey)).Return([]*storage.EvaluationRollout{
		{
			NamespaceKey: namespaceKey,
			Rank:         1,
			RolloutType:  core.RolloutType_THRESHOLD_ROLLOUT_TYPE,
			Threshold: &storage.RolloutThreshold{
				Percentage: 80,
				Value:      true,
			},
		},
	}, nil)

	store.On("GetEvaluationRules", mock.Anything, storage.NewResource(namespaceKey, variantFlagKey)).Return(
		[]*storage.EvaluationRule{
			{
				ID:      "1",
				FlagKey: variantFlagKey,
				Rank:    0,
				Segments: map[string]*storage.EvaluationSegment{
					"bar": {
						SegmentKey: "bar",
						MatchType:  core.MatchType_ALL_MATCH_TYPE,
						Constraints: []storage.EvaluationConstraint{
							{
								ID:       "2",
								Type:     core.ComparisonType_STRING_COMPARISON_TYPE,
								Property: "hello",
								Operator: flipt.OpEQ,
								Value:    "world",
							},
						},
					},
				},
			},
		}, nil)

	store.On(
		"GetEvaluationDistributions",
		mock.Anything,
		storage.NewResource(namespaceKey, variantFlagKey),
		storage.NewID("1"),
	).Return([]*storage.EvaluationDistribution{}, nil)

	res, err := s.Batch(context.TODO(), &rpcevaluation.BatchEvaluationRequest{
		Requests: []*rpcevaluation.EvaluationRequest{
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

	b, ok := res.Responses[0].Response.(*rpcevaluation.EvaluationResponse_BooleanResponse)
	assert.True(t, ok, "response should be a boolean evaluation response")
	assert.True(t, b.BooleanResponse.Enabled, "value should be true from match")
	assert.Equal(t, rpcevaluation.EvaluationReason_MATCH_EVALUATION_REASON, b.BooleanResponse.Reason)
	assert.Equal(t, rpcevaluation.EvaluationResponseType_BOOLEAN_EVALUATION_RESPONSE_TYPE, res.Responses[0].Type)
	assert.Equal(t, "1", b.BooleanResponse.RequestId)
	assert.Equal(t, flagKey, b.BooleanResponse.FlagKey)

	e, ok := res.Responses[1].Response.(*rpcevaluation.EvaluationResponse_ErrorResponse)
	assert.True(t, ok, "response should be a error evaluation response")
	assert.Equal(t, anotherFlagKey, e.ErrorResponse.FlagKey)
	assert.Equal(t, namespaceKey, e.ErrorResponse.NamespaceKey)
	assert.Equal(t, rpcevaluation.ErrorEvaluationReason_NOT_FOUND_ERROR_EVALUATION_REASON, e.ErrorResponse.Reason)
	assert.Equal(t, rpcevaluation.EvaluationResponseType_ERROR_EVALUATION_RESPONSE_TYPE, res.Responses[1].Type)

	v, ok := res.Responses[2].Response.(*rpcevaluation.EvaluationResponse_VariantResponse)
	assert.True(t, ok, "response should be a variant evaluation response")
	assert.True(t, v.VariantResponse.Match, "variant response should have matched")
	assert.Contains(t, v.VariantResponse.SegmentKeys, "bar")
	assert.Equal(t, rpcevaluation.EvaluationReason_MATCH_EVALUATION_REASON, v.VariantResponse.Reason)
	assert.Equal(t, rpcevaluation.EvaluationResponseType_VARIANT_EVALUATION_RESPONSE_TYPE, res.Responses[2].Type)
	assert.Equal(t, "3", v.VariantResponse.RequestId)
	assert.Equal(t, variantFlagKey, v.VariantResponse.FlagKey)
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
		{
			name: "is one of",
			constraint: storage.EvaluationConstraint{
				Property: "foo",
				Operator: "isoneof",
				Value:    "[\"bar\", \"baz\"]",
			},
			value:     "baz",
			wantMatch: true,
		},
		{
			name: "negative is one of",
			constraint: storage.EvaluationConstraint{
				Property: "foo",
				Operator: "isoneof",
				Value:    "[\"bar\", \"baz\"]",
			},
			value: "nope",
		},
		{
			name: "negative is one of (invalid json)",
			constraint: storage.EvaluationConstraint{
				Property: "foo",
				Operator: "isoneof",
				Value:    "[\"bar\", \"baz\"",
			},
			value: "bar",
		},
		{
			name: "negative is one of (non-string values)",
			constraint: storage.EvaluationConstraint{
				Property: "foo",
				Operator: "isoneof",
				Value:    "[\"bar\", 5]",
			},
			value: "bar",
		},
		{
			name: "is not one of",
			constraint: storage.EvaluationConstraint{
				Property: "foo",
				Operator: "isnotoneof",
				Value:    "[\"bar\", \"baz\"]",
			},
			value: "baz",
		},
		{
			name: "negative is not one of",
			constraint: storage.EvaluationConstraint{
				Property: "foo",
				Operator: "isnotoneof",
				Value:    "[\"bar\", \"baz\"]",
			},
			value:     "nope",
			wantMatch: true,
		},
		{

			name: "contains",
			constraint: storage.EvaluationConstraint{
				Property: "foo",
				Operator: "contains",
				Value:    "bar",
			},
			value:     "foobar",
			wantMatch: true,
		},
		{

			name: "negative contains",
			constraint: storage.EvaluationConstraint{
				Property: "foo",
				Operator: "contains",
				Value:    "bar",
			},
			value: "nope",
		},
		{
			name: "not contains",
			constraint: storage.EvaluationConstraint{
				Property: "foo",
				Operator: "notcontains",
				Value:    "bar",
			},
			value:     "nope",
			wantMatch: true,
		},
		{
			name: "negative not contains",
			constraint: storage.EvaluationConstraint{
				Property: "foo",
				Operator: "notcontains",
				Value:    "bar",
			},
			value: "foobar",
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
		{
			name: "is one of",
			constraint: storage.EvaluationConstraint{
				Property: "foo",
				Operator: "isoneof",
				Value:    "[3, 3.14159, 4]",
			},
			value:     "3.14159",
			wantMatch: true,
		},
		{
			name: "negative is one of",
			constraint: storage.EvaluationConstraint{
				Property: "foo",
				Operator: "isoneof",
				Value:    "[5, 3.14159, 4]",
			},
			value: "9",
		},
		{
			name: "negative is one of (non-number values)",
			constraint: storage.EvaluationConstraint{
				Property: "foo",
				Operator: "isoneof",
				Value:    "[5, \"str\"]",
			},
			value:   "5",
			wantErr: true,
		},
		{
			name: "is not one of",
			constraint: storage.EvaluationConstraint{
				Property: "foo",
				Operator: "isnotoneof",
				Value:    "[5, 3.14159, 4]",
			},
			value:     "3",
			wantMatch: true,
		},
		{
			name: "negative is not one of",
			constraint: storage.EvaluationConstraint{
				Property: "foo",
				Operator: "isnotoneof",
				Value:    "[5, 3.14159, 4]",
			},
			value:     "3.14159",
			wantMatch: false,
		},
		{
			name: "negative is not one of (invalid json)",
			constraint: storage.EvaluationConstraint{
				Property: "foo",
				Operator: "isnotoneof",
				Value:    "[5, 6",
			},
			value:   "5",
			wantErr: true,
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
				var ierr errs.ErrInvalid
				assert.ErrorAs(t, err, &ierr)
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
				var ierr errs.ErrInvalid
				assert.ErrorAs(t, err, &ierr)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, wantMatch, match)
		})
	}
}

func Test_matchesDateTime(t *testing.T) {
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
			value:     "2006-01-02T15:04:05Z",
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
			value: "2006-01-02T15:04:05Z",
		},
		{
			name: "not a datetime constraint value",
			constraint: storage.EvaluationConstraint{
				Property: "foo",
				Operator: "eq",
				Value:    "bar",
			},
			value:   "2006-01-02T15:04:05Z",
			wantErr: true,
		},
		{
			name: "not a datetime context value",
			constraint: storage.EvaluationConstraint{
				Property: "foo",
				Operator: "eq",
				Value:    "2006-01-02T15:04:05Z",
			},
			value:   "foo",
			wantErr: true,
		},
		{
			name: "eq",
			constraint: storage.EvaluationConstraint{
				Property: "foo",
				Operator: "eq",
				Value:    "2006-01-02T15:04:05Z",
			},
			value:     "2006-01-02T15:04:05Z",
			wantMatch: true,
		},
		{
			name: "eq date only",
			constraint: storage.EvaluationConstraint{
				Property: "foo",
				Operator: "eq",
				Value:    "2006-01-02",
			},
			value:     "2006-01-02",
			wantMatch: true,
		},
		{
			name: "negative eq",
			constraint: storage.EvaluationConstraint{
				Property: "foo",
				Operator: "eq",
				Value:    "2006-01-02T15:04:05Z",
			},
			value: "2007-01-02T15:04:05Z",
		},
		{
			name: "neq",
			constraint: storage.EvaluationConstraint{
				Property: "foo",
				Operator: "neq",
				Value:    "2006-01-02T15:04:05Z",
			},
			value:     "2007-01-02T15:04:05Z",
			wantMatch: true,
		},
		{
			name: "negative neq",
			constraint: storage.EvaluationConstraint{
				Property: "foo",
				Operator: "neq",
				Value:    "2006-01-02T15:04:05Z",
			},
			value: "2006-01-02T15:04:05Z",
		},
		{
			name: "negative neq date only",
			constraint: storage.EvaluationConstraint{
				Property: "foo",
				Operator: "neq",
				Value:    "2006-01-02",
			},
			value: "2006-01-02",
		},
		{
			name: "lt",
			constraint: storage.EvaluationConstraint{
				Property: "foo",
				Operator: "lt",
				Value:    "2006-01-02T15:04:05Z",
			},
			value:     "2005-01-02T15:04:05Z",
			wantMatch: true,
		},
		{
			name: "negative lt",
			constraint: storage.EvaluationConstraint{
				Property: "foo",
				Operator: "lt",
				Value:    "2005-01-02T15:04:05Z",
			},
			value: "2006-01-02T15:04:05Z",
		},
		{
			name: "lte",
			constraint: storage.EvaluationConstraint{
				Property: "foo",
				Operator: "lte",
				Value:    "2006-01-02T15:04:05Z",
			},
			value:     "2006-01-02T15:04:05Z",
			wantMatch: true,
		},
		{
			name: "negative lte",
			constraint: storage.EvaluationConstraint{
				Property: "foo",
				Operator: "lte",
				Value:    "2006-01-02T15:04:05Z",
			},
			value: "2007-01-02T15:04:05Z",
		},
		{
			name: "gt",
			constraint: storage.EvaluationConstraint{
				Property: "foo",
				Operator: "gt",
				Value:    "2006-01-02T15:04:05Z",
			},
			value:     "2007-01-02T15:04:05Z",
			wantMatch: true,
		},
		{
			name: "negative gt",
			constraint: storage.EvaluationConstraint{
				Property: "foo",
				Operator: "gt",
				Value:    "2007-01-02T15:04:05Z",
			},
			value: "2006-01-02T15:04:05Z",
		},
		{
			name: "gte",
			constraint: storage.EvaluationConstraint{
				Property: "foo",
				Operator: "gte",
				Value:    "2006-01-02T15:04:05Z",
			},
			value:     "2006-01-02T15:04:05Z",
			wantMatch: true,
		},
		{
			name: "negative gte",
			constraint: storage.EvaluationConstraint{
				Property: "foo",
				Operator: "gte",
				Value:    "2006-01-02T15:04:05Z",
			},
			value: "2005-01-02T15:04:05Z",
		},
		{
			name: "empty value",
			constraint: storage.EvaluationConstraint{
				Property: "foo",
				Operator: "eq",
				Value:    "2006-01-02T15:04:05Z",
			},
		},
		{
			name: "unknown operator",
			constraint: storage.EvaluationConstraint{
				Property: "foo",
				Operator: "foo",
				Value:    "2006-01-02T15:04:05Z",
			},
			value: "2006-01-02T15:04:05Z",
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
			match, err := matchesDateTime(constraint, value)

			if wantErr {
				require.Error(t, err)
				var ierr errs.ErrInvalid
				assert.ErrorAs(t, err, &ierr)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, wantMatch, match)
		})
	}
}
