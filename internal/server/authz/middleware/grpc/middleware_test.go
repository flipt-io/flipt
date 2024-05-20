package grpc_middleware

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
	authmiddlewaregrpc "go.flipt.io/flipt/internal/server/authn/middleware/grpc"
	"go.flipt.io/flipt/rpc/flipt"
	authrpc "go.flipt.io/flipt/rpc/flipt/auth"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

type mockPolicyVerifier struct {
	isAllowed bool
	wantErr   error
}

func (v *mockPolicyVerifier) IsAllowed(ctx context.Context, input map[string]interface{}) (bool, error) {
	return v.isAllowed, v.wantErr
}

// mockServer is used to test skipping authz
type mockServer struct {
	skipsAuthz bool
}

func (s *mockServer) SkipsAuthorization(ctx context.Context) bool {
	return s.skipsAuthz
}

type mockRequester struct {
	action  flipt.Action
	subject flipt.Subject
}

func (r *mockRequester) Request() flipt.Request {
	return flipt.Request{
		Action:  r.action,
		Subject: r.subject,
	}
}

func TestAuthorizationRequiredInterceptor(t *testing.T) {
	var tests = []struct {
		name             string
		server           any
		req              any
		authn            *authrpc.Authentication
		validatorAllowed bool
		validatorErr     error
		wantAllowed      bool
	}{
		{
			name: "allowed",
			authn: &authrpc.Authentication{
				Metadata: map[string]string{
					"io.flipt.auth.role": "admin",
				},
			},
			req:              &flipt.CreateFlagRequest{},
			validatorAllowed: true,
			wantAllowed:      true,
		},
		{
			name: "not allowed",
			authn: &authrpc.Authentication{
				Metadata: map[string]string{
					"io.flipt.auth.role": "admin",
				},
			},
			req:              &flipt.CreateFlagRequest{},
			validatorAllowed: false,
			wantAllowed:      false,
		},
		{
			name: "skips authz",
			server: &mockServer{
				skipsAuthz: true,
			},
			req:         &flipt.CreateFlagRequest{},
			wantAllowed: true,
		},
		{
			name:        "no auth",
			req:         &flipt.CreateFlagRequest{},
			wantAllowed: false,
		},
		{
			name: "no role",
			authn: &authrpc.Authentication{
				Metadata: map[string]string{},
			},
			req:         &flipt.CreateFlagRequest{},
			wantAllowed: false,
		},
		{
			name: "invalid request",
			authn: &authrpc.Authentication{
				Metadata: map[string]string{
					"io.flipt.auth.role": "admin",
				},
			},
			req:         struct{}{},
			wantAllowed: false,
		},
		{
			name: "missing action",
			authn: &authrpc.Authentication{
				Metadata: map[string]string{
					"io.flipt.auth.role": "admin",
				},
			},
			req: &mockRequester{
				subject: "subject",
			},
			wantAllowed: false,
		},
		{
			name: "missing subject",
			authn: &authrpc.Authentication{
				Metadata: map[string]string{
					"io.flipt.auth.role": "admin",
				},
			},
			req: &mockRequester{
				action: "action",
			},
			wantAllowed: false,
		},
		{
			name: "validator error",
			authn: &authrpc.Authentication{
				Metadata: map[string]string{
					"io.flipt.auth.role": "admin",
				},
			},
			req:          &flipt.CreateFlagRequest{},
			validatorErr: errors.New("error"),
			wantAllowed:  false,
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
				policyVerfier = &mockPolicyVerifier{isAllowed: tt.validatorAllowed, wantErr: tt.validatorErr}
			)

			if tt.server != nil {
				srv.Server = tt.server
			}

			_, err := AuthorizationRequiredInterceptor(logger, policyVerfier)(ctx, tt.req, srv, handler)

			require.Equal(t, tt.wantAllowed, allowed)

			if tt.wantAllowed {
				require.NoError(t, err)
				return
			}

			require.EqualError(t, err, errUnauthorized.Error())
		})
	}
}
