package evaluation

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.flipt.io/flipt/internal/server/environments"
	"go.flipt.io/flipt/internal/server/tracing"
	"go.flipt.io/flipt/internal/storage"
	"go.flipt.io/flipt/rpc/flipt/core"
	"go.flipt.io/flipt/rpc/flipt/ofrep"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
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

	envStore.On("GetFromContext", mock.Anything).Return(environment, nil)
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

	ctx := metadata.NewIncomingContext(t.Context(), metadata.New(map[string]string{
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

	envStore.On("GetFromContext", mock.Anything).Return(environment, nil)
	environment.On("EvaluationStore").Return(store, nil)

	store.On("GetFlag", mock.Anything, storage.NewResource(namespaceKey, flagKey)).Return(flag, nil)

	store.On("GetEvaluationRollouts", mock.Anything, storage.NewResource(namespaceKey, flagKey)).Return([]*storage.EvaluationRollout{}, nil)

	ctx := metadata.NewIncomingContext(t.Context(), metadata.New(map[string]string{
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

func TestOFREPEvaluationWithTracing(t *testing.T) {
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
				envStore.On("GetFromContext", mock.Anything).Return(environment, nil)
				environment.On("EvaluationStore").Return(store, nil)

				store.On("GetFlag", mock.Anything, storage.NewResource(namespaceKey, flagKey)).Return(&core.Flag{
					Key:     flagKey,
					Enabled: true,
					Type:    core.FlagType_VARIANT_FLAG_TYPE,
				}, nil)

				store.On("GetEvaluationRules", mock.Anything, storage.NewResource(namespaceKey, flagKey)).Return([]*storage.EvaluationRule{
					{
						ID:      "1",
						FlagKey: flagKey,
						Rank:    0,
						Segments: map[string]*storage.EvaluationSegment{
							"test-segment": {
								SegmentKey: "test-segment",
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
						VariantKey: "test-variant",
					},
				}, nil)
			},
			runEvaluation: func(ctx context.Context, s *Server) error {
				_, err := s.OFREPFlagEvaluation(ctx, &ofrep.EvaluateFlagRequest{
					Key: "test-flag",
					Context: map[string]string{
						"hello":        "world",
						"targetingKey": "test-entity",
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
				assertAttributeNotPresent(t, attrs, tracing.AttributeRequestID)
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
				envStore.On("GetFromContext", mock.Anything).Return(environment, nil)
				environment.On("EvaluationStore").Return(store, nil)

				store.On("GetFlag", mock.Anything, storage.NewResource(namespaceKey, flagKey)).Return(&core.Flag{
					Key:     flagKey,
					Enabled: true,
					Type:    core.FlagType_BOOLEAN_FLAG_TYPE,
				}, nil)

				store.On("GetEvaluationRollouts", mock.Anything, storage.NewResource(namespaceKey, flagKey)).Return([]*storage.EvaluationRollout{}, nil)
			},
			runEvaluation: func(ctx context.Context, s *Server) error {
				_, err := s.OFREPFlagEvaluation(ctx, &ofrep.EvaluateFlagRequest{
					Key: "test-flag",
					Context: map[string]string{
						"targetingKey": "test-entity",
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
				assertAttributeNotPresent(t, attrs, tracing.AttributeRequestID)
				assertAttributeValue(t, attrs, tracing.AttributeVariant, true)
				assertAttributeValue(t, attrs, tracing.AttributeReason, "default")
				assertAttributeValue(t, attrs, tracing.AttributeFlagType, "boolean")
			},
		},
		{
			name:               "tracing_disabled",
			tracingEnabled:     false,
			expectedEventCount: 0,
			setupMocks: func(t *testing.T, envStore *MockEnvironmentStore, environment *environments.MockEnvironment, store *storage.MockReadOnlyStore, namespaceKey string) {
				flagKey := "test-flag"
				environment.On("Key").Return("test-environment")
				envStore.On("GetFromContext", mock.Anything).Return(environment, nil)
				environment.On("EvaluationStore").Return(store, nil)

				store.On("GetFlag", mock.Anything, storage.NewResource(namespaceKey, flagKey)).Return(&core.Flag{
					Key:     flagKey,
					Enabled: true,
					Type:    core.FlagType_BOOLEAN_FLAG_TYPE,
				}, nil)

				store.On("GetEvaluationRollouts", mock.Anything, storage.NewResource(namespaceKey, flagKey)).Return([]*storage.EvaluationRollout{}, nil)
			},
			runEvaluation: func(ctx context.Context, s *Server) error {
				_, err := s.OFREPFlagEvaluation(ctx, &ofrep.EvaluateFlagRequest{
					Key: "test-flag",
					Context: map[string]string{
						"targetingKey": "test-entity",
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

			mdCtx := metadata.NewIncomingContext(t.Context(), metadata.New(map[string]string{
				"x-flipt-environment": environmentKey,
				"x-flipt-namespace":   namespaceKey,
			}))
			ctx, span := tracer.Start(mdCtx, "test-span")

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

func TestOFREPFlagEvaluationBulkWithEtag(t *testing.T) {
	tests := []struct {
		name              string
		metadata          map[string]string
		flags             []*core.Flag
		expectListFlags   bool
		expectNotModified bool
		run               func(ctx context.Context, s *Server) (*ofrep.BulkEvaluationResponse, error)
		validate          func(t *testing.T, resp *ofrep.BulkEvaluationResponse, err error)
	}{
		{
			name: "success returns flags",
			metadata: map[string]string{
				"x-flipt-environment": "test-environment",
				"x-flipt-namespace":   "test-namespace",
			},
			flags: []*core.Flag{{
				Key:     "test-flag",
				Enabled: true,
				Type:    core.FlagType_VARIANT_FLAG_TYPE,
			}},
			expectListFlags: true,
			run: func(ctx context.Context, s *Server) (*ofrep.BulkEvaluationResponse, error) {
				return s.OFREPFlagEvaluationBulk(ctx, &ofrep.EvaluateBulkRequest{
					Context: map[string]string{"targetingKey": "12345"},
				})
			},
			validate: func(t *testing.T, resp *ofrep.BulkEvaluationResponse, err error) {
				require.NoError(t, err)
				require.Len(t, resp.Flags, 1)
				assert.Equal(t, "test-flag", resp.Flags[0].Key)
				assert.Equal(t, "boz", resp.Flags[0].Variant)
			},
		},
		{
			name: "returns not modified when etag matches",
			metadata: map[string]string{
				"x-flipt-environment":       "test-environment",
				"x-flipt-namespace":         "test-namespace",
				"GrpcGateway-If-None-Match": "7e427de9ffaef7ba",
			},
			flags: []*core.Flag{{
				Key:     "test-flag",
				Enabled: true,
				Type:    core.FlagType_VARIANT_FLAG_TYPE,
			}},
			expectListFlags:   true,
			expectNotModified: true,
			run: func(ctx context.Context, s *Server) (*ofrep.BulkEvaluationResponse, error) {
				return s.OFREPFlagEvaluationBulk(ctx, &ofrep.EvaluateBulkRequest{
					Context: map[string]string{"targetingKey": "12345"},
				})
			},
			validate: func(t *testing.T, resp *ofrep.BulkEvaluationResponse, err error) {
				require.Error(t, err)
				assert.Contains(t, err.Error(), "not modified")
			},
		},
		{
			name: "returns flags when etag does not match",
			metadata: map[string]string{
				"x-flipt-environment":       "test-environment",
				"x-flipt-namespace":         "test-namespace",
				"GrpcGateway-If-None-Match": "different-etag",
			},
			flags: []*core.Flag{{
				Key:     "test-flag",
				Enabled: true,
				Type:    core.FlagType_VARIANT_FLAG_TYPE,
			}},
			expectListFlags: true,
			run: func(ctx context.Context, s *Server) (*ofrep.BulkEvaluationResponse, error) {
				return s.OFREPFlagEvaluationBulk(ctx, &ofrep.EvaluateBulkRequest{
					Context: map[string]string{"targetingKey": "12345"},
				})
			},
			validate: func(t *testing.T, resp *ofrep.BulkEvaluationResponse, err error) {
				require.NoError(t, err)
				require.Len(t, resp.Flags, 1)
				assert.Equal(t, "boz", resp.Flags[0].Variant)
			},
		},
		{
			name: "uses flags from context instead of listing",
			metadata: map[string]string{
				"x-flipt-environment": "test-environment",
				"x-flipt-namespace":   "test-namespace",
			},
			flags: []*core.Flag{{
				Key:     "test-flag",
				Enabled: true,
				Type:    core.FlagType_VARIANT_FLAG_TYPE,
			}},
			expectListFlags: false,
			run: func(ctx context.Context, s *Server) (*ofrep.BulkEvaluationResponse, error) {
				return s.OFREPFlagEvaluationBulk(ctx, &ofrep.EvaluateBulkRequest{
					Context: map[string]string{
						"flags":        "test-flag",
						"targetingKey": "entity-123",
					},
				})
			},
			validate: func(t *testing.T, resp *ofrep.BulkEvaluationResponse, err error) {
				require.NoError(t, err)
				require.Len(t, resp.Flags, 1)
				assert.Equal(t, "test-flag", resp.Flags[0].Key)
				assert.Equal(t, "variant-a", resp.Flags[0].Variant)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var (
				environmentKey = "test-environment"
				envStore       = NewMockEnvironmentStore(t)
				environment    = environments.NewMockEnvironment(t)
				store          = storage.NewMockReadOnlyStore(t)
				logger         = zaptest.NewLogger(t)
				s              = New(logger, envStore)
			)

			environment.On("Key").Return(environmentKey)
			envStore.On("GetFromContext", mock.Anything).Return(environment, nil)
			environment.On("EvaluationStore").Return(store, nil)

			for _, flag := range tt.flags {
				if flag.Key == "" {
					continue
				}

				store.On("GetFlag", mock.Anything, mock.Anything).Return(flag, nil)
				store.On("GetEvaluationRules", mock.Anything, mock.Anything).Return([]*storage.EvaluationRule{
					{ID: "1", FlagKey: flag.Key, Rank: 0, Segments: map[string]*storage.EvaluationSegment{
						"bar": {SegmentKey: "bar", MatchType: core.MatchType_ANY_MATCH_TYPE},
					}},
				}, nil)

				variantKey := "boz"
				if flag.Key == "test-flag" && tt.name == "uses flags from context instead of listing" {
					variantKey = "variant-a"
				}
				store.On("GetEvaluationDistributions", mock.Anything, mock.Anything, storage.NewID("1")).Return([]*storage.EvaluationDistribution{
					{ID: "4", RuleID: "1", VariantID: "5", Rollout: 100, VariantKey: variantKey},
				}, nil)
			}

			if tt.expectListFlags && len(tt.flags) > 0 {
				store.On("ListFlags", mock.Anything, mock.Anything).Return(storage.ResultSet[*core.Flag]{Results: tt.flags}, nil)
			}

			ctx := metadata.NewIncomingContext(t.Context(), metadata.New(tt.metadata))

			resp, err := tt.run(ctx, s)

			if tt.expectNotModified {
				require.Error(t, err)
				assert.Contains(t, err.Error(), "not modified")
			} else {
				tt.validate(t, resp, err)
			}
		})
	}
}

func TestOFREPFlagEvaluationBulk(t *testing.T) {
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

	envStore.On("GetFromContext", mock.Anything).Return(environment, nil)
	environment.On("EvaluationStore").Return(store, nil)

	store.On("GetFlag", mock.Anything, mock.Anything).Return(flag, nil)

	store.On("ListFlags", mock.Anything, mock.Anything).Return(storage.ResultSet[*core.Flag]{
		Results: []*core.Flag{flag},
	}, nil)

	store.On("GetEvaluationRules", mock.Anything, mock.Anything).Return([]*storage.EvaluationRule{
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

	result, err := s.OFREPFlagEvaluationBulk(ctx, &ofrep.EvaluateBulkRequest{
		Context: map[string]string{
			"targetingKey": "12345",
		},
	})
	require.NoError(t, err)

	assert.Len(t, result.Flags, 1)
	evaluation := result.Flags[0]
	assert.Equal(t, flagKey, evaluation.Key)
	assert.Equal(t, ofrep.EvaluateReason_TARGETING_MATCH, evaluation.Reason)
	assert.Equal(t, "boz", evaluation.Variant)

	require.Len(t, result.EventStreams, 1)

	stream := result.EventStreams[0]
	assert.Equal(t, "sse", stream.Type)
	assert.NotNil(t, stream.Endpoint)
	assert.Equal(t, "/client/v2/environments/test-environment/namespaces/test-namespace/stream", stream.Endpoint.GetRequestUri())
	assert.Empty(t, stream.Endpoint.GetOrigin())
}
