package grpc_middleware

import (
	"context"
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"encoding/json"
	"fmt"
	"regexp"
	"testing"
	"time"

	jjwt "github.com/go-jose/go-jose/v4/jwt"
	"github.com/hashicorp/cap/jwt"
	"github.com/hashicorp/cap/oidc"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	authjwt "go.flipt.io/flipt/internal/server/authn/method/jwt"
	"go.flipt.io/flipt/internal/storage/authn"
	"go.flipt.io/flipt/internal/storage/authn/memory"
	authrpc "go.flipt.io/flipt/rpc/flipt/auth"
	"go.uber.org/zap/zaptest"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// mockServer is used to test skipping authn
type mockServer struct {
	skipsAuthn bool
}

func (s *mockServer) SkipsAuthentication(ctx context.Context) bool {
	return s.skipsAuthn
}

type mockStream struct {
	grpc.ServerStream
	ctx context.Context
}

func (m *mockStream) Context() context.Context { return m.ctx }

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

func TestAuthenticationRequiredUnaryInterceptor(t *testing.T) {
	logger := zaptest.NewLogger(t)

	for _, tt := range []struct {
		name         string
		ctx          context.Context
		server       any
		expectErr    error
		expectCalled bool
	}{
		{
			name:         "authenticated context",
			ctx:          ContextWithAuthentication(context.Background(), &authrpc.Authentication{Method: authrpc.Method_METHOD_TOKEN}),
			expectErr:    nil,
			expectCalled: true,
		},
		{
			name:         "unauthenticated context",
			ctx:          context.Background(),
			expectErr:    errUnauthenticated,
			expectCalled: false,
		},
		{
			name:         "skipped server",
			ctx:          context.Background(),
			server:       &mockServer{skipsAuthn: true},
			expectErr:    nil,
			expectCalled: true,
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			called := false
			handler := func(ctx context.Context, req any) (any, error) {
				called = true
				return "ok", nil
			}
			info := &grpc.UnaryServerInfo{Server: tt.server}
			resp, err := AuthenticationRequiredUnaryInterceptor(logger)(tt.ctx, nil, info, handler)
			assert.Equal(t, tt.expectErr, err)
			assert.Equal(t, tt.expectCalled, called)
			if tt.expectCalled {
				assert.Equal(t, "ok", resp)
			}
		})
	}
}

func TestAuthenticationRequiredStreamInterceptor(t *testing.T) {
	logger := zaptest.NewLogger(t)

	for _, tt := range []struct {
		name         string
		ctx          context.Context
		server       any
		expectErr    error
		expectCalled bool
	}{
		{
			name:         "authenticated context",
			ctx:          ContextWithAuthentication(context.Background(), &authrpc.Authentication{Method: authrpc.Method_METHOD_TOKEN}),
			expectErr:    nil,
			expectCalled: true,
		},
		{
			name:         "unauthenticated context",
			ctx:          context.Background(),
			expectErr:    errUnauthenticated,
			expectCalled: false,
		},
		{
			name:         "skipped server",
			ctx:          context.Background(),
			server:       &mockServer{skipsAuthn: true},
			expectErr:    nil,
			expectCalled: true,
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			called := false
			handler := func(srv any, stream grpc.ServerStream) error {
				called = true
				return nil
			}
			info := &grpc.StreamServerInfo{}
			stream := &mockStream{ctx: tt.ctx}
			err := AuthenticationRequiredStreamInterceptor(logger)(tt.server, stream, info, handler)
			assert.Equal(t, tt.expectErr, err)
			assert.Equal(t, tt.expectCalled, called)
		})
	}
}

