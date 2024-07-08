package ofrep

import (
	"context"

	"go.flipt.io/flipt/rpc/flipt/ofrep"
)

// GetProviderConfiguration returns the configuration set by the running flipt instance.
func (s *Server) GetProviderConfiguration(_ context.Context, _ *ofrep.GetProviderConfigurationRequest) (*ofrep.GetProviderConfigurationResponse, error) {
	var pollingInterval uint32
	if s.cacheCfg.Enabled {
		pollingInterval = uint32(s.cacheCfg.TTL.Milliseconds())
	}

	return &ofrep.GetProviderConfigurationResponse{
		Name: "flipt",
		Capabilities: &ofrep.Capabilities{
			CacheInvalidation: &ofrep.CacheInvalidation{
				Polling: &ofrep.Polling{
					Enabled:              s.cacheCfg.Enabled,
					MinPollingIntervalMs: pollingInterval,
				},
			},
			FlagEvaluation: &ofrep.FlagEvaluation{
				SupportedTypes: []string{"string", "boolean"},
			},
		},
	}, nil
}
