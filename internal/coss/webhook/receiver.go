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

	"go.uber.org/zap"

	errs "go.flipt.io/flipt/errors"
	"go.flipt.io/flipt/internal/config"
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
	ctx context.Context
	// fetchCtx outlives ctx so fetches already running or accepted when ctx
	// is cancelled complete; Shutdown cancels it when its own ctx expires.
	fetchCtx    context.Context
	cancelFetch context.CancelFunc
	logger      *zap.Logger
	targets     map[string]Target
	syncers     map[Repository]*syncer
	metrics     receiverMetrics
	wg          sync.WaitGroup
}

// NewReceiver returns a Receiver for the given targets, keyed by environment
// name. Cancelling ctx stops new fetches from being scheduled, but fetches
// already running or accepted still complete; call Shutdown to wait for them.
func NewReceiver(ctx context.Context, logger *zap.Logger, targets map[string]Target) *Receiver {
	fetchCtx, cancelFetch := context.WithCancel(context.WithoutCancel(ctx))

	r := &Receiver{
		ctx:         ctx,
		fetchCtx:    fetchCtx,
		cancelFetch: cancelFetch,
		logger:      logger.With(zap.String("component", "webhook")),
		targets:     targets,
		syncers:     map[Repository]*syncer{},
		metrics:     newReceiverMetrics(),
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
// It responds 401 when the request fails authentication, 413 for an
// oversized body, 400 for a malformed or unsupported request, 202 when a
// fetch was scheduled and 200 for authentic events that need no fetch. Only
// the 202 path fetches.
//
// An unknown environment, or one without a webhook configured, gets the same
// 401 response as an authentication failure without its body being read, so
// responses don't reveal which environments accept webhooks. It's still
// counted as not_found in the requests metric.
//
// Once the receiver's context is cancelled (Flipt is shutting down), an
// authentic push that would otherwise fetch gets 503 and isn't counted in
// the requests metric, since no fetch was scheduled; the sender can retry
// against the restarted server.
func (r *Receiver) ServeWebhook(w http.ResponseWriter, req *http.Request, environment string) {
	var (
		ctx      = req.Context()
		received = time.Now()
	)

	target, ok := r.targets[environment]
	if !ok {
		r.metrics.recordRequest(ctx, unknownLabel, unknownLabel, resultNotFound)
		r.logger.Debug("webhook rejected",
			zap.String("environment", environment),
			zap.String("reason", "no webhook configured for environment"))
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	var (
		scm    = string(target.SCM)
		logger = r.logger.With(zap.String("environment", environment), zap.String("scm", scm))
	)

	event, err := handle(target.SCM, target.Secret, req)
	if err != nil {
		switch {
		case errs.AsMatch[errs.ErrUnauthenticated](err):
			r.metrics.recordRequest(ctx, environment, scm, resultUnauthorized)
			logger.Debug("webhook rejected", zap.String("reason", "authentication failed"), zap.Error(err))
			http.Error(w, "unauthorized", http.StatusUnauthorized)
		case errors.Is(err, ErrBodyTooLarge):
			r.metrics.recordRequest(ctx, environment, scm, resultInvalid)
			logger.Debug("webhook rejected", zap.String("reason", "body too large"), zap.Error(err))
			http.Error(w, "request body too large", http.StatusRequestEntityTooLarge)
		default:
			r.metrics.recordRequest(ctx, environment, scm, resultInvalid)
			logger.Debug("webhook rejected", zap.String("reason", "invalid request"), zap.Error(err))
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
		logger.Debug("webhook ignored",
			zap.String("kind", string(event.Kind)),
			zap.Strings("branches", event.Branches))
		w.WriteHeader(http.StatusOK)
		return
	}

	if !r.syncers[target.Repository].trigger(pendingSync{
		environment: environment,
		scm:         scm,
		received:    received,
	}) {
		logger.Debug("webhook rejected", zap.String("reason", "shutting down"))
		http.Error(w, "shutting down", http.StatusServiceUnavailable)
		return
	}

	r.metrics.recordRequest(ctx, environment, scm, resultAccepted)
	logger.Debug("webhook accepted; fetch scheduled", zap.Strings("branches", tracked))

	w.WriteHeader(http.StatusAccepted)
}

// Shutdown waits for webhook-triggered fetches, running or accepted and
// pending, to finish. If ctx is done first, it cancels them, waits for them
// to unwind and returns ctx's error. Call it after cancelling the context
// given to NewReceiver, so no new fetch is scheduled while it waits.
func (r *Receiver) Shutdown(ctx context.Context) error {
	defer r.cancelFetch()

	done := make(chan struct{})
	go func() {
		r.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		return nil
	case <-ctx.Done():
		r.cancelFetch()
		<-done
		return ctx.Err()
	}
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

// trigger schedules a fetch serving p. It reports false, scheduling nothing,
// once the receiver's context is cancelled.
func (s *syncer) trigger(p pendingSync) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	// checked under mu so no fetch is started after cancellation, which
	// keeps wg.Go from racing a Shutdown that began after cancellation
	if s.receiver.ctx.Err() != nil {
		return false
	}

	s.pending = append(s.pending, p)
	if s.running {
		return true
	}

	s.running = true
	s.receiver.wg.Go(s.run)

	return true
}

// run fetches until no triggers are pending. Each fetch serves every trigger
// accepted before it started. Triggers still pending once Shutdown cancels
// the fetch context are dropped.
func (s *syncer) run() {
	r := s.receiver

	for {
		s.mu.Lock()
		batch := s.pending
		s.pending = nil
		if len(batch) == 0 || r.fetchCtx.Err() != nil {
			s.running = false
			s.mu.Unlock()
			if len(batch) > 0 {
				r.logger.Debug("webhook-triggered fetch dropped at shutdown", zap.Int("webhooks", len(batch)))
			}
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
		logger.Debug("webhook-triggered fetch started")

		start := time.Now()
		err := s.repo.Fetch(r.fetchCtx)
		completed := time.Now()

		if err != nil && r.fetchCtx.Err() != nil {
			// aborted by a Shutdown timeout rather than failed
			logger.Debug("webhook-triggered fetch aborted", zap.Error(err))
			continue
		}

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
				r.metrics.recordFetchError(r.fetchCtx, p.environment, p.scm)
			}
			continue
		}

		logger.Debug("webhook-triggered fetch completed", zap.Duration("duration", completed.Sub(start)))

		for _, p := range batch {
			r.metrics.recordSync(r.fetchCtx, p.environment, p.scm, completed.Sub(p.received))
		}
	}
}
