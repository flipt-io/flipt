package analytics

import (
	"context"
	"time"

	"go.flipt.io/flipt/rpc/flipt/analytics"
)

// GetFlagEvaluationsCount is the implemented RPC method that will return aggregated flag evaluation counts.
func (s *Server) GetFlagEvaluationsCount(ctx context.Context, req *analytics.GetFlagEvaluationsCountRequest) (*analytics.GetFlagEvaluationsCountResponse, error) {
	duration, err := time.ParseDuration(req.From)
	if err != nil {
		return nil, err
	}

	timestamps, values, err := s.client.GetFlagEvaluationsCount(ctx, req.NamespaceKey, req.FlagKey, duration)
	if err != nil {
		return nil, err
	}

	return &analytics.GetFlagEvaluationsCountResponse{
		Timestamps: timestamps,
		Values:     values,
	}, nil
}
