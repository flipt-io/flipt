package grpc_middleware

import (
	"context"
	"testing"
	"time"

	"go.flipt.io/flipt/errors"
	"go.flipt.io/flipt/internal/config"
	"go.flipt.io/flipt/internal/server"
	"go.flipt.io/flipt/internal/server/auth"
	"go.flipt.io/flipt/internal/server/cache/memory"
	"go.flipt.io/flipt/internal/storage"
	flipt "go.flipt.io/flipt/rpc/flipt"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.uber.org/zap/zaptest"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	authrpc "go.flipt.io/flipt/rpc/flipt/auth"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type validatable struct {
	err error
}

func (v *validatable) Validate() error {
	return v.err
}

func TestValidationUnaryInterceptor(t *testing.T) {
	tests := []struct {
		name       string
		req        interface{}
		wantCalled int
	}{
		{
			name:       "does not implement Validate",
			req:        struct{}{},
			wantCalled: 1,
		},
		{
			name:       "implements validate no error",
			req:        &validatable{},
			wantCalled: 1,
		},
		{
			name: "implements validate error",
			req:  &validatable{err: errors.New("invalid")},
		},
	}

	for _, tt := range tests {
		var (
			req        = tt.req
			wantCalled = tt.wantCalled
			called     int
		)

		t.Run(tt.name, func(t *testing.T) {
			var (
				spyHandler = grpc.UnaryHandler(func(ctx context.Context, req interface{}) (interface{}, error) {
					called++
					return nil, nil
				})
			)

			_, _ = ValidationUnaryInterceptor(context.Background(), req, nil, spyHandler)
			assert.Equal(t, wantCalled, called)
		})
	}
}

func TestErrorUnaryInterceptor(t *testing.T) {
	tests := []struct {
		name     string
		wantErr  error
		wantCode codes.Code
	}{
		{
			name:     "not found error",
			wantErr:  errors.ErrNotFound("foo"),
			wantCode: codes.NotFound,
		},
		{
			name:     "invalid error",
			wantErr:  errors.ErrInvalid("foo"),
			wantCode: codes.InvalidArgument,
		},
		{
			name:     "invalid field",
			wantErr:  errors.InvalidFieldError("bar", "is wrong"),
			wantCode: codes.InvalidArgument,
		},
		{
			name:     "empty field",
			wantErr:  errors.EmptyFieldError("bar"),
			wantCode: codes.InvalidArgument,
		},
		{
			name:     "unauthenticated error",
			wantErr:  errors.NewErrorf[errors.ErrUnauthenticated]("user %q not found", "foo"),
			wantCode: codes.Unauthenticated,
		},
		{
			name:     "other error",
			wantErr:  errors.New("foo"),
			wantCode: codes.Internal,
		},
		{
			name: "no error",
		},
	}

	for _, tt := range tests {
		var (
			wantErr  = tt.wantErr
			wantCode = tt.wantCode
		)

		t.Run(tt.name, func(t *testing.T) {
			var (
				spyHandler = grpc.UnaryHandler(func(ctx context.Context, req interface{}) (interface{}, error) {
					return nil, wantErr
				})
			)

			_, err := ErrorUnaryInterceptor(context.Background(), nil, nil, spyHandler)
			if wantErr != nil {
				require.Error(t, err)
				status := status.Convert(err)
				assert.Equal(t, wantCode, status.Code())
				return
			}

			require.NoError(t, err)
		})
	}
}

func TestEvaluationUnaryInterceptor_Noop(t *testing.T) {
	var (
		req = &flipt.GetFlagRequest{
			Key: "foo",
		}

		handler = func(ctx context.Context, r interface{}) (interface{}, error) {
			return &flipt.Flag{
				Key: "foo",
			}, nil
		}

		info = &grpc.UnaryServerInfo{
			FullMethod: "FakeMethod",
		}
	)

	got, err := EvaluationUnaryInterceptor(context.Background(), req, info, handler)
	require.NoError(t, err)

	assert.NotNil(t, got)

	resp, ok := got.(*flipt.Flag)
	assert.True(t, ok)
	assert.NotNil(t, resp)
	assert.Equal(t, "foo", resp.Key)
}

func TestEvaluationUnaryInterceptor_Evaluation(t *testing.T) {
	var (
		req = &flipt.EvaluationRequest{
			FlagKey: "foo",
		}

		handler = func(ctx context.Context, r interface{}) (interface{}, error) {
			return &flipt.EvaluationResponse{
				FlagKey: "foo",
			}, nil
		}

		info = &grpc.UnaryServerInfo{
			FullMethod: "FakeMethod",
		}
	)

	got, err := EvaluationUnaryInterceptor(context.Background(), req, info, handler)
	require.NoError(t, err)

	assert.NotNil(t, got)

	resp, ok := got.(*flipt.EvaluationResponse)
	assert.True(t, ok)
	assert.NotNil(t, resp)

	assert.Equal(t, "foo", resp.FlagKey)
	// check that the requestID was created and set
	assert.NotEmpty(t, resp.RequestId)
	assert.NotZero(t, resp.Timestamp)
	assert.NotZero(t, resp.RequestDurationMillis)

	req = &flipt.EvaluationRequest{
		FlagKey:   "foo",
		RequestId: "bar",
	}

	got, err = EvaluationUnaryInterceptor(context.Background(), req, info, handler)
	require.NoError(t, err)

	assert.NotNil(t, got)

	resp, ok = got.(*flipt.EvaluationResponse)
	assert.True(t, ok)
	assert.NotNil(t, resp)

	assert.Equal(t, "foo", resp.FlagKey)
	// check that the requestID was propagated
	assert.NotEmpty(t, resp.RequestId)
	assert.Equal(t, "bar", resp.RequestId)
	assert.NotZero(t, resp.Timestamp)
	assert.NotZero(t, resp.RequestDurationMillis)
}

