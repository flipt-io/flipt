package grpc_middleware

import (
	"context"
	"fmt"
	"net/http/httptest"
	"testing"

	"go.flipt.io/flipt/errors"
	"go.flipt.io/flipt/internal/common"
	"go.flipt.io/flipt/internal/server"
	"go.flipt.io/flipt/internal/server/authn/method/token"
	authmiddlewaregrpc "go.flipt.io/flipt/internal/server/authn/middleware/grpc"
	"go.flipt.io/flipt/internal/storage"
	flipt "go.flipt.io/flipt/rpc/flipt"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.uber.org/zap/zaptest"

	"github.com/blang/semver/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	storageauth "go.flipt.io/flipt/internal/storage/authn"
	authrpc "go.flipt.io/flipt/rpc/flipt/auth"
	"go.flipt.io/flipt/rpc/flipt/evaluation"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
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
			name:     "deadline exceeded error",
			wantErr:  fmt.Errorf("foo: %w", context.DeadlineExceeded),
			wantCode: codes.DeadlineExceeded,
		},
		{
			name:     "context cancelled error",
			wantErr:  fmt.Errorf("foo: %w", context.Canceled),
			wantCode: codes.Canceled,
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
			wantErr:  errors.ErrUnauthenticatedf("user %q not found", "foo"),
			wantCode: codes.Unauthenticated,
		},
		{
			name:     "unauthorized error",
			wantErr:  errors.ErrUnauthorizedf("user %q not authorized, missing role", "foo"),
			wantCode: codes.PermissionDenied,
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

	got, err := EvaluationUnaryInterceptor(false)(context.Background(), req, info, handler)
	require.NoError(t, err)

	assert.NotNil(t, got)

	resp, ok := got.(*flipt.Flag)
	assert.True(t, ok)
	assert.NotNil(t, resp)
	assert.Equal(t, "foo", resp.Key)
}

func TestEvaluationUnaryInterceptor_Evaluation(t *testing.T) {
	type request interface {
		GetRequestId() string
	}

	type response interface {
		request
		GetTimestamp() *timestamppb.Timestamp
		GetRequestDurationMillis() float64
	}

	for _, test := range []struct {
		name      string
		requestID string
		req       request
		resp      response
	}{
		{
			name:      "flipt.EvaluationRequest without request ID",
			requestID: "",
			req: &flipt.EvaluationRequest{
				FlagKey: "foo",
			},
			resp: &flipt.EvaluationResponse{
				FlagKey: "foo",
			},
		},
		{
			name:      "flipt.EvaluationRequest with request ID",
			requestID: "bar",
			req: &flipt.EvaluationRequest{
				FlagKey:   "foo",
				RequestId: "bar",
			},
			resp: &flipt.EvaluationResponse{
				FlagKey: "foo",
			},
		},
		{
			name:      "Variant evaluation.EvaluationRequest without request ID",
			requestID: "",
			req: &evaluation.EvaluationRequest{
				FlagKey: "foo",
			},
			resp: &evaluation.EvaluationResponse{
				Response: &evaluation.EvaluationResponse_VariantResponse{
					VariantResponse: &evaluation.VariantEvaluationResponse{},
				},
			},
		},
		{
			name:      "Boolean evaluation.EvaluationRequest with request ID",
			requestID: "baz",
			req: &evaluation.EvaluationRequest{
				FlagKey:   "foo",
				RequestId: "baz",
			},
			resp: &evaluation.EvaluationResponse{
				Response: &evaluation.EvaluationResponse_BooleanResponse{
					BooleanResponse: &evaluation.BooleanEvaluationResponse{},
				},
			},
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			var (
				handler = func(ctx context.Context, r interface{}) (interface{}, error) {
					// ensure request has ID once it reaches the handler
					req, ok := r.(request)
					require.True(t, ok)

					if test.requestID == "" {
						assert.NotEmpty(t, req.GetRequestId())
						return test.resp, nil
					}

					assert.Equal(t, test.requestID, req.GetRequestId())

					return test.resp, nil
				}

				info = &grpc.UnaryServerInfo{
					FullMethod: "FakeMethod",
				}
			)

			got, err := EvaluationUnaryInterceptor(true)(context.Background(), test.req, info, handler)
			require.NoError(t, err)

			assert.NotNil(t, got)

			resp, ok := got.(response)
			assert.True(t, ok)
			assert.NotNil(t, resp)

			// check that the requestID is either non-empty
			// or explicitly what was provided for the test
			if test.requestID == "" {
				assert.NotEmpty(t, resp.GetRequestId())
			} else {
				assert.Equal(t, test.requestID, resp.GetRequestId())

			}

			assert.NotZero(t, resp.GetTimestamp())
			assert.NotZero(t, resp.GetRequestDurationMillis())
		})
	}
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

	got, err := EvaluationUnaryInterceptor(false)(context.Background(), req, info, handler)
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

	got, err = EvaluationUnaryInterceptor(false)(context.Background(), req, info, handler)
	require.NoError(t, err)

	assert.NotNil(t, got)

	resp, ok = got.(*flipt.BatchEvaluationResponse)
	assert.True(t, ok)
	assert.NotNil(t, resp)
	assert.NotEmpty(t, resp.Responses)
	assert.Equal(t, "foo", resp.Responses[0].FlagKey)
	assert.NotNil(t, resp.Responses[0].Timestamp)
	// check that the requestID was propagated
	assert.NotEmpty(t, resp.RequestId)
	assert.Equal(t, "bar", resp.RequestId)

	// TODO(yquansah): flakey assertion
	// assert.NotZero(t, resp.RequestDurationMillis)
}

