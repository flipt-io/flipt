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
	"go.flipt.io/flipt/internal/server/tracing"
	"go.flipt.io/flipt/internal/storage"
	"go.flipt.io/flipt/rpc/flipt"
	"go.flipt.io/flipt/rpc/flipt/core"
	rpcevaluation "go.flipt.io/flipt/rpc/flipt/evaluation"
	"go.opentelemetry.io/otel/attribute"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
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

	envStore.On("Get", mock.Anything, mock.Anything).Return(environment, nil)
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

	envStore.On("Get", mock.Anything, mock.Anything).Return(environment, nil)
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
		flagKey        = "test-flag"
		environmentKey = "test-environment"
		namespaceKey   = "test-namespace"
		envStore       = NewMockEnvironmentStore(t)
		environment    = environments.NewMockEnvironment(t)
		store          = storage.NewMockReadOnlyStore(t)
		logger         = zaptest.NewLogger(t)
		s              = New(logger, envStore)
	)

	environment.On("Key").Return(environmentKey)
	envStore.On("Get", mock.Anything, mock.Anything).Return(environment, nil)
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
		flagKey        = "test-flag"
		environmentKey = "test-environment"
		namespaceKey   = "test-namespace"
		envStore       = NewMockEnvironmentStore(t)
		environment    = environments.NewMockEnvironment(t)
		store          = storage.NewMockReadOnlyStore(t)
		logger         = zaptest.NewLogger(t)
		s              = New(logger, envStore)
		flag           = &core.Flag{
			Key:     flagKey,
			Enabled: true,
			Type:    core.FlagType_VARIANT_FLAG_TYPE,
		}
	)

	environment.On("Key").Return(environmentKey)
	envStore.On("Get", mock.Anything, mock.Anything).Return(environment, nil)
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
		flagKey        = "test-flag"
		environmentKey = "test-environment"
		namespaceKey   = "test-namespace"
		envStore       = NewMockEnvironmentStore(t)
		environment    = environments.NewMockEnvironment(t)
		store          = storage.NewMockReadOnlyStore(t)
		logger         = zaptest.NewLogger(t)
		s              = New(logger, envStore)
		flag           = &core.Flag{
			Key:     flagKey,
			Enabled: true,
			Type:    core.FlagType_VARIANT_FLAG_TYPE,
		}
	)

	environment.On("Key").Return(environmentKey)
	envStore.On("Get", mock.Anything, mock.Anything).Return(environment, nil)
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

	envStore.On("Get", mock.Anything, mock.Anything).Return(environment, nil)
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

	envStore.On("Get", mock.Anything, mock.Anything).Return(environment, nil)
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
		flagKey        = "test-flag"
		environmentKey = "test-environment"
		namespaceKey   = "test-namespace"
		envStore       = NewMockEnvironmentStore(t)
		environment    = environments.NewMockEnvironment(t)
		store          = storage.NewMockReadOnlyStore(t)
		logger         = zaptest.NewLogger(t)
		s              = New(logger, envStore)
	)

	environment.On("Key").Return(environmentKey)
	envStore.On("Get", mock.Anything, mock.Anything).Return(environment, nil)
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
		flagKey        = "test-flag"
		environmentKey = "test-environment"
		namespaceKey   = "test-namespace"
		envStore       = NewMockEnvironmentStore(t)
		environment    = environments.NewMockEnvironment(t)
		store          = storage.NewMockReadOnlyStore(t)
		logger         = zaptest.NewLogger(t)
		s              = New(logger, envStore)
	)

	environment.On("Key").Return(environmentKey)
	envStore.On("Get", mock.Anything, mock.Anything).Return(environment, nil)
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
		flagKey        = "test-flag"
		environmentKey = "test-environment"
		namespaceKey   = "test-namespace"
		envStore       = NewMockEnvironmentStore(t)
		environment    = environments.NewMockEnvironment(t)
		store          = storage.NewMockReadOnlyStore(t)
		logger         = zaptest.NewLogger(t)
		s              = New(logger, envStore)
	)

	environment.On("Key").Return(environmentKey)
	envStore.On("Get", mock.Anything, mock.Anything).Return(environment, nil)
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
		flagKey        = "test-flag"
		environmentKey = "test-environment"
		namespaceKey   = "test-namespace"
		envStore       = NewMockEnvironmentStore(t)
		environment    = environments.NewMockEnvironment(t)
		store          = storage.NewMockReadOnlyStore(t)
		logger         = zaptest.NewLogger(t)
		s              = New(logger, envStore)
	)

	environment.On("Key").Return(environmentKey)
	envStore.On("Get", mock.Anything, mock.Anything).Return(environment, nil)
	environment.On("EvaluationStore").Return(store, nil)

	store.On("GetFlag", mock.Anything, storage.NewResource(namespaceKey, flagKey)).Return(&core.Flag{
		Key:     flagKey,
		Enabled: true,
		Type:    core.FlagType_BOOLEAN_FLAG_TYPE,
	}, nil)

	store.On("GetEvaluationRollouts", mock.Anything, storage.NewResource(namespaceKey, flagKey)).Return([]*storage.EvaluationRollout{
		{
			NamespaceKey: namespaceKey,
			RolloutType:  core.RolloutType_THRESHOLD_ROLLOUT_TYPE,
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
	assert.Equal(t, []string{"test-segment"}, res.SegmentKeys)
}

func TestBoolean_SegmentMatch_MultipleConstraints(t *testing.T) {
	var (
		flagKey        = "test-flag"
		environmentKey = "test-environment"
		namespaceKey   = "test-namespace"
		envStore       = NewMockEnvironmentStore(t)
		environment    = environments.NewMockEnvironment(t)
		store          = storage.NewMockReadOnlyStore(t)
		logger         = zaptest.NewLogger(t)
		s              = New(logger, envStore)
	)

	environment.On("Key").Return(environmentKey)
	envStore.On("Get", mock.Anything, mock.Anything).Return(environment, nil)
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
		flagKey        = "test-flag"
		environmentKey = "test-environment"
		namespaceKey   = "test-namespace"
		envStore       = NewMockEnvironmentStore(t)
		environment    = environments.NewMockEnvironment(t)
		store          = storage.NewMockReadOnlyStore(t)
		logger         = zaptest.NewLogger(t)
		s              = New(logger, envStore)
	)

	environment.On("Key").Return(environmentKey)
	envStore.On("Get", mock.Anything, mock.Anything).Return(environment, nil)
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
		flagKey        = "test-flag"
		environmentKey = "test-environment"
		namespaceKey   = "test-namespace"
		envStore       = NewMockEnvironmentStore(t)
		environment    = environments.NewMockEnvironment(t)
		store          = storage.NewMockReadOnlyStore(t)
		logger         = zaptest.NewLogger(t)
		s              = New(logger, envStore)
	)

	environment.On("Key").Return(environmentKey)
	envStore.On("Get", mock.Anything, mock.Anything).Return(environment, nil)
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
		flagKey        = "test-flag"
		environmentKey = "test-environment"
		namespaceKey   = "test-namespace"
		envStore       = NewMockEnvironmentStore(t)
		environment    = environments.NewMockEnvironment(t)
		store          = storage.NewMockReadOnlyStore(t)
		logger         = zaptest.NewLogger(t)
		s              = New(logger, envStore)
	)

	environment.On("Key").Return(environmentKey)
	envStore.On("Get", mock.Anything, mock.Anything).Return(environment, nil)
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

	envStore.On("Get", mock.Anything, mock.Anything).Return(environment, nil)
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

	envStore.On("Get", mock.Anything, mock.Anything).Return(environment, nil)
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
		environmentKey = "test-environment"
		anotherFlagKey = "another-test-flag"
		variantFlagKey = "variant-test-flag"
		namespaceKey   = "test-namespace"
		envStore       = NewMockEnvironmentStore(t)
		environment    = environments.NewMockEnvironment(t)
		store          = storage.NewMockReadOnlyStore(t)
		logger         = zaptest.NewLogger(t)
		s              = New(logger, envStore)
	)

	environment.On("Key").Return(environmentKey)
	envStore.On("Get", mock.Anything, mock.Anything).Return(environment, nil)
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