func TestEvaluationUnaryInterceptor_BatchEvaluation(t *testing.T) {
	var (
		req = &flipt.BatchEvaluationRequest{
			Requests: []*flipt.EvaluationRequest{
				{
					FlagKey: "foo",
				},
			},
		}

		handler = func(ctx context.Context, r interface{}) (interface{}, error) {
			return &flipt.BatchEvaluationResponse{
				Responses: []*flipt.EvaluationResponse{
					{
						FlagKey: "foo",
					},
				},
			}, nil
		}

		info = &grpc.UnaryServerInfo{
			FullMethod: "FakeMethod",
		}
	)

	got, err := EvaluationUnaryInterceptor(context.Background(), req, info, handler)
	require.NoError(t, err)

	assert.NotNil(t, got)

	resp, ok := got.(*flipt.BatchEvaluationResponse)
	assert.True(t, ok)
	assert.NotNil(t, resp)
	assert.NotEmpty(t, resp.Responses)
	assert.Equal(t, "foo", resp.Responses[0].FlagKey)
	// check that the requestID was created and set
	assert.NotEmpty(t, resp.RequestId)
	assert.NotZero(t, resp.RequestDurationMillis)

	req = &flipt.BatchEvaluationRequest{
		RequestId: "bar",
		Requests: []*flipt.EvaluationRequest{
			{
				FlagKey: "foo",
			},
		},
	}

	got, err = EvaluationUnaryInterceptor(context.Background(), req, info, handler)
	require.NoError(t, err)

	assert.NotNil(t, got)

	resp, ok = got.(*flipt.BatchEvaluationResponse)
	assert.True(t, ok)
	assert.NotNil(t, resp)
	assert.NotEmpty(t, resp.Responses)
	assert.Equal(t, "foo", resp.Responses[0].FlagKey)
	// check that the requestID was propagated
	assert.NotEmpty(t, resp.RequestId)
	assert.Equal(t, "bar", resp.RequestId)
	assert.NotZero(t, resp.RequestDurationMillis)
}

func TestCacheUnaryInterceptor_GetFlag(t *testing.T) {
	var (
		store = &storeMock{}
		cache = memory.NewCache(config.CacheConfig{
			TTL:     time.Second,
			Enabled: true,
			Backend: config.CacheMemory,
		})
		cacheSpy = newCacheSpy(cache)
		logger   = zaptest.NewLogger(t)
		s        = server.New(logger, store)
	)

	store.On("GetFlag", mock.Anything, mock.Anything, "foo").Return(&flipt.Flag{
		NamespaceKey: storage.DefaultNamespace,
		Key:          "foo",
		Enabled:      true,
	}, nil)

	unaryInterceptor := CacheUnaryInterceptor(cacheSpy, logger)

	handler := func(ctx context.Context, r interface{}) (interface{}, error) {
		return s.GetFlag(ctx, r.(*flipt.GetFlagRequest))
	}

	info := &grpc.UnaryServerInfo{
		FullMethod: "FakeMethod",
	}

	for i := 0; i < 10; i++ {
		req := &flipt.GetFlagRequest{Key: "foo"}
		got, err := unaryInterceptor(context.Background(), req, info, handler)
		require.NoError(t, err)
		assert.NotNil(t, got)
	}

	assert.Equal(t, 10, cacheSpy.getCalled)
	assert.NotEmpty(t, cacheSpy.getKeys)

	// cache key is flipt:(md5(f:foo))
	const cacheKey = "flipt:864ce319cc64891a59e4745fbe7ecc47"
	_, ok := cacheSpy.getKeys[cacheKey]
	assert.True(t, ok)

	assert.Equal(t, 1, cacheSpy.setCalled)
	assert.NotEmpty(t, cacheSpy.setItems)
	assert.NotEmpty(t, cacheSpy.setItems[cacheKey])
}

func TestCacheUnaryInterceptor_UpdateFlag(t *testing.T) {
	var (
		store = &storeMock{}
		cache = memory.NewCache(config.CacheConfig{
			TTL:     time.Second,
			Enabled: true,
			Backend: config.CacheMemory,
		})
		cacheSpy = newCacheSpy(cache)
		logger   = zaptest.NewLogger(t)
		s        = server.New(logger, store)
		req      = &flipt.UpdateFlagRequest{
			Key:         "key",
			Name:        "name",
			Description: "desc",
			Enabled:     true,
		}
	)

	store.On("UpdateFlag", mock.Anything, req).Return(&flipt.Flag{
		Key:         req.Key,
		Name:        req.Name,
		Description: req.Description,
		Enabled:     req.Enabled,
	}, nil)

	unaryInterceptor := CacheUnaryInterceptor(cacheSpy, logger)

	handler := func(ctx context.Context, r interface{}) (interface{}, error) {
		return s.UpdateFlag(ctx, r.(*flipt.UpdateFlagRequest))
	}

	info := &grpc.UnaryServerInfo{
		FullMethod: "FakeMethod",
	}

	got, err := unaryInterceptor(context.Background(), req, info, handler)
	require.NoError(t, err)
	assert.NotNil(t, got)

	assert.Equal(t, 1, cacheSpy.deleteCalled)
	assert.NotEmpty(t, cacheSpy.deleteKeys)
}

func TestCacheUnaryInterceptor_DeleteFlag(t *testing.T) {
	var (
		store = &storeMock{}
		cache = memory.NewCache(config.CacheConfig{
			TTL:     time.Second,
			Enabled: true,
			Backend: config.CacheMemory,
		})
		cacheSpy = newCacheSpy(cache)
		logger   = zaptest.NewLogger(t)
		s        = server.New(logger, store)
		req      = &flipt.DeleteFlagRequest{
			Key: "key",
		}
	)

	store.On("DeleteFlag", mock.Anything, req).Return(nil)

	unaryInterceptor := CacheUnaryInterceptor(cacheSpy, logger)

	handler := func(ctx context.Context, r interface{}) (interface{}, error) {
		return s.DeleteFlag(ctx, r.(*flipt.DeleteFlagRequest))
	}

	info := &grpc.UnaryServerInfo{
		FullMethod: "FakeMethod",
	}

	got, err := unaryInterceptor(context.Background(), req, info, handler)
	require.NoError(t, err)
	assert.NotNil(t, got)

	assert.Equal(t, 1, cacheSpy.deleteCalled)
	assert.NotEmpty(t, cacheSpy.deleteKeys)
}

func TestCacheUnaryInterceptor_CreateVariant(t *testing.T) {
	var (
		store = &storeMock{}
		cache = memory.NewCache(config.CacheConfig{
			TTL:     time.Second,
			Enabled: true,
			Backend: config.CacheMemory,
		})
		cacheSpy = newCacheSpy(cache)
		logger   = zaptest.NewLogger(t)
		s        = server.New(logger, store)
		req      = &flipt.CreateVariantRequest{
			FlagKey:     "flagKey",
			Key:         "key",
			Name:        "name",
			Description: "desc",
		}
	)

	store.On("CreateVariant", mock.Anything, req).Return(&flipt.Variant{
		Id:          "1",
		FlagKey:     req.FlagKey,
		Key:         req.Key,
		Name:        req.Name,
		Description: req.Description,
		Attachment:  req.Attachment,
	}, nil)

	unaryInterceptor := CacheUnaryInterceptor(cacheSpy, logger)

	handler := func(ctx context.Context, r interface{}) (interface{}, error) {
		return s.CreateVariant(ctx, r.(*flipt.CreateVariantRequest))
	}

	info := &grpc.UnaryServerInfo{
		FullMethod: "FakeMethod",
	}

	got, err := unaryInterceptor(context.Background(), req, info, handler)
	require.NoError(t, err)
	assert.NotNil(t, got)

	assert.Equal(t, 1, cacheSpy.deleteCalled)
	assert.NotEmpty(t, cacheSpy.deleteKeys)
}

