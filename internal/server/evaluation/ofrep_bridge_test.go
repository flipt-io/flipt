package evaluation

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.flipt.io/flipt/internal/server/ofrep"
	"go.flipt.io/flipt/internal/storage"
	"go.flipt.io/flipt/rpc/flipt/core"
	rpcevaluation "go.flipt.io/flipt/rpc/flipt/evaluation"
	"go.uber.org/zap/zaptest"
)

func TestOFREPFlagEvaluation_Variant(t *testing.T) {
	var (
		flagKey      = "test-flag"
		namespaceKey = "test-namespace"
		store        = &evaluationStoreMock{}
		logger       = zaptest.NewLogger(t)
		s            = New(logger, store)
		flag         = &core.Flag{
			Key:     flagKey,
			Enabled: true,
			Type:    core.FlagType_VARIANT_FLAG_TYPE,
		}
	)

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

	store.On("GetEvaluationDistributions", mock.Anything, storage.NewID("1")).Return([]*storage.EvaluationDistribution{
		{
			ID:         "4",
			RuleID:     "1",
			VariantID:  "5",
			Rollout:    100,
			VariantKey: "boz",
		},
	}, nil)

	output, err := s.OFREPFlagEvaluation(context.TODO(), ofrep.EvaluationBridgeInput{
		FlagKey:      flagKey,
		NamespaceKey: namespaceKey,
		Context: map[string]string{
			"hello":        "world",
			"targetingKey": "12345",
		},
	})
	require.NoError(t, err)

	assert.Equal(t, flagKey, output.FlagKey)
	assert.Equal(t, rpcevaluation.EvaluationReason_MATCH_EVALUATION_REASON, output.Reason)
	assert.Equal(t, "boz", output.Variant)
	assert.Equal(t, "boz", output.Value)
}

func TestOFREPFlagEvaluation_Boolean(t *testing.T) {
	var (
		flagKey      = "test-flag"
		namespaceKey = "test-namespace"
		store        = &evaluationStoreMock{}
		logger       = zaptest.NewLogger(t)
		s            = New(logger, store)
		flag         = &core.Flag{
			Key:     flagKey,
			Enabled: true,
			Type:    core.FlagType_BOOLEAN_FLAG_TYPE,
		}
	)

	store.On("GetFlag", mock.Anything, storage.NewResource(namespaceKey, flagKey)).Return(flag, nil)

	store.On("GetEvaluationRollouts", mock.Anything, storage.NewResource(namespaceKey, flagKey)).Return([]*storage.EvaluationRollout{}, nil)

	output, err := s.OFREPFlagEvaluation(context.TODO(), ofrep.EvaluationBridgeInput{
		FlagKey:      flagKey,
		NamespaceKey: namespaceKey,
		Context: map[string]string{
			"targetingKey": "12345",
		},
	})
	require.NoError(t, err)

	assert.Equal(t, flagKey, output.FlagKey)
	assert.Equal(t, rpcevaluation.EvaluationReason_DEFAULT_EVALUATION_REASON, output.Reason)
	assert.Equal(t, "true", output.Variant)
	assert.Equal(t, true, output.Value)
}
