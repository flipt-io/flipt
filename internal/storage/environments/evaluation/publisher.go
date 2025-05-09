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
	rpcevaluation "go.flipt.io/flipt/rpc/v2/evaluation"
	"go.uber.org/zap"
)

type subscription struct {
	mu     sync.Mutex
	finish func()
	ch     chan<- *rpcevaluation.EvaluationNamespaceSnapshot
	id     string
}

func (s *subscription) send(ctx context.Context, snap *rpcevaluation.EvaluationNamespaceSnapshot) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.ch == nil {
		return errors.New("subscription has been closed")
	}

	if snap == nil {
		return nil
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

// publishOptions provides configuration options for the publish operation
type publishOptions struct {
	// timeout is the maximum time to wait for a subscriber to accept a message
	// Default is 5 seconds if not specified
	timeout time.Duration
}

type SnapshotPublisher struct {
	logger *zap.Logger

	mu   sync.Mutex
	last *rpcevaluation.EvaluationSnapshot
	subs map[string][]*subscription

	options publishOptions
}

type OptionsFunc func(options *publishOptions)

func NewSnapshotPublisher(logger *zap.Logger, opts ...OptionsFunc) *SnapshotPublisher {
	options := publishOptions{
		timeout: 5 * time.Second,
	}

	for _, opt := range opts {
		opt(&options)
	}

	return &SnapshotPublisher{
		logger:  logger,
		subs:    make(map[string][]*subscription),
		options: options,
	}
}

func WithTimeout(timeout time.Duration) OptionsFunc {
	return func(options *publishOptions) {
		options.timeout = timeout
	}
}

func (p *SnapshotPublisher) Publish(ctx context.Context, snap *storagefs.Snapshot) error {
	last, err := snap.EvaluationSnapshot(ctx)
	if err != nil {
		return err
	}

	p.mu.Lock()
	p.last = last
	// TODO: this is a shallow copy, unsure if we should deep copy
	subscriptions := maps.Clone(p.subs)
	timeout := p.options.timeout
	p.mu.Unlock()

	var (
		wg          sync.WaitGroup
		publishErrs []error
		errMu       sync.Mutex
	)

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

				// Create a timeout context for just this subscriber
				subCtx, cancel := context.WithTimeout(ctx, timeout)
				defer cancel()

				p.logger.Debug("sending update",
					zap.String("subscription", sub.id),
					zap.String("namespace", ns))

				if err := sub.send(subCtx, last.Namespaces[ns]); err != nil {
					p.logger.Error("error sending update",
						zap.String("subscription", sub.id),
						zap.String("namespace", ns),
						zap.Error(err))

					// Record the error
					errMu.Lock()
					publishErrs = append(publishErrs, err)
					errMu.Unlock()
				}
			}()
		}
	}

	// Wait for all goroutines to complete
	wg.Wait()

	// Check for errors
	if len(publishErrs) > 0 {
		p.logger.Warn("some subscriptions failed to receive updates",
			zap.Int("error_count", len(publishErrs)))

		// Just return the first error
		return publishErrs[0]
	}

	return nil
}

func (p *SnapshotPublisher) Subscribe(ctx context.Context, ns string, ch chan<- *rpcevaluation.EvaluationNamespaceSnapshot) (io.Closer, error) {
	id := uuid.New().String()
	sub := &subscription{id: id, finish: func() {
		p.mu.Lock()
		defer p.mu.Unlock()

		delete(p.subs, ns)

		p.logger.Debug("Subscription canceled", zap.String("subscription", id))
	}, ch: ch}

	p.mu.Lock()
	// send initial snapshot if one has already been observed
	if p.last != nil {
		if err := sub.send(ctx, p.last.Namespaces[ns]); err != nil {
			p.logger.Debug("error sending to subscriber", zap.Error(err))
		}
	}
	p.subs[ns] = append(p.subs[ns], sub)
	p.mu.Unlock()

	p.logger.Debug("Subscription created", zap.String("subscription", id))
	return sub, nil
}
