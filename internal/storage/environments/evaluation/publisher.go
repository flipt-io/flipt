package evaluation

import (
	"context"
	"errors"
	"io"
	"maps"
	"slices"
	"sync"
	"time"

	"github.com/google/uuid"
	"go.flipt.io/flipt/internal/storage"
	"go.flipt.io/flipt/rpc/flipt/evaluation"
	"go.uber.org/zap"
)

type SnapshotPublisher struct {
	logger *zap.Logger
	ref    string

	mu   sync.Mutex
	last *evaluation.EvaluationSnapshot
	subs map[string]*subscription
}

func NewPublisher(logger *zap.Logger, ref string) *SnapshotPublisher {
	return &SnapshotPublisher{
		logger: logger,
		ref:    ref,
		subs:   map[string]*subscription{},
	}
}

func (p *SnapshotPublisher) BuildAndPublish(ctx context.Context, store storage.ReadOnlyStore) (*evaluation.EvaluationSnapshot, error) {
	p.logger.Debug("BuildAndPublish", zap.String("reference", p.ref))

	snap, err := createStoreSnapshot(ctx, store, p.ref)
	if err != nil {
		return nil, err
	}

	p.mu.Lock()
	p.last = snap
	subs := slices.Collect(maps.Values(p.subs))
	p.mu.Unlock()

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	var wg sync.WaitGroup
	for _, sub := range subs {
		sub := sub

		wg.Add(1)
		go func() {
			defer wg.Done()

			p.logger.Debug("sending update")

			if err := sub.send(ctx, snap); err != nil {
				p.logger.Debug("error sending to subscriber", zap.Error(err))
			}
		}()
	}

	wg.Wait()

	return snap, nil
}

func (p *SnapshotPublisher) Subscribe(ctx context.Context, ch chan<- *evaluation.EvaluationSnapshot) (io.Closer, error) {
	p.logger.Debug("Subscribe", zap.String("reference", p.ref))

	id := uuid.New().String()
	sub := &subscription{finish: func() {
		p.mu.Lock()
		defer p.mu.Unlock()

		delete(p.subs, id)
		p.logger.Debug("Subscription canceled", zap.String("reference", p.ref))
	}, ch: ch}

	p.mu.Lock()
	// send initial snapshot if one has already been observed
	if p.last != nil {
		if err := sub.send(ctx, p.last); err != nil {
			p.logger.Debug("error sending to subscriber", zap.Error(err))
		}
	}

	p.subs[id] = sub
	p.mu.Unlock()

	return sub, nil
}

type subscription struct {
	mu     sync.Mutex
	finish func()
	ch     chan<- *evaluation.EvaluationSnapshot
}

func (s *subscription) send(ctx context.Context, snap *evaluation.EvaluationSnapshot) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.ch == nil {
		return errors.New("subscription has been closed")
	}

	select {
	case s.ch <- snap:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (s *subscription) Close() error {
	// remove the subscription from the publishable list
	s.finish()

	// close and remove the channel from the subscription
	s.mu.Lock()
	defer s.mu.Unlock()

	close(s.ch)
	s.ch = nil

	return nil
}
