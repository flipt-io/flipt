// Flipt Commercial Open Source Feature
// This file contains functionality that is licensed under the Flipt Fair Core License (FCL).
// You may NOT use, modify, or distribute this file or its contents without a valid paid license.
// For details: https://github.com/flipt-io/flipt/blob/v2/LICENSE

package webhook

import (
	"context"
	"errors"
	"net/http"
	"slices"
	"sync"
	"time"

	errs "go.flipt.io/flipt/errors"
	"go.flipt.io/flipt/internal/config"
	"go.uber.org/zap"
)

// Repository is the part of a storage git repository the receiver drives.
// It is satisfied by *storagegit.Repository.
type Repository interface {
	// Fetch fetches every tracked branch and notifies the environments
	// subscribed to the repository.
	Fetch(ctx context.Context, heads ...string) error
	// Tracks reports whether a fetch would update the given branch.
	Tracks(branch string) bool
}

// Target is an environment that accepts webhooks.
type Target struct {
	SCM        config.SCMType
	Secret     []byte
	Repository Repository
}

// Receiver serves incoming SCM webhooks. An authenticated push to a branch
// tracked by the target environment's repository triggers a fetch of that
// repository.
//
// Fetches are collapsed per repository: at most one runs at a time, and any
// webhooks accepted while it runs are served by a single follow-up fetch.
type Receiver struct {
	ctx     context.Context
	logger  *zap.Logger
	targets map[string]Target
	syncers map[Repository]*syncer
	metrics receiverMetrics
	wg      sync.WaitGroup
}

// NewReceiver returns a Receiver for the given targets, keyed by environment
// name. Fetches run under ctx, so cancelling it aborts in-flight fetches.
func NewReceiver(ctx context.Context, logger *zap.Logger, targets map[string]Target) *Receiver {
	r := &Receiver{
		ctx:     ctx,
		logger:  logger.With(zap.String("component", "webhook")),
		targets: targets,
		syncers: map[Repository]*syncer{},
		metrics: newReceiverMetrics(),
	}

	for _, t := range targets {
		if _, ok := r.syncers[t.Repository]; !ok {
			r.syncers[t.Repository] = &syncer{receiver: r, repo: t.Repository}
		}
	}

	return r
}

// Environments returns the names of the environments accepting webhooks.
func (r *Receiver) Environments() []string {
	names := make([]string, 0, len(r.targets))
	for name := range r.targets {
		names = append(names, name)
	}

	slices.Sort(names)

	return names
}

// ServeWebhook handles a webhook request addressed to the named environment.
//
// It responds 404 for an environment without a webhook configured, 401 when
// the request fails authentication, 202 when a fetch was scheduled and 200
// for authentic events that need no fetch. Only the 202 path fetches.
func (r *Receiver) ServeWebhook(w http.ResponseWriter, req *http.Request, environment string) {
	var (
		ctx      = req.Context()
		received = time.Now()
	)

	target, ok := r.targets[environment]
	if !ok {
		r.metrics.recordRequest(ctx, unknownLabel, unknownLabel, resultNotFound)
		r.logger.Info("webhook rejected",
			zap.String("environment", environment),
			zap.String("reason", "no webhook configured for environment"))
		http.Error(w, "not found", http.StatusNotFound)
		return
	}

	var (
		scm    = string(target.SCM)
		logger = r.logger.With(zap.String("environment", environment), zap.String("scm", scm))
	)

	event, err := Handle(target.SCM, target.Secret, req)
	if err != nil {
		switch {
		case errs.AsMatch[errs.ErrUnauthenticated](err):
			r.metrics.recordRequest(ctx, environment, scm, resultUnauthorized)
			logger.Info("webhook rejected", zap.String("reason", "authentication failed"), zap.Error(err))
			http.Error(w, "unauthorized", http.StatusUnauthorized)
		case errors.Is(err, ErrBodyTooLarge):
			logger.Info("webhook rejected", zap.String("reason", "body too large"), zap.Error(err))
			http.Error(w, "request body too large", http.StatusRequestEntityTooLarge)
		default:
			logger.Info("webhook rejected", zap.String("reason", "invalid request"), zap.Error(err))
			http.Error(w, "bad request", http.StatusBadRequest)
		}
		return
	}

	logger = logger.With(zap.String("event", event.Name))

	var tracked []string
	if event.Kind == KindPush {
		for _, branch := range event.Branches {
			if target.Repository.Tracks(branch) {
				tracked = append(tracked, branch)
			}
		}
	}

	if len(tracked) == 0 {
		r.metrics.recordRequest(ctx, environment, scm, resultIgnored)
		logger.Info("webhook ignored",
			zap.String("kind", string(event.Kind)),
			zap.Strings("branches", event.Branches))
		w.WriteHeader(http.StatusOK)
		return
	}

	r.metrics.recordRequest(ctx, environment, scm, resultAccepted)
	logger.Info("webhook accepted; scheduling fetch", zap.Strings("branches", tracked))

	r.syncers[target.Repository].trigger(pendingSync{
		environment: environment,
		scm:         scm,
		received:    received,
	})

	w.WriteHeader(http.StatusAccepted)
}

// Wait blocks until no webhook-triggered fetch is running or pending.
func (r *Receiver) Wait() {
	r.wg.Wait()
}

type pendingSync struct {
	environment string
	scm         string
	received    time.Time
}

// syncer collapses fetch triggers for a single repository.
type syncer struct {
	receiver *Receiver
	repo     Repository

	mu      sync.Mutex
	running bool
	pending []pendingSync
}

func (s *syncer) trigger(p pendingSync) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.pending = append(s.pending, p)
	if s.running {
		return
	}

	s.running = true
	s.receiver.wg.Go(s.run)
}

// run fetches until no triggers are pending. Each fetch serves every trigger
// accepted before it started.
func (s *syncer) run() {
	r := s.receiver

	for {
		s.mu.Lock()
		batch := s.pending
		s.pending = nil
		if len(batch) == 0 {
			s.running = false
			s.mu.Unlock()
			return
		}
		s.mu.Unlock()

		environments := make([]string, 0, len(batch))
		for _, p := range batch {
			if !slices.Contains(environments, p.environment) {
				environments = append(environments, p.environment)
			}
		}

		logger := r.logger.With(zap.Strings("environments", environments), zap.Int("webhooks", len(batch)))
		logger.Info("webhook-triggered fetch started")

		start := time.Now()
		err := s.repo.Fetch(r.ctx)
		completed := time.Now()

		if err != nil {
			logger.Error("webhook-triggered fetch failed", zap.Error(err))

			// count one failure per environment and SCM served by the fetch
			type key struct{ environment, scm string }
			seen := map[key]struct{}{}
			for _, p := range batch {
				k := key{p.environment, p.scm}
				if _, ok := seen[k]; ok {
					continue
				}
				seen[k] = struct{}{}
				r.metrics.recordFetchError(r.ctx, p.environment, p.scm)
			}
			continue
		}

		logger.Info("webhook-triggered fetch completed", zap.Duration("duration", completed.Sub(start)))

		for _, p := range batch {
			r.metrics.recordSync(r.ctx, p.environment, p.scm, completed.Sub(p.received))
		}
	}
}
