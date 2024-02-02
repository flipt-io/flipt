package evaluation

import (
	"context"

	"go.flipt.io/flipt/internal/storage"
	"go.flipt.io/flipt/rpc/flipt"
	"go.flipt.io/flipt/rpc/flipt/evaluation"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

// Storer is the minimal abstraction for interacting with the storage layer for evaluation.
type Storer interface {
	GetFlag(ctx context.Context, flag storage.ResourceRequest) (*flipt.Flag, error)
	GetEvaluationRules(ctx context.Context, flag storage.ResourceRequest) ([]*storage.EvaluationRule, error)
	GetEvaluationDistributions(ctx context.Context, ruleID storage.IDRequest) ([]*storage.EvaluationDistribution, error)
	GetEvaluationRollouts(ctx context.Context, flag storage.ResourceRequest) ([]*storage.EvaluationRollout, error)
}

type AnalyticsStoreMutator interface {
	IncrementBooleanFlagEvaluation(ctx context.Context, namespaceKey string, resp *evaluation.BooleanEvaluationResponse) error
	IncrementVariantFlagEvaluation(ctx context.Context, namspaceKey string, resp *evaluation.VariantEvaluationResponse) error
}

// Option defines a function acting as a functional option to modify server properties.
type Option func(*Server)

// Server serves the Flipt evaluate v2 gRPC Server.
type Server struct {
	logger         *zap.Logger
	store          Storer
	evaluator      *Evaluator
	analyticsStore AnalyticsStoreMutator
	evaluation.UnimplementedEvaluationServiceServer
}

// New is constructs a new Server.
func New(logger *zap.Logger, store Storer, opts ...Option) *Server {
	s := &Server{
		logger:    logger,
		store:     store,
		evaluator: NewEvaluator(logger, store),
	}

	for _, opt := range opts {
		opt(s)
	}

	return s
}

// WithAnalyticsStoreMutator configures a analytics store for the Server for
// storing analytics into a pre-configured store.
func WithAnalyticsStoreMutator(as AnalyticsStoreMutator) Option {
	return func(s *Server) {
		s.analyticsStore = as
	}
}

// RegisterGRPC registers the EvaluateServer onto the provided gRPC Server.
func (s *Server) RegisterGRPC(server *grpc.Server) {
	evaluation.RegisterEvaluationServiceServer(server, s)
}

func (s *Server) AllowsNamespaceScopedAuthentication(ctx context.Context) bool {
	return true
}
