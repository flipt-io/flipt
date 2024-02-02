package analytics

import (
	"context"
	"fmt"

	"go.flipt.io/flipt/rpc/flipt/analytics"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

// Client is a contract that each analytics store needs to conform to for
// getting analytics from the implemented store.
type Client interface {
	GetFlagEvaluationsCount(ctx context.Context, req *analytics.GetFlagEvaluationsCountRequest) ([]string, []float32, error)
	fmt.Stringer
}

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
