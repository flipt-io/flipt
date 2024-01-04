package fs

import (
	"context"
	"testing"
	"time"

	"go.flipt.io/flipt/internal/containers"
	"go.uber.org/zap"
)

type Poller struct {
	logger *zap.Logger

	interval time.Duration
	update   UpdateFunc

	ctx    context.Context
	cancel func()
	done   chan struct{}
	notify func(modified bool)
}

func WithInterval(interval time.Duration) containers.Option[Poller] {
	return func(p *Poller) {
		p.interval = interval
	}
}

func WithNotify(t *testing.T, n func(modified bool)) containers.Option[Poller] {
	t.Helper()
	return func(p *Poller) {
		p.notify = n
	}
}

type UpdateFunc func(context.Context) (bool, error)

func NewPoller(logger *zap.Logger, ctx context.Context, update UpdateFunc, opts ...containers.Option[Poller]) *Poller {
	p := &Poller{
		logger:   logger,
		done:     make(chan struct{}),
		interval: 30 * time.Second,
		update:   update,
	}

	containers.ApplyAll(p, opts...)

	p.ctx, p.cancel = context.WithCancel(ctx)

	return p
}

func (p *Poller) Close() error {
	p.cancel()

	<-p.done

	return nil
}

// Poll is a utility function for a common polling strategy used by lots of declarative
// store implementations.
func (p *Poller) Poll() {
	defer close(p.done)

	ticker := time.NewTicker(p.interval)
	for {
		select {
		case <-p.ctx.Done():
			return
		case <-ticker.C:
			modified, err := p.update(p.ctx)
			if err != nil {
				p.logger.Error("error getting file system from directory", zap.Error(err))
				continue
			}

			if p.notify != nil {
				p.notify(modified)
			}

			if !modified {
				p.logger.Debug("skipping snapshot update as it has not been modified")
				continue
			}

			p.logger.Debug("snapshot updated")
		}
	}
}