func TestCacheUnaryInterceptor_UpdateVariant(t *testing.T) {
	var (
		store = &storeMock{}
		cache = memory.NewCache(config.CacheConfig{
			TTL:     time.Second,
			Enabled: true,
			Backend: config.CacheMemory,
		})
		cacheSpy = newCacheSpy(cache)
		logger   = zaptest.NewLogger(t)
		s        = server.New(logger, store)
		req      = &flipt.UpdateVariantRequest{
			Id:          "1",
			FlagKey:     "flagKey",
			Key:         "key",
			Name:        "name",
			Description: "desc",
		}
	)

	store.On("UpdateVariant", mock.Anything, req).Return(&flipt.Variant{
		Id:          req.Id,
		FlagKey:     req.FlagKey,
		Key:         req.Key,
		Name:        req.Name,
		Description: req.Description,
		Attachment:  req.Attachment,
	}, nil)

	unaryInterceptor := CacheUnaryInterceptor(cacheSpy, logger)

	handler := func(ctx context.Context, r interface{}) (interface{}, error) {
		return s.UpdateVariant(ctx, r.(*flipt.UpdateVariantRequest))
	}

	info := &grpc.UnaryServerInfo{
		FullMethod: "FakeMethod",
	}

	got, err := unaryInterceptor(context.Background(), req, info, handler)
	require.NoError(t, err)
	assert.NotNil(t, got)

	assert.Equal(t, 1, cacheSpy.deleteCalled)
	assert.NotEmpty(t, cacheSpy.deleteKeys)
}

func TestCacheUnaryInterceptor_DeleteVariant(t *testing.T) {
	var (
		store = &storeMock{}
		cache = memory.NewCache(config.CacheConfig{
			TTL:     time.Second,
			Enabled: true,
			Backend: config.CacheMemory,
		})
		cacheSpy = newCacheSpy(cache)
		logger   = zaptest.NewLogger(t)
		s        = server.New(logger, store)
		req      = &flipt.DeleteVariantRequest{
			Id: "1",
		}
	)

	store.On("DeleteVariant", mock.Anything, req).Return(nil)

	unaryInterceptor := CacheUnaryInterceptor(cacheSpy, logger)

	handler := func(ctx context.Context, r interface{}) (interface{}, error) {
		return s.DeleteVariant(ctx, r.(*flipt.DeleteVariantRequest))
	}

	info := &grpc.UnaryServerInfo{
		FullMethod: "FakeMethod",
	}

	got, err := unaryInterceptor(context.Background(), req, info, handler)
	require.NoError(t, err)
	assert.NotNil(t, got)

	assert.Equal(t, 1, cacheSpy.deleteCalled)
	assert.NotEmpty(t, cacheSpy.deleteKeys)
}

func TestCacheUnaryInterceptor_Evaluate(t *testing.T) {
	var (
		store = &storeMock{}
		cache = memory.NewCache(config.CacheConfig{
			TTL:     time.Second,
			Enabled: true,
			Backend: config.CacheMemory,
		})
		cacheSpy = newCacheSpy(cache)
		logger   = zaptest.NewLogger(t)
		s        = server.New(logger, store)
	)

	store.On("GetFlag", mock.Anything, mock.Anything, "foo").Return(&flipt.Flag{
		Key:     "foo",
		Enabled: true,
	}, nil)

	store.On("GetEvaluationRules", mock.Anything, mock.Anything, "foo").Return(
		[]*storage.EvaluationRule{
			{
				ID:               "1",
				FlagKey:          "foo",
				SegmentKey:       "bar",
				SegmentMatchType: flipt.MatchType_ALL_MATCH_TYPE,
				Rank:             0,
				Constraints: []storage.EvaluationConstraint{
					// constraint: bar (string) == baz
					{
						ID:       "2",
						Type:     flipt.ComparisonType_STRING_COMPARISON_TYPE,
						Property: "bar",
						Operator: flipt.OpEQ,
						Value:    "baz",
					},
					// constraint: admin (bool) == true
					{
						ID:       "3",
						Type:     flipt.ComparisonType_BOOLEAN_COMPARISON_TYPE,
						Property: "admin",
						Operator: flipt.OpTrue,
					},
				},
			},
		}, nil)

	store.On("GetEvaluationDistributions", mock.Anything, "1").Return(
		[]*storage.EvaluationDistribution{
			{
				ID:                "4",
				RuleID:            "1",
				VariantID:         "5",
				Rollout:           100,
				VariantKey:        "boz",
				VariantAttachment: `{"key":"value"}`,
			},
		}, nil)

	tests := []struct {
		name      string
		req       *flipt.EvaluationRequest
		wantMatch bool
	}{
		{
			name: "matches all",
			req: &flipt.EvaluationRequest{
				FlagKey:  "foo",
				EntityId: "1",
				Context: map[string]string{
					"bar":   "baz",
					"admin": "true",
				},
			},
			wantMatch: true,
		},
		{
			name: "no match all",
			req: &flipt.EvaluationRequest{
				FlagKey:  "foo",
				EntityId: "1",
				Context: map[string]string{
					"bar":   "boz",
					"admin": "true",
				},
			},
		},
		{
			name: "no match just bool value",
			req: &flipt.EvaluationRequest{
				FlagKey:  "foo",
				EntityId: "1",
				Context: map[string]string{
					"admin": "true",
				},
			},
		},
		{
			name: "no match just string value",
			req: &flipt.EvaluationRequest{
				FlagKey:  "foo",
				EntityId: "1",
				Context: map[string]string{
					"bar": "baz",
				},
			},
		},
	}

	unaryInterceptor := CacheUnaryInterceptor(cacheSpy, logger)

	handler := func(ctx context.Context, r interface{}) (interface{}, error) {
		return s.Evaluate(ctx, r.(*flipt.EvaluationRequest))
	}

	info := &grpc.UnaryServerInfo{
		FullMethod: "FakeMethod",
	}

	for i, tt := range tests {
		var (
			i         = i + 1
			req       = tt.req
			wantMatch = tt.wantMatch
		)

		t.Run(tt.name, func(t *testing.T) {
			got, err := unaryInterceptor(context.Background(), req, info, handler)
			require.NoError(t, err)
			assert.NotNil(t, got)

			resp := got.(*flipt.EvaluationResponse)
			assert.NotNil(t, resp)
			assert.Equal(t, "foo", resp.FlagKey)
			assert.Equal(t, req.Context, resp.RequestContext)

			assert.Equal(t, i, cacheSpy.getCalled)
			assert.NotEmpty(t, cacheSpy.getKeys)

			if !wantMatch {
				assert.False(t, resp.Match)
				assert.Empty(t, resp.SegmentKey)
				return
			}

			assert.True(t, resp.Match)
			assert.Equal(t, "bar", resp.SegmentKey)
			assert.Equal(t, "boz", resp.Value)
			assert.Equal(t, `{"key":"value"}`, resp.Attachment)
		})
	}
}

