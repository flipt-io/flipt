package client

import (
	"context"

	"go.flipt.io/flipt/errors"
	"go.flipt.io/flipt/internal/server/environments"
	"go.flipt.io/flipt/rpc/v2/evaluation/client"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

var _ client.EvaluationServiceServer = (*Server)(nil)

type Server struct {
	logger *zap.Logger
	envs   *environments.EnvironmentStore

	client.UnimplementedEvaluationServiceServer
}

func NewServer(logger *zap.Logger, envs *environments.EnvironmentStore) *Server {
	return &Server{logger: logger, envs: envs}
}

// RegisterGRPC registers the *Server onto the provided grpc Server.
func (s *Server) RegisterGRPC(server *grpc.Server) {
	client.RegisterEvaluationServiceServer(server, s)
}

func (s *Server) EvaluationSnapshotNamespace(ctx context.Context, r *client.EvaluationNamespaceSnapshotRequest) (*client.EvaluationNamespaceSnapshot, error) {
	env := s.envs.GetFromContext(ctx)

	snap, err := env.EvaluationNamespaceSnapshot(ctx, r.Key)
	if err != nil {
		return nil, err
	}

	md, ok := metadata.FromIncomingContext(ctx)
	if ok && snap.Digest != "" {
		etag := snap.Digest
		// set etag header in the response
		_ = grpc.SetHeader(ctx, metadata.Pairs("x-etag", etag))
		// get If-None-Match header from request
		if vals := md.Get("GrpcGateway-If-None-Match"); len(vals) > 0 && etag == vals[0] {
			return &client.EvaluationNamespaceSnapshot{}, errors.ErrNotModifiedf("namespace %q", r.Key)
		}
	}

	return &client.EvaluationNamespaceSnapshot{
		Digest:    snap.Digest,
		Namespace: snap.Namespace,
		Flags:     snap.Flags,
	}, nil
}

func (s *Server) EvaluationSnapshotNamespaceStream(req *client.EvaluationNamespaceSnapshotStreamRequest, stream client.EvaluationService_EvaluationSnapshotNamespaceStreamServer) error {
	return status.Errorf(codes.Unimplemented, "method EvaluationSnapshotNamespaceStream not implemented")
}

func (s *Server) SkipsAuthorization(ctx context.Context) bool {
	return true
}
