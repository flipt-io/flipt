package data

import (
	"context"

	"go.flipt.io/flipt/errors"
	"go.flipt.io/flipt/internal/server/environments"
	"go.flipt.io/flipt/rpc/flipt/evaluation"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

type Server struct {
	logger *zap.Logger
	envs   *environments.EnvironmentStore

	evaluation.UnimplementedDataServiceServer
}

func New(logger *zap.Logger, envs *environments.EnvironmentStore) *Server {
	return &Server{
		logger: logger,
		envs:   envs,
	}
}

// RegisterGRPC registers the *Server onto the provided grpc Server.
func (srv *Server) RegisterGRPC(server *grpc.Server) {
	evaluation.RegisterDataServiceServer(server, srv)
}

func (srv *Server) AllowsNamespaceScopedAuthentication(ctx context.Context) bool {
	return true
}

func (s *Server) EvaluationSnapshotNamespace(ctx context.Context, r *evaluation.EvaluationNamespaceSnapshotRequest) (*evaluation.EvaluationNamespaceSnapshot, error) {
	// TODO(georgemac): support overriding via configuration and or metadata header
	environment := "default"

	env, err := s.envs.Get(ctx, environment)
	if err != nil {
		return nil, err
	}

	snap, err := env.EvaluationNamespaceSnapshot(ctx, r.Key)
	if err != nil {
		return nil, err
	}

	md, ok := metadata.FromIncomingContext(ctx)
	if ok && snap.Digest != "" {
		// get If-None-Match header from request
		if vals := md.Get("GrpcGateway-If-None-Match"); len(vals) > 0 && snap.Digest == vals[0] {
			return nil, errors.ErrNotModifiedf("namespace %q", r.Key)
		}
	}

	return snap, nil
}