func TestAuditUnaryInterceptor_CreateFlag(t *testing.T) {
	var (
		store       = &storeMock{}
		logger      = zaptest.NewLogger(t)
		exporterSpy = newAuditExporterSpy(logger)
		s           = server.New(logger, store)
		req         = &flipt.CreateFlagRequest{
			Key:         "key",
			Name:        "name",
			Description: "desc",
		}
	)

	store.On("CreateFlag", mock.Anything, req).Return(&flipt.Flag{
		Key:         req.Key,
		Name:        req.Name,
		Description: req.Description,
	}, nil)

	unaryInterceptor := AuditUnaryInterceptor(logger)

	handler := func(ctx context.Context, r interface{}) (interface{}, error) {
		return s.CreateFlag(ctx, r.(*flipt.CreateFlagRequest))
	}

	info := &grpc.UnaryServerInfo{
		FullMethod: "FakeMethod",
	}

	tp := sdktrace.NewTracerProvider(sdktrace.WithSampler(sdktrace.AlwaysSample()))
	tp.RegisterSpanProcessor(sdktrace.NewSimpleSpanProcessor(exporterSpy))

	tr := tp.Tracer("SpanProcessor")
	ctx, span := tr.Start(context.Background(), "OnStart")

	got, err := unaryInterceptor(ctx, req, info, handler)
	require.NoError(t, err)
	assert.NotNil(t, got)

	span.End()

	assert.Equal(t, exporterSpy.GetSendAuditsCalled(), 1)
}

func TestAuditUnaryInterceptor_UpdateFlag(t *testing.T) {
	var (
		store       = &storeMock{}
		logger      = zaptest.NewLogger(t)
		exporterSpy = newAuditExporterSpy(logger)
		s           = server.New(logger, store)
		req         = &flipt.UpdateFlagRequest{
			Key:         "key",
			Name:        "name",
			Description: "desc",
			Enabled:     true,
		}
	)

	store.On("UpdateFlag", mock.Anything, req).Return(&flipt.Flag{
		Key:         req.Key,
		Name:        req.Name,
		Description: req.Description,
		Enabled:     req.Enabled,
	}, nil)

	unaryInterceptor := AuditUnaryInterceptor(logger)

	handler := func(ctx context.Context, r interface{}) (interface{}, error) {
		return s.UpdateFlag(ctx, r.(*flipt.UpdateFlagRequest))
	}

	info := &grpc.UnaryServerInfo{
		FullMethod: "FakeMethod",
	}

	tp := sdktrace.NewTracerProvider(sdktrace.WithSampler(sdktrace.AlwaysSample()))
	tp.RegisterSpanProcessor(sdktrace.NewSimpleSpanProcessor(exporterSpy))

	tr := tp.Tracer("SpanProcessor")
	ctx, span := tr.Start(context.Background(), "OnStart")

	got, err := unaryInterceptor(ctx, req, info, handler)
	require.NoError(t, err)
	assert.NotNil(t, got)

	span.End()

	assert.Equal(t, exporterSpy.GetSendAuditsCalled(), 1)
}

func TestAuditUnaryInterceptor_DeleteFlag(t *testing.T) {
	var (
		store       = &storeMock{}
		logger      = zaptest.NewLogger(t)
		exporterSpy = newAuditExporterSpy(logger)
		s           = server.New(logger, store)
		req         = &flipt.DeleteFlagRequest{
			Key: "key",
		}
	)

	store.On("DeleteFlag", mock.Anything, req).Return(nil)

	unaryInterceptor := AuditUnaryInterceptor(logger)

	handler := func(ctx context.Context, r interface{}) (interface{}, error) {
		return s.DeleteFlag(ctx, r.(*flipt.DeleteFlagRequest))
	}

	info := &grpc.UnaryServerInfo{
		FullMethod: "FakeMethod",
	}

	tp := sdktrace.NewTracerProvider(sdktrace.WithSampler(sdktrace.AlwaysSample()))
	tp.RegisterSpanProcessor(sdktrace.NewSimpleSpanProcessor(exporterSpy))

	tr := tp.Tracer("SpanProcessor")
	ctx, span := tr.Start(context.Background(), "OnStart")

	got, err := unaryInterceptor(ctx, req, info, handler)
	require.NoError(t, err)
	assert.NotNil(t, got)

	span.End()
	assert.Equal(t, exporterSpy.GetSendAuditsCalled(), 1)
}

func TestAuditUnaryInterceptor_CreateVariant(t *testing.T) {
	var (
		store       = &storeMock{}
		logger      = zaptest.NewLogger(t)
		exporterSpy = newAuditExporterSpy(logger)
		s           = server.New(logger, store)
		req         = &flipt.CreateVariantRequest{
			FlagKey:     "flagKey",
			Key:         "key",
			Name:        "name",
			Description: "desc",
		}
	)

	store.On("CreateVariant", mock.Anything, req).Return(&flipt.Variant{
		Id:          "1",
		FlagKey:     req.FlagKey,
		Key:         req.Key,
		Name:        req.Name,
		Description: req.Description,
		Attachment:  req.Attachment,
	}, nil)

	unaryInterceptor := AuditUnaryInterceptor(logger)

	handler := func(ctx context.Context, r interface{}) (interface{}, error) {
		return s.CreateVariant(ctx, r.(*flipt.CreateVariantRequest))
	}

	info := &grpc.UnaryServerInfo{
		FullMethod: "FakeMethod",
	}

	tp := sdktrace.NewTracerProvider(sdktrace.WithSampler(sdktrace.AlwaysSample()))
	tp.RegisterSpanProcessor(sdktrace.NewSimpleSpanProcessor(exporterSpy))

	tr := tp.Tracer("SpanProcessor")
	ctx, span := tr.Start(context.Background(), "OnStart")

	got, err := unaryInterceptor(ctx, req, info, handler)
	require.NoError(t, err)
	assert.NotNil(t, got)

	span.End()
	assert.Equal(t, exporterSpy.GetSendAuditsCalled(), 1)
}