func TestEvaluationWithTracing(t *testing.T) {
	tests := []struct {
		name               string
		tracingEnabled     bool
		expectedEventCount int
		setupMocks         func(t *testing.T, envStore *MockEnvironmentStore, environment *environments.MockEnvironment, store *storage.MockReadOnlyStore, namespaceKey string)
		runEvaluation      func(ctx context.Context, s *Server) error
		validateEvents     func(t *testing.T, events []sdktrace.Event, environmentKey, namespaceKey string)
	}{
		{
			name:               "variant_flag_with_tracing",
			tracingEnabled:     true,
			expectedEventCount: 1,
			setupMocks: func(t *testing.T, envStore *MockEnvironmentStore, environment *environments.MockEnvironment, store *storage.MockReadOnlyStore, namespaceKey string) {
				flagKey := "test-flag"
				environment.On("Key").Return("test-environment")
				envStore.On("Get", mock.Anything, mock.Anything).Return(environment, nil)
				environment.On("EvaluationStore").Return(store, nil)

				store.On("GetFlag", mock.Anything, storage.NewResource(namespaceKey, flagKey)).Return(&core.Flag{
					Key:     flagKey,
					Enabled: true,
					Type:    core.FlagType_VARIANT_FLAG_TYPE,
				}, nil)

				store.On("GetEvaluationRules", mock.Anything, storage.NewResource(namespaceKey, flagKey)).Return(
					[]*storage.EvaluationRule{
						{
							ID:      "1",
							FlagKey: flagKey,
							Rank:    0,
							Segments: map[string]*storage.EvaluationSegment{
								"segment-1": {
									SegmentKey: "segment-1",
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
				).Return([]*storage.EvaluationDistribution{
					{
						RuleID:     "1",
						VariantID:  "variant-1",
						VariantKey: "test-variant",
						Rollout:    100,
					},
				}, nil)
			},
			runEvaluation: func(ctx context.Context, s *Server) error {
				_, err := s.Variant(ctx, &rpcevaluation.EvaluationRequest{
					FlagKey:        "test-flag",
					EnvironmentKey: "test-environment",
					EntityId:       "test-entity",
					RequestId:      "test-request-id",
					NamespaceKey:   "test-namespace",
					Context: map[string]string{
						"hello": "world",
					},
				})
				return err
			},
			validateEvents: func(t *testing.T, events []sdktrace.Event, environmentKey, namespaceKey string) {
				require.Len(t, events, 1)
				event := events[0]
				assert.Equal(t, tracing.Event, event.Name)
				attrs := event.Attributes
				assertAttributeValue(t, attrs, tracing.AttributeEnvironment, environmentKey)
				assertAttributeValue(t, attrs, tracing.AttributeNamespace, namespaceKey)
				assertAttributeValue(t, attrs, tracing.AttributeFlag, "test-flag")
				assertAttributeValue(t, attrs, tracing.AttributeEntityID, "test-entity")
				assertAttributeValue(t, attrs, tracing.AttributeRequestID, "test-request-id")
				assertAttributeValue(t, attrs, tracing.AttributeMatch, true)
				assertAttributeValue(t, attrs, tracing.AttributeVariant, "test-variant")
				assertAttributeValue(t, attrs, tracing.AttributeReason, "match")
				assertAttributeValue(t, attrs, tracing.AttributeFlagType, "variant")
			},
		},
		{
			name:               "boolean_flag_with_tracing",
			tracingEnabled:     true,
			expectedEventCount: 1,
			setupMocks: func(t *testing.T, envStore *MockEnvironmentStore, environment *environments.MockEnvironment, store *storage.MockReadOnlyStore, namespaceKey string) {
				flagKey := "test-flag"
				environment.On("Key").Return("test-environment")
				envStore.On("Get", mock.Anything, mock.Anything).Return(environment, nil)
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
							Percentage: 100,
							Value:      true,
						},
					},
				}, nil)
			},
			runEvaluation: func(ctx context.Context, s *Server) error {
				_, err := s.Boolean(ctx, &rpcevaluation.EvaluationRequest{
					FlagKey:        "test-flag",
					EnvironmentKey: "test-environment",
					EntityId:       "test-entity",
					RequestId:      "test-request-id",
					NamespaceKey:   "test-namespace",
					Context: map[string]string{
						"hello": "world",
					},
				})
				return err
			},
			validateEvents: func(t *testing.T, events []sdktrace.Event, environmentKey, namespaceKey string) {
				require.Len(t, events, 1)
				event := events[0]
				assert.Equal(t, tracing.Event, event.Name)
				attrs := event.Attributes
				assertAttributeValue(t, attrs, tracing.AttributeEnvironment, environmentKey)
				assertAttributeValue(t, attrs, tracing.AttributeNamespace, namespaceKey)
				assertAttributeValue(t, attrs, tracing.AttributeFlag, "test-flag")
				assertAttributeValue(t, attrs, tracing.AttributeEntityID, "test-entity")
				assertAttributeValue(t, attrs, tracing.AttributeRequestID, "test-request-id")
				assertAttributeValue(t, attrs, tracing.AttributeVariant, true)
				assertAttributeValue(t, attrs, tracing.AttributeReason, "match")
				assertAttributeValue(t, attrs, tracing.AttributeFlagType, "boolean")
			},
		},
		{
			name:               "batch_evaluation_with_tracing",
			tracingEnabled:     true,
			expectedEventCount: 2,
			setupMocks: func(t *testing.T, envStore *MockEnvironmentStore, environment *environments.MockEnvironment, store *storage.MockReadOnlyStore, namespaceKey string) {
				booleanFlagKey := "boolean-flag"
				variantFlagKey := "variant-flag"

				environment.On("Key").Return("test-environment")
				envStore.On("Get", mock.Anything, mock.Anything).Return(environment, nil)
				environment.On("EvaluationStore").Return(store, nil)

				// Setup boolean flag
				store.On("GetFlag", mock.Anything, storage.NewResource(namespaceKey, booleanFlagKey)).Return(&core.Flag{
					Key:     booleanFlagKey,
					Enabled: true,
					Type:    core.FlagType_BOOLEAN_FLAG_TYPE,
				}, nil)

				store.On("GetEvaluationRollouts", mock.Anything, storage.NewResource(namespaceKey, booleanFlagKey)).Return([]*storage.EvaluationRollout{
					{
						NamespaceKey: namespaceKey,
						Rank:         1,
						RolloutType:  core.RolloutType_THRESHOLD_ROLLOUT_TYPE,
						Threshold: &storage.RolloutThreshold{
							Percentage: 100,
							Value:      true,
						},
					},
				}, nil)

				// Setup variant flag
				store.On("GetFlag", mock.Anything, storage.NewResource(namespaceKey, variantFlagKey)).Return(&core.Flag{
					Key:     variantFlagKey,
					Enabled: true,
					Type:    core.FlagType_VARIANT_FLAG_TYPE,
				}, nil)

				store.On("GetEvaluationRules", mock.Anything, storage.NewResource(namespaceKey, variantFlagKey)).Return(
					[]*storage.EvaluationRule{
						{
							ID:      "1",
							FlagKey: variantFlagKey,
							Rank:    0,
							Segments: map[string]*storage.EvaluationSegment{
								"segment-1": {
									SegmentKey: "segment-1",
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
				).Return([]*storage.EvaluationDistribution{
					{
						RuleID:     "1",
						VariantID:  "variant-1",
						VariantKey: "test-variant",
						Rollout:    100,
					},
				}, nil)
			},
			runEvaluation: func(ctx context.Context, s *Server) error {
				_, err := s.Batch(ctx, &rpcevaluation.BatchEvaluationRequest{
					Requests: []*rpcevaluation.EvaluationRequest{
						{
							RequestId:      "request-1",
							FlagKey:        "boolean-flag",
							EntityId:       "test-entity",
							NamespaceKey:   "test-namespace",
							EnvironmentKey: "test-environment",
							Context: map[string]string{
								"hello": "world",
							},
						},
						{
							RequestId:      "request-2",
							FlagKey:        "variant-flag",
							EntityId:       "test-entity",
							NamespaceKey:   "test-namespace",
							EnvironmentKey: "test-environment",
							Context: map[string]string{
								"hello": "world",
							},
						},
					},
				})
				return err
			},
			validateEvents: func(t *testing.T, events []sdktrace.Event, environmentKey, namespaceKey string) {
				require.Len(t, events, 2, "expected two events (one per flag)")

				// Verify boolean flag event
				booleanEvent := events[0]
				assert.Equal(t, tracing.Event, booleanEvent.Name)
				booleanAttrs := booleanEvent.Attributes
				assertAttributeValue(t, booleanAttrs, tracing.AttributeEnvironment, environmentKey)
				assertAttributeValue(t, booleanAttrs, tracing.AttributeNamespace, namespaceKey)
				assertAttributeValue(t, booleanAttrs, tracing.AttributeFlag, "boolean-flag")
				assertAttributeValue(t, booleanAttrs, tracing.AttributeEntityID, "test-entity")
				assertAttributeValue(t, booleanAttrs, tracing.AttributeRequestID, "request-1")
				assertAttributeValue(t, booleanAttrs, tracing.AttributeVariant, true)
				assertAttributeValue(t, booleanAttrs, tracing.AttributeFlagType, "boolean")

				// Verify variant flag event
				variantEvent := events[1]
				assert.Equal(t, tracing.Event, variantEvent.Name)
				variantAttrs := variantEvent.Attributes
				assertAttributeValue(t, variantAttrs, tracing.AttributeEnvironment, environmentKey)
				assertAttributeValue(t, variantAttrs, tracing.AttributeNamespace, namespaceKey)
				assertAttributeValue(t, variantAttrs, tracing.AttributeFlag, "variant-flag")
				assertAttributeValue(t, variantAttrs, tracing.AttributeEntityID, "test-entity")
				assertAttributeValue(t, variantAttrs, tracing.AttributeRequestID, "request-2")
				assertAttributeValue(t, variantAttrs, tracing.AttributeMatch, true)
				assertAttributeValue(t, variantAttrs, tracing.AttributeVariant, "test-variant")
				assertAttributeValue(t, variantAttrs, tracing.AttributeFlagType, "variant")
			},
		},
		{
			name:               "tracing_disabled",
			tracingEnabled:     false,
			expectedEventCount: 0,
			setupMocks: func(t *testing.T, envStore *MockEnvironmentStore, environment *environments.MockEnvironment, store *storage.MockReadOnlyStore, namespaceKey string) {
				flagKey := "test-flag"
				environment.On("Key").Return("test-environment")
				envStore.On("Get", mock.Anything, mock.Anything).Return(environment, nil)
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
							Percentage: 100,
							Value:      true,
						},
					},
				}, nil)
			},
			runEvaluation: func(ctx context.Context, s *Server) error {
				_, err := s.Boolean(ctx, &rpcevaluation.EvaluationRequest{
					FlagKey:        "test-flag",
					EnvironmentKey: "test-environment",
					EntityId:       "test-entity",
					RequestId:      "test-request-id",
					NamespaceKey:   "test-namespace",
					Context: map[string]string{
						"hello": "world",
					},
				})
				return err
			},
			validateEvents: func(t *testing.T, events []sdktrace.Event, environmentKey, namespaceKey string) {
				assert.Empty(t, events, "expected no events when tracing is disabled")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var (
				environmentKey = "test-environment"
				namespaceKey   = "test-namespace"
				envStore       = NewMockEnvironmentStore(t)
				environment    = environments.NewMockEnvironment(t)
				store          = storage.NewMockReadOnlyStore(t)
				logger         = zaptest.NewLogger(t)
			)

			spanRecorder := tracetest.NewSpanRecorder()
			tp := sdktrace.NewTracerProvider(
				sdktrace.WithSpanProcessor(spanRecorder),
				sdktrace.WithSampler(sdktrace.AlwaysSample()),
			)
			tracer := tp.Tracer("test")

			s := New(logger, envStore, WithTracing(tt.tracingEnabled))

			tt.setupMocks(t, envStore, environment, store, namespaceKey)

			ctx, span := tracer.Start(context.Background(), "test-span")
			err := tt.runEvaluation(ctx, s)
			span.End()
			require.NoError(t, err)

			spans := spanRecorder.Ended()
			require.Len(t, spans, 1, "expected one span to be recorded")

			events := spans[0].Events()
			require.Len(t, events, tt.expectedEventCount)

			if tt.expectedEventCount > 0 {
				tt.validateEvents(t, events, environmentKey, namespaceKey)
			}
		})
	}
}

// Helper function to assert attribute values
func assertAttributeValue(t *testing.T, attrs []attribute.KeyValue, key attribute.Key, expectedValue interface{}) {
	t.Helper()
	found := false
	for _, attr := range attrs {
		if attr.Key == key {
			found = true
			switch v := expectedValue.(type) {
			case string:
				assert.Equal(t, v, attr.Value.AsString(), "attribute %s value mismatch", key)
			case bool:
				assert.Equal(t, v, attr.Value.AsBool(), "attribute %s value mismatch", key)
			case []string:
				actualSlice := attr.Value.AsStringSlice()
				assert.ElementsMatch(t, v, actualSlice, "attribute %s value mismatch", key)
			default:
				t.Fatalf("unsupported type for attribute %s: %T", key, v)
			}
			break
		}
	}
	assert.True(t, found, "attribute %s not found in event", key)
}

// Helper function to assert attribute is not present
func assertAttributeNotPresent(t *testing.T, attrs []attribute.KeyValue, key attribute.Key) {
	t.Helper()
	for _, attr := range attrs {
		if attr.Key == key {
			assert.Failf(t, "unexpected attribute", "attribute %s should not be present in event", string(key))
			return
		}
	}
}
