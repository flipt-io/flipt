package evaluation

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.flipt.io/flipt/internal/server/environments"
	"go.flipt.io/flipt/internal/storage"
	"go.flipt.io/flipt/rpc/flipt/core"
	"go.flipt.io/flipt/rpc/flipt/ofrep"
	"go.uber.org/zap/zaptest"
	"google.golang.org/grpc/metadata"
)

func TestOFREPFlagEvaluation_Variant(t *testing.T) {
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

	envStore.On("GetFromContext", mock.Anything).Return(environment)
	environment.On("EvaluationStore").Return(store, nil)

	store.On("GetFlag", mock.Anything, storage.NewResource(namespaceKey, flagKey)).Return(flag, nil)

	store.On("GetEvaluationRules", mock.Anything, storage.NewResource(namespaceKey, flagKey)).Return([]*storage.EvaluationRule{
		{
			ID:      "1",
			FlagKey: flagKey,
			Rank:    0,
			Segments: map[string]*storage.EvaluationSegment{
				"bar": {
					SegmentKey: "bar",
					MatchType:  core.MatchType_ANY_MATCH_TYPE,
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
			ID:         "4",
			RuleID:     "1",
			VariantID:  "5",
			Rollout:    100,
			VariantKey: "boz",
		},
	}, nil)

	ctx := metadata.NewIncomingContext(context.TODO(), metadata.New(map[string]string{
		"x-flipt-environment": environmentKey,
		"x-flipt-namespace":   namespaceKey,
	}))

	output, err := s.OFREPFlagEvaluation(ctx, &ofrep.EvaluateFlagRequest{
		Key: flagKey,
		Context: map[string]string{
			"hello":        "world",
			"targetingKey": "12345",
		},
	})
	require.NoError(t, err)

	assert.Equal(t, flagKey, output.Key)
	assert.Equal(t, ofrep.EvaluateReason_TARGETING_MATCH, output.Reason)
	assert.Equal(t, "boz", output.Variant)
	assert.Equal(t, "boz", output.Value.GetStringValue())
}

func TestOFREPFlagEvaluation_Boolean(t *testing.T) {
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
			Type:    core.FlagType_BOOLEAN_FLAG_TYPE,
		}
	)

	environment.On("Key").Return(environmentKey)

	envStore.On("GetFromContext", mock.Anything).Return(environment)
	environment.On("EvaluationStore").Return(store, nil)

	store.On("GetFlag", mock.Anything, storage.NewResource(namespaceKey, flagKey)).Return(flag, nil)

	store.On("GetEvaluationRollouts", mock.Anything, storage.NewResource(namespaceKey, flagKey)).Return([]*storage.EvaluationRollout{}, nil)

	ctx := metadata.NewIncomingContext(context.TODO(), metadata.New(map[string]string{
		"x-flipt-environment": environmentKey,
		"x-flipt-namespace":   namespaceKey,
	}))

	output, err := s.OFREPFlagEvaluation(ctx, &ofrep.EvaluateFlagRequest{
		Key: flagKey,
		Context: map[string]string{
			"targetingKey": "12345",
		},
	})
	require.NoError(t, err)

	assert.Equal(t, flagKey, output.Key)
	assert.Equal(t, ofrep.EvaluateReason_DEFAULT, output.Reason)
	assert.Equal(t, "true", output.Variant)
	assert.True(t, output.Value.GetBoolValue())
}
