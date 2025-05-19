package evaluation

import (
	"context"
	"io"

	storagefs "go.flipt.io/flipt/internal/storage/fs"
	rpcevaluation "go.flipt.io/flipt/rpc/v2/evaluation"
)

var NoopPublisher = &noopPublisher{}

// noopPublisher is a publisher that does nothing.
// It is used in branched environments where evaluation is not enabled.
type noopPublisher struct {
}

func (p *noopPublisher) Publish(snap *storagefs.Snapshot) error {
	return nil
}

func (p *noopPublisher) Subscribe(ctx context.Context, ns string, ch chan<- *rpcevaluation.EvaluationNamespaceSnapshot) (io.Closer, error) {
	return nil, nil
}