func TestAuditUnaryInterceptor_UpdateVariant(t *testing.T) {
	var (
		store       = &storeMock{}
		logger      = zaptest.NewLogger(t)
		exporterSpy = newAuditExporterSpy(logger)
		s           = server.New(logger, store)
		req         = &flipt.UpdateVariantRequest{
			Id:          "1",
			FlagKey:     "flagKey",
			Key:         "key",
			Name:        "name",
			Description: "desc",
		}
	)

	store.On("UpdateVariant", mock.Anything, req).Return(&flipt.Variant{
		Id:          req.Id,
		FlagKey:     req.FlagKey,
		Key:         req.Key,
		Name:        req.Name,
		Description: req.Description,
		Attachment:  req.Attachment,
	}, nil)

	unaryInterceptor := AuditUnaryInterceptor(logger)

	handler := func(ctx context.Context, r interface{}) (interface{}, error) {
		return s.UpdateVariant(ctx, r.(*flipt.UpdateVariantRequest))
	}

	info := &grpc.UnaryServerInfo{
		FullMethod: "FakeMethod",
	}

	tp := sdktrace.NewTracerProvider(sdktrace.WithSampler(sdktrace.AlwaysSample()))
	tp.RegisterSpanProcessor(sdktrace.NewSimpleSpanProcessor(exporterSpy))

	tr := tp.Tracer("SpanProcessor")
	ctx, span := tr.Start(context.Background(), "OnStart")

	got, err := unaryInterceptor(ctx, req, info, handler)
	require.NoError(t, err)
	assert.NotNil(t, got)

	span.End()
	assert.Equal(t, exporterSpy.GetSendAuditsCalled(), 1)
}

func TestAuditUnaryInterceptor_DeleteVariant(t *testing.T) {
	var (
		store       = &storeMock{}
		logger      = zaptest.NewLogger(t)
		exporterSpy = newAuditExporterSpy(logger)
		s           = server.New(logger, store)
		req         = &flipt.DeleteVariantRequest{
			Id: "1",
		}
	)

	store.On("DeleteVariant", mock.Anything, req).Return(nil)

	unaryInterceptor := AuditUnaryInterceptor(logger)

	handler := func(ctx context.Context, r interface{}) (interface{}, error) {
		return s.DeleteVariant(ctx, r.(*flipt.DeleteVariantRequest))
	}

	info := &grpc.UnaryServerInfo{
		FullMethod: "FakeMethod",
	}

	tp := sdktrace.NewTracerProvider(sdktrace.WithSampler(sdktrace.AlwaysSample()))
	tp.RegisterSpanProcessor(sdktrace.NewSimpleSpanProcessor(exporterSpy))

	tr := tp.Tracer("SpanProcessor")
	ctx, span := tr.Start(context.Background(), "OnStart")

	got, err := unaryInterceptor(ctx, req, info, handler)
	require.NoError(t, err)
	assert.NotNil(t, got)

	span.End()
	assert.Equal(t, exporterSpy.GetSendAuditsCalled(), 1)
}

func TestAuditUnaryInterceptor_CreateDistribution(t *testing.T) {
	var (
		store       = &storeMock{}
		logger      = zaptest.NewLogger(t)
		exporterSpy = newAuditExporterSpy(logger)
		s           = server.New(logger, store)
		req         = &flipt.CreateDistributionRequest{
			FlagKey:   "flagKey",
			RuleId:    "1",
			VariantId: "2",
			Rollout:   25,
		}
	)

	store.On("CreateDistribution", mock.Anything, req).Return(&flipt.Distribution{
		Id:        "1",
		RuleId:    req.RuleId,
		VariantId: req.VariantId,
		Rollout:   req.Rollout,
	}, nil)

	unaryInterceptor := AuditUnaryInterceptor(logger)

	handler := func(ctx context.Context, r interface{}) (interface{}, error) {
		return s.CreateDistribution(ctx, r.(*flipt.CreateDistributionRequest))
	}

	info := &grpc.UnaryServerInfo{
		FullMethod: "FakeMethod",
	}

	tp := sdktrace.NewTracerProvider(sdktrace.WithSampler(sdktrace.AlwaysSample()))
	tp.RegisterSpanProcessor(sdktrace.NewSimpleSpanProcessor(exporterSpy))

	tr := tp.Tracer("SpanProcessor")
	ctx, span := tr.Start(context.Background(), "OnStart")

	got, err := unaryInterceptor(ctx, req, info, handler)
	require.NoError(t, err)
	assert.NotNil(t, got)

	span.End()
	assert.Equal(t, exporterSpy.GetSendAuditsCalled(), 1)
}

func TestAuditUnaryInterceptor_UpdateDistribution(t *testing.T) {
	var (
		store       = &storeMock{}
		logger      = zaptest.NewLogger(t)
		exporterSpy = newAuditExporterSpy(logger)
		s           = server.New(logger, store)
		req         = &flipt.UpdateDistributionRequest{
			Id:        "1",
			FlagKey:   "flagKey",
			RuleId:    "1",
			VariantId: "2",
			Rollout:   25,
		}
	)

	store.On("UpdateDistribution", mock.Anything, req).Return(&flipt.Distribution{
		Id:        req.Id,
		RuleId:    req.RuleId,
		VariantId: req.VariantId,
		Rollout:   req.Rollout,
	}, nil)

	unaryInterceptor := AuditUnaryInterceptor(logger)

	handler := func(ctx context.Context, r interface{}) (interface{}, error) {
		return s.UpdateDistribution(ctx, r.(*flipt.UpdateDistributionRequest))
	}

	info := &grpc.UnaryServerInfo{
		FullMethod: "FakeMethod",
	}

	tp := sdktrace.NewTracerProvider(sdktrace.WithSampler(sdktrace.AlwaysSample()))
	tp.RegisterSpanProcessor(sdktrace.NewSimpleSpanProcessor(exporterSpy))

	tr := tp.Tracer("SpanProcessor")
	ctx, span := tr.Start(context.Background(), "OnStart")

	got, err := unaryInterceptor(ctx, req, info, handler)
	require.NoError(t, err)
	assert.NotNil(t, got)

	span.End()
	assert.Equal(t, exporterSpy.GetSendAuditsCalled(), 1)
}

func TestAuditUnaryInterceptor_DeleteDistribution(t *testing.T) {
	var (
		store       = &storeMock{}
		logger      = zaptest.NewLogger(t)
		exporterSpy = newAuditExporterSpy(logger)
		s           = server.New(logger, store)
		req         = &flipt.DeleteDistributionRequest{
			Id:        "1",
			FlagKey:   "flagKey",
			RuleId:    "1",
			VariantId: "2",
		}
	)

	store.On("DeleteDistribution", mock.Anything, req).Return(nil)

	unaryInterceptor := AuditUnaryInterceptor(logger)

	handler := func(ctx context.Context, r interface{}) (interface{}, error) {
		return s.DeleteDistribution(ctx, r.(*flipt.DeleteDistributionRequest))
	}

	info := &grpc.UnaryServerInfo{
		FullMethod: "FakeMethod",
	}

	tp := sdktrace.NewTracerProvider(sdktrace.WithSampler(sdktrace.AlwaysSample()))
	tp.RegisterSpanProcessor(sdktrace.NewSimpleSpanProcessor(exporterSpy))

	tr := tp.Tracer("SpanProcessor")
	ctx, span := tr.Start(context.Background(), "OnStart")

	got, err := unaryInterceptor(ctx, req, info, handler)
	require.NoError(t, err)
	assert.NotNil(t, got)

	span.End()
	assert.Equal(t, exporterSpy.GetSendAuditsCalled(), 1)
}

