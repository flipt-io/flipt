package evaluation

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	errs "go.flipt.io/flipt/errors"
	"go.flipt.io/flipt/rpc/flipt"
	rpcEvaluation "go.flipt.io/flipt/rpc/flipt/evaluation"
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

	v, err := s.Variant(context.TODO(), &rpcEvaluation.EvaluationRequest{
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

	v, err := s.Variant(context.TODO(), &rpcEvaluation.EvaluationRequest{
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

func TestVariant_EvaluateFailure(t *testing.T) {
	var (
		flagKey      = "test-flag"
		namespaceKey = "test-namespace"
		store        = &evaluationStoreMock{}
		evaluator    = &evaluatorMock{}
		logger       = zaptest.NewLogger(t)
		s            = &Server{
			logger:    logger,
			store:     store,
			evaluator: evaluator,
		}
		flag = &flipt.Flag{
			NamespaceKey: namespaceKey,
			Key:          flagKey,
			Enabled:      true,
			Type:         flipt.FlagType_VARIANT_FLAG_TYPE,
		}
	)

	store.On("GetFlag", mock.Anything, namespaceKey, flagKey).Return(flag, nil)

	evaluator.On("Evaluate", mock.Anything, flag, &flipt.EvaluationRequest{
		FlagKey:      flagKey,
		NamespaceKey: namespaceKey,
		EntityId:     "test-entity",
		Context: map[string]string{
			"hello": "world",
		},
	}).Return(&flipt.EvaluationResponse{}, errs.ErrInvalid("some error"))

	v, err := s.Variant(context.TODO(), &rpcEvaluation.EvaluationRequest{
		FlagKey:      flagKey,
		EntityId:     "test-entity",
		NamespaceKey: namespaceKey,
		Context: map[string]string{
			"hello": "world",
		},
	})

	require.Nil(t, v)

	assert.EqualError(t, err, "some error")
}

func TestVariant_Success(t *testing.T) {
	var (
		flagKey      = "test-flag"
		namespaceKey = "test-namespace"
		store        = &evaluationStoreMock{}
		evaluator    = &evaluatorMock{}
		logger       = zaptest.NewLogger(t)
		s            = &Server{
			logger:    logger,
			store:     store,
			evaluator: evaluator,
		}
		flag = &flipt.Flag{
			NamespaceKey: namespaceKey,
			Key:          flagKey,
			Enabled:      true,
			Type:         flipt.FlagType_VARIANT_FLAG_TYPE,
		}
	)

	store.On("GetFlag", mock.Anything, namespaceKey, flagKey).Return(flag, nil)

	evaluator.On("Evaluate", mock.Anything, flag, &flipt.EvaluationRequest{
		FlagKey:      flagKey,
		NamespaceKey: namespaceKey,
		EntityId:     "test-entity",
		Context: map[string]string{
			"hello": "world",
		},
	}).Return(
		&flipt.EvaluationResponse{
			FlagKey:      flagKey,
			NamespaceKey: namespaceKey,
			Value:        "foo",
			Match:        true,
			SegmentKey:   "segment",
			Reason:       flipt.EvaluationReason(rpcEvaluation.EvaluationReason_MATCH_EVALUATION_REASON),
		}, nil)

	v, err := s.Variant(context.TODO(), &rpcEvaluation.EvaluationRequest{
		FlagKey:      flagKey,
		EntityId:     "test-entity",
		NamespaceKey: namespaceKey,
		Context: map[string]string{
			"hello": "world",
		},
	})

	require.NoError(t, err)

	assert.Equal(t, true, v.Match)
	assert.Equal(t, "foo", v.VariantKey)
	assert.Equal(t, "segment", v.SegmentKey)
	assert.Equal(t, rpcEvaluation.EvaluationReason_MATCH_EVALUATION_REASON, v.Reason)
}
