package ofrep

import (
	"context"
	"fmt"
	"io"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	flipterrors "go.flipt.io/flipt/errors"
	"google.golang.org/grpc/codes"

	"google.golang.org/grpc/metadata"

	"google.golang.org/protobuf/proto"

	"github.com/stretchr/testify/assert"
	"go.flipt.io/flipt/internal/config"
	rpcevaluation "go.flipt.io/flipt/rpc/flipt/evaluation"
	"go.flipt.io/flipt/rpc/flipt/ofrep"
	"go.uber.org/zap/zaptest"
	"google.golang.org/protobuf/types/known/structpb"
)

func TestEvaluateFlag_Success(t *testing.T) {
	t.Run("should use the default namespace when no one was provided", func(t *testing.T) {
		ctx := context.TODO()
		flagKey := "flag-key"
		expectedResponse := &ofrep.EvaluatedFlag{
			Key:      flagKey,
			Reason:   ofrep.EvaluateReason_DEFAULT,
			Variant:  "false",
			Value:    structpb.NewBoolValue(false),
			Metadata: &structpb.Struct{Fields: make(map[string]*structpb.Value)},
		}
		bridge := NewMockBridge(t)
		s := New(zaptest.NewLogger(t), config.CacheConfig{}, bridge)

		bridge.On("OFREPFlagEvaluation", ctx, EvaluationBridgeInput{
			FlagKey:      flagKey,
			NamespaceKey: "default",
			EntityId:     "testing-key",
			Context: map[string]string{
				ofrepCtxTargetingKey: "testing-key",
				"hello":              "world",
			},
		}).Return(EvaluationBridgeOutput{
			FlagKey: flagKey,
			Reason:  rpcevaluation.EvaluationReason_DEFAULT_EVALUATION_REASON,
			Variant: "false",
			Value:   false,
		}, nil)

		actualResponse, err := s.EvaluateFlag(ctx, &ofrep.EvaluateFlagRequest{
			Key: flagKey,
			Context: map[string]string{
				ofrepCtxTargetingKey: "testing-key",
				"hello":              "world",
			},
		})
		require.NoError(t, err)
		assert.True(t, proto.Equal(expectedResponse, actualResponse))
	})

	t.Run("should use the given namespace when one was provided", func(t *testing.T) {
		namespace := "test-namespace"
		ctx := metadata.NewIncomingContext(context.TODO(), metadata.New(map[string]string{
			"x-flipt-namespace": namespace,
		}))
		flagKey := "flag-key"
		expectedResponse := &ofrep.EvaluatedFlag{
			Key:      flagKey,
			Reason:   ofrep.EvaluateReason_DISABLED,
			Variant:  "true",
			Value:    structpb.NewBoolValue(true),
			Metadata: &structpb.Struct{Fields: make(map[string]*structpb.Value)},
		}
		bridge := NewMockBridge(t)
		s := New(zaptest.NewLogger(t), config.CacheConfig{}, bridge)

		bridge.On("OFREPFlagEvaluation", ctx, EvaluationBridgeInput{
			FlagKey:      flagKey,
			EntityId:     "string",
			NamespaceKey: namespace,
			Context: map[string]string{
				ofrepCtxTargetingKey: "string",
			},
		}).Return(EvaluationBridgeOutput{
			FlagKey: flagKey,
			Reason:  rpcevaluation.EvaluationReason_FLAG_DISABLED_EVALUATION_REASON,
			Variant: "true",
			Value:   true,
		}, nil)

		actualResponse, err := s.EvaluateFlag(ctx, &ofrep.EvaluateFlagRequest{
			Key:     flagKey,
			Context: map[string]string{ofrepCtxTargetingKey: "string"},
		})
		require.NoError(t, err)
		assert.True(t, proto.Equal(expectedResponse, actualResponse))
	})
}

func TestEvaluateFlag_Failure(t *testing.T) {
	testCases := []struct {
		name         string
		req          *ofrep.EvaluateFlagRequest
		err          error
		expectedCode codes.Code
		expectedErr  error
	}{
		{
			name:        "should return a targeting key missing error when a key is not provided",
			req:         &ofrep.EvaluateFlagRequest{},
			expectedErr: newFlagMissingError(),
		},
		{
			name:        "should return a bad request error when an invalid is returned by the bridge",
			req:         &ofrep.EvaluateFlagRequest{Key: "test-flag"},
			err:         flipterrors.ErrInvalid("invalid"),
			expectedErr: newBadRequestError("test-flag", flipterrors.ErrInvalid("invalid")),
		},
		{
			name:        "should return a bad request error when a validation error is returned by the bridge",
			req:         &ofrep.EvaluateFlagRequest{Key: "test-flag"},
			err:         flipterrors.InvalidFieldError("field", "reason"),
			expectedErr: newBadRequestError("test-flag", flipterrors.InvalidFieldError("field", "reason")),
		},
		{
			name:        "should return a not found error when a flag not found error is returned by the bridge",
			req:         &ofrep.EvaluateFlagRequest{Key: "test-flag"},
			err:         flipterrors.ErrNotFound("test-flag"),
			expectedErr: newFlagNotFoundError("test-flag"),
		},
		{
			name:        "should return a general error",
			req:         &ofrep.EvaluateFlagRequest{Key: "test-flag"},
			err:         io.ErrNoProgress,
			expectedErr: io.ErrNoProgress,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctx := context.TODO()
			bridge := NewMockBridge(t)
			s := New(zaptest.NewLogger(t), config.CacheConfig{}, bridge)
			if tc.req.Key != "" {
				bridge.On("OFREPFlagEvaluation", ctx, mock.Anything).Return(EvaluationBridgeOutput{}, tc.err)
			}

			_, err := s.EvaluateFlag(ctx, tc.req)

			assert.Equal(t, tc.expectedErr, err)
		})
	}
}

func TestEvaluateBulkSuccess(t *testing.T) {
	t.Run("should use the default namespace when no one was provided", func(t *testing.T) {
		ctx := context.TODO()
		flagKey := "flag-key"
		expectedResponse := []*ofrep.EvaluatedFlag{{
			Key:     flagKey,
			Reason:  ofrep.EvaluateReason_DEFAULT,
			Variant: "false",
			Value:   structpb.NewBoolValue(false),
			Metadata: &structpb.Struct{
				Fields: map[string]*structpb.Value{"attachment": structpb.NewStringValue("my value")},
			},
		}}
		bridge := NewMockBridge(t)
		s := New(zaptest.NewLogger(t), config.CacheConfig{}, bridge)

		bridge.On("OFREPFlagEvaluation", ctx, EvaluationBridgeInput{
			FlagKey:      flagKey,
			NamespaceKey: "default",
			EntityId:     "targeting",
			Context: map[string]string{
				ofrepCtxTargetingKey: "targeting",
				"flags":              flagKey,
			},
		}).Return(EvaluationBridgeOutput{
			FlagKey:  flagKey,
			Reason:   rpcevaluation.EvaluationReason_DEFAULT_EVALUATION_REASON,
			Variant:  "false",
			Value:    false,
			Metadata: map[string]any{"attachment": "my value"},
		}, nil)

		actualResponse, err := s.EvaluateBulk(ctx, &ofrep.EvaluateBulkRequest{
			Context: map[string]string{
				ofrepCtxTargetingKey: "targeting",
				"flags":              flagKey,
			},
		})
		require.NoError(t, err)
		require.Len(t, actualResponse.Flags, len(expectedResponse))
		for i, expected := range expectedResponse {
			fmt.Println(actualResponse.Flags)
			assert.True(t, proto.Equal(expected, actualResponse.Flags[i]))
		}
	})
}
