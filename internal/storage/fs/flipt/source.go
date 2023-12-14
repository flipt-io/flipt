package flipt

import (
	"context"
	"errors"
	"io"
	"sync"

	storagefs "go.flipt.io/flipt/internal/storage/fs"
	"go.flipt.io/flipt/rpc/flipt/evaluation"
	"go.uber.org/zap"
)

// Source is an implementation fs.SnapshotSource backed by OCI repositories
// It fetches instances of OCI manifests and uses them to build snapshots from their contents
type Source struct {
	logger *zap.Logger

	cli evaluation.DataService_EvaluationSnapshotStreamClient

	mu      sync.Mutex
	current *storagefs.StoreSnapshot
}

// NewSource constructs and configures a Source.
// The source uses the connection and credential details provided to build
// *storagefs.StoreSnapshot implementations around a target git repository.
func NewSource(logger *zap.Logger, ctx context.Context, client evaluation.DataServiceClient) (_ *Source, err error) {
	src := &Source{
		logger: logger,
	}

	cli, err := client.EvaluationSnapshotStream(ctx, &evaluation.EvaluationSnapshotStreamRequest{})
	if err != nil {
		return nil, err
	}

	src.cli = cli

	return src, nil
}

func (s *Source) String() string {
	return "flipt"
}

// Get builds a single instance of an *storagefs.StoreSnapshot
func (s *Source) Get(context.Context) (*storagefs.StoreSnapshot, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	return s.current, nil
}

// Subscribe feeds implementations of *storagefs.StoreSnapshot onto the provided channel.
// It should block until the provided context is cancelled (it will be called in a goroutine).
// It should close the provided channel before it returns.
func (s *Source) Subscribe(ctx context.Context, ch chan<- *storagefs.StoreSnapshot) {
	defer close(ch)

	for {
		if err := s.cli.Context().Err(); err != nil {
			// context is closed so client has stopped
			return
		}

		snap, err := s.cli.Recv()
		if err != nil {
			if !errors.Is(err, io.EOF) {
				s.logger.Error("receiving snapshot", zap.Error(err))
				continue
			}

			return
		}

		current, err := storagefs.SnapshotFromEvaluationSnapshot(snap)
		if err != nil {
			s.logger.Error("adapting snapshot", zap.Error(err))
			continue
		}

		select {
		case <-ctx.Done():
			return
		case ch <- current:
		}

		s.mu.Lock()
		s.current = current
		s.mu.Unlock()
	}
}
