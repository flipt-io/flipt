package ofrep

import (
	"context"

	"go.flipt.io/flipt/rpc/flipt/ofrep"
)

func (s *Server) GetProviderConfiguration(_ context.Context, _ *ofrep.GetProviderConfigurationRequest) (*ofrep.GetProviderConfigurationResponse, error) {
	var pollingInterval int64
	if s.cacheCfg.Enabled {
		pollingInterval = s.cacheCfg.TTL.Milliseconds()
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
