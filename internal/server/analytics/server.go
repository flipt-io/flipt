package analytics

import (
	"context"

	"go.flipt.io/flipt/rpc/v2/analytics"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

// Server is a grpc server for Flipt analytics.
type Server struct {
	logger *zap.Logger
	client Client
	analytics.UnimplementedAnalyticsServiceServer
}

// New constructs a new server for Flipt analytics.
func New(logger *zap.Logger, client Client) *Server {
	return &Server{
		logger: logger,
		client: client,
	}
}

func (s *Server) RegisterGRPC(server *grpc.Server) {
	analytics.RegisterAnalyticsServiceServer(server, s)
}

func (s *Server) SkipsAuthorization(ctx context.Context) bool {
	return true
}
