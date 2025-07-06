package kubernetes

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.flipt.io/flipt/internal/config"
	"go.flipt.io/flipt/rpc/flipt/auth"
	"go.uber.org/zap/zaptest"
	"google.golang.org/protobuf/types/known/timestamppb"
)

var (
	errVerify        = errors.New("token invalid")
	staticTime       = timestamppb.New(time.Date(2023, 2, 17, 8, 0, 0, 0, time.UTC))
	staticExpiration = timestamppb.New(time.Date(2023, 2, 17, 12, 0, 0, 0, time.UTC))
)

const (
	staticToken = "somevalidtoken"
)

func Test_Server_VerifyServiceAccount(t *testing.T) {
	var (
		ctx    = context.Background()
		logger = zaptest.NewLogger(t)
	)

	for _, test := range []struct {
		name          string
		validator     mockTokenValidator
		req           *auth.VerifyServiceAccountRequest
		expectedResp  *auth.VerifyServiceAccountResponse
		expectedErrIs error
	}{
		{
			name:      "invalid service account token",
			validator: mockTokenValidator{},
			req: &auth.VerifyServiceAccountRequest{
				ServiceAccountToken: "someinvalidtoken",
			},
			expectedErrIs: errVerify,
		},
		{
			name: "valid service account token",
			validator: mockTokenValidator{
				staticToken: claims{
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
		t.Run(test.name, func(t *testing.T) {

			conf := config.AuthenticationConfig{}

			server := &Server{
				logger: logger,
				config: conf,
				// override now for unit test
				now: func() *timestamppb.Timestamp {
					return staticTime
				},
				// override token validator for unit test
				validator: test.validator,
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

type mockTokenValidator map[string]claims

func (m mockTokenValidator) Validate(_ context.Context, jwt string) (map[string]any, error) {
	claims, ok := m[jwt]
	if !ok {
		return nil, errVerify
	}

	jsonBytes, err := json.Marshal(claims)
	if err != nil {
		return nil, err
	}

	var result map[string]any
	if err := json.Unmarshal(jsonBytes, &result); err != nil {
		return nil, err
	}

	return result, nil
}
