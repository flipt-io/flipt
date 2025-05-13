package analytics

import (
	"context"
	"fmt"
	"time"

	"go.flipt.io/flipt/rpc/v2/analytics"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

type FlagEvaluationsCountRequest struct {
	EnvironmentKey string
	NamespaceKey   string
	FlagKey        string
	From           time.Time
	To             time.Time
	StepMinutes    int
}

// Client is a contract that each analytics store needs to conform to for
// getting analytics from the implemented store.
type Client interface {
	GetFlagEvaluationsCount(ctx context.Context, req *FlagEvaluationsCountRequest) ([]string, []float32, error)
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

func (s *Server) SkipsAuthorization(ctx context.Context) bool {
	return true
}
