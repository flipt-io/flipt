package grpc_middleware

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	authmiddlewaregrpc "go.flipt.io/flipt/internal/server/authn/middleware/grpc"
	"go.flipt.io/flipt/rpc/flipt"
	authrpc "go.flipt.io/flipt/rpc/flipt/auth"
	"go.flipt.io/flipt/rpc/v2/environments"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

type mockPolicyVerifier struct {
	isAllowed          bool
	wantErr            error
	input              map[string]any
	viewableNamespaces []string
}

func (v *mockPolicyVerifier) IsAllowed(ctx context.Context, input map[string]any) (bool, error) {
	v.input = input
	return v.isAllowed, v.wantErr
}

func (v *mockPolicyVerifier) Namespaces(_ context.Context, _ map[string]any) ([]string, error) {
	if v.viewableNamespaces != nil {
		return v.viewableNamespaces, nil
	}
	return nil, errors.ErrUnsupported
}

func (v *mockPolicyVerifier) Shutdown(_ context.Context) error {
	return nil
}

// mockServer is used to test skipping authz
type mockServer struct {
	skipsAuthz bool
}

func (s *mockServer) SkipsAuthorization(ctx context.Context) bool {
	return s.skipsAuthz
}

var adminAuth = &authrpc.Authentication{
	Metadata: map[string]string{
		"io.flipt.auth.role": "admin",
	},
}

func TestAuthorizationRequiredInterceptor(t *testing.T) {
	tests := []struct {
		name               string
		server             any
		req                any
		authn              *authrpc.Authentication
		validatorAllowed   bool
		validatorErr       error
		wantAllowed        bool
		authzInput         map[string]any
		viewableNamespaces []string
		serverFullMethod   string
	}{
		{
			name:  "allowed",
			authn: adminAuth,
			req: &environments.UpdateResourceRequest{
				Environment: "default",
				Namespace:   "default",
				Key:         "some_flag",
			},
			validatorAllowed: true,
			wantAllowed:      true,
			authzInput: map[string]any{
				"request": flipt.Request{
					Namespace: "default",
					Resource:  flipt.ResourceFlag,
					Subject:   flipt.SubjectFlag,
					Action:    flipt.ActionCreate,
					Status:    flipt.StatusSuccess,
				},
				"authentication": adminAuth,
			},
		},
		{
			name:  "not allowed",
			authn: adminAuth,
			req: &environments.UpdateResourceRequest{
				Environment: "default",
				Namespace:   "default",
				Key:         "some_other_flag",
			},
			validatorAllowed: false,
			wantAllowed:      false,
			authzInput: map[string]any{
				"request": flipt.Request{
					Namespace: "default",
					Resource:  flipt.ResourceFlag,
					Subject:   flipt.SubjectFlag,
					Action:    flipt.ActionCreate,
				},
				"authentication": adminAuth,
			},
		},
		{
			name: "skips authz",
			server: &mockServer{
				skipsAuthz: true,
			},
			req:         &environments.UpdateResourceRequest{},
			wantAllowed: true,
		},
		{
			name:        "no auth",
			req:         &environments.UpdateResourceRequest{},
			wantAllowed: false,
		},
		{
			name:        "invalid request",
			authn:       adminAuth,
			req:         struct{}{},
			wantAllowed: false,
		},
		{
			name:         "validator error",
			authn:        adminAuth,
			req:          &environments.UpdateResourceRequest{},
			validatorErr: errors.New("error"),
			wantAllowed:  false,
		},
		{
			name:               "list namespaces",
			authn:              adminAuth,
			req:                &environments.ListNamespacesRequest{},
			wantAllowed:        true,
			viewableNamespaces: []string{"special"},
			serverFullMethod:   "/environments.EnvironmentsService/ListNamespaces",
			authzInput: map[string]any{
				"request": flipt.Request{
					Resource: flipt.ResourceNamespace,
					Action:   flipt.ActionRead,
					Status:   flipt.StatusSuccess,
				},
				"authentication": adminAuth,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var (
				logger  = zap.NewNop()
				allowed = false

				ctx     = authmiddlewaregrpc.ContextWithAuthentication(context.Background(), tt.authn)
				handler = func(ctx context.Context, req interface{}) (interface{}, error) {
					allowed = true
					return nil, nil
				}

				srv           = &grpc.UnaryServerInfo{Server: &mockServer{}}
				policyVerfier = &mockPolicyVerifier{
					isAllowed:          tt.validatorAllowed,
					wantErr:            tt.validatorErr,
					viewableNamespaces: tt.viewableNamespaces,
				}
			)

			if tt.server != nil {
				srv.Server = tt.server
			}
			srv.FullMethod = tt.serverFullMethod

			_, err := AuthorizationRequiredInterceptor(logger, policyVerfier)(ctx, tt.req, srv, handler)

			require.Equal(t, tt.wantAllowed, allowed)

			if tt.wantAllowed {
				require.NoError(t, err)
				assert.Equal(t, tt.authzInput, policyVerfier.input)
				return
			}

			require.Error(t, err)
		})
	}
}
