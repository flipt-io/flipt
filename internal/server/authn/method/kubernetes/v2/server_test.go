package v2

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.flipt.io/flipt/internal/config"
	"go.flipt.io/flipt/rpc/flipt/auth"
	"go.uber.org/zap/zaptest"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type mockAuthenticator struct {
	authentication *auth.Authentication
	err            error
}

func (m *mockAuthenticator) Authenticate(ctx context.Context, token string) (*auth.Authentication, error) {
	return m.authentication, m.err
}

func TestServer_VerifyServiceAccount(t *testing.T) {
	logger := zaptest.NewLogger(t)

	tests := []struct {
		name           string
		token          string
		authentication *auth.Authentication
		err            error
		want           *auth.VerifyServiceAccountResponse
		wantErr        bool
	}{
		{
			name:  "valid token",
			token: "test-token",
			authentication: &auth.Authentication{
				Method:    auth.Method_METHOD_KUBERNETES,
				ExpiresAt: timestamppb.New(time.Now().Add(time.Hour)),
				Metadata: map[string]string{
					metadataKeyNamespace:          "default",
					metadataKeyPodName:            "test-pod",
					metadataKeyPodUID:             "pod-uid",
					metadataKeyServiceAccountName: "test-sa",
					metadataKeyServiceAccountUID:  "sa-uid",
				},
			},
			want: &auth.VerifyServiceAccountResponse{
				ClientToken: "test-token",
				Authentication: &auth.Authentication{
					Method:    auth.Method_METHOD_KUBERNETES,
					ExpiresAt: timestamppb.New(time.Now().Add(time.Hour)),
					Metadata: map[string]string{
						metadataKeyNamespace:          "default",
						metadataKeyPodName:            "test-pod",
						metadataKeyPodUID:             "pod-uid",
						metadataKeyServiceAccountName: "test-sa",
						metadataKeyServiceAccountUID:  "sa-uid",
					},
				},
			},
		},
		{
			name:    "invalid token",
			token:   "invalid-token",
			err:     assert.AnError,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Server{
				logger: logger,
				authenticator: &mockAuthenticator{
					authentication: tt.authentication,
					err:            tt.err,
				},
				config: config.AuthenticationConfig{},
			}

			got, err := s.VerifyServiceAccount(context.Background(), &auth.VerifyServiceAccountRequest{
				ServiceAccountToken: tt.token,
			})

			if tt.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			// Compare everything except ExpiresAt timestamp which will be different
			got.Authentication.ExpiresAt = tt.want.Authentication.ExpiresAt
			assert.Equal(t, tt.want, got)
		})
	}
}
