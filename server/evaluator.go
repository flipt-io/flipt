package server

import (
	"context"
	"time"

	"github.com/gofrs/uuid"
	flipt "github.com/markphelps/flipt/rpc"
)

// Evaluate evaluates a request for a given flag and entity
func (s *Server) Evaluate(ctx context.Context, req *flipt.EvaluationRequest) (*flipt.EvaluationResponse, error) {
	startTime := time.Now()

	// set request ID if not present
	if req.RequestId == "" {
		req.RequestId = uuid.Must(uuid.NewV4()).String()
	}

	resp, err := s.Evaluator.Evaluate(ctx, req)
	if err != nil {
		return nil, err
	}

	if resp != nil {
		resp.RequestDurationMillis = float64(time.Since(startTime)) / float64(time.Millisecond)
	}

	return resp, nil
}
