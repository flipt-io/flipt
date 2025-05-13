package client

import (
	"bytes"
	"context"
	"time"

	"crypto/sha1" //nolint:gosec

	"go.flipt.io/flipt/errors"
	"go.flipt.io/flipt/internal/server/evaluation"
	"go.flipt.io/flipt/internal/server/metrics"
	rpcevaluation "go.flipt.io/flipt/rpc/v2/evaluation"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

var _ rpcevaluation.ClientEvaluationServiceServer = (*Server)(nil)

type Server struct {
	logger *zap.Logger
	envs   evaluation.EnvironmentStore

	rpcevaluation.UnimplementedClientEvaluationServiceServer
}

func NewServer(logger *zap.Logger, envs evaluation.EnvironmentStore) *Server {
	return &Server{logger: logger, envs: envs}
}

// RegisterGRPC registers the *Server onto the provided grpc Server.
func (s *Server) RegisterGRPC(server *grpc.Server) {
	rpcevaluation.RegisterClientEvaluationServiceServer(server, s)
}

func (s *Server) EvaluationSnapshotNamespace(ctx context.Context, r *rpcevaluation.EvaluationNamespaceSnapshotRequest) (*rpcevaluation.EvaluationNamespaceSnapshot, error) {
	start := time.Now()

	env, err := s.envs.Get(ctx, r.EnvironmentKey)
	if err != nil {
		// try to get the environment from the context
		// this is for backwards compatibility with v1
		env = s.envs.GetFromContext(ctx)
	}

	var (
		environmentKey = env.Key()
		namespaceKey   = r.Key

		environmentAttr = metrics.AttributeEnvironment.String(environmentKey)
		namespaceAttr   = metrics.AttributeNamespace.String(namespaceKey)

		attrSet = attribute.NewSet(environmentAttr, namespaceAttr)
	)

	defer func() {
		metrics.EvaluationsSnapshotLatency.Record(ctx, float64(time.Since(start).Milliseconds()), metric.WithAttributeSet(attrSet))
	}()

	metrics.EvaluationsSnapshotRequestsTotal.Add(ctx, 1, metric.WithAttributeSet(attrSet))

	snap, err := env.EvaluationNamespaceSnapshot(ctx, namespaceKey)
	if err != nil {
		metrics.EvaluationsSnapshotErrorsTotal.Add(ctx, 1, metric.WithAttributeSet(attrSet))
		return nil, err
	}

	md, ok := metadata.FromIncomingContext(ctx)
	if ok && snap.Digest != "" {
		etag := snap.Digest
		// set etag header in the response
		_ = grpc.SetHeader(ctx, metadata.Pairs("x-etag", etag))
		// get If-None-Match header from request
		if vals := md.Get("GrpcGateway-If-None-Match"); len(vals) > 0 && etag == vals[0] {
			return &rpcevaluation.EvaluationNamespaceSnapshot{}, errors.ErrNotModifiedf("namespace %q", namespaceKey)
		}
	}

	return &rpcevaluation.EvaluationNamespaceSnapshot{
		Digest:    snap.Digest,
		Namespace: snap.Namespace,
		Flags:     snap.Flags,
	}, nil
}

func (s *Server) EvaluationSnapshotNamespaceStream(req *rpcevaluation.EvaluationNamespaceSnapshotStreamRequest, stream rpcevaluation.ClientEvaluationService_EvaluationSnapshotNamespaceStreamServer) error {
	setupStart := time.Now()

	var (
		ctx = stream.Context()
		env = s.envs.GetFromContext(ctx)
		//nolint:gosec // this is a hash for a stream
		hash = sha1.New()
		// lastDigest is the digest of the last snapshot we sent
		// this includes all namespaces
		lastDigest []byte

		environmentKey = env.Key()
		namespaceKey   = req.Key

		environmentAttr = metrics.AttributeEnvironment.String(environmentKey)
		namespaceAttr   = metrics.AttributeNamespace.String(namespaceKey)

		attrSet = attribute.NewSet(environmentAttr, namespaceAttr)
	)

	metrics.EvaluationsStreamRequestsTotal.Add(ctx, 1, metric.WithAttributeSet(attrSet))
	metrics.EvaluationsStreamsInProgress.Add(ctx, 1, metric.WithAttributeSet(attrSet))

	// start subscription with a channel with a buffer of one
	// to allow the subscription to preload the last observed snapshot
	ch := make(chan *rpcevaluation.EvaluationNamespaceSnapshot, 1)
	closer, err := env.EvaluationNamespaceSnapshotSubscribe(ctx, namespaceKey, ch, func() {
		metrics.EvaluationsStreamsInProgress.Add(ctx, -1, metric.WithAttributeSet(attrSet))
	})

	if err != nil {
		metrics.EvaluationsStreamErrorsTotal.Add(ctx, 1, metric.WithAttributeSet(attrSet))
		metrics.EvaluationsStreamLatency.Record(ctx, float64(time.Since(setupStart).Milliseconds()), metric.WithAttributeSet(attrSet))
		s.logger.Error("error subscribing to environment evaluation namespace snapshot", zap.Error(err), zap.String("namespace", namespaceKey), zap.String("environment", environmentKey))
		return err
	}

	metrics.EvaluationsStreamLatency.Record(ctx, float64(time.Since(setupStart).Milliseconds()), metric.WithAttributeSet(attrSet))

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
					metrics.EvaluationsStreamErrorsTotal.Add(ctx, 1, metric.WithAttributeSet(attrSet))
					s.logger.Error("error sending evaluation namespace snapshot", zap.Error(err), zap.String("namespace", namespaceKey), zap.String("environment", environmentKey))
					return err
				}
				metrics.EvaluationsStreamMessagesTotal.Add(ctx, 1, metric.WithAttributeSet(attrSet))
				lastDigest = digest
			}

			hash.Reset()
		}
	}
}

func (s *Server) SkipsAuthorization(ctx context.Context) bool {
	return true
}
