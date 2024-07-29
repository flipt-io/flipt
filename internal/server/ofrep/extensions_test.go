package ofrep

import (
	"context"
	"testing"
	"time"

	"go.uber.org/zap/zaptest"

	"github.com/stretchr/testify/require"
	"go.flipt.io/flipt/internal/config"
	"go.flipt.io/flipt/rpc/flipt/ofrep"
)

func TestGetProviderConfiguration(t *testing.T) {
	testCases := []struct {
		name         string
		cfg          config.CacheConfig
		expectedResp *ofrep.GetProviderConfigurationResponse
	}{
		{
			name: "should return the configuration correctly",
			cfg: config.CacheConfig{
				Enabled: true,
				TTL:     time.Second,
			},
			expectedResp: &ofrep.GetProviderConfigurationResponse{
				Name: "flipt",
				Capabilities: &ofrep.Capabilities{
					CacheInvalidation: &ofrep.CacheInvalidation{
						Polling: &ofrep.Polling{
							Enabled:              true,
							MinPollingIntervalMs: 1000,
						},
					},
					FlagEvaluation: &ofrep.FlagEvaluation{
						SupportedTypes: []string{"string", "boolean"},
					},
				},
			},
		},
		{
			name: "should return the min pool interval with zero value when the cache is not enabled",
			cfg: config.CacheConfig{
				Enabled: false,
				TTL:     time.Second,
			},
			expectedResp: &ofrep.GetProviderConfigurationResponse{
				Name: "flipt",
				Capabilities: &ofrep.Capabilities{
					CacheInvalidation: &ofrep.CacheInvalidation{
						Polling: &ofrep.Polling{
							Enabled:              false,
							MinPollingIntervalMs: 0,
						},
					},
					FlagEvaluation: &ofrep.FlagEvaluation{
						SupportedTypes: []string{"string", "boolean"},
					},
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			s := New(zaptest.NewLogger(t), tc.cfg, &bridgeMock{})

			resp, err := s.GetProviderConfiguration(context.TODO(), &ofrep.GetProviderConfigurationRequest{})

			require.NoError(t, err)
			require.Equal(t, tc.expectedResp, resp)
		})
	}
}
