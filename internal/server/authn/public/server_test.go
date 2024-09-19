package public

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.flipt.io/flipt/internal/config"
	"go.flipt.io/flipt/rpc/flipt/auth"
	"go.uber.org/zap/zaptest"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/structpb"
)

func Test_Server(t *testing.T) {
	for _, test := range []struct {
		name         string
		ctx          context.Context
		conf         config.AuthenticationConfig
		expectedErr  error
		expectedResp *auth.ListAuthenticationMethodsResponse
	}{
		{
			name: "github and single oidc provider methods enabled",
			ctx:  context.Background(),
			conf: config.AuthenticationConfig{
				Required: true,
				Methods: config.AuthenticationMethods{
					Github: config.AuthenticationMethod[config.AuthenticationMethodGithubConfig]{
						Enabled: true,
						Method: config.AuthenticationMethodGithubConfig{
							RedirectAddress: "some.host.com",
						},
					},
					OIDC: config.AuthenticationMethod[config.AuthenticationMethodOIDCConfig]{
						Enabled: true,
						Method: config.AuthenticationMethodOIDCConfig{
							Providers: map[string]config.AuthenticationMethodOIDCProvider{
								"someprovider": {
									IssuerURL:       "some.issuer.com",
									RedirectAddress: "some.host.com",
								},
							},
						},
					},
				},
			},
			expectedResp: responseBuilder(t, func(m *methods) {
				m.github.Enabled = true
				m.oidc.Enabled = true
				m.oidc.Metadata.Fields["providers"] = structpb.NewStructValue(&structpb.Struct{
					Fields: map[string]*structpb.Value{
						"someprovider": structpb.NewStructValue(&structpb.Struct{
							Fields: map[string]*structpb.Value{
								"authorize_url": structpb.NewStringValue("/auth/v1/method/oidc/someprovider/authorize"),
								"callback_url":  structpb.NewStringValue("/auth/v1/method/oidc/someprovider/callback"),
							},
						}),
					},
				})
			}),
		},
		{
			name: "github and single oidc provider methods enabled with prefix",
			ctx: metadata.NewIncomingContext(context.Background(), metadata.New(
				map[string]string{
					"x-forwarded-prefix": "/someprefix",
				},
			)),
			conf: config.AuthenticationConfig{
				Required: true,
				Methods: config.AuthenticationMethods{
					Github: config.AuthenticationMethod[config.AuthenticationMethodGithubConfig]{
						Enabled: true,
						Method: config.AuthenticationMethodGithubConfig{
							RedirectAddress: "some.host.com",
						},
					},
					OIDC: config.AuthenticationMethod[config.AuthenticationMethodOIDCConfig]{
						Enabled: true,
						Method: config.AuthenticationMethodOIDCConfig{
							Providers: map[string]config.AuthenticationMethodOIDCProvider{
								"someprovider": {
									IssuerURL:       "some.issuer.com",
									RedirectAddress: "some.host.com",
								},
							},
						},
					},
				},
			},
			expectedResp: responseBuilder(t, func(m *methods) {
				m.github.Enabled = true
				m.github.Metadata.Fields = map[string]*structpb.Value{
					"authorize_url": structpb.NewStringValue("/someprefix/auth/v1/method/github/authorize"),
					"callback_url":  structpb.NewStringValue("/someprefix/auth/v1/method/github/callback"),
				}

				m.oidc.Enabled = true
				m.oidc.Metadata.Fields["providers"] = structpb.NewStructValue(&structpb.Struct{
					Fields: map[string]*structpb.Value{
						"someprovider": structpb.NewStructValue(&structpb.Struct{
							Fields: map[string]*structpb.Value{
								"authorize_url": structpb.NewStringValue("/someprefix/auth/v1/method/oidc/someprovider/authorize"),
								"callback_url":  structpb.NewStringValue("/someprefix/auth/v1/method/oidc/someprovider/callback"),
							},
						}),
					},
				})
			}),
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			server := NewServer(zaptest.NewLogger(t), test.conf)
			resp, err := server.ListAuthenticationMethods(test.ctx, &emptypb.Empty{})
			if test.expectedErr != nil {
				assert.Equal(t, test.expectedErr, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, test.expectedResp, resp)
		})
	}
}

type methods struct {
	token      *auth.MethodInfo
	github     *auth.MethodInfo
	oidc       *auth.MethodInfo
	kubernetes *auth.MethodInfo
	jwt        *auth.MethodInfo
	cloud      *auth.MethodInfo
}

func responseBuilder(t testing.TB, fn func(*methods)) *auth.ListAuthenticationMethodsResponse {
	t.Helper()
	newInfo := func(method auth.Method, session bool, meta *structpb.Struct) *auth.MethodInfo {
		return &auth.MethodInfo{
			Method:            method,
			Enabled:           false,
			SessionCompatible: session,
			Metadata:          meta,
		}
	}

	methods := methods{
		token: newInfo(auth.Method_METHOD_TOKEN, false, nil),
		github: newInfo(auth.Method_METHOD_GITHUB, true, &structpb.Struct{
			Fields: map[string]*structpb.Value{
				"authorize_url": structpb.NewStringValue("/auth/v1/method/github/authorize"),
				"callback_url":  structpb.NewStringValue("/auth/v1/method/github/callback"),
			},
		}),
		oidc: newInfo(auth.Method_METHOD_OIDC, true, &structpb.Struct{
			Fields: map[string]*structpb.Value{
				"providers": structpb.NewStructValue(&structpb.Struct{
					Fields: map[string]*structpb.Value{},
				}),
			},
		}),
		kubernetes: newInfo(auth.Method_METHOD_KUBERNETES, false, nil),
		jwt:        newInfo(auth.Method_METHOD_JWT, false, nil),
		cloud:      newInfo(auth.Method_METHOD_CLOUD, true, nil),
	}

	fn(&methods)

	return &auth.ListAuthenticationMethodsResponse{
		Methods: []*auth.MethodInfo{
			methods.token,
			methods.github,
			methods.oidc,
			methods.kubernetes,
			methods.jwt,
			methods.cloud,
		},
	}
}
