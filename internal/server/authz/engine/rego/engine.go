package rego

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/open-policy-agent/opa/rego"
	"github.com/open-policy-agent/opa/storage"
	"github.com/open-policy-agent/opa/storage/inmem"
	"go.flipt.io/flipt/internal/config"
	"go.flipt.io/flipt/internal/containers"
	"go.flipt.io/flipt/internal/server/authz"
	_ "go.flipt.io/flipt/internal/server/authz/engine/ext"
	"go.flipt.io/flipt/internal/server/authz/engine/rego/source"
	"go.flipt.io/flipt/internal/server/authz/engine/rego/source/filesystem"
	"go.uber.org/zap"
)

var (
	_                         authz.Verifier = (*Engine)(nil)
	defaultPolicyPollDuration                = 5 * time.Minute
)

type CachedSource[T any] interface {
	Get(_ context.Context, hash source.Hash) (T, source.Hash, error)
}

type PolicySource CachedSource[[]byte]

type DataSource CachedSource[map[string]any]

type Engine struct {
	logger *zap.Logger

	mu                sync.RWMutex
	queryAllow        rego.PreparedEvalQuery
	queryEnvironments *rego.PreparedEvalQuery
	queryNamespaces   *rego.PreparedEvalQuery
	store             storage.Store

	policySource PolicySource
	policyHash   source.Hash

	dataSource DataSource
	dataHash   source.Hash

	policySourcePollDuration time.Duration
	dataSourcePollDuration   time.Duration
}

func withPolicySource(source PolicySource) containers.Option[Engine] {
	return func(e *Engine) {
		e.policySource = source
	}
}

func withDataSource(source DataSource, pollDuration time.Duration) containers.Option[Engine] {
	return func(e *Engine) {
		e.dataSource = source
		e.dataSourcePollDuration = pollDuration
	}
}

func withPolicySourcePollDuration(dur time.Duration) containers.Option[Engine] {
	return func(e *Engine) {
		e.policySourcePollDuration = dur
	}
}

// NewEngine creates a new local authorization engine
func NewEngine(ctx context.Context, logger *zap.Logger, cfg *config.Config) (*Engine, error) {
	var (
		opts       []containers.Option[Engine]
		authConfig = cfg.Authorization
	)

	switch authConfig.Backend {
	case config.AuthorizationBackendLocal:
		opts = []containers.Option[Engine]{
			withPolicySource(filesystem.PolicySourceFromPath(authConfig.Local.Policy.Path)),
		}

		if authConfig.Local.Policy.PollInterval > 0 {
			opts = append(opts, withPolicySourcePollDuration(authConfig.Local.Policy.PollInterval))
		}

		if authConfig.Local.Data != nil {
			opts = append(opts, withDataSource(
				filesystem.DataSourceFromPath(authConfig.Local.Data.Path),
				authConfig.Local.Data.PollInterval,
			))
		}

	default:
		return nil, fmt.Errorf("unsupported authorization backend: %s", authConfig.Backend)
	}

	return newEngine(ctx, logger, opts...)
}

// newEngine creates a new engine with the provided options, visible for testing
func newEngine(ctx context.Context, logger *zap.Logger, opts ...containers.Option[Engine]) (*Engine, error) {
	engine := &Engine{
		logger:                   logger,
		store:                    inmem.New(),
		policySourcePollDuration: defaultPolicyPollDuration,
	}

	containers.ApplyAll(engine, opts...)

	// update data store with initial data if source is configured
	if err := engine.updateData(ctx, storage.AddOp); err != nil {
		return nil, err
	}

	// fetch policy and then compile and set query engine
	if err := engine.updatePolicy(ctx); err != nil {
		return nil, err
	}

	// begin polling for updates for policy
	go poll(ctx, engine.policySourcePollDuration, func() {
		if err := engine.updatePolicy(ctx); err != nil {
			engine.logger.Error("updating policy", zap.Error(err))
		}
	})

	// being polling for updates to data if source configured
	if engine.dataSource != nil {
		go poll(ctx, engine.dataSourcePollDuration, func() {
			if err := engine.updateData(ctx, storage.ReplaceOp); err != nil {
				engine.logger.Error("updating data", zap.Error(err))
			}
		})
	}

	return engine, nil
}

func (e *Engine) IsAllowed(ctx context.Context, input map[string]any) (bool, error) {
	e.mu.RLock()
	defer e.mu.RUnlock()

	e.logger.Debug("evaluating policy", zap.Any("input", input))
	results, err := e.queryAllow.Eval(ctx, rego.EvalInput(input))
	if err != nil {
		return false, err
	}

	if len(results) == 0 {
		return false, nil
	}

	return results[0].Expressions[0].Value.(bool), nil
}

