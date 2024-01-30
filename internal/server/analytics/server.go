package analytics

import (
	"context"
	"time"

	"go.flipt.io/flipt/rpc/flipt/analytics"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

// Client is a contract that each metrics store needs to conform to for
// getting metrics from the implemented store.
type Client interface {
	GetFlagEvaluationsCount(ctx context.Context, flagKey string, from time.Duration) ([]string, []float32, error)
}

// Server is a grpc server for Flipt metrics.
type Server struct {
	logger *zap.Logger
	client Client
	analytics.UnimplementedAnalyticsServiceServer
}

// New constructs a new server for Flipt metrics.
func New(logger *zap.Logger, client Client) *Server {
	return &Server{
		logger: logger,
		client: client,
	}
}

func (s *Server) RegisterGRPC(server *grpc.Server) {
	analytics.RegisterAnalyticsServiceServer(server, s)
}