func TestAuditUnaryInterceptor_CreateSegment(t *testing.T) {
	var (
		store       = &storeMock{}
		logger      = zaptest.NewLogger(t)
		exporterSpy = newAuditExporterSpy(logger)
		s           = server.New(logger, store)
		req         = &flipt.CreateSegmentRequest{
			Key:         "segmentkey",
			Name:        "segment",
			Description: "segment description",
			MatchType:   25,
		}
	)

	store.On("CreateSegment", mock.Anything, req).Return(&flipt.Segment{
		Key:         req.Key,
		Name:        req.Name,
		Description: req.Description,
		MatchType:   req.MatchType,
	}, nil)

	unaryInterceptor := AuditUnaryInterceptor(logger)

	handler := func(ctx context.Context, r interface{}) (interface{}, error) {
		return s.CreateSegment(ctx, r.(*flipt.CreateSegmentRequest))
	}

	info := &grpc.UnaryServerInfo{
		FullMethod: "FakeMethod",
	}

	tp := sdktrace.NewTracerProvider(sdktrace.WithSampler(sdktrace.AlwaysSample()))
	tp.RegisterSpanProcessor(sdktrace.NewSimpleSpanProcessor(exporterSpy))

	tr := tp.Tracer("SpanProcessor")
	ctx, span := tr.Start(context.Background(), "OnStart")

	got, err := unaryInterceptor(ctx, req, info, handler)
	require.NoError(t, err)
	assert.NotNil(t, got)

	span.End()
	assert.Equal(t, exporterSpy.GetSendAuditsCalled(), 1)
}

func TestAuditUnaryInterceptor_UpdateSegment(t *testing.T) {
	var (
		store       = &storeMock{}
		logger      = zaptest.NewLogger(t)
		exporterSpy = newAuditExporterSpy(logger)
		s           = server.New(logger, store)
		req         = &flipt.UpdateSegmentRequest{
			Key:         "segmentkey",
			Name:        "segment",
			Description: "segment description",
			MatchType:   25,
		}
	)

	store.On("UpdateSegment", mock.Anything, req).Return(&flipt.Segment{
		Key:         req.Key,
		Name:        req.Name,
		Description: req.Description,
		MatchType:   req.MatchType,
	}, nil)

	unaryInterceptor := AuditUnaryInterceptor(logger)

	handler := func(ctx context.Context, r interface{}) (interface{}, error) {
		return s.UpdateSegment(ctx, r.(*flipt.UpdateSegmentRequest))
	}

	info := &grpc.UnaryServerInfo{
		FullMethod: "FakeMethod",
	}

	tp := sdktrace.NewTracerProvider(sdktrace.WithSampler(sdktrace.AlwaysSample()))
	tp.RegisterSpanProcessor(sdktrace.NewSimpleSpanProcessor(exporterSpy))

	tr := tp.Tracer("SpanProcessor")
	ctx, span := tr.Start(context.Background(), "OnStart")

	got, err := unaryInterceptor(ctx, req, info, handler)
	require.NoError(t, err)
	assert.NotNil(t, got)

	span.End()
	assert.Equal(t, exporterSpy.GetSendAuditsCalled(), 1)
}

func TestAuditUnaryInterceptor_DeleteSegment(t *testing.T) {
	var (
		store       = &storeMock{}
		logger      = zaptest.NewLogger(t)
		exporterSpy = newAuditExporterSpy(logger)
		s           = server.New(logger, store)
		req         = &flipt.DeleteSegmentRequest{
			Key: "segment",
		}
	)

	store.On("DeleteSegment", mock.Anything, req).Return(nil)

	unaryInterceptor := AuditUnaryInterceptor(logger)

	handler := func(ctx context.Context, r interface{}) (interface{}, error) {
		return s.DeleteSegment(ctx, r.(*flipt.DeleteSegmentRequest))
	}

	info := &grpc.UnaryServerInfo{
		FullMethod: "FakeMethod",
	}

	tp := sdktrace.NewTracerProvider(sdktrace.WithSampler(sdktrace.AlwaysSample()))
	tp.RegisterSpanProcessor(sdktrace.NewSimpleSpanProcessor(exporterSpy))

	tr := tp.Tracer("SpanProcessor")
	ctx, span := tr.Start(context.Background(), "OnStart")

	got, err := unaryInterceptor(ctx, req, info, handler)
	require.NoError(t, err)
	assert.NotNil(t, got)

	span.End()
	assert.Equal(t, exporterSpy.GetSendAuditsCalled(), 1)
}

func TestAuditUnaryInterceptor_CreateConstraint(t *testing.T) {
	var (
		store       = &storeMock{}
		logger      = zaptest.NewLogger(t)
		exporterSpy = newAuditExporterSpy(logger)
		s           = server.New(logger, store)
		req         = &flipt.CreateConstraintRequest{
			SegmentKey: "constraintsegmentkey",
			Type:       32,
			Property:   "constraintproperty",
			Operator:   "eq",
			Value:      "thisvalue",
		}
	)

	store.On("CreateConstraint", mock.Anything, req).Return(&flipt.Constraint{
		Id:         "1",
		SegmentKey: req.SegmentKey,
		Type:       req.Type,
		Property:   req.Property,
		Operator:   req.Operator,
		Value:      req.Value,
	}, nil)

	unaryInterceptor := AuditUnaryInterceptor(logger)

	handler := func(ctx context.Context, r interface{}) (interface{}, error) {
		return s.CreateConstraint(ctx, r.(*flipt.CreateConstraintRequest))
	}

	info := &grpc.UnaryServerInfo{
		FullMethod: "FakeMethod",
	}

	tp := sdktrace.NewTracerProvider(sdktrace.WithSampler(sdktrace.AlwaysSample()))
	tp.RegisterSpanProcessor(sdktrace.NewSimpleSpanProcessor(exporterSpy))

	tr := tp.Tracer("SpanProcessor")
	ctx, span := tr.Start(context.Background(), "OnStart")

	got, err := unaryInterceptor(ctx, req, info, handler)
	require.NoError(t, err)
	assert.NotNil(t, got)

	span.End()
	assert.Equal(t, exporterSpy.GetSendAuditsCalled(), 1)
}

