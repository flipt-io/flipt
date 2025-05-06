package client

import (
	"bytes"
	"context"

	"crypto/sha1" //nolint:gosec

	"go.flipt.io/flipt/errors"
	"go.flipt.io/flipt/internal/server/environments"
	"go.flipt.io/flipt/rpc/v2/evaluation"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

var _ evaluation.ClientEvaluationServiceServer = (*Server)(nil)

type Server struct {
	logger *zap.Logger
	envs   *environments.EnvironmentStore

	evaluation.UnimplementedClientEvaluationServiceServer
}

func NewServer(logger *zap.Logger, envs *environments.EnvironmentStore) *Server {
	return &Server{logger: logger, envs: envs}
}

// RegisterGRPC registers the *Server onto the provided grpc Server.
func (s *Server) RegisterGRPC(server *grpc.Server) {
	evaluation.RegisterClientEvaluationServiceServer(server, s)
}

func (s *Server) EvaluationSnapshotNamespace(ctx context.Context, r *evaluation.EvaluationNamespaceSnapshotRequest) (*evaluation.EvaluationNamespaceSnapshot, error) {
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
			return &evaluation.EvaluationNamespaceSnapshot{}, errors.ErrNotModifiedf("namespace %q", r.Key)
		}
	}

	return &evaluation.EvaluationNamespaceSnapshot{
		Digest:    snap.Digest,
		Namespace: snap.Namespace,
		Flags:     snap.Flags,
	}, nil
}

func (s *Server) EvaluationSnapshotNamespaceStream(req *evaluation.EvaluationNamespaceSnapshotStreamRequest, stream evaluation.ClientEvaluationService_EvaluationSnapshotNamespaceStreamServer) error {
	var (
		ctx = stream.Context()
		env = s.envs.GetFromContext(ctx)
		//nolint:gosec // this is a hash for a stream
		hash = sha1.New()
		// lastDigest is the digest of the last snapshot we sent
		// this includes all namespaces
		lastDigest []byte
	)

	// start subscription with a channel with a buffer of one
	// to allow the subscription to preload the last observed snapshot
	ch := make(chan *evaluation.EvaluationNamespaceSnapshot, 1)
	closer, err := env.EvaluationNamespaceSnapshotSubscribe(ctx, req.Key, ch)
	if err != nil {
		s.logger.Error("error subscribing to environment evaluation namespace snapshot", zap.Error(err), zap.String("namespace", req.Key), zap.String("environment", req.EnvironmentKey))
		return err
	}

	// close removes the channel from the subscribers
	defer closer.Close()

	for {
		select {
		case <-ctx.Done():
			return nil
		case snap, ok := <-ch:
			if !ok {
				return nil
			}

			if snap == nil {
				s.logger.Debug("received nil snapshot, skipping")
				continue
			}

			hash.Write([]byte(snap.Digest))

			// only send the snapshot if we have a new digest
			if digest := hash.Sum(nil); !bytes.Equal(lastDigest, digest) {
				if err := stream.Send(snap); err != nil {
					s.logger.Error("error sending evaluation namespace snapshot", zap.Error(err), zap.String("namespace", req.Key), zap.String("environment", req.EnvironmentKey))
					return err
				}
				lastDigest = digest
			}

			hash.Reset()
		}
	}
}

func (s *Server) SkipsAuthorization(ctx context.Context) bool {
	return true
}
