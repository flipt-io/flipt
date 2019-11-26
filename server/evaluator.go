package server

import (
	"context"
	"time"

	"github.com/gofrs/uuid"
	"github.com/markphelps/flipt/errors"
	flipt "github.com/markphelps/flipt/rpc"
)

// Evaluate evaluates a request for a given flag and entity
func (s *Server) Evaluate(ctx context.Context, req *flipt.EvaluationRequest) (*flipt.EvaluationResponse, error) {
	if req.FlagKey == "" {
		return nil, errors.EmptyFieldError("flagKey")
	}

	if req.EntityId == "" {
		return nil, errors.EmptyFieldError("entityId")
	}

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
