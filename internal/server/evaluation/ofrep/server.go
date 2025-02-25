package ofrep

import (
	"context"

	"go.flipt.io/flipt/internal/server/environments"
	"go.uber.org/zap"

	"go.flipt.io/flipt/rpc/flipt/ofrep"
	"google.golang.org/grpc"
)

// Bridge is the interface between the OFREP specification to Flipt internals.
type Bridge interface {
	// OFREPFlagEvaluation evaluates a single flag.
	OFREPFlagEvaluation(ctx context.Context, input *ofrep.EvaluateFlagRequest) (*ofrep.EvaluationResponse, error)
	// OFREPFlagEvaluationBulk evaluates a list of flags.
	OFREPFlagEvaluationBulk(ctx context.Context, input *ofrep.EvaluateBulkRequest) (*ofrep.BulkEvaluationResponse, error)
}

// Server servers the methods used by the OpenFeature Remote Evaluation Protocol.
// It will be used only with gRPC Gateway as there's no specification for gRPC itself.
type Server struct {
	logger *zap.Logger
	bridge Bridge
	envs   *environments.EnvironmentStore
	ofrep.UnimplementedOFREPServiceServer
}

// New constructs a new Server.
func New(logger *zap.Logger, bridge Bridge, envs *environments.EnvironmentStore) *Server {
	return &Server{
		logger: logger,
		bridge: bridge,
		envs:   envs,
	}
}

// RegisterGRPC registers the EvaluateServer onto the provided gRPC Server.
func (s *Server) RegisterGRPC(server *grpc.Server) {
	ofrep.RegisterOFREPServiceServer(server, s)
}

func (s *Server) SkipsAuthorization(ctx context.Context) bool {
	return true
}