func TestJWTAuthenticationUnaryInterceptor(t *testing.T) {
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
				claims := map[string]any{
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
				claims := map[string]any{
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
				claims := map[string]any{
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
				claims := map[string]any{
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
				claims := map[string]any{
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
				claims := map[string]any{
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

		t.Run(fmt.Sprintf("%s/static", tt.name), func(t *testing.T) {
			ks, err := jwt.NewStaticKeySet(pub)
			require.NoError(t, err)

			validator, err := jwt.NewValidator(ks)
			require.NoError(t, err)

			jwtValidator := authjwt.NewValidator(validator, tt.expectedJWT)

			var (
				logger = zaptest.NewLogger(t)

				ctx     = context.Background()
				handler = func(ctx context.Context, req any) (any, error) {
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

			_, err = JWTAuthenticationUnaryInterceptor(logger, jwtValidator, nil)(
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

			jwtValidator := authjwt.NewValidator(validator, tt.expectedJWT)

			var (
				logger = zaptest.NewLogger(t)

				ctx     = context.Background()
				handler = func(ctx context.Context, req any) (any, error) {
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

			_, err = JWTAuthenticationUnaryInterceptor(logger, jwtValidator, nil)(
				ctx,
				nil,
				srv,
				handler,
			)
			assert.Equal(t, tt.expectedErr, err)
		})
	}
}

func TestJWTAuthenticationStreamInterceptor(t *testing.T) {
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
				claims := map[string]any{
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
			name: "invalid issuer",
			metadataFunc: func() metadata.MD {
				claims := map[string]any{
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
		{
			name:   "skipped server",
			server: &mockServer{skipsAuthn: true},
		},
	} {

		t.Run(tt.name, func(t *testing.T) {
			ks, err := jwt.NewStaticKeySet(pub)
			require.NoError(t, err)

			validator, err := jwt.NewValidator(ks)
			require.NoError(t, err)

			jwtValidator := authjwt.NewValidator(validator, tt.expectedJWT)

			logger := zaptest.NewLogger(t)

			ctx := context.Background()
			if tt.metadataFunc != nil {
				ctx = metadata.NewIncomingContext(ctx, tt.metadataFunc())
			}

			stream := &mockStream{ctx: ctx}
			called := false
			handler := func(srv any, stream grpc.ServerStream) error {
				called = true
				if tt.expectedMetadata != nil {
					authentication := GetAuthenticationFrom(stream.Context())
					for k, v := range tt.expectedMetadata {
						assert.Equal(t, v, authentication.Metadata[k])
					}
				}
				return nil
			}
			info := &grpc.StreamServerInfo{}

			var server any = &mockServer{}
			if tt.server != nil {
				server = tt.server
			}

			err = JWTAuthenticationStreamInterceptor(logger, jwtValidator, nil)(server, stream, info, handler)
			assert.Equal(t, tt.expectedErr, err)
			if tt.expectedErr == nil || tt.server != nil {
				assert.True(t, called)
			} else {
				assert.False(t, called)
			}
		})
	}
}

func TestClientTokenAuthenticationUnaryInterceptor(t *testing.T) {
	authenticator := memory.NewStore(zaptest.NewLogger(t))

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
		t.Run(tt.name, func(t *testing.T) {
			var (
				logger = zaptest.NewLogger(t)

				ctx          = context.Background()
				retrievedCtx = ctx
				handler      = func(ctx context.Context, req any) (any, error) {
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

			_, err := ClientTokenAuthenticationUnaryInterceptor(logger, authenticator)(
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

func TestClientTokenAuthenticationStreamInterceptor(t *testing.T) {
	authenticator := memory.NewStore(zaptest.NewLogger(t))

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
			server:   &mockServer{skipsAuthn: true},
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
		t.Run(tt.name, func(t *testing.T) {
			logger := zaptest.NewLogger(t)

			ctx := context.Background()
			if tt.metadata != nil {
				ctx = metadata.NewIncomingContext(ctx, tt.metadata)
			}

			stream := &mockStream{ctx: ctx}
			var retrievedAuth *authrpc.Authentication
			called := false
			handler := func(srv any, stream grpc.ServerStream) error {
				called = true
				retrievedAuth = GetAuthenticationFrom(stream.Context())
				return nil
			}
			info := &grpc.StreamServerInfo{}

			var server any = &mockServer{}
			if tt.server != nil {
				server = tt.server
			}

			err := ClientTokenStreamInterceptor(logger, authenticator)(server, stream, info, handler)
			assert.Equal(t, tt.expectedErr, err)
			if tt.expectedErr == nil || tt.server != nil {
				assert.True(t, called)
			} else {
				assert.False(t, called)
			}
			assert.Equal(t, tt.expectedAuth, retrievedAuth)
		})
	}
}

func TestEmailMatchingUnaryInterceptorWithNoAuth(t *testing.T) {
	var (
		logger = zaptest.NewLogger(t)

		ctx     = context.Background()
		handler = func(ctx context.Context, req any) (any, error) {
			return nil, nil
		}
	)

	require.Panics(t, func() {
		_, _ = EmailMatchingUnaryInterceptor(logger, []*regexp.Regexp{regexp.MustCompile("^.*@flipt.io$")})(
			ctx,
			nil,
			&grpc.UnaryServerInfo{Server: &mockServer{}},
			handler,
		)
	})
}

func TestEmailMatchingUnaryInterceptor(t *testing.T) {
	var (
		logger = zaptest.NewLogger(t)

		authenticator = memory.NewStore(logger)
	)

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
		t.Run(tt.name, func(t *testing.T) {
			var (
				ctx     = ContextWithAuthentication(context.Background(), tt.auth)
				handler = func(ctx context.Context, req any) (any, error) {
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

			_, err := EmailMatchingUnaryInterceptor(logger, rgxs)(
				ctx,
				nil,
				srv,
				handler,
			)
			assert.Equal(t, tt.expectedErr, err)
		})
	}
}

func TestJwtClaimsToMetadata(t *testing.T) {
	tests := []struct {
		name     string
		claims   map[string]interface{}
		expected map[string]string
	}{
		{
			name:     "empty claims",
			claims:   map[string]interface{}{},
			expected: map[string]string{},
		},
		{
			name: "issuer claim",
			claims: map[string]interface{}{
				"iss": "flipt.io",
			},
			expected: map[string]string{
				"io.flipt.auth.jwt.issuer": "flipt.io",
			},
		},
		{
			name: "flipt auth prefixed claims",
			claims: map[string]interface{}{
				"io.flipt.auth.role":   "admin",
				"io.flipt.auth.tenant": "acme",
			},
			expected: map[string]string{
				"io.flipt.auth.role":   "admin",
				"io.flipt.auth.tenant": "acme",
			},
		},
		{
			name: "user claims",
			claims: map[string]interface{}{
				"user": map[string]interface{}{
					"email": "user@example.com",
					"sub":   "12345",
					"image": "https://example.com/avatar.jpg",
					"name":  "John Doe",
					"role":  "user",
				},
			},
			expected: map[string]string{
				"io.flipt.auth.jwt.email":   "user@example.com",
				"io.flipt.auth.jwt.sub":     "12345",
				"io.flipt.auth.jwt.picture": "https://example.com/avatar.jpg",
				"io.flipt.auth.jwt.name":    "John Doe",
				"io.flipt.auth.jwt.role":    "user",
			},
		},
		{
			name: "combined claims",
			claims: map[string]interface{}{
				"iss":                 "flipt.io",
				"io.flipt.auth.scope": "read",
				"user": map[string]interface{}{
					"email": "admin@flipt.io",
					"name":  "Admin User",
				},
			},
			expected: map[string]string{
				"io.flipt.auth.jwt.issuer": "flipt.io",
				"io.flipt.auth.scope":      "read",
				"io.flipt.auth.jwt.email":  "admin@flipt.io",
				"io.flipt.auth.jwt.name":   "Admin User",
			},
		},
		{
			name: "non-string issuer ignored",
			claims: map[string]interface{}{
				"iss": 12345,
			},
			expected: map[string]string{},
		},
		{
			name: "user claim not a map",
			claims: map[string]interface{}{
				"user": "not-a-map",
			},
			expected: map[string]string{},
		},
		{
			name: "mixed user claim types",
			claims: map[string]interface{}{
				"user": map[string]interface{}{
					"email":   "test@example.com",
					"sub":     123,
					"name":    "Test User",
					"unknown": "ignored",
					"picture": nil,
				},
			},
			expected: map[string]string{
				"io.flipt.auth.jwt.email": "test@example.com",
				"io.flipt.auth.jwt.sub":   "123",
				"io.flipt.auth.jwt.name":  "Test User",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := jwtClaimsToMetadata(tt.claims, nil)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestJwtClaimsToMetadataWithClaimsMapping(t *testing.T) {
	tests := []struct {
		name          string
		claimsJSON    string
		claimsMapping map[string]string
		expected      map[string]string
	}{
		{
			name:          "empty claims mapping uses defaults",
			claimsJSON:    `{"user": {"email": "test@example.com"}}`,
			claimsMapping: map[string]string{},
			expected:      map[string]string{"io.flipt.auth.jwt.email": "test@example.com"},
		},
		{
			name: "simple JSON pointer mapping",
			claimsJSON: `{
                                "user": {
                                        "email": "user@example.com",
                                        "name": "John Doe"
                                }
                        }`,
			claimsMapping: map[string]string{
				"email": "/user/email",
				"name":  "/user/name",
			},
			expected: map[string]string{
				"io.flipt.auth.jwt.email": "user@example.com",
				"io.flipt.auth.jwt.name":  "John Doe",
			},
		},
		{
			name: "nested JSON pointer mapping",
			claimsJSON: `{
                                "profile": {
                                        "personal": {
                                                "email": "deep@example.com"
                                        }
                                },
                                "sub": "123456"
                        }`,
			claimsMapping: map[string]string{
				"email": "/profile/personal/email",
				"sub":   "/sub",
			},
			expected: map[string]string{
				"io.flipt.auth.jwt.email": "deep@example.com",
				"io.flipt.auth.jwt.sub":   "123456",
			},
		},
		{
			name: "invalid JSON pointer ignored",
			claimsJSON: `{
                                "user": {
                                        "email": "valid@example.com"
                                }
                        }`,
			claimsMapping: map[string]string{
				"email": "/user/email",
				"name":  "/user/nonexistent",
			},
			expected: map[string]string{
				"io.flipt.auth.jwt.email": "valid@example.com",
			},
		},
		{
			name: "preserve flipt auth prefixed claims",
			claimsJSON: `{
                                "io.flipt.auth.role": "admin",
                                "user": {
                                        "email": "admin@example.com"
                                }
                        }`,
			claimsMapping: map[string]string{
				"email": "/user/email",
			},
			expected: map[string]string{
				"io.flipt.auth.role":      "admin",
				"io.flipt.auth.jwt.email": "admin@example.com",
			},
		},
		{
			name: "preserve issuer claim",
			claimsJSON: `{
                                "iss": "flipt.io",
                                "user": {
                                        "email": "user@flipt.io"
                                }
                        }`,
			claimsMapping: map[string]string{
				"email": "/user/email",
			},
			expected: map[string]string{
				"io.flipt.auth.jwt.issuer": "flipt.io",
				"io.flipt.auth.jwt.email":  "user@flipt.io",
			},
		},
		{
			name: "custom mapping merged with defaults",
			claimsJSON: `{
                                "user": {
                                        "email": "user@example.com",
                                        "name": "John Doe"
                                },
                                "name": "Custom Name"
                        }`,
			claimsMapping: map[string]string{
				"name": "/name", // Override default path for name
			},
			expected: map[string]string{
				"io.flipt.auth.jwt.email": "user@example.com", // From default path
				"io.flipt.auth.jwt.name":  "Custom Name",      // From custom path
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var claims map[string]interface{}
			err := json.Unmarshal([]byte(tt.claimsJSON), &claims)
			require.NoError(t, err)

			result := jwtClaimsToMetadata(claims, tt.claimsMapping)
			assert.Equal(t, tt.expected, result)
		})
	}
}
