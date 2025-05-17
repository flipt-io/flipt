package grpc_middleware

import (
	"context"
	"fmt"
	"net/http/httptest"
	"testing"

	"go.flipt.io/flipt/errors"
	flipt "go.flipt.io/flipt/rpc/flipt"
	"go.uber.org/zap"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	cctx "go.flipt.io/flipt/internal/common"
	"go.flipt.io/flipt/internal/server/common"
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

type mockStream struct {
	grpc.ServerStream
	ctx context.Context
}

func (m *mockStream) Context() context.Context { return m.ctx }

func TestValidationUnaryInterceptor(t *testing.T) {
	tests := []struct {
		name       string
		req        any
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
			handler := grpc.UnaryHandler(func(ctx context.Context, req any) (any, error) {
				called++
				return nil, nil
			})

			_, _ = ValidationUnaryInterceptor(context.Background(), req, nil, handler)
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
			name:     "already exists error",
			wantErr:  errors.ErrAlreadyExists("foo"),
			wantCode: codes.AlreadyExists,
		},
		{
			name:     "conflict error",
			wantErr:  errors.ErrConflict("foo"),
			wantCode: codes.Aborted,
		},
		{
			name:     "not implemented error",
			wantErr:  errors.ErrNotImplemented("foo"),
			wantCode: codes.Unimplemented,
		},
		{
			name:     "not modified error",
			wantCode: codes.OK,
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
			handler := grpc.UnaryHandler(func(ctx context.Context, req any) (any, error) {
				return nil, wantErr
			})

			_, err := ErrorUnaryInterceptor(context.Background(), nil, nil, handler)
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

func TestErrorStreamInterceptor(t *testing.T) {
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
			name:     "unauthorized error",
			wantErr:  errors.ErrUnauthorizedf("user %q not authorized, missing role", "foo"),
			wantCode: codes.PermissionDenied,
		},
		{
			name:     "already exists error",
			wantErr:  errors.ErrAlreadyExists("foo"),
			wantCode: codes.AlreadyExists,
		},
		{
			name:     "conflict error",
			wantErr:  errors.ErrConflict("foo"),
			wantCode: codes.Aborted,
		},
		{
			name:     "not implemented error",
			wantErr:  errors.ErrNotImplemented("foo"),
			wantCode: codes.Unimplemented,
		},
		{
			name:     "not modified error",
			wantCode: codes.OK,
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

		called := false
		t.Run(tt.name, func(t *testing.T) {
			handler := grpc.StreamHandler(func(srv any, stream grpc.ServerStream) error {
				called = true
				return wantErr
			})

			stream := &mockStream{ctx: context.Background()}
			info := &grpc.StreamServerInfo{}

			err := ErrorStreamInterceptor(nil, stream, info, handler)
			if wantErr != nil {
				require.Error(t, err)
				status := status.Convert(err)
				assert.Equal(t, wantCode, status.Code())
				return
			}

			require.NoError(t, err)
			assert.True(t, called)
		})
	}
}

func TestFliptHeadersUnaryInterceptor(t *testing.T) {
	type want struct {
		environment string
		namespace   string
	}

	tests := []struct {
		name string
		md   metadata.MD
		want want
	}{
		{
			name: "no flipt environment",
			want: want{
				environment: flipt.DefaultEnvironment,
				namespace:   flipt.DefaultNamespace,
			},
		},
		{
			name: "flipt environment",
			md:   metadata.New(map[string]string{common.HeaderFliptEnvironment: "foo"}),
			want: want{
				environment: "foo",
				namespace:   flipt.DefaultNamespace,
			},
		},
		{
			name: "no flipt namespace",
			want: want{
				environment: flipt.DefaultEnvironment,
				namespace:   flipt.DefaultNamespace,
			},
		},
		{
			name: "flipt namespace",
			md:   metadata.New(map[string]string{common.HeaderFliptNamespace: "bar"}),
			want: want{
				environment: flipt.DefaultEnvironment,
				namespace:   "bar",
			},
		},
		{
			name: "flipt environment and namespace",
			md:   metadata.New(map[string]string{common.HeaderFliptEnvironment: "foo", common.HeaderFliptNamespace: "bar"}),
			want: want{
				environment: "foo",
				namespace:   "bar",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			if tt.md != nil {
				ctx = metadata.NewIncomingContext(ctx, tt.md)
			}

			//nolint:all
			handler := grpc.UnaryHandler(func(ctx context.Context, req any) (any, error) {
				environment, _ := cctx.FliptEnvironmentFromContext(ctx)
				assert.Equal(t, tt.want.environment, environment)

				namespace, _ := cctx.FliptNamespaceFromContext(ctx)
				assert.Equal(t, tt.want.namespace, namespace)
				return nil, nil
			})

			_, err := FliptHeadersUnaryInterceptor(zap.NewNop())(ctx, nil, nil, handler)
			require.NoError(t, err)
		})
	}
}

func TestFliptHeadersStreamInterceptor(t *testing.T) {
	type want struct {
		environment string
		namespace   string
	}

	tests := []struct {
		name string
		md   metadata.MD
		want want
	}{
		{
			name: "no flipt environment",
			want: want{
				environment: flipt.DefaultEnvironment,
				namespace:   flipt.DefaultNamespace,
			},
		},
		{
			name: "flipt environment",
			md:   metadata.New(map[string]string{common.HeaderFliptEnvironment: "foo"}),
			want: want{
				environment: "foo",
				namespace:   flipt.DefaultNamespace,
			},
		},
		{
			name: "no flipt namespace",
			want: want{
				environment: flipt.DefaultEnvironment,
				namespace:   flipt.DefaultNamespace,
			},
		},
		{
			name: "flipt namespace",
			md:   metadata.New(map[string]string{common.HeaderFliptNamespace: "bar"}),
			want: want{
				environment: flipt.DefaultEnvironment,
				namespace:   "bar",
			},
		},
		{
			name: "flipt environment and namespace",
			md:   metadata.New(map[string]string{common.HeaderFliptEnvironment: "foo", common.HeaderFliptNamespace: "bar"}),
			want: want{
				environment: "foo",
				namespace:   "bar",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			if tt.md != nil {
				ctx = metadata.NewIncomingContext(ctx, tt.md)
			}

			called := false
			handler := func(srv any, stream grpc.ServerStream) error {
				called = true
				environment, _ := cctx.FliptEnvironmentFromContext(stream.Context())
				assert.Equal(t, tt.want.environment, environment)

				namespace, _ := cctx.FliptNamespaceFromContext(stream.Context())
				assert.Equal(t, tt.want.namespace, namespace)
				return nil
			}

			stream := &mockStream{ctx: ctx}
			info := &grpc.StreamServerInfo{}
			logger := zap.NewNop()

			err := FliptHeadersStreamInterceptor(logger)(nil, stream, info, handler)
			require.NoError(t, err)
			assert.True(t, called)
		})
	}
}

func TestEvaluationUnaryInterceptor_Noop(t *testing.T) {
	var (
		req = &flipt.ListFlagRequest{
			NamespaceKey: "foo",
		}

		//nolint:all
		handler = func(ctx context.Context, r any) (any, error) {
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

func TestEvaluationUnaryInterceptor_EnvironmentAndNamespace(t *testing.T) {
	tests := []struct {
		name string
		req  *evaluation.EvaluationRequest
		want *evaluation.EvaluationRequest
	}{
		{
			name: "no environment or namespace",
			req: &evaluation.EvaluationRequest{
				FlagKey: "foo",
			},
			want: &evaluation.EvaluationRequest{
				EnvironmentKey: "test-environment",
				NamespaceKey:   "test-namespace",
				FlagKey:        "foo",
			},
		},
		{
			name: "existing environment and namespace",
			req: &evaluation.EvaluationRequest{
				EnvironmentKey: "foo-environment",
				NamespaceKey:   "foo-namespace",
				FlagKey:        "foo",
			},
			want: &evaluation.EvaluationRequest{
				EnvironmentKey: "foo-environment",
				NamespaceKey:   "foo-namespace",
				FlagKey:        "foo",
			},
		},
		{
			name: "existing environment and no namespace",
			req: &evaluation.EvaluationRequest{
				EnvironmentKey: "foo-environment",
				FlagKey:        "foo",
			},
			want: &evaluation.EvaluationRequest{
				EnvironmentKey: "foo-environment",
				NamespaceKey:   "test-namespace",
				FlagKey:        "foo",
			},
		},
		{
			name: "existing namespace and no environment",
			req: &evaluation.EvaluationRequest{
				NamespaceKey: "foo-namespace",
				FlagKey:      "foo",
			},
			want: &evaluation.EvaluationRequest{
				EnvironmentKey: "test-environment",
				NamespaceKey:   "foo-namespace",
				FlagKey:        "foo",
			},
		},
	}

	for _, tt := range tests {
		ctx := context.Background()
		ctx = cctx.WithFliptEnvironment(ctx, "test-environment")
		ctx = cctx.WithFliptNamespace(ctx, "test-namespace")

		var (
			//nolint:all
			handler = func(_ context.Context, r any) (any, error) {
				// ensure request has ID once it reaches the handler
				req, ok := r.(*evaluation.EvaluationRequest)
				require.True(t, ok)

				assert.Equal(t, tt.want.EnvironmentKey, req.GetEnvironmentKey())
				assert.Equal(t, tt.want.NamespaceKey, req.GetNamespaceKey())

				return &evaluation.EvaluationResponse{}, nil
			}

			info = &grpc.UnaryServerInfo{
				FullMethod: "FakeMethod",
			}
		)

		t.Run(tt.name, func(t *testing.T) {
			_, err := EvaluationUnaryInterceptor(false)(ctx, tt.req, info, handler)
			require.NoError(t, err)
		})
	}
}

func TestEvaluationUnaryInterceptor_RequestID(t *testing.T) {
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
				//nolint:all
				handler = func(ctx context.Context, r any) (any, error) {
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

func TestForwardFliptEnvironment(t *testing.T) {
	req := httptest.NewRequest("GET", "/", nil)
	md := ForwardFliptEnvironment(context.Background(), req)
	assert.Empty(t, md.Get(common.HeaderFliptEnvironment))
	req.Header.Add(common.HeaderFliptEnvironment, "extra-environment")

	md = ForwardFliptEnvironment(context.Background(), req)
	assert.Equal(t, []string{"extra-environment"}, md.Get(common.HeaderFliptEnvironment))
}

func TestForwardFliptNamespace(t *testing.T) {
	req := httptest.NewRequest("GET", "/", nil)
	md := ForwardFliptNamespace(context.Background(), req)
	assert.Empty(t, md.Get(common.HeaderFliptNamespace))
	req.Header.Add(common.HeaderFliptNamespace, "extra-namespace")

	md = ForwardFliptNamespace(context.Background(), req)
	assert.Equal(t, []string{"extra-namespace"}, md.Get(common.HeaderFliptNamespace))
}
