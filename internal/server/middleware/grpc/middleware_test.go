package grpc_middleware

import (
	"context"
	"fmt"
	"net/http/httptest"
	"testing"

	"go.flipt.io/flipt/errors"
	flipt "go.flipt.io/flipt/rpc/flipt"
	"go.uber.org/zap/zaptest"

	"github.com/blang/semver/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
			spyHandler := grpc.UnaryHandler(func(ctx context.Context, req interface{}) (interface{}, error) {
				called++
				return nil, nil
			})

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
			spyHandler := grpc.UnaryHandler(func(ctx context.Context, req interface{}) (interface{}, error) {
				return nil, wantErr
			})

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
		req = &flipt.ListFlagRequest{
			NamespaceKey: "foo",
		}

		handler = func(ctx context.Context, r interface{}) (interface{}, error) {
			return &flipt.FlagList{
				Flags: []*flipt.Flag{
					{Key: "foo"},
				},
				TotalCount: 1,
			}, nil
		}

		info = &grpc.UnaryServerInfo{
			FullMethod: "FakeMethod",
		}
	)

	got, err := EvaluationUnaryInterceptor(false)(context.Background(), req, info, handler)
	require.NoError(t, err)

	assert.NotNil(t, got)

	resp, ok := got.(*flipt.FlagList)
	assert.True(t, ok)
	assert.NotNil(t, resp)
	assert.Equal(t, "foo", resp.Flags[0].Key)
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

func TestForwardFliptNamespace(t *testing.T) {
	req := httptest.NewRequest("GET", "/", nil)
	md := ForwardFliptNamespace(context.Background(), req)
	assert.Empty(t, md.Get(fliptAcceptServerVersionHeaderKey))
	req.Header.Add(fliptNamespaceHeaderKey, "extra-namespace")

	md = ForwardFliptNamespace(context.Background(), req)
	assert.Equal(t, []string{"extra-namespace"}, md.Get(fliptNamespaceHeaderKey))
}
