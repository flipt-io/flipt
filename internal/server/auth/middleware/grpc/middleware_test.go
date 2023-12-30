package grpc_middleware

import (
	"context"
	"regexp"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.flipt.io/flipt/internal/storage/auth"
	"go.flipt.io/flipt/internal/storage/auth/memory"
	authrpc "go.flipt.io/flipt/rpc/flipt/auth"
	"go.flipt.io/flipt/rpc/flipt/evaluation"
	"go.uber.org/zap/zaptest"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// mockServer is used to test skipping auth
type mockServer struct {
	skipsAuth           bool
	allowNamespacedAuth bool
}

func (s *mockServer) SkipsAuthentication(ctx context.Context) bool {
	return s.skipsAuth
}

func (s *mockServer) AllowsNamespaceScopedAuthentication(ctx context.Context) bool {
	return s.allowNamespacedAuth
}

func TestUnaryInterceptor(t *testing.T) {
	authenticator := memory.NewStore()
	clientToken, storedAuth, err := authenticator.CreateAuthentication(
		context.TODO(),
		&auth.CreateAuthenticationRequest{Method: authrpc.Method_METHOD_TOKEN},
	)
	require.NoError(t, err)

	// expired auth
	expiredToken, _, err := authenticator.CreateAuthentication(
		context.TODO(),
		&auth.CreateAuthenticationRequest{
			Method:    authrpc.Method_METHOD_TOKEN,
			ExpiresAt: timestamppb.New(time.Now().UTC().Add(-time.Hour)),
		},
	)
	require.NoError(t, err)

	for _, test := range []struct {
		name         string
		metadata     metadata.MD
		server       any
		expectedErr  error
		expectedAuth *authrpc.Authentication
	}{
		{
			name: "successful authentication (authorization header)",
			metadata: metadata.MD{
				"Authorization": []string{"Bearer " + clientToken},
			},
			expectedAuth: storedAuth,
		},
		{
			name: "successful authentication (cookie header)",
			metadata: metadata.MD{
				"grpcgateway-cookie": []string{"flipt_client_token=" + clientToken},
			},
			expectedAuth: storedAuth,
		},
		{
			name:     "successful authentication (skipped)",
			metadata: metadata.MD{},
			server: &mockServer{
				skipsAuth: true,
			},
		},
		{
			name: "token has expired",
			metadata: metadata.MD{
				"Authorization": []string{"Bearer " + expiredToken},
			},
			expectedErr: ErrUnauthenticated,
		},
		{
			name: "client token not found in store",
			metadata: metadata.MD{
				"Authorization": []string{"Bearer unknowntoken"},
			},
			expectedErr: ErrUnauthenticated,
		},
		{
			name: "client token missing Bearer prefix",
			metadata: metadata.MD{
				"Authorization": []string{clientToken},
			},
			expectedErr: ErrUnauthenticated,
		},
		{
			name: "authorization header empty",
			metadata: metadata.MD{
				"Authorization": []string{},
			},
			expectedErr: ErrUnauthenticated,
		},
		{
			name: "cookie header with no flipt_client_token",
			metadata: metadata.MD{
				"grpcgateway-cookie": []string{"blah"},
			},
			expectedErr: ErrUnauthenticated,
		},
		{
			name:        "authorization header not set",
			metadata:    metadata.MD{},
			expectedErr: ErrUnauthenticated,
		},
		{
			name:        "no metadata on context",
			metadata:    nil,
			expectedErr: ErrUnauthenticated,
		},
	} {
		test := test
		t.Run(test.name, func(t *testing.T) {
			var (
				logger = zaptest.NewLogger(t)

				ctx          = context.Background()
				retrievedCtx = ctx
				handler      = func(ctx context.Context, req interface{}) (interface{}, error) {
					// update retrievedCtx to the one delegated to the handler
					retrievedCtx = ctx
					return nil, nil
				}
			)

			if test.metadata != nil {
				ctx = metadata.NewIncomingContext(ctx, test.metadata)
			}

			_, err := UnaryInterceptor(logger, authenticator)(
				ctx,
				nil,
				&grpc.UnaryServerInfo{Server: test.server},
				handler,
			)
			require.Equal(t, test.expectedErr, err)
			assert.Equal(t, test.expectedAuth, GetAuthenticationFrom(retrievedCtx))
		})
	}
}

func TestEmailMatchingInterceptorWithNoAuth(t *testing.T) {
	var (
		logger = zaptest.NewLogger(t)

		ctx     = context.Background()
		handler = func(ctx context.Context, req interface{}) (interface{}, error) {
			return nil, nil
		}
	)

	require.Panics(t, func() {
		_, _ = EmailMatchingInterceptor(logger, []*regexp.Regexp{regexp.MustCompile("^.*@flipt.io$")})(
			ctx,
			nil,
			&grpc.UnaryServerInfo{Server: &mockServer{}},
			handler,
		)
	})
}

func TestEmailMatchingInterceptor(t *testing.T) {
	authenticator := memory.NewStore()
	clientToken, storedAuth, err := authenticator.CreateAuthentication(
		context.TODO(),
		&auth.CreateAuthenticationRequest{
			Method: authrpc.Method_METHOD_OIDC,
			Metadata: map[string]string{
				"io.flipt.auth.oidc.email": "foo@flipt.io",
			},
		},
	)
	require.NoError(t, err)

	nonEmailClientToken, nonEmailStoredAuth, err := authenticator.CreateAuthentication(
		context.TODO(),
		&auth.CreateAuthenticationRequest{
			Method:   authrpc.Method_METHOD_OIDC,
			Metadata: map[string]string{},
		},
	)
	require.NoError(t, err)

	staticClientToken, staticStoreAuth, err := authenticator.CreateAuthentication(
		context.TODO(),
		&auth.CreateAuthenticationRequest{
			Method:   authrpc.Method_METHOD_TOKEN,
			Metadata: map[string]string{},
		},
	)
	require.NoError(t, err)

	for _, test := range []struct {
		name         string
		metadata     metadata.MD
		server       any
		auth         *authrpc.Authentication
		emailMatches []string
		expectedErr  error
	}{
		{
			name: "successful email match (regular string)",
			metadata: metadata.MD{
				"Authorization": []string{"Bearer " + clientToken},
			},
			emailMatches: []string{
				"foo@flipt.io",
			},
			auth: storedAuth,
		},
		{
			name: "successful email match (regex)",
			metadata: metadata.MD{
				"Authorization": []string{"Bearer " + clientToken},
			},
			emailMatches: []string{
				"^.*@flipt.io$",
			},
			auth: storedAuth,
		},
		{
			name: "successful token was not generated via OIDC method",
			metadata: metadata.MD{
				"Authorization": []string{"Bearer " + staticClientToken},
			},
			emailMatches: []string{
				"foo@flipt.io",
			},
			auth: staticStoreAuth,
		},
		{
			name: "email does not match (regular string)",
			metadata: metadata.MD{
				"Authorization": []string{"Bearer " + clientToken},
			},
			emailMatches: []string{
				"bar@flipt.io",
			},
			auth:        storedAuth,
			expectedErr: ErrUnauthenticated,
		},
		{
			name: "email does not match (regex)",
			metadata: metadata.MD{
				"Authorization": []string{"Bearer " + clientToken},
			},
			emailMatches: []string{
				"^.*@gmail.com$",
			},
			auth:        storedAuth,
			expectedErr: ErrUnauthenticated,
		},
		{
			name: "email not provided by oidc provider",
			metadata: metadata.MD{
				"Authorization": []string{"Bearer " + nonEmailClientToken},
			},
			emailMatches: []string{
				"foo@flipt.io",
			},
			auth:        nonEmailStoredAuth,
			expectedErr: ErrUnauthenticated,
		},
	} {
		test := test
		t.Run(test.name, func(t *testing.T) {
			var (
				logger = zaptest.NewLogger(t)

				ctx     = ContextWithAuthentication(context.Background(), test.auth)
				handler = func(ctx context.Context, req interface{}) (interface{}, error) {
					return nil, nil
				}
			)

			if test.metadata != nil {
				ctx = metadata.NewIncomingContext(ctx, test.metadata)
			}

			rgxs := make([]*regexp.Regexp, 0, len(test.emailMatches))

			for _, em := range test.emailMatches {
				rgx, err := regexp.Compile(em)
				require.NoError(t, err)

				rgxs = append(rgxs, rgx)
			}

			_, err := EmailMatchingInterceptor(logger, rgxs)(
				ctx,
				nil,
				&grpc.UnaryServerInfo{Server: test.server},
				handler,
			)
			require.Equal(t, test.expectedErr, err)
		})
	}
}

func TestNamespaceMatchingInterceptorWithNoAuth(t *testing.T) {
	var (
		logger = zaptest.NewLogger(t)

		ctx     = context.Background()
		handler = func(ctx context.Context, req interface{}) (interface{}, error) {
			return nil, nil
		}
	)

	require.Panics(t, func() {
		_, _ = NamespaceMatchingInterceptor(logger)(
			ctx,
			nil,
			&grpc.UnaryServerInfo{Server: &mockServer{}},
			handler,
		)
	})
}

func TestNamespaceMatchingInterceptor(t *testing.T) {
	for _, tt := range []struct {
		name        string
		authReq     *auth.CreateAuthenticationRequest
		req         any
		srv         any
		wantCalled  bool
		expectedErr error
	}{
		{
			name: "successful namespace match",
			authReq: &auth.CreateAuthenticationRequest{
				Method: authrpc.Method_METHOD_TOKEN,
				Metadata: map[string]string{
					"io.flipt.auth.token.namespace": "foo",
				},
			},
			req: &evaluation.EvaluationRequest{
				NamespaceKey: "foo",
			},
			wantCalled: true,
		},
		{
			name: "successful namespace match on batch",
			authReq: &auth.CreateAuthenticationRequest{
				Method: authrpc.Method_METHOD_TOKEN,
				Metadata: map[string]string{
					"io.flipt.auth.token.namespace": "foo",
				},
			},
			req: &evaluation.BatchEvaluationRequest{
				Requests: []*evaluation.EvaluationRequest{
					{
						NamespaceKey: "foo",
					},
					{
						NamespaceKey: "foo",
					},
				},
			},
			wantCalled: true,
		},
		{
			name: "successful namespace (default) match on batch",
			authReq: &auth.CreateAuthenticationRequest{
				Method: authrpc.Method_METHOD_TOKEN,
				Metadata: map[string]string{
					"io.flipt.auth.token.namespace": "default",
				},
			},
			req: &evaluation.BatchEvaluationRequest{
				Requests: []*evaluation.EvaluationRequest{
					{},
					{},
				},
			},
			wantCalled: true,
		},
		{
			name: "successful skips auth",
			authReq: &auth.CreateAuthenticationRequest{
				Method: authrpc.Method_METHOD_TOKEN,
				Metadata: map[string]string{
					"io.flipt.auth.token.namespace": "foo",
				},
			},
			req: &struct{}{},
			srv: &mockServer{
				skipsAuth: true,
			},
			wantCalled: true,
		},
		{
			name: "not a token authentication",
			authReq: &auth.CreateAuthenticationRequest{
				Method: authrpc.Method_METHOD_OIDC,
				Metadata: map[string]string{
					"io.flipt.auth.github.sub": "foo",
				},
			},
			req: &evaluation.EvaluationRequest{
				NamespaceKey: "foo",
			},
			wantCalled: true,
		},
		{
			name: "namespace does not match",
			authReq: &auth.CreateAuthenticationRequest{
				Method: authrpc.Method_METHOD_TOKEN,
				Metadata: map[string]string{
					"io.flipt.auth.token.namespace": "bar",
				},
			},
			req: &evaluation.EvaluationRequest{
				NamespaceKey: "foo",
			},
			expectedErr: ErrUnauthenticated,
		},
		{
			name: "namespace not provided by token authentication",
			authReq: &auth.CreateAuthenticationRequest{
				Method: authrpc.Method_METHOD_TOKEN,
			},
			req: &evaluation.EvaluationRequest{
				NamespaceKey: "foo",
			},
			wantCalled: true,
		},
		{
			name: "empty namespace provided by token authentication",
			authReq: &auth.CreateAuthenticationRequest{
				Method: authrpc.Method_METHOD_TOKEN,
				Metadata: map[string]string{
					"io.flipt.auth.token.namespace": "",
				},
			},
			req: &evaluation.EvaluationRequest{
				NamespaceKey: "foo",
			},
			wantCalled: true,
		},
		{
			name: "namespace not provided by request",
			authReq: &auth.CreateAuthenticationRequest{
				Method: authrpc.Method_METHOD_TOKEN,
				Metadata: map[string]string{
					"io.flipt.auth.token.namespace": "foo",
				},
			},
			req:         &evaluation.EvaluationRequest{},
			expectedErr: ErrUnauthenticated,
		},
		{
			name: "namespace not available",
			authReq: &auth.CreateAuthenticationRequest{
				Method: authrpc.Method_METHOD_TOKEN,
				Metadata: map[string]string{
					"io.flipt.auth.token.namespace": "foo",
				},
			},
			req:         &evaluation.BatchEvaluationRequest{},
			expectedErr: ErrUnauthenticated,
		},
		{
			name: "namespace not consistent for batch",
			authReq: &auth.CreateAuthenticationRequest{
				Method: authrpc.Method_METHOD_TOKEN,
				Metadata: map[string]string{
					"io.flipt.auth.token.namespace": "foo",
				},
			},
			req: &evaluation.BatchEvaluationRequest{
				Requests: []*evaluation.EvaluationRequest{
					{
						NamespaceKey: "foo",
					},
					{
						NamespaceKey: "bar",
					},
				},
			},
			expectedErr: ErrUnauthenticated,
		},
		{
			name: "non-namespaced request",
			authReq: &auth.CreateAuthenticationRequest{
				Method: authrpc.Method_METHOD_TOKEN,
				Metadata: map[string]string{
					"io.flipt.auth.token.namespace": "foo",
				},
			},
			req:         &struct{}{},
			expectedErr: ErrUnauthenticated,
		},
		{
			name: "non-namespaced scoped server",
			authReq: &auth.CreateAuthenticationRequest{
				Method: authrpc.Method_METHOD_TOKEN,
				Metadata: map[string]string{
					"io.flipt.auth.token.namespace": "foo",
				},
			},
			req: &evaluation.EvaluationRequest{
				NamespaceKey: "foo",
			},
			srv: &mockServer{
				allowNamespacedAuth: false,
			},
			expectedErr: ErrUnauthenticated,
		},
	} {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			var (
				logger        = zaptest.NewLogger(t)
				authenticator = memory.NewStore()
			)

			clientToken, storedAuth, err := authenticator.CreateAuthentication(
				context.TODO(),
				tt.authReq,
			)

			require.NoError(t, err)

			var (
				ctx     = ContextWithAuthentication(context.Background(), storedAuth)
				handler = func(ctx context.Context, req interface{}) (interface{}, error) {
					assert.True(t, tt.wantCalled)
					return nil, nil
				}

				srv = &grpc.UnaryServerInfo{Server: &mockServer{
					allowNamespacedAuth: true,
				}}

				unaryInterceptor = NamespaceMatchingInterceptor(logger)
			)

			if tt.srv != nil {
				srv.Server = tt.srv
			}

			ctx = metadata.NewIncomingContext(ctx, metadata.MD{
				"Authorization": []string{"Bearer " + clientToken},
			})

			_, err = unaryInterceptor(ctx, tt.req, srv, handler)
			require.Equal(t, tt.expectedErr, err)
		})
	}
}