func TestAuditUnaryInterceptor_UpdateConstraint(t *testing.T) {
	var (
		store       = &storeMock{}
		logger      = zaptest.NewLogger(t)
		exporterSpy = newAuditExporterSpy(logger)
		s           = server.New(logger, store)
		req         = &flipt.UpdateConstraintRequest{
			Id:         "1",
			SegmentKey: "constraintsegmentkey",
			Type:       32,
			Property:   "constraintproperty",
			Operator:   "eq",
			Value:      "thisvalue",
		}
	)

	store.On("UpdateConstraint", mock.Anything, req).Return(&flipt.Constraint{
		Id:         "1",
		SegmentKey: req.SegmentKey,
		Type:       req.Type,
		Property:   req.Property,
		Operator:   req.Operator,
		Value:      req.Value,
	}, nil)

	unaryInterceptor := AuditUnaryInterceptor(logger)

	handler := func(ctx context.Context, r interface{}) (interface{}, error) {
		return s.UpdateConstraint(ctx, r.(*flipt.UpdateConstraintRequest))
	}

	info := &grpc.UnaryServerInfo{
		FullMethod: "FakeMethod",
	}

	tp := sdktrace.NewTracerProvider(sdktrace.WithSampler(sdktrace.AlwaysSample()))
	tp.RegisterSpanProcessor(sdktrace.NewSimpleSpanProcessor(exporterSpy))

	tr := tp.Tracer("SpanProcessor")
	ctx, span := tr.Start(context.Background(), "OnStart")

	got, err := unaryInterceptor(ctx, req, info, handler)
	require.NoError(t, err)
	assert.NotNil(t, got)

	span.End()
	assert.Equal(t, exporterSpy.GetSendAuditsCalled(), 1)
}

func TestAuditUnaryInterceptor_DeleteConstraint(t *testing.T) {
	var (
		store       = &storeMock{}
		logger      = zaptest.NewLogger(t)
		exporterSpy = newAuditExporterSpy(logger)
		s           = server.New(logger, store)
		req         = &flipt.DeleteConstraintRequest{
			Id:         "1",
			SegmentKey: "constraintsegmentkey",
		}
	)

	store.On("DeleteConstraint", mock.Anything, req).Return(nil)

	unaryInterceptor := AuditUnaryInterceptor(logger)

	handler := func(ctx context.Context, r interface{}) (interface{}, error) {
		return s.DeleteConstraint(ctx, r.(*flipt.DeleteConstraintRequest))
	}

	info := &grpc.UnaryServerInfo{
		FullMethod: "FakeMethod",
	}

	tp := sdktrace.NewTracerProvider(sdktrace.WithSampler(sdktrace.AlwaysSample()))
	tp.RegisterSpanProcessor(sdktrace.NewSimpleSpanProcessor(exporterSpy))

	tr := tp.Tracer("SpanProcessor")
	ctx, span := tr.Start(context.Background(), "OnStart")

	got, err := unaryInterceptor(ctx, req, info, handler)
	require.NoError(t, err)
	assert.NotNil(t, got)

	span.End()
	assert.Equal(t, exporterSpy.GetSendAuditsCalled(), 1)
}

func TestAuditUnaryInterceptor_CreateRule(t *testing.T) {
	var (
		store       = &storeMock{}
		logger      = zaptest.NewLogger(t)
		exporterSpy = newAuditExporterSpy(logger)
		s           = server.New(logger, store)
		req         = &flipt.CreateRuleRequest{
			FlagKey:    "flagkey",
			SegmentKey: "segmentkey",
			Rank:       1,
		}
	)

	store.On("CreateRule", mock.Anything, req).Return(&flipt.Rule{
		Id:         "1",
		SegmentKey: req.SegmentKey,
		FlagKey:    req.FlagKey,
	}, nil)

	unaryInterceptor := AuditUnaryInterceptor(logger)

	handler := func(ctx context.Context, r interface{}) (interface{}, error) {
		return s.CreateRule(ctx, r.(*flipt.CreateRuleRequest))
	}

	info := &grpc.UnaryServerInfo{
		FullMethod: "FakeMethod",
	}

	tp := sdktrace.NewTracerProvider(sdktrace.WithSampler(sdktrace.AlwaysSample()))
	tp.RegisterSpanProcessor(sdktrace.NewSimpleSpanProcessor(exporterSpy))

	tr := tp.Tracer("SpanProcessor")
	ctx, span := tr.Start(context.Background(), "OnStart")

	got, err := unaryInterceptor(ctx, req, info, handler)
	require.NoError(t, err)
	assert.NotNil(t, got)

	span.End()
	assert.Equal(t, exporterSpy.GetSendAuditsCalled(), 1)
}

func TestAuditUnaryInterceptor_UpdateRule(t *testing.T) {
	var (
		store       = &storeMock{}
		logger      = zaptest.NewLogger(t)
		exporterSpy = newAuditExporterSpy(logger)
		s           = server.New(logger, store)
		req         = &flipt.UpdateRuleRequest{
			Id:         "1",
			FlagKey:    "flagkey",
			SegmentKey: "segmentkey",
		}
	)

	store.On("UpdateRule", mock.Anything, req).Return(&flipt.Rule{
		Id:         "1",
		SegmentKey: req.SegmentKey,
		FlagKey:    req.FlagKey,
	}, nil)

	unaryInterceptor := AuditUnaryInterceptor(logger)

	handler := func(ctx context.Context, r interface{}) (interface{}, error) {
		return s.UpdateRule(ctx, r.(*flipt.UpdateRuleRequest))
	}

	info := &grpc.UnaryServerInfo{
		FullMethod: "FakeMethod",
	}

	tp := sdktrace.NewTracerProvider(sdktrace.WithSampler(sdktrace.AlwaysSample()))
	tp.RegisterSpanProcessor(sdktrace.NewSimpleSpanProcessor(exporterSpy))

	tr := tp.Tracer("SpanProcessor")
	ctx, span := tr.Start(context.Background(), "OnStart")

	got, err := unaryInterceptor(ctx, req, info, handler)
	require.NoError(t, err)
	assert.NotNil(t, got)

	span.End()
	assert.Equal(t, exporterSpy.GetSendAuditsCalled(), 1)
}

func TestAuditUnaryInterceptor_DeleteRule(t *testing.T) {
	var (
		store       = &storeMock{}
		logger      = zaptest.NewLogger(t)
		exporterSpy = newAuditExporterSpy(logger)
		s           = server.New(logger, store)
		req         = &flipt.DeleteRuleRequest{
			Id:      "1",
			FlagKey: "flagkey",
		}
	)

	store.On("DeleteRule", mock.Anything, req).Return(nil)

	unaryInterceptor := AuditUnaryInterceptor(logger)

	handler := func(ctx context.Context, r interface{}) (interface{}, error) {
		return s.DeleteRule(ctx, r.(*flipt.DeleteRuleRequest))
	}

	info := &grpc.UnaryServerInfo{
		FullMethod: "FakeMethod",
	}

	tp := sdktrace.NewTracerProvider(sdktrace.WithSampler(sdktrace.AlwaysSample()))
	tp.RegisterSpanProcessor(sdktrace.NewSimpleSpanProcessor(exporterSpy))

	tr := tp.Tracer("SpanProcessor")
	ctx, span := tr.Start(context.Background(), "OnStart")

	got, err := unaryInterceptor(ctx, req, info, handler)
	require.NoError(t, err)
	assert.NotNil(t, got)

	span.End()
	assert.Equal(t, exporterSpy.GetSendAuditsCalled(), 1)
}

