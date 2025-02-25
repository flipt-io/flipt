package ofrep

import (
	"context"
	"testing"

	"go.uber.org/zap/zaptest"

	"github.com/stretchr/testify/require"
	"go.flipt.io/flipt/rpc/flipt/ofrep"
)

func TestGetProviderConfiguration(t *testing.T) {
	testCases := []struct {
		name         string
		expectedResp *ofrep.GetProviderConfigurationResponse
	}{
		{
			name: "should return the configuration correctly",
			expectedResp: &ofrep.GetProviderConfigurationResponse{
				Name: "flipt",
				Capabilities: &ofrep.Capabilities{
					FlagEvaluation: &ofrep.FlagEvaluation{
						SupportedTypes: []string{"string", "boolean"},
					},
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			s := New(zaptest.NewLogger(t), nil, nil)

			resp, err := s.GetProviderConfiguration(context.TODO(), &ofrep.GetProviderConfigurationRequest{})

			require.NoError(t, err)
			require.Equal(t, tc.expectedResp, resp)
		})
	}
}