type checkerDummy struct{}

func (c *checkerDummy) Check(e string) bool {
	return true
}

func (c *checkerDummy) Events() []string {
	return []string{"event"}
}

func TestAuditUnaryInterceptor_CreateFlag(t *testing.T) {
	var (
		store       = &common.StoreMock{}
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

	unaryInterceptor := AuditEventUnaryInterceptor(logger, &checkerDummy{})

	handler := func(ctx context.Context, r interface{}) (interface{}, error) {
		return s.CreateFlag(ctx, r.(*flipt.CreateFlagRequest))
	}

	info := &grpc.UnaryServerInfo{
		FullMethod: "CreateFlag",
	}

	tp := sdktrace.NewTracerProvider(sdktrace.WithSampler(sdktrace.AlwaysSample()))
	tp.RegisterSpanProcessor(sdktrace.NewSimpleSpanProcessor(exporterSpy))

	tr := tp.Tracer("SpanProcessor")
	ctx, span := tr.Start(context.Background(), "OnStart")

	got, err := unaryInterceptor(ctx, req, info, handler)
	require.NoError(t, err)
	assert.NotNil(t, got)

	span.End()

	assert.Equal(t, 1, exporterSpy.GetSendAuditsCalled())
}

func TestAuditUnaryInterceptor_UpdateFlag(t *testing.T) {
	var (
		store       = &common.StoreMock{}
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

	unaryInterceptor := AuditEventUnaryInterceptor(logger, &checkerDummy{})

	handler := func(ctx context.Context, r interface{}) (interface{}, error) {
		return s.UpdateFlag(ctx, r.(*flipt.UpdateFlagRequest))
	}

	info := &grpc.UnaryServerInfo{
		FullMethod: "UpdateFlag",
	}

	tp := sdktrace.NewTracerProvider(sdktrace.WithSampler(sdktrace.AlwaysSample()))
	tp.RegisterSpanProcessor(sdktrace.NewSimpleSpanProcessor(exporterSpy))

	tr := tp.Tracer("SpanProcessor")
	ctx, span := tr.Start(context.Background(), "OnStart")

	got, err := unaryInterceptor(ctx, req, info, handler)
	require.NoError(t, err)
	assert.NotNil(t, got)

	span.End()

	assert.Equal(t, 1, exporterSpy.GetSendAuditsCalled())
}

func TestAuditUnaryInterceptor_DeleteFlag(t *testing.T) {
	var (
		store       = &common.StoreMock{}
		logger      = zaptest.NewLogger(t)
		exporterSpy = newAuditExporterSpy(logger)
		s           = server.New(logger, store)
		req         = &flipt.DeleteFlagRequest{
			Key: "key",
		}
	)

	store.On("DeleteFlag", mock.Anything, req).Return(nil)

	unaryInterceptor := AuditEventUnaryInterceptor(logger, &checkerDummy{})

	handler := func(ctx context.Context, r interface{}) (interface{}, error) {
		return s.DeleteFlag(ctx, r.(*flipt.DeleteFlagRequest))
	}

	info := &grpc.UnaryServerInfo{
		FullMethod: "DeleteFlag",
	}

	tp := sdktrace.NewTracerProvider(sdktrace.WithSampler(sdktrace.AlwaysSample()))
	tp.RegisterSpanProcessor(sdktrace.NewSimpleSpanProcessor(exporterSpy))

	tr := tp.Tracer("SpanProcessor")
	ctx, span := tr.Start(context.Background(), "OnStart")

	got, err := unaryInterceptor(ctx, req, info, handler)
	require.NoError(t, err)
	assert.NotNil(t, got)

	span.End()
	assert.Equal(t, 1, exporterSpy.GetSendAuditsCalled())
}

func TestAuditUnaryInterceptor_CreateVariant(t *testing.T) {
	var (
		store       = &common.StoreMock{}
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

	unaryInterceptor := AuditEventUnaryInterceptor(logger, &checkerDummy{})

	handler := func(ctx context.Context, r interface{}) (interface{}, error) {
		return s.CreateVariant(ctx, r.(*flipt.CreateVariantRequest))
	}

	info := &grpc.UnaryServerInfo{
		FullMethod: "CreateVariant",
	}

	tp := sdktrace.NewTracerProvider(sdktrace.WithSampler(sdktrace.AlwaysSample()))
	tp.RegisterSpanProcessor(sdktrace.NewSimpleSpanProcessor(exporterSpy))

	tr := tp.Tracer("SpanProcessor")
	ctx, span := tr.Start(context.Background(), "OnStart")

	got, err := unaryInterceptor(ctx, req, info, handler)
	require.NoError(t, err)
	assert.NotNil(t, got)

	span.End()
	assert.Equal(t, 1, exporterSpy.GetSendAuditsCalled())
}

func TestAuditUnaryInterceptor_UpdateVariant(t *testing.T) {
	var (
		store       = &common.StoreMock{}
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

	unaryInterceptor := AuditEventUnaryInterceptor(logger, &checkerDummy{})

	handler := func(ctx context.Context, r interface{}) (interface{}, error) {
		return s.UpdateVariant(ctx, r.(*flipt.UpdateVariantRequest))
	}

	info := &grpc.UnaryServerInfo{
		FullMethod: "UpdateVariant",
	}

	tp := sdktrace.NewTracerProvider(sdktrace.WithSampler(sdktrace.AlwaysSample()))
	tp.RegisterSpanProcessor(sdktrace.NewSimpleSpanProcessor(exporterSpy))

	tr := tp.Tracer("SpanProcessor")
	ctx, span := tr.Start(context.Background(), "OnStart")

	got, err := unaryInterceptor(ctx, req, info, handler)
	require.NoError(t, err)
	assert.NotNil(t, got)

	span.End()
	assert.Equal(t, 1, exporterSpy.GetSendAuditsCalled())
}

func TestAuditUnaryInterceptor_DeleteVariant(t *testing.T) {
	var (
		store       = &common.StoreMock{}
		logger      = zaptest.NewLogger(t)
		exporterSpy = newAuditExporterSpy(logger)
		s           = server.New(logger, store)
		req         = &flipt.DeleteVariantRequest{
			Id: "1",
		}
	)

	store.On("DeleteVariant", mock.Anything, req).Return(nil)

	unaryInterceptor := AuditEventUnaryInterceptor(logger, &checkerDummy{})

	handler := func(ctx context.Context, r interface{}) (interface{}, error) {
		return s.DeleteVariant(ctx, r.(*flipt.DeleteVariantRequest))
	}

	info := &grpc.UnaryServerInfo{
		FullMethod: "DeleteVariant",
	}

	tp := sdktrace.NewTracerProvider(sdktrace.WithSampler(sdktrace.AlwaysSample()))
	tp.RegisterSpanProcessor(sdktrace.NewSimpleSpanProcessor(exporterSpy))

	tr := tp.Tracer("SpanProcessor")
	ctx, span := tr.Start(context.Background(), "OnStart")

	got, err := unaryInterceptor(ctx, req, info, handler)
	require.NoError(t, err)
	assert.NotNil(t, got)

	span.End()
	assert.Equal(t, 1, exporterSpy.GetSendAuditsCalled())
}

func TestAuditUnaryInterceptor_CreateDistribution(t *testing.T) {
	var (
		store       = &common.StoreMock{}
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

	unaryInterceptor := AuditEventUnaryInterceptor(logger, &checkerDummy{})

	handler := func(ctx context.Context, r interface{}) (interface{}, error) {
		return s.CreateDistribution(ctx, r.(*flipt.CreateDistributionRequest))
	}

	info := &grpc.UnaryServerInfo{
		FullMethod: "CreateDistribution",
	}

	tp := sdktrace.NewTracerProvider(sdktrace.WithSampler(sdktrace.AlwaysSample()))
	tp.RegisterSpanProcessor(sdktrace.NewSimpleSpanProcessor(exporterSpy))

	tr := tp.Tracer("SpanProcessor")
	ctx, span := tr.Start(context.Background(), "OnStart")

	got, err := unaryInterceptor(ctx, req, info, handler)
	require.NoError(t, err)
	assert.NotNil(t, got)

	span.End()
	assert.Equal(t, 1, exporterSpy.GetSendAuditsCalled())
}

func TestAuditUnaryInterceptor_UpdateDistribution(t *testing.T) {
	var (
		store       = &common.StoreMock{}
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

	unaryInterceptor := AuditEventUnaryInterceptor(logger, &checkerDummy{})

	handler := func(ctx context.Context, r interface{}) (interface{}, error) {
		return s.UpdateDistribution(ctx, r.(*flipt.UpdateDistributionRequest))
	}

	info := &grpc.UnaryServerInfo{
		FullMethod: "UpdateDistribution",
	}

	tp := sdktrace.NewTracerProvider(sdktrace.WithSampler(sdktrace.AlwaysSample()))
	tp.RegisterSpanProcessor(sdktrace.NewSimpleSpanProcessor(exporterSpy))

	tr := tp.Tracer("SpanProcessor")
	ctx, span := tr.Start(context.Background(), "OnStart")

	got, err := unaryInterceptor(ctx, req, info, handler)
	require.NoError(t, err)
	assert.NotNil(t, got)

	span.End()
	assert.Equal(t, 1, exporterSpy.GetSendAuditsCalled())
}

func TestAuditUnaryInterceptor_DeleteDistribution(t *testing.T) {
	var (
		store       = &common.StoreMock{}
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

	unaryInterceptor := AuditEventUnaryInterceptor(logger, &checkerDummy{})

	handler := func(ctx context.Context, r interface{}) (interface{}, error) {
		return s.DeleteDistribution(ctx, r.(*flipt.DeleteDistributionRequest))
	}

	info := &grpc.UnaryServerInfo{
		FullMethod: "DeleteDistribution",
	}

	tp := sdktrace.NewTracerProvider(sdktrace.WithSampler(sdktrace.AlwaysSample()))
	tp.RegisterSpanProcessor(sdktrace.NewSimpleSpanProcessor(exporterSpy))

	tr := tp.Tracer("SpanProcessor")
	ctx, span := tr.Start(context.Background(), "OnStart")

	got, err := unaryInterceptor(ctx, req, info, handler)
	require.NoError(t, err)
	assert.NotNil(t, got)

	span.End()
	assert.Equal(t, 1, exporterSpy.GetSendAuditsCalled())
}

func TestAuditUnaryInterceptor_CreateSegment(t *testing.T) {
	var (
		store       = &common.StoreMock{}
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

	unaryInterceptor := AuditEventUnaryInterceptor(logger, &checkerDummy{})

	handler := func(ctx context.Context, r interface{}) (interface{}, error) {
		return s.CreateSegment(ctx, r.(*flipt.CreateSegmentRequest))
	}

	info := &grpc.UnaryServerInfo{
		FullMethod: "CreateSegment",
	}

	tp := sdktrace.NewTracerProvider(sdktrace.WithSampler(sdktrace.AlwaysSample()))
	tp.RegisterSpanProcessor(sdktrace.NewSimpleSpanProcessor(exporterSpy))

	tr := tp.Tracer("SpanProcessor")
	ctx, span := tr.Start(context.Background(), "OnStart")

	got, err := unaryInterceptor(ctx, req, info, handler)
	require.NoError(t, err)
	assert.NotNil(t, got)

	span.End()
	assert.Equal(t, 1, exporterSpy.GetSendAuditsCalled())
}

func TestAuditUnaryInterceptor_UpdateSegment(t *testing.T) {
	var (
		store       = &common.StoreMock{}
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

	unaryInterceptor := AuditEventUnaryInterceptor(logger, &checkerDummy{})

	handler := func(ctx context.Context, r interface{}) (interface{}, error) {
		return s.UpdateSegment(ctx, r.(*flipt.UpdateSegmentRequest))
	}

	info := &grpc.UnaryServerInfo{
		FullMethod: "UpdateSegment",
	}

	tp := sdktrace.NewTracerProvider(sdktrace.WithSampler(sdktrace.AlwaysSample()))
	tp.RegisterSpanProcessor(sdktrace.NewSimpleSpanProcessor(exporterSpy))

	tr := tp.Tracer("SpanProcessor")
	ctx, span := tr.Start(context.Background(), "OnStart")

	got, err := unaryInterceptor(ctx, req, info, handler)
	require.NoError(t, err)
	assert.NotNil(t, got)

	span.End()
	assert.Equal(t, 1, exporterSpy.GetSendAuditsCalled())
}

func TestAuditUnaryInterceptor_DeleteSegment(t *testing.T) {
	var (
		store       = &common.StoreMock{}
		logger      = zaptest.NewLogger(t)
		exporterSpy = newAuditExporterSpy(logger)
		s           = server.New(logger, store)
		req         = &flipt.DeleteSegmentRequest{
			Key: "segment",
		}
	)

	store.On("DeleteSegment", mock.Anything, req).Return(nil)

	unaryInterceptor := AuditEventUnaryInterceptor(logger, &checkerDummy{})

	handler := func(ctx context.Context, r interface{}) (interface{}, error) {
		return s.DeleteSegment(ctx, r.(*flipt.DeleteSegmentRequest))
	}

	info := &grpc.UnaryServerInfo{
		FullMethod: "DeleteSegment",
	}

	tp := sdktrace.NewTracerProvider(sdktrace.WithSampler(sdktrace.AlwaysSample()))
	tp.RegisterSpanProcessor(sdktrace.NewSimpleSpanProcessor(exporterSpy))

	tr := tp.Tracer("SpanProcessor")
	ctx, span := tr.Start(context.Background(), "OnStart")

	got, err := unaryInterceptor(ctx, req, info, handler)
	require.NoError(t, err)
	assert.NotNil(t, got)

	span.End()
	assert.Equal(t, 1, exporterSpy.GetSendAuditsCalled())
}

func TestAuditUnaryInterceptor_CreateConstraint(t *testing.T) {
	var (
		store       = &common.StoreMock{}
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

	unaryInterceptor := AuditEventUnaryInterceptor(logger, &checkerDummy{})

	handler := func(ctx context.Context, r interface{}) (interface{}, error) {
		return s.CreateConstraint(ctx, r.(*flipt.CreateConstraintRequest))
	}

	info := &grpc.UnaryServerInfo{
		FullMethod: "CreateConstraint",
	}

	tp := sdktrace.NewTracerProvider(sdktrace.WithSampler(sdktrace.AlwaysSample()))
	tp.RegisterSpanProcessor(sdktrace.NewSimpleSpanProcessor(exporterSpy))

	tr := tp.Tracer("SpanProcessor")
	ctx, span := tr.Start(context.Background(), "OnStart")

	got, err := unaryInterceptor(ctx, req, info, handler)
	require.NoError(t, err)
	assert.NotNil(t, got)

	span.End()
	assert.Equal(t, 1, exporterSpy.GetSendAuditsCalled())
}

func TestAuditUnaryInterceptor_UpdateConstraint(t *testing.T) {
	var (
		store       = &common.StoreMock{}
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

	unaryInterceptor := AuditEventUnaryInterceptor(logger, &checkerDummy{})

	handler := func(ctx context.Context, r interface{}) (interface{}, error) {
		return s.UpdateConstraint(ctx, r.(*flipt.UpdateConstraintRequest))
	}

	info := &grpc.UnaryServerInfo{
		FullMethod: "UpdateConstraint",
	}

	tp := sdktrace.NewTracerProvider(sdktrace.WithSampler(sdktrace.AlwaysSample()))
	tp.RegisterSpanProcessor(sdktrace.NewSimpleSpanProcessor(exporterSpy))

	tr := tp.Tracer("SpanProcessor")
	ctx, span := tr.Start(context.Background(), "OnStart")

	got, err := unaryInterceptor(ctx, req, info, handler)
	require.NoError(t, err)
	assert.NotNil(t, got)

	span.End()
	assert.Equal(t, 1, exporterSpy.GetSendAuditsCalled())
}

func TestAuditUnaryInterceptor_DeleteConstraint(t *testing.T) {
	var (
		store       = &common.StoreMock{}
		logger      = zaptest.NewLogger(t)
		exporterSpy = newAuditExporterSpy(logger)
		s           = server.New(logger, store)
		req         = &flipt.DeleteConstraintRequest{
			Id:         "1",
			SegmentKey: "constraintsegmentkey",
		}
	)

	store.On("DeleteConstraint", mock.Anything, req).Return(nil)

	unaryInterceptor := AuditEventUnaryInterceptor(logger, &checkerDummy{})

	handler := func(ctx context.Context, r interface{}) (interface{}, error) {
		return s.DeleteConstraint(ctx, r.(*flipt.DeleteConstraintRequest))
	}

	info := &grpc.UnaryServerInfo{
		FullMethod: "DeleteConstraint",
	}

	tp := sdktrace.NewTracerProvider(sdktrace.WithSampler(sdktrace.AlwaysSample()))
	tp.RegisterSpanProcessor(sdktrace.NewSimpleSpanProcessor(exporterSpy))

	tr := tp.Tracer("SpanProcessor")
	ctx, span := tr.Start(context.Background(), "OnStart")

	got, err := unaryInterceptor(ctx, req, info, handler)
	require.NoError(t, err)
	assert.NotNil(t, got)

	span.End()
	assert.Equal(t, 1, exporterSpy.GetSendAuditsCalled())
}

func TestAuditUnaryInterceptor_CreateRollout(t *testing.T) {
	var (
		store       = &common.StoreMock{}
		logger      = zaptest.NewLogger(t)
		exporterSpy = newAuditExporterSpy(logger)
		s           = server.New(logger, store)
		req         = &flipt.CreateRolloutRequest{
			FlagKey: "flagkey",
			Rank:    1,
			Rule: &flipt.CreateRolloutRequest_Threshold{
				Threshold: &flipt.RolloutThreshold{
					Percentage: 50.0,
					Value:      true,
				},
			},
		}
	)

	store.On("CreateRollout", mock.Anything, req).Return(&flipt.Rollout{
		Id:           "1",
		NamespaceKey: "default",
		Rank:         1,
		FlagKey:      req.FlagKey,
	}, nil)

	unaryInterceptor := AuditEventUnaryInterceptor(logger, &checkerDummy{})

	handler := func(ctx context.Context, r interface{}) (interface{}, error) {
		return s.CreateRollout(ctx, r.(*flipt.CreateRolloutRequest))
	}

	info := &grpc.UnaryServerInfo{
		FullMethod: "CreateRollout",
	}

	tp := sdktrace.NewTracerProvider(sdktrace.WithSampler(sdktrace.AlwaysSample()))
	tp.RegisterSpanProcessor(sdktrace.NewSimpleSpanProcessor(exporterSpy))

	tr := tp.Tracer("SpanProcessor")
	ctx, span := tr.Start(context.Background(), "OnStart")

	got, err := unaryInterceptor(ctx, req, info, handler)
	require.NoError(t, err)
	assert.NotNil(t, got)

	span.End()
	assert.Equal(t, 1, exporterSpy.GetSendAuditsCalled())
}

func TestAuditUnaryInterceptor_UpdateRollout(t *testing.T) {
	var (
		store       = &common.StoreMock{}
		logger      = zaptest.NewLogger(t)
		exporterSpy = newAuditExporterSpy(logger)
		s           = server.New(logger, store)
		req         = &flipt.UpdateRolloutRequest{
			Description: "desc",
		}
	)

	store.On("UpdateRollout", mock.Anything, req).Return(&flipt.Rollout{
		Description:  "desc",
		FlagKey:      "flagkey",
		NamespaceKey: "default",
		Rank:         1,
	}, nil)

	unaryInterceptor := AuditEventUnaryInterceptor(logger, &checkerDummy{})

	handler := func(ctx context.Context, r interface{}) (interface{}, error) {
		return s.UpdateRollout(ctx, r.(*flipt.UpdateRolloutRequest))
	}

	info := &grpc.UnaryServerInfo{
		FullMethod: "UpdateRollout",
	}

	tp := sdktrace.NewTracerProvider(sdktrace.WithSampler(sdktrace.AlwaysSample()))
	tp.RegisterSpanProcessor(sdktrace.NewSimpleSpanProcessor(exporterSpy))

	tr := tp.Tracer("SpanProcessor")
	ctx, span := tr.Start(context.Background(), "OnStart")

	got, err := unaryInterceptor(ctx, req, info, handler)
	require.NoError(t, err)
	assert.NotNil(t, got)

	span.End()
	assert.Equal(t, 1, exporterSpy.GetSendAuditsCalled())
}

func TestAuditUnaryInterceptor_OrderRollout(t *testing.T) {
	var (
		store       = &common.StoreMock{}
		logger      = zaptest.NewLogger(t)
		exporterSpy = newAuditExporterSpy(logger)
		s           = server.New(logger, store)
		req         = &flipt.OrderRolloutsRequest{
			NamespaceKey: "default",
			FlagKey:      "flagkey",
			RolloutIds:   []string{"1", "2"},
		}
	)

	store.On("OrderRollouts", mock.Anything, req).Return(nil)

	unaryInterceptor := AuditEventUnaryInterceptor(logger, &checkerDummy{})

	handler := func(ctx context.Context, r interface{}) (interface{}, error) {
		return s.OrderRollouts(ctx, r.(*flipt.OrderRolloutsRequest))
	}

	info := &grpc.UnaryServerInfo{
		FullMethod: "OrderRollouts",
	}

	tp := sdktrace.NewTracerProvider(sdktrace.WithSampler(sdktrace.AlwaysSample()))
	tp.RegisterSpanProcessor(sdktrace.NewSimpleSpanProcessor(exporterSpy))

	tr := tp.Tracer("SpanProcessor")
	ctx, span := tr.Start(context.Background(), "OnStart")

	got, err := unaryInterceptor(ctx, req, info, handler)
	require.NoError(t, err)
	assert.NotNil(t, got)
	span.End()
	assert.Equal(t, 1, exporterSpy.GetSendAuditsCalled())
}

func TestAuditUnaryInterceptor_DeleteRollout(t *testing.T) {
	var (
		store       = &common.StoreMock{}
		logger      = zaptest.NewLogger(t)
		exporterSpy = newAuditExporterSpy(logger)
		s           = server.New(logger, store)
		req         = &flipt.DeleteRolloutRequest{
			Id:      "1",
			FlagKey: "flagKey",
		}
	)

	store.On("DeleteRollout", mock.Anything, req).Return(nil)

	unaryInterceptor := AuditEventUnaryInterceptor(logger, &checkerDummy{})

	handler := func(ctx context.Context, r interface{}) (interface{}, error) {
		return s.DeleteRollout(ctx, r.(*flipt.DeleteRolloutRequest))
	}

	info := &grpc.UnaryServerInfo{
		FullMethod: "DeleteRollout",
	}

	tp := sdktrace.NewTracerProvider(sdktrace.WithSampler(sdktrace.AlwaysSample()))
	tp.RegisterSpanProcessor(sdktrace.NewSimpleSpanProcessor(exporterSpy))

	tr := tp.Tracer("SpanProcessor")
	ctx, span := tr.Start(context.Background(), "OnStart")

	got, err := unaryInterceptor(ctx, req, info, handler)
	require.NoError(t, err)
	assert.NotNil(t, got)

	span.End()
	assert.Equal(t, 1, exporterSpy.GetSendAuditsCalled())
}

func TestAuditUnaryInterceptor_CreateRule(t *testing.T) {
	var (
		store       = &common.StoreMock{}
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

	unaryInterceptor := AuditEventUnaryInterceptor(logger, &checkerDummy{})

	handler := func(ctx context.Context, r interface{}) (interface{}, error) {
		return s.CreateRule(ctx, r.(*flipt.CreateRuleRequest))
	}

	info := &grpc.UnaryServerInfo{
		FullMethod: "CreateRule",
	}

	tp := sdktrace.NewTracerProvider(sdktrace.WithSampler(sdktrace.AlwaysSample()))
	tp.RegisterSpanProcessor(sdktrace.NewSimpleSpanProcessor(exporterSpy))

	tr := tp.Tracer("SpanProcessor")
	ctx, span := tr.Start(context.Background(), "OnStart")

	got, err := unaryInterceptor(ctx, req, info, handler)
	require.NoError(t, err)
	assert.NotNil(t, got)

	span.End()
	assert.Equal(t, 1, exporterSpy.GetSendAuditsCalled())
}

func TestAuditUnaryInterceptor_UpdateRule(t *testing.T) {
	var (
		store       = &common.StoreMock{}
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

	unaryInterceptor := AuditEventUnaryInterceptor(logger, &checkerDummy{})

	handler := func(ctx context.Context, r interface{}) (interface{}, error) {
		return s.UpdateRule(ctx, r.(*flipt.UpdateRuleRequest))
	}

	info := &grpc.UnaryServerInfo{
		FullMethod: "UpdateRule",
	}

	tp := sdktrace.NewTracerProvider(sdktrace.WithSampler(sdktrace.AlwaysSample()))
	tp.RegisterSpanProcessor(sdktrace.NewSimpleSpanProcessor(exporterSpy))

	tr := tp.Tracer("SpanProcessor")
	ctx, span := tr.Start(context.Background(), "OnStart")

	got, err := unaryInterceptor(ctx, req, info, handler)
	require.NoError(t, err)
	assert.NotNil(t, got)

	span.End()
	assert.Equal(t, 1, exporterSpy.GetSendAuditsCalled())
}

func TestAuditUnaryInterceptor_OrderRule(t *testing.T) {
	var (
		store       = &common.StoreMock{}
		logger      = zaptest.NewLogger(t)
		exporterSpy = newAuditExporterSpy(logger)
		s           = server.New(logger, store)
		req         = &flipt.OrderRulesRequest{
			FlagKey: "flagkey",
			RuleIds: []string{"1", "2"},
		}
	)

	store.On("OrderRules", mock.Anything, req).Return(nil)

	unaryInterceptor := AuditEventUnaryInterceptor(logger, &checkerDummy{})

	handler := func(ctx context.Context, r interface{}) (interface{}, error) {
		return s.OrderRules(ctx, r.(*flipt.OrderRulesRequest))
	}

	info := &grpc.UnaryServerInfo{
		FullMethod: "OrderRules",
	}

	tp := sdktrace.NewTracerProvider(sdktrace.WithSampler(sdktrace.AlwaysSample()))
	tp.RegisterSpanProcessor(sdktrace.NewSimpleSpanProcessor(exporterSpy))

	tr := tp.Tracer("SpanProcessor")
	ctx, span := tr.Start(context.Background(), "OnStart")

	got, err := unaryInterceptor(ctx, req, info, handler)
	require.NoError(t, err)
	assert.NotNil(t, got)

	span.End()
	assert.Equal(t, 1, exporterSpy.GetSendAuditsCalled())
}

func TestAuditUnaryInterceptor_DeleteRule(t *testing.T) {
	var (
		store       = &common.StoreMock{}
		logger      = zaptest.NewLogger(t)
		exporterSpy = newAuditExporterSpy(logger)
		s           = server.New(logger, store)
		req         = &flipt.DeleteRuleRequest{
			Id:      "1",
			FlagKey: "flagkey",
		}
	)

	store.On("DeleteRule", mock.Anything, req).Return(nil)

	unaryInterceptor := AuditEventUnaryInterceptor(logger, &checkerDummy{})

	handler := func(ctx context.Context, r interface{}) (interface{}, error) {
		return s.DeleteRule(ctx, r.(*flipt.DeleteRuleRequest))
	}

	info := &grpc.UnaryServerInfo{
		FullMethod: "DeleteRule",
	}

	tp := sdktrace.NewTracerProvider(sdktrace.WithSampler(sdktrace.AlwaysSample()))
	tp.RegisterSpanProcessor(sdktrace.NewSimpleSpanProcessor(exporterSpy))

	tr := tp.Tracer("SpanProcessor")
	ctx, span := tr.Start(context.Background(), "OnStart")

	got, err := unaryInterceptor(ctx, req, info, handler)
	require.NoError(t, err)
	assert.NotNil(t, got)

	span.End()
	assert.Equal(t, 1, exporterSpy.GetSendAuditsCalled())
}

func TestAuditUnaryInterceptor_CreateNamespace(t *testing.T) {
	var (
		store       = &common.StoreMock{}
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

	unaryInterceptor := AuditEventUnaryInterceptor(logger, &checkerDummy{})

	handler := func(ctx context.Context, r interface{}) (interface{}, error) {
		return s.CreateNamespace(ctx, r.(*flipt.CreateNamespaceRequest))
	}

	info := &grpc.UnaryServerInfo{
		FullMethod: "CreateNamespace",
	}

	tp := sdktrace.NewTracerProvider(sdktrace.WithSampler(sdktrace.AlwaysSample()))
	tp.RegisterSpanProcessor(sdktrace.NewSimpleSpanProcessor(exporterSpy))

	tr := tp.Tracer("SpanProcessor")
	ctx, span := tr.Start(context.Background(), "OnStart")

	got, err := unaryInterceptor(ctx, req, info, handler)
	require.NoError(t, err)
	assert.NotNil(t, got)

	span.End()
	assert.Equal(t, 1, exporterSpy.GetSendAuditsCalled())
}

func TestAuditUnaryInterceptor_UpdateNamespace(t *testing.T) {
	var (
		store       = &common.StoreMock{}
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

	unaryInterceptor := AuditEventUnaryInterceptor(logger, &checkerDummy{})

	handler := func(ctx context.Context, r interface{}) (interface{}, error) {
		return s.UpdateNamespace(ctx, r.(*flipt.UpdateNamespaceRequest))
	}

	info := &grpc.UnaryServerInfo{
		FullMethod: "UpdateNamespace",
	}

	tp := sdktrace.NewTracerProvider(sdktrace.WithSampler(sdktrace.AlwaysSample()))
	tp.RegisterSpanProcessor(sdktrace.NewSimpleSpanProcessor(exporterSpy))

	tr := tp.Tracer("SpanProcessor")
	ctx, span := tr.Start(context.Background(), "OnStart")

	got, err := unaryInterceptor(ctx, req, info, handler)
	require.NoError(t, err)
	assert.NotNil(t, got)

	span.End()
	assert.Equal(t, 1, exporterSpy.GetSendAuditsCalled())
}

func TestAuditUnaryInterceptor_DeleteNamespace(t *testing.T) {
	var (
		store       = &common.StoreMock{}
		logger      = zaptest.NewLogger(t)
		exporterSpy = newAuditExporterSpy(logger)
		s           = server.New(logger, store)
		req         = &flipt.DeleteNamespaceRequest{
			Key: "namespacekey",
		}
	)

	store.On("GetNamespace", mock.Anything, storage.NewNamespace(req.Key)).Return(&flipt.Namespace{
		Key: req.Key,
	}, nil)

	store.On("CountFlags", mock.Anything, storage.NewNamespace(req.Key)).Return(uint64(0), nil)

	store.On("DeleteNamespace", mock.Anything, req).Return(nil)

	unaryInterceptor := AuditEventUnaryInterceptor(logger, &checkerDummy{})

	handler := func(ctx context.Context, r interface{}) (interface{}, error) {
		return s.DeleteNamespace(ctx, r.(*flipt.DeleteNamespaceRequest))
	}

	info := &grpc.UnaryServerInfo{
		FullMethod: "DeleteNamespace",
	}

	tp := sdktrace.NewTracerProvider(sdktrace.WithSampler(sdktrace.AlwaysSample()))
	tp.RegisterSpanProcessor(sdktrace.NewSimpleSpanProcessor(exporterSpy))

	tr := tp.Tracer("SpanProcessor")
	ctx, span := tr.Start(context.Background(), "OnStart")

	got, err := unaryInterceptor(ctx, req, info, handler)
	require.NoError(t, err)
	assert.NotNil(t, got)

	span.End()
	assert.Equal(t, 1, exporterSpy.GetSendAuditsCalled())
}

func TestAuthMetadataAuditUnaryInterceptor(t *testing.T) {
	var (
		store       = &common.StoreMock{}
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

	unaryInterceptor := AuditEventUnaryInterceptor(logger, &checkerDummy{})

	handler := func(ctx context.Context, r interface{}) (interface{}, error) {
		return s.CreateFlag(ctx, r.(*flipt.CreateFlagRequest))
	}

	info := &grpc.UnaryServerInfo{
		FullMethod: "CreateFlag",
	}

	tp := sdktrace.NewTracerProvider(sdktrace.WithSampler(sdktrace.AlwaysSample()))
	tp.RegisterSpanProcessor(sdktrace.NewSimpleSpanProcessor(exporterSpy))

	tr := tp.Tracer("SpanProcessor")
	ctx, span := tr.Start(context.Background(), "OnStart")

	ctx = authmiddlewaregrpc.ContextWithAuthentication(ctx, &authrpc.Authentication{
		Method: authrpc.Method_METHOD_OIDC,
		Metadata: map[string]string{
			"io.flipt.auth.oidc.email": "example@flipt.com",
		},
	})

	got, err := unaryInterceptor(ctx, req, info, handler)
	require.NoError(t, err)
	assert.NotNil(t, got)

	span.End()

	event := exporterSpy.GetEvents()[0]
	assert.Equal(t, "example@flipt.com", event.Metadata.Actor.Email)
	assert.Equal(t, "oidc", event.Metadata.Actor.Authentication)
}

func TestAuditUnaryInterceptor_CreateToken(t *testing.T) {
	var (
		store       = &authStoreMock{}
		logger      = zaptest.NewLogger(t)
		exporterSpy = newAuditExporterSpy(logger)
		s           = token.NewServer(logger, store)
		req         = &authrpc.CreateTokenRequest{
			Name: "token",
		}
	)

	store.On("CreateAuthentication", mock.Anything, &storageauth.CreateAuthenticationRequest{
		Method: authrpc.Method_METHOD_TOKEN,
		Metadata: map[string]string{
			"io.flipt.auth.token.name": "token",
		},
	}).Return("", &authrpc.Authentication{Metadata: map[string]string{
		"email": "example@flipt.io",
	}}, nil)

	unaryInterceptor := AuditEventUnaryInterceptor(logger, &checkerDummy{})

	handler := func(ctx context.Context, r interface{}) (interface{}, error) {
		return s.CreateToken(ctx, r.(*authrpc.CreateTokenRequest))
	}

	info := &grpc.UnaryServerInfo{
		FullMethod: "CreateToken",
	}

	tp := sdktrace.NewTracerProvider(sdktrace.WithSampler(sdktrace.AlwaysSample()))
	tp.RegisterSpanProcessor(sdktrace.NewSimpleSpanProcessor(exporterSpy))

	tr := tp.Tracer("SpanProcessor")
	ctx, span := tr.Start(context.Background(), "OnStart")

	got, err := unaryInterceptor(ctx, req, info, handler)
	require.NoError(t, err)
	assert.NotNil(t, got)

	span.End()
	assert.Equal(t, 1, exporterSpy.GetSendAuditsCalled())
}

func TestFliptAcceptServerVersionUnaryInterceptor(t *testing.T) {
	tests := []struct {
		name          string
		metadata      metadata.MD
		expectVersion string
	}{
		{
			name:     "does not contain x-flipt-accept-server-version header",
			metadata: metadata.MD{},
		},
		{
			name: "contains valid x-flipt-accept-server-version header with v prefix",
			metadata: metadata.MD{
				"x-flipt-accept-server-version": []string{"v1.0.0"},
			},
			expectVersion: "1.0.0",
		},
		{
			name: "contains valid x-flipt-accept-server-version header no prefix",
			metadata: metadata.MD{
				"x-flipt-accept-server-version": []string{"1.0.0"},
			},
			expectVersion: "1.0.0",
		},
		{
			name: "contains invalid x-flipt-accept-server-version header",
			metadata: metadata.MD{
				"x-flipt-accept-server-version": []string{"invalid"},
			},
		},
	}

	for _, tt := range tests {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			var (
				ctx          = context.Background()
				retrievedCtx = ctx
				called       int

				spyHandler = grpc.UnaryHandler(func(ctx context.Context, req interface{}) (interface{}, error) {
					called++
					retrievedCtx = ctx
					return nil, nil
				})
				logger      = zaptest.NewLogger(t)
				interceptor = FliptAcceptServerVersionUnaryInterceptor(logger)
			)

			if tt.metadata != nil {
				ctx = metadata.NewIncomingContext(context.Background(), tt.metadata)
			}

			_, _ = interceptor(ctx, nil, nil, spyHandler)
			assert.Equal(t, 1, called)

			v := FliptAcceptServerVersionFromContext(retrievedCtx)
			require.NotNil(t, v)

			if tt.expectVersion != "" {
				assert.Equal(t, semver.MustParse(tt.expectVersion), v)
			} else {
				assert.Equal(t, preFliptAcceptServerVersion, v)
			}
		})
	}
}

func TestForwardFliptAcceptServerVersion(t *testing.T) {
	req := httptest.NewRequest("GET", "/", nil)
	md := ForwardFliptAcceptServerVersion(context.Background(), req)
	assert.Empty(t, md.Get(fliptAcceptServerVersionHeaderKey))
	req.Header.Add(fliptAcceptServerVersionHeaderKey, "v1.32.0")

	ctx := metadata.NewIncomingContext(context.Background(), metadata.Pairs("key", "value"))
	md = ForwardFliptAcceptServerVersion(ctx, req)
	assert.Equal(t, []string{"v1.32.0"}, md.Get(fliptAcceptServerVersionHeaderKey))
	assert.Equal(t, []string{"value"}, md.Get("key"))
}