func (e *Engine) ViewableEnvironments(ctx context.Context, input map[string]any) ([]string, error) {
	e.mu.RLock()
	defer e.mu.RUnlock()

	e.logger.Debug("evaluating viewable environments", zap.Any("input", input))

	if e.queryEnvironments == nil {
		e.logger.Debug("environments query not prepared, skipping evaluation")
		return nil, nil
	}

	results, err := e.queryEnvironments.Eval(ctx, rego.EvalInput(input))
	if err != nil {
		return nil, fmt.Errorf("evaluating viewable environments: %w", err)
	}

	if len(results) == 0 {
		return nil, nil
	}

	values, ok := results[0].Expressions[0].Value.([]any)
	if !ok {
		return nil, fmt.Errorf("unexpected result type: %T", results[0].Expressions[0].Value)
	}

	environments := make([]string, len(values))
	for i, env := range values {
		environments[i] = fmt.Sprintf("%s", env)
	}
	return environments, nil
}

func (e *Engine) ViewableNamespaces(ctx context.Context, env string, input map[string]any) ([]string, error) {
	e.mu.RLock()
	defer e.mu.RUnlock()

	e.logger.Debug("evaluating viewable namespaces",
		zap.String("environment", env),
		zap.Any("input", input))

	if e.queryNamespaces == nil {
		e.logger.Debug("namespaces query not prepared, skipping evaluation")
		return nil, nil
	}

	// Add environment to input for Rego evaluation
	input["environment"] = env

	results, err := e.queryNamespaces.Eval(ctx, rego.EvalInput(input))
	if err != nil {
		return nil, fmt.Errorf("evaluating viewable namespaces: %w", err)
	}

	if len(results) == 0 {
		return nil, nil
	}

	// The result will be in the "x" variable from our query
	values, ok := results[0].Bindings["x"].([]any)
	if !ok {
		return nil, fmt.Errorf("unexpected result type: %T", results[0].Bindings["x"])
	}

	namespaces := make([]string, len(values))
	for i, ns := range values {
		namespaces[i] = fmt.Sprintf("%s", ns)
	}
	return namespaces, nil
}

func (e *Engine) Shutdown(_ context.Context) error {
	return nil
}

func poll(ctx context.Context, d time.Duration, fn func()) {
	ticker := time.NewTicker(d)
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			fn()
		}
	}
}

func (e *Engine) updatePolicy(ctx context.Context) error {
	e.mu.RLock()
	policyHash := e.policyHash
	e.mu.RUnlock()

	policy, hash, err := e.policySource.Get(ctx, policyHash)
	if err != nil {
		if errors.Is(err, source.ErrNotModified) {
			return nil
		}

		return fmt.Errorf("getting policy definition: %w", err)
	}

	m := rego.Module("policy.rego", string(policy))
	s := rego.Store(e.store)

	// Prepare allow query
	r := rego.New(
		rego.Query("data.flipt.authz.v2.allow"),
		m,
		s,
	)

	queryAllow, err := r.PrepareForEval(ctx)
	if err != nil {
		return fmt.Errorf("preparing policy allow: %w", err)
	}

	// Prepare environments query
	r = rego.New(
		rego.Query("data.flipt.authz.v2.viewable_environments"),
		m,
		s,
	)

	queryEnvironments, err := r.PrepareForEval(ctx)
	if err == nil {
		// queryEnvironments is optional, so we dont error here
		e.queryEnvironments = &queryEnvironments
	}

	// Prepare namespaces query
	r = rego.New(
		rego.Query("x = data.flipt.authz.v2.viewable_namespaces(input.environment)"),
		m,
		s,
	)

	queryNamespaces, err := r.PrepareForEval(ctx)
	if err == nil {
		// queryNamespaces is optional, so we dont error here
		e.queryNamespaces = &queryNamespaces
	}

	e.mu.Lock()
	defer e.mu.Unlock()
	if !bytes.Equal(e.policyHash, policyHash) {
		e.logger.Warn("policy hash doesn't match original one. skipping updating")
		return nil
	}
	e.policyHash = hash
	e.queryAllow = queryAllow

	return nil
}

func (e *Engine) updateData(ctx context.Context, op storage.PatchOp) (err error) {
	if e.dataSource == nil {
		return nil
	}

	data, hash, err := e.dataSource.Get(ctx, e.dataHash)
	if err != nil {
		if errors.Is(err, source.ErrNotModified) {
			return nil
		}

		return fmt.Errorf("getting data for policy evaluation: %w", err)
	}

	e.dataHash = hash

	txn, err := e.store.NewTransaction(ctx, storage.WriteParams)
	if err != nil {
		return err
	}

	if err := e.store.Write(ctx, txn, op, storage.Path{}, data); err != nil {
		return err
	}

	return e.store.Commit(ctx, txn)
}
