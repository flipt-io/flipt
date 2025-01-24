package ofrep

import (
	"context"

	"go.flipt.io/flipt/rpc/flipt/ofrep"
)

// GetProviderConfiguration returns the configuration set by the running flipt instance.
func (s *Server) GetProviderConfiguration(_ context.Context, _ *ofrep.GetProviderConfigurationRequest) (*ofrep.GetProviderConfigurationResponse, error) {
	return &ofrep.GetProviderConfigurationResponse{
		Name: "flipt",
		Capabilities: &ofrep.Capabilities{
			FlagEvaluation: &ofrep.FlagEvaluation{
				SupportedTypes: []string{"string", "boolean"},
			},
		},
	}, nil
}
