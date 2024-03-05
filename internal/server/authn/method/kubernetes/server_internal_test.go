package kubernetes

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.flipt.io/flipt/internal/config"
	"go.flipt.io/flipt/internal/storage/authn/memory"
	"go.flipt.io/flipt/rpc/flipt/auth"
	"go.uber.org/zap/zaptest"
	"google.golang.org/protobuf/types/known/timestamppb"
)

var (
	errVerify        = errors.New("token invalid")
	staticID         = "staticID"
	staticToken      = "staticToken"
	staticTime       = timestamppb.New(time.Date(2023, 2, 17, 8, 0, 0, 0, time.UTC))
	staticExpiration = timestamppb.New(time.Date(2023, 2, 17, 12, 0, 0, 0, time.UTC))
)

func Test_Server_VerifyServiceAccount(t *testing.T) {
	var (
		ctx    = context.Background()
		logger = zaptest.NewLogger(t)
	)

	for _, test := range []struct {
		name          string
		verifier      mockTokenVerifier
		req           *auth.VerifyServiceAccountRequest
		expectedResp  *auth.VerifyServiceAccountResponse
		expectedErrIs error
	}{
		{
			name:     "invalid service account token",
			verifier: mockTokenVerifier{},
			req: &auth.VerifyServiceAccountRequest{
				ServiceAccountToken: "someinvalidtoken",
			},
			expectedErrIs: errVerify,
		},
		{
			name: "valid service account token",
			verifier: mockTokenVerifier{
				"somevalidtoken": claims{
					Expiration: staticExpiration.AsTime().Unix(),
					Identity: identity{
						Namespace: "applications",
						Pod: resource{
							Name: "booking-7d26f049-kdurb",
							UID:  "bd8299f9-c50f-4b76-af33-9d8e3ef2b850",
						},
						ServiceAccount: resource{
							Name: "booking",
							UID:  "4f18914e-f276-44b2-aebd-27db1d8f8def",
						},
					},
				},
			},
			req: &auth.VerifyServiceAccountRequest{
				ServiceAccountToken: "somevalidtoken",
			},
			expectedResp: &auth.VerifyServiceAccountResponse{
				ClientToken: staticToken,
				Authentication: &auth.Authentication{
					Id:        staticID,
					Method:    auth.Method_METHOD_KUBERNETES,
					ExpiresAt: staticExpiration,
					CreatedAt: staticTime,
					UpdatedAt: staticTime,
					Metadata: map[string]string{
						"io.flipt.auth.k8s.namespace":           "applications",
						"io.flipt.auth.k8s.pod.name":            "booking-7d26f049-kdurb",
						"io.flipt.auth.k8s.pod.uid":             "bd8299f9-c50f-4b76-af33-9d8e3ef2b850",
						"io.flipt.auth.k8s.serviceaccount.name": "booking",
						"io.flipt.auth.k8s.serviceaccount.uid":  "4f18914e-f276-44b2-aebd-27db1d8f8def",
					},
				},
			},
		},
	} {
		test := test
		t.Run(test.name, func(t *testing.T) {
			var (
				store = memory.NewStore(
					memory.WithIDGeneratorFunc(func() string { return staticID }),
					memory.WithTokenGeneratorFunc(func() string { return staticToken }),
					memory.WithNowFunc(func() *timestamppb.Timestamp { return staticTime }),
				)
				conf = config.AuthenticationConfig{}
			)

			server := &Server{
				logger: logger,
				store:  store,
				config: conf,
				// override token verifier for unit test
				verifier: test.verifier,
			}

			resp, err := server.VerifyServiceAccount(ctx, test.req)
			if test.expectedErrIs != nil {
				assert.ErrorIs(t, err, test.expectedErrIs)
				return
			}
			require.NoError(t, err)

			assert.Equal(t, test.expectedResp, resp)
		})
	}
}

func Test_Server_SkipsAuthentication(t *testing.T) {
	server := &Server{}
	assert.True(t, server.SkipsAuthentication(context.Background()))
}

type mockTokenVerifier map[string]claims

func (m mockTokenVerifier) verify(_ context.Context, jwt string) (claims, error) {
	claims, ok := m[jwt]
	if !ok {
		return claims, errVerify
	}

	return claims, nil
}
