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

type mockVerifier struct {
	claims claims
	err    error
}

func (m *mockVerifier) verify(ctx context.Context, jwt string) (claims, error) {
	return m.claims, m.err
}

func TestKubernetesAuthenticator_Authenticate(t *testing.T) {
	logger := zaptest.NewLogger(t)

	tests := []struct {
		name    string
		claims  claims
		err     error
		want    *auth.Authentication
		wantErr bool
	}{
		{
			name: "valid token",
			claims: claims{
				Expiration: time.Now().Add(time.Hour).Unix(),
				Identity: identity{
					Namespace: "default",
					Pod: resource{
						Name: "test-pod",
						UID:  "pod-uid",
					},
					ServiceAccount: resource{
						Name: "test-sa",
						UID:  "sa-uid",
					},
				},
			},
			want: &auth.Authentication{
				Method:    auth.Method_METHOD_KUBERNETES,
				ExpiresAt: timestamppb.New(time.Unix(time.Now().Add(time.Hour).Unix(), 0)),
				Metadata: map[string]string{
					metadataKeyNamespace:          "default",
					metadataKeyPodName:            "test-pod",
					metadataKeyPodUID:             "pod-uid",
					metadataKeyServiceAccountName: "test-sa",
					metadataKeyServiceAccountUID:  "sa-uid",
				},
			},
		},
		{
			name:    "invalid token",
			err:     assert.AnError,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ka := &KubernetesAuthenticator{
				logger: logger,
				verifier: &mockVerifier{
					claims: tt.claims,
					err:    tt.err,
				},
				config: config.AuthenticationMethodKubernetesConfig{},
			}

			got, err := ka.Authenticate(context.Background(), "test-token")
			if tt.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			// Compare everything except ExpiresAt timestamp which will be different
			got.ExpiresAt = tt.want.ExpiresAt
			assert.Equal(t, tt.want, got)
		})
	}
}
