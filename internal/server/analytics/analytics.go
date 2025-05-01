package analytics

import (
	"context"
	"time"

	"go.flipt.io/flipt/rpc/v2/analytics"
)

var _ analytics.AnalyticsServiceServer = (*Server)(nil)

// FlagEvaluationsCountRequest represents a request to get flag evaluation counts.
type FlagEvaluationsCountRequest struct {
	EnvironmentKey string
	NamespaceKey   string
	FlagKey        string
	From           time.Time
	To             time.Time
	StepMinutes    int
}

// BatchFlagEvaluationsCountRequest represents a request to get evaluation counts for multiple flags.
type BatchFlagEvaluationsCountRequest struct {
	EnvironmentKey string
	NamespaceKey   string
	FlagKeys       []string
	From           time.Time
	To             time.Time
	StepMinutes    int
	Limit          int
}

// FlagEvaluationData represents time series data for a single flag
type FlagEvaluationData struct {
	Timestamps []string
	Values     []float32
}

// Client interface for analytics operations
type Client interface {
	GetFlagEvaluationsCount(ctx context.Context, req *FlagEvaluationsCountRequest) ([]string, []float32, error)
	GetBatchFlagEvaluationsCount(ctx context.Context, req *BatchFlagEvaluationsCountRequest) (map[string]FlagEvaluationData, error)
	String() string
}

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
		EnvironmentKey: req.EnvironmentKey,
		NamespaceKey:   req.NamespaceKey,
		FlagKey:        req.FlagKey,
		From:           fromTime,
		To:             toTime,
		StepMinutes:    getStepFromDuration(toTime.Sub(fromTime)),
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

// GetBatchFlagEvaluationsCount handles requests for evaluation counts of multiple flags in a single request.
func (s *Server) GetBatchFlagEvaluationsCount(ctx context.Context, req *analytics.GetBatchFlagEvaluationsCountRequest) (*analytics.GetBatchFlagEvaluationsCountResponse, error) {
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

	const stepMinutes = 15 // hardcoded to 15 minutes for now

	r := &BatchFlagEvaluationsCountRequest{
		EnvironmentKey: req.EnvironmentKey,
		NamespaceKey:   req.NamespaceKey,
		FlagKeys:       req.FlagKeys,
		From:           fromTime,
		To:             toTime,
		StepMinutes:    stepMinutes,
		Limit:          int(req.Limit),
	}

	flagData, err := s.client.GetBatchFlagEvaluationsCount(ctx, r)
	if err != nil {
		return nil, err
	}

	// Convert to proto response format
	protoFlagData := make(map[string]*analytics.FlagEvaluationData)

	for flagKey, data := range flagData {
		protoFlagData[flagKey] = &analytics.FlagEvaluationData{
			Timestamps: data.Timestamps,
			Values:     data.Values,
		}
	}

	return &analytics.GetBatchFlagEvaluationsCountResponse{
		FlagEvaluations: protoFlagData,
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
