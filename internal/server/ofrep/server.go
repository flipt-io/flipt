package ofrep

import (
	"context"

	"go.flipt.io/flipt/internal/storage"
	"go.flipt.io/flipt/rpc/flipt"
	rpcevaluation "go.flipt.io/flipt/rpc/flipt/evaluation"
	"go.uber.org/zap"

	"go.flipt.io/flipt/rpc/flipt/ofrep"
	"google.golang.org/grpc"
)

// EvaluationBridgeInput is the input for the bridge between OFREP specficiation to Flipt internals.
type EvaluationBridgeInput struct {
	FlagKey      string
	NamespaceKey string
	EntityId     string
	Context      map[string]string
}

// EvaluationBridgeOutput is the input for the bridge between Flipt internals and OFREP specficiation.
type EvaluationBridgeOutput struct {
	FlagKey  string
	Reason   rpcevaluation.EvaluationReason
	Variant  string
	Value    any
	Metadata map[string]any
}

// Bridge is the interface between the OFREP specification to Flipt internals.
type Bridge interface {
	// OFREPFlagEvaluation evaluates a single flag.
	OFREPFlagEvaluation(ctx context.Context, input EvaluationBridgeInput) (EvaluationBridgeOutput, error)
}

type Storer interface {
	ListFlags(ctx context.Context, req *storage.ListRequest[storage.NamespaceRequest]) (storage.ResultSet[*flipt.Flag], error)
}

// Server servers the methods used by the OpenFeature Remote Evaluation Protocol.
// It will be used only with gRPC Gateway as there's no specification for gRPC itself.
type Server struct {
	logger *zap.Logger
	bridge Bridge
	store  Storer
	ofrep.UnimplementedOFREPServiceServer
}

// New constructs a new Server.
func New(logger *zap.Logger, bridge Bridge, store Storer) *Server {
	return &Server{
		logger: logger,
		bridge: bridge,
		store:  store,
	}
}

// RegisterGRPC registers the EvaluateServer onto the provided gRPC Server.
func (s *Server) RegisterGRPC(server *grpc.Server) {
	ofrep.RegisterOFREPServiceServer(server, s)
}

func (s *Server) SkipsAuthorization(ctx context.Context) bool {
	return true
}
