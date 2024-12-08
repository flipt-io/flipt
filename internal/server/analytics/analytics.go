package analytics

import (
	"context"
	"time"

	"go.flipt.io/flipt/rpc/flipt/analytics"
)

// GetFlagEvaluationsCount is the implemented RPC method that will return aggregated flag evaluation counts.
func (s *Server) GetFlagEvaluationsCount(ctx context.Context, req *analytics.GetFlagEvaluationsCountRequest) (*analytics.GetFlagEvaluationsCountResponse, error) {
	fromTime, err := time.Parse(time.RFC3339, req.From)
	if err != nil {
		var innerErr error
		fromTime, innerErr = time.Parse(time.DateTime, req.From)
		if innerErr != nil {
			return nil, err
		}
	}

	toTime, err := time.Parse(time.RFC3339, req.To)
	if err != nil {
		var innerErr error
		toTime, innerErr = time.Parse(time.DateTime, req.To)
		if innerErr != nil {
			return nil, err
		}
	}

	r := &FlagEvaluationsCountRequest{
		NamespaceKey: req.NamespaceKey,
		FlagKey:      req.FlagKey,
		From:         fromTime,
		To:           toTime,
		StepMinutes:  getStepFromDuration(toTime.Sub(fromTime)),
	}

	timestamps, values, err := s.client.GetFlagEvaluationsCount(ctx, r)
	if err != nil {
		return nil, err
	}

	return &analytics.GetFlagEvaluationsCountResponse{
		Timestamps: timestamps,
		Values:     values,
	}, nil
}

// getStepFromDuration is a utility function that translates the duration passed in from the client
// to determine the interval steps we should use in minutes
func getStepFromDuration(duration time.Duration) int {
	switch {
	case duration >= 24*time.Hour:
		return 30
	case duration >= 12*time.Hour:
		return 15
	case duration >= 4*time.Hour:
		return 5
	default:
		return 1
	}
}
