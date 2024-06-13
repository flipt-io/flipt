package grpc_middleware

import (
	"context"
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"fmt"
	"regexp"
	"testing"
	"time"

	jjwt "github.com/go-jose/go-jose/v3/jwt"
	"github.com/hashicorp/cap/jwt"
	"github.com/hashicorp/cap/oidc"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.flipt.io/flipt/internal/storage/authn"
	"go.flipt.io/flipt/internal/storage/authn/memory"
	authrpc "go.flipt.io/flipt/rpc/flipt/auth"
	"go.flipt.io/flipt/rpc/flipt/evaluation"
	"go.uber.org/zap/zaptest"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// mockServer is used to test skipping authn
type mockServer struct {
	skipsAuthn           bool
	allowNamespacedAuthn bool
}

func (s *mockServer) SkipsAuthentication(ctx context.Context) bool {
	return s.skipsAuthn
}

func (s *mockServer) AllowsNamespaceScopedAuthentication(ctx context.Context) bool {
	return s.allowNamespacedAuthn
}

var priv *rsa.PrivateKey

func init() {
	// Generate a key to sign JWTs with throughout most test cases.
	// It can be slow sometimes to generate a 4096-bit RSA key, so we only do it once.
	var err error
	priv, err = rsa.GenerateKey(rand.Reader, 4096)
	if err != nil {
		panic(err)
	}
}

func TestJWTAuthenticationInterceptor(t *testing.T) {
	var (
		now        = time.Now()
		nowUnix    = float64(now.Unix())
		futureUnix = float64(now.Add(2 * jjwt.DefaultLeeway).Unix())
		pub        = []crypto.PublicKey{priv.Public()}
	)

	for _, tt := range []struct {
		name             string
		metadataFunc     func() metadata.MD
		server           any
		expectedJWT      jwt.Expected
		expectedErr      error
		expectedMetadata map[string]string
	}{
		{
			name: "successful authentication",
			metadataFunc: func() metadata.MD {
				claims := map[string]interface{}{
					"iss": "flipt.io",
					"aud": "flipt",
					"sub": "sunglasses",
					"iat": nowUnix,
					"exp": futureUnix,
				}

				token := oidc.TestSignJWT(t, priv, string(jwt.RS256), claims, []byte("test-key"))
				return metadata.MD{
					"Authorization": []string{"JWT " + token},
				}
			},
			expectedJWT: jwt.Expected{
				Issuer:    "flipt.io",
				Audiences: []string{"flipt"},
				Subject:   "sunglasses",
			},
		},
		{
			name: "successful authentication (with custom user claims)",
			metadataFunc: func() metadata.MD {
				claims := map[string]interface{}{
					"iss": "flipt.io",
					"aud": "flipt",
					"iat": nowUnix,
					"exp": futureUnix,
					"user": map[string]string{
						"sub":   "sub",
						"email": "email",
						"image": "image",
						"name":  "name",
					},
				}

				token := oidc.TestSignJWT(t, priv, string(jwt.RS256), claims, []byte("test-key"))
				return metadata.MD{
					"Authorization": []string{"JWT " + token},
				}
			},
			expectedJWT: jwt.Expected{
				Issuer:    "flipt.io",
				Audiences: []string{"flipt"},
			},
			expectedMetadata: map[string]string{
				"io.flipt.auth.jwt.sub":     "sub",
				"io.flipt.auth.jwt.email":   "email",
				"io.flipt.auth.jwt.picture": "image",
				"io.flipt.auth.jwt.name":    "name",
				"io.flipt.auth.jwt.issuer":  "flipt.io",
			},
		},
		{
			name: "successful authentication (with custom user claims and arbitrary role)",
			metadataFunc: func() metadata.MD {
				claims := map[string]interface{}{
					"iss": "flipt.io",
					"aud": "flipt",
					"iat": nowUnix,
					"exp": futureUnix,
					"user": map[string]string{
						"sub":   "sub",
						"email": "email",
						"image": "image",
						"name":  "name",
					},
					"io.flipt.auth.role": "admin",
				}

				token := oidc.TestSignJWT(t, priv, string(jwt.RS256), claims, []byte("test-key"))
				return metadata.MD{
					"Authorization": []string{"JWT " + token},
				}
			},
			expectedJWT: jwt.Expected{
				Issuer:    "flipt.io",
				Audiences: []string{"flipt"},
			},
			expectedMetadata: map[string]string{
				"io.flipt.auth.jwt.sub":     "sub",
				"io.flipt.auth.jwt.email":   "email",
				"io.flipt.auth.jwt.picture": "image",
				"io.flipt.auth.jwt.name":    "name",
				"io.flipt.auth.jwt.issuer":  "flipt.io",
				"io.flipt.auth.role":        "admin",
			},
		},
		{
			name: "invalid issuer",
			metadataFunc: func() metadata.MD {
				claims := map[string]interface{}{
					"iss": "foo.com",
					"iat": nowUnix,
					"exp": futureUnix,
				}

				token := oidc.TestSignJWT(t, priv, string(jwt.RS256), claims, []byte("test-key"))
				return metadata.MD{
					"Authorization": []string{"JWT " + token},
				}
			},
			expectedJWT: jwt.Expected{
				Issuer: "flipt.io",
			},
			expectedErr: errUnauthenticated,
		},
		{
			name: "invalid subject",
			metadataFunc: func() metadata.MD {
				claims := map[string]interface{}{
					"iss": "flipt.io",
					"iat": nowUnix,
					"exp": futureUnix,
					"sub": "bar",
				}

				token := oidc.TestSignJWT(t, priv, string(jwt.RS256), claims, []byte("test-key"))
				return metadata.MD{
					"Authorization": []string{"JWT " + token},
				}
			},
			expectedJWT: jwt.Expected{
				Issuer:  "flipt.io",
				Subject: "flipt",
			},
			expectedErr: errUnauthenticated,
		},
		{
			name: "invalid audience",
			metadataFunc: func() metadata.MD {
				claims := map[string]interface{}{
					"iss": "flipt.io",
					"iat": nowUnix,
					"exp": futureUnix,
					"aud": "bar",
				}

				token := oidc.TestSignJWT(t, priv, string(jwt.RS256), claims, []byte("test-key"))
				return metadata.MD{
					"Authorization": []string{"JWT " + token},
				}
			},
			expectedJWT: jwt.Expected{
				Issuer:    "flipt.io",
				Audiences: []string{"flipt"},
			},
			expectedErr: errUnauthenticated,
		},
		{
			name: "successful authentication (skipped)",
			server: &mockServer{
				skipsAuthn: true,
			},
		},
		{
			name: "client token missing JWT prefix",
			metadataFunc: func() metadata.MD {
				return metadata.MD{
					"Authorization": []string{"blah"},
				}
			},
			expectedErr: errUnauthenticated,
		},
		{
			name: "authorization header not set",
			metadataFunc: func() metadata.MD {
				return metadata.MD{}
			},
			expectedErr: errUnauthenticated,
		},
		{
			name:        "no metadata on context",
			expectedErr: errUnauthenticated,
		},
	} {
		tt := tt

		t.Run(fmt.Sprintf("%s/static", tt.name), func(t *testing.T) {
			ks, err := jwt.NewStaticKeySet(pub)
			require.NoError(t, err)

			validator, err := jwt.NewValidator(ks)
			require.NoError(t, err)

			var (
				logger = zaptest.NewLogger(t)

				ctx     = context.Background()
				handler = func(ctx context.Context, req interface{}) (interface{}, error) {
					if tt.expectedMetadata != nil {
						authentication := GetAuthenticationFrom(ctx)

						for k, v := range authentication.Metadata {
							assert.Equal(t, tt.expectedMetadata[k], v)
						}
					}
					return nil, nil
				}
				srv = &grpc.UnaryServerInfo{Server: &mockServer{}}
			)

			if tt.metadataFunc != nil {
				ctx = metadata.NewIncomingContext(ctx, tt.metadataFunc())
			}

			if tt.server != nil {
				srv.Server = tt.server
			}

			_, err = JWTAuthenticationInterceptor(logger, *validator, tt.expectedJWT)(
				ctx,
				nil,
				srv,
				handler,
			)
			assert.Equal(t, tt.expectedErr, err)
		})

		t.Run(fmt.Sprintf("%s/remote", tt.name), func(t *testing.T) {
			tp := oidc.StartTestProvider(t, oidc.WithNoTLS())
			tp.SetSigningKeys(priv, priv.Public(), oidc.RS256, "test")

			ks, err := jwt.NewJSONWebKeySet(context.Background(), tp.Addr()+"/.well-known/jwks.json", "")
			require.NoError(t, err)

			validator, err := jwt.NewValidator(ks)
			require.NoError(t, err)

			var (
				logger = zaptest.NewLogger(t)

				ctx     = context.Background()
				handler = func(ctx context.Context, req interface{}) (interface{}, error) {
					if tt.expectedMetadata != nil {
						authentication := GetAuthenticationFrom(ctx)

						for k, v := range authentication.Metadata {
							assert.Equal(t, tt.expectedMetadata[k], v)
						}
					}
					return nil, nil
				}
				srv = &grpc.UnaryServerInfo{Server: &mockServer{}}
			)

			if tt.metadataFunc != nil {
				ctx = metadata.NewIncomingContext(ctx, tt.metadataFunc())
			}

			if tt.server != nil {
				srv.Server = tt.server
			}

			_, err = JWTAuthenticationInterceptor(logger, *validator, tt.expectedJWT)(
				ctx,
				nil,
				srv,
				handler,
			)
			assert.Equal(t, tt.expectedErr, err)
		})
	}
}

func TestClientTokenAuthenticationInterceptor(t *testing.T) {
	authenticator := memory.NewStore()

	clientToken, storedAuth, err := authenticator.CreateAuthentication(
		context.TODO(),
		&authn.CreateAuthenticationRequest{Method: authrpc.Method_METHOD_TOKEN},
	)

	require.NoError(t, err)

	// expired auth
	expiredToken, _, err := authenticator.CreateAuthentication(
		context.TODO(),
		&authn.CreateAuthenticationRequest{
			Method:    authrpc.Method_METHOD_TOKEN,
			ExpiresAt: timestamppb.New(time.Now().UTC().Add(-time.Hour)),
		},
	)

	require.NoError(t, err)

	for _, tt := range []struct {
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
				skipsAuthn: true,
			},
		},
		{
			name: "token has expired",
			metadata: metadata.MD{
				"Authorization": []string{"Bearer " + expiredToken},
			},
			expectedErr: errUnauthenticated,
		},
		{
			name: "client token not found in store",
			metadata: metadata.MD{
				"Authorization": []string{"Bearer unknowntoken"},
			},
			expectedErr: errUnauthenticated,
		},
		{
			name: "client token missing Bearer prefix",
			metadata: metadata.MD{
				"Authorization": []string{clientToken},
			},
			expectedErr: errUnauthenticated,
		},
		{
			name: "authorization header empty",
			metadata: metadata.MD{
				"Authorization": []string{},
			},
			expectedErr: errUnauthenticated,
		},
		{
			name: "cookie header with no flipt_client_token",
			metadata: metadata.MD{
				"grpcgateway-cookie": []string{"blah"},
			},
			expectedErr: errUnauthenticated,
		},
		{
			name:        "authorization header not set",
			metadata:    metadata.MD{},
			expectedErr: errUnauthenticated,
		},
		{
			name:        "no metadata on context",
			metadata:    nil,
			expectedErr: errUnauthenticated,
		},
	} {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			var (
				logger = zaptest.NewLogger(t)

				ctx          = context.Background()
				retrievedCtx = ctx
				handler      = func(ctx context.Context, req interface{}) (interface{}, error) {
					// update retrievedCtx to the one delegated to the handler
					retrievedCtx = ctx
					return nil, nil
				}
				srv = &grpc.UnaryServerInfo{Server: &mockServer{}}
			)

			if tt.metadata != nil {
				ctx = metadata.NewIncomingContext(ctx, tt.metadata)
			}

			if tt.server != nil {
				srv.Server = tt.server
			}

			_, err := ClientTokenAuthenticationInterceptor(logger, authenticator)(
				ctx,
				nil,
				srv,
				handler,
			)
			assert.Equal(t, tt.expectedErr, err)
			assert.Equal(t, tt.expectedAuth, GetAuthenticationFrom(retrievedCtx))
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
		&authn.CreateAuthenticationRequest{
			Method: authrpc.Method_METHOD_OIDC,
			Metadata: map[string]string{
				"io.flipt.auth.oidc.email": "foo@flipt.io",
			},
		},
	)
	require.NoError(t, err)

	nonEmailClientToken, nonEmailStoredAuth, err := authenticator.CreateAuthentication(
		context.TODO(),
		&authn.CreateAuthenticationRequest{
			Method:   authrpc.Method_METHOD_OIDC,
			Metadata: map[string]string{},
		},
	)
	require.NoError(t, err)

	staticClientToken, staticStoreAuth, err := authenticator.CreateAuthentication(
		context.TODO(),
		&authn.CreateAuthenticationRequest{
			Method:   authrpc.Method_METHOD_TOKEN,
			Metadata: map[string]string{},
		},
	)
	require.NoError(t, err)

	for _, tt := range []struct {
		name         string
		metadata     metadata.MD
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
			expectedErr: errUnauthenticated,
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
			expectedErr: errUnauthenticated,
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
			expectedErr: errUnauthenticated,
		},
	} {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			var (
				logger = zaptest.NewLogger(t)

				ctx     = ContextWithAuthentication(context.Background(), tt.auth)
				handler = func(ctx context.Context, req interface{}) (interface{}, error) {
					return nil, nil
				}
				srv = &grpc.UnaryServerInfo{Server: &mockServer{}}
			)

			if tt.metadata != nil {
				ctx = metadata.NewIncomingContext(ctx, tt.metadata)
			}

			rgxs := make([]*regexp.Regexp, 0, len(tt.emailMatches))

			for _, em := range tt.emailMatches {
				rgx, err := regexp.Compile(em)
				require.NoError(t, err)

				rgxs = append(rgxs, rgx)
			}

			_, err := EmailMatchingInterceptor(logger, rgxs)(
				ctx,
				nil,
				srv,
				handler,
			)
			assert.Equal(t, tt.expectedErr, err)
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
		authReq     *authn.CreateAuthenticationRequest
		req         any
		srv         any
		wantCalled  bool
		expectedErr error
	}{
		{
			name: "successful namespace match",
			authReq: &authn.CreateAuthenticationRequest{
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
			authReq: &authn.CreateAuthenticationRequest{
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
			authReq: &authn.CreateAuthenticationRequest{
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
			authReq: &authn.CreateAuthenticationRequest{
				Method: authrpc.Method_METHOD_TOKEN,
				Metadata: map[string]string{
					"io.flipt.auth.token.namespace": "foo",
				},
			},
			req: &struct{}{},
			srv: &mockServer{
				skipsAuthn: true,
			},
			wantCalled: true,
		},
		{
			name: "not a token authentication",
			authReq: &authn.CreateAuthenticationRequest{
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
			authReq: &authn.CreateAuthenticationRequest{
				Method: authrpc.Method_METHOD_TOKEN,
				Metadata: map[string]string{
					"io.flipt.auth.token.namespace": "bar",
				},
			},
			req: &evaluation.EvaluationRequest{
				NamespaceKey: "foo",
			},
			expectedErr: errUnauthenticated,
		},
		{
			name: "namespace not provided by token authentication",
			authReq: &authn.CreateAuthenticationRequest{
				Method: authrpc.Method_METHOD_TOKEN,
			},
			req: &evaluation.EvaluationRequest{
				NamespaceKey: "foo",
			},
			wantCalled: true,
		},
		{
			name: "empty namespace provided by token authentication",
			authReq: &authn.CreateAuthenticationRequest{
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
			authReq: &authn.CreateAuthenticationRequest{
				Method: authrpc.Method_METHOD_TOKEN,
				Metadata: map[string]string{
					"io.flipt.auth.token.namespace": "foo",
				},
			},
			req:         &evaluation.EvaluationRequest{},
			expectedErr: errUnauthenticated,
		},
		{
			name: "namespace not available",
			authReq: &authn.CreateAuthenticationRequest{
				Method: authrpc.Method_METHOD_TOKEN,
				Metadata: map[string]string{
					"io.flipt.auth.token.namespace": "foo",
				},
			},
			req:         &evaluation.BatchEvaluationRequest{},
			expectedErr: errUnauthenticated,
		},
		{
			name: "namespace not consistent for batch",
			authReq: &authn.CreateAuthenticationRequest{
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
			expectedErr: errUnauthenticated,
		},
		{
			name: "non-namespaced request",
			authReq: &authn.CreateAuthenticationRequest{
				Method: authrpc.Method_METHOD_TOKEN,
				Metadata: map[string]string{
					"io.flipt.auth.token.namespace": "foo",
				},
			},
			req:         &struct{}{},
			expectedErr: errUnauthenticated,
		},
		{
			name: "non-namespaced scoped server",
			authReq: &authn.CreateAuthenticationRequest{
				Method: authrpc.Method_METHOD_TOKEN,
				Metadata: map[string]string{
					"io.flipt.auth.token.namespace": "foo",
				},
			},
			req: &evaluation.EvaluationRequest{
				NamespaceKey: "foo",
			},
			srv: &mockServer{
				allowNamespacedAuthn: false,
			},
			expectedErr: errUnauthenticated,
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
					allowNamespacedAuthn: true,
				}}
			)

			if tt.srv != nil {
				srv.Server = tt.srv
			}

			ctx = metadata.NewIncomingContext(ctx, metadata.MD{
				"Authorization": []string{"Bearer " + clientToken},
			})

			_, err = NamespaceMatchingInterceptor(logger)(ctx, tt.req, srv, handler)
			assert.Equal(t, tt.expectedErr, err)
		})
	}
}
