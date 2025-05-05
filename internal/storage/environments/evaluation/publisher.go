package evaluation

import (
	"context"
	"errors"
	"io"
	"maps"
	"sync"
	"time"

	"github.com/google/uuid"
	storagefs "go.flipt.io/flipt/internal/storage/fs"
	"go.flipt.io/flipt/rpc/flipt/evaluation"
	"go.uber.org/zap"
)

type subscription struct {
	mu     sync.Mutex
	finish func()
	ch     chan<- *evaluation.EvaluationNamespaceSnapshot
	id     string
}

func (s *subscription) send(ctx context.Context, snap *evaluation.EvaluationNamespaceSnapshot) error {
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

type SnapshotPublisher struct {
	logger *zap.Logger

	mu sync.Mutex
	// last is the last snapshot published which contains all the namespace data
	last *evaluation.EvaluationSnapshot
	// subs is a map of namespace to subscriptions
	subs map[string][]*subscription
}

func NewSnapshotPublisher(logger *zap.Logger) *SnapshotPublisher {
	return &SnapshotPublisher{logger: logger}
}

func (p *SnapshotPublisher) Publish(ctx context.Context, snap *storagefs.Snapshot) error {
	last, err := snap.EvaluationSnapshot(ctx)
	if err != nil {
		return err
	}

	p.mu.Lock()
	p.last = last
	// TODO: not sure if this shallow copy is enough
	subscriptions := maps.Clone(p.subs)
	p.mu.Unlock()

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	var wg sync.WaitGroup

	for ns, subs := range subscriptions {
		ns := ns
		subs := subs
		for _, sub := range subs {
			sub := sub
			if sub == nil {
				continue
			}
			wg.Add(1)
			go func() {
				defer wg.Done()
				p.logger.Debug("sending update", zap.String("subscription", sub.id), zap.String("namespace", ns))

				if err := sub.send(ctx, last.Namespaces[ns]); err != nil {
					p.logger.Error("error sending update", zap.String("subscription", sub.id), zap.String("namespace", ns), zap.Error(err))
				}
			}()
		}
	}

	wg.Wait()

	return nil
}

func (p *SnapshotPublisher) Subscribe(ctx context.Context, ns string, ch chan<- *evaluation.EvaluationNamespaceSnapshot) (io.Closer, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	id := uuid.New().String()
	sub := &subscription{id: id, finish: func() {
		p.mu.Lock()
		defer p.mu.Unlock()

		delete(p.subs, id)
		p.logger.Debug("Subscription canceled", zap.String("subscription", id))
	}, ch: ch}

	p.mu.Lock()
	p.subs[ns] = append(p.subs[ns], sub)
	p.mu.Unlock()

	return sub, nil
}