func TestAuditUnaryInterceptor_CreateNamespace(t *testing.T) {
	var (
		store       = &storeMock{}
		logger      = zaptest.NewLogger(t)
		exporterSpy = newAuditExporterSpy(logger)
		s           = server.New(logger, store)
		req         = &flipt.CreateNamespaceRequest{
			Key:  "namespacekey",
			Name: "namespaceKey",
		}
	)

	store.On("CreateNamespace", mock.Anything, req).Return(&flipt.Namespace{
		Key:  req.Key,
		Name: req.Name,
	}, nil)

	unaryInterceptor := AuditUnaryInterceptor(logger)

	handler := func(ctx context.Context, r interface{}) (interface{}, error) {
		return s.CreateNamespace(ctx, r.(*flipt.CreateNamespaceRequest))
	}

	info := &grpc.UnaryServerInfo{
		FullMethod: "FakeMethod",
	}

	tp := sdktrace.NewTracerProvider(sdktrace.WithSampler(sdktrace.AlwaysSample()))
	tp.RegisterSpanProcessor(sdktrace.NewSimpleSpanProcessor(exporterSpy))

	tr := tp.Tracer("SpanProcessor")
	ctx, span := tr.Start(context.Background(), "OnStart")

	got, err := unaryInterceptor(ctx, req, info, handler)
	require.NoError(t, err)
	assert.NotNil(t, got)

	span.End()
	assert.Equal(t, exporterSpy.GetSendAuditsCalled(), 1)
}

func TestAuditUnaryInterceptor_UpdateNamespace(t *testing.T) {
	var (
		store       = &storeMock{}
		logger      = zaptest.NewLogger(t)
		exporterSpy = newAuditExporterSpy(logger)
		s           = server.New(logger, store)
		req         = &flipt.UpdateNamespaceRequest{
			Key:         "namespacekey",
			Name:        "namespaceKey",
			Description: "namespace description",
		}
	)

	store.On("UpdateNamespace", mock.Anything, req).Return(&flipt.Namespace{
		Key:         req.Key,
		Name:        req.Name,
		Description: req.Description,
	}, nil)

	unaryInterceptor := AuditUnaryInterceptor(logger)

	handler := func(ctx context.Context, r interface{}) (interface{}, error) {
		return s.UpdateNamespace(ctx, r.(*flipt.UpdateNamespaceRequest))
	}

	info := &grpc.UnaryServerInfo{
		FullMethod: "FakeMethod",
	}

	tp := sdktrace.NewTracerProvider(sdktrace.WithSampler(sdktrace.AlwaysSample()))
	tp.RegisterSpanProcessor(sdktrace.NewSimpleSpanProcessor(exporterSpy))

	tr := tp.Tracer("SpanProcessor")
	ctx, span := tr.Start(context.Background(), "OnStart")

	got, err := unaryInterceptor(ctx, req, info, handler)
	require.NoError(t, err)
	assert.NotNil(t, got)

	span.End()
	assert.Equal(t, exporterSpy.GetSendAuditsCalled(), 1)
}

func TestAuditUnaryInterceptor_DeleteNamespace(t *testing.T) {
	var (
		store       = &storeMock{}
		logger      = zaptest.NewLogger(t)
		exporterSpy = newAuditExporterSpy(logger)
		s           = server.New(logger, store)
		req         = &flipt.DeleteNamespaceRequest{
			Key: "namespacekey",
		}
	)

	store.On("GetNamespace", mock.Anything, req.Key).Return(&flipt.Namespace{
		Key: req.Key,
	}, nil)

	store.On("CountFlags", mock.Anything, req.Key).Return(uint64(0), nil)

	store.On("DeleteNamespace", mock.Anything, req).Return(nil)

	unaryInterceptor := AuditUnaryInterceptor(logger)

	handler := func(ctx context.Context, r interface{}) (interface{}, error) {
		return s.DeleteNamespace(ctx, r.(*flipt.DeleteNamespaceRequest))
	}

	info := &grpc.UnaryServerInfo{
		FullMethod: "FakeMethod",
	}

	tp := sdktrace.NewTracerProvider(sdktrace.WithSampler(sdktrace.AlwaysSample()))
	tp.RegisterSpanProcessor(sdktrace.NewSimpleSpanProcessor(exporterSpy))

	tr := tp.Tracer("SpanProcessor")
	ctx, span := tr.Start(context.Background(), "OnStart")

	got, err := unaryInterceptor(ctx, req, info, handler)
	require.NoError(t, err)
	assert.NotNil(t, got)

	span.End()
	assert.Equal(t, exporterSpy.GetSendAuditsCalled(), 1)
}

func TestAuthMetadataAuditUnaryInterceptor(t *testing.T) {
	var (
		store       = &storeMock{}
		logger      = zaptest.NewLogger(t)
		exporterSpy = newAuditExporterSpy(logger)
		s           = server.New(logger, store)
		req         = &flipt.CreateFlagRequest{
			Key:         "key",
			Name:        "name",
			Description: "desc",
		}
	)

	store.On("CreateFlag", mock.Anything, req).Return(&flipt.Flag{
		Key:         req.Key,
		Name:        req.Name,
		Description: req.Description,
	}, nil)

	unaryInterceptor := AuditUnaryInterceptor(logger)

	handler := func(ctx context.Context, r interface{}) (interface{}, error) {
		return s.CreateFlag(ctx, r.(*flipt.CreateFlagRequest))
	}

	info := &grpc.UnaryServerInfo{
		FullMethod: "FakeMethod",
	}

	tp := sdktrace.NewTracerProvider(sdktrace.WithSampler(sdktrace.AlwaysSample()))
	tp.RegisterSpanProcessor(sdktrace.NewSimpleSpanProcessor(exporterSpy))

	tr := tp.Tracer("SpanProcessor")
	ctx, span := tr.Start(context.Background(), "OnStart")

	ctx = auth.ContextWithAuthentication(ctx, &authrpc.Authentication{
		Method: authrpc.Method_METHOD_OIDC,
		Metadata: map[string]string{
			"email": "example@flipt.com",
		},
	})

	got, err := unaryInterceptor(ctx, req, info, handler)
	require.NoError(t, err)
	assert.NotNil(t, got)

	span.End()

	event := exporterSpy.GetEvents()[0]
	assert.Equal(t, event.Metadata.Actor["email"], "example@flipt.com")
	assert.Equal(t, event.Metadata.Actor["method"], "oidc")
}
