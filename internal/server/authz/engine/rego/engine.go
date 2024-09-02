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
	"go.flipt.io/flipt/internal/server/authz/engine/rego/source/cloud"
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

	mu    sync.RWMutex
	query rego.PreparedEvalQuery
	store storage.Store

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

	case config.AuthorizationBackendCloud:
		opts = []containers.Option[Engine]{
			withPolicySource(cloud.PolicySourceFromCloud(cfg.Cloud.Host, cfg.Cloud.Authentication.ApiKey)),
			withDataSource(cloud.DataSourceFromCloud(cfg.Cloud.Host, cfg.Cloud.Authentication.ApiKey), authConfig.Cloud.PollInterval),
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

func (e *Engine) IsAllowed(ctx context.Context, input map[string]interface{}) (bool, error) {
	e.mu.RLock()
	defer e.mu.RUnlock()

	e.logger.Debug("evaluating policy", zap.Any("input", input))
	results, err := e.query.Eval(ctx, rego.EvalInput(input))
	if err != nil {
		return false, err
	}

	if len(results) == 0 {
		return false, nil
	}

	return results[0].Expressions[0].Value.(bool), nil
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

	r := rego.New(
		rego.Query("data.flipt.authz.v1.allow"),
		rego.Module("policy.rego", string(policy)),
		rego.Store(e.store),
	)

	query, err := r.PrepareForEval(ctx)
	if err != nil {
		return fmt.Errorf("preparing policy: %w", err)
	}

	e.mu.Lock()
	defer e.mu.Unlock()
	if !bytes.Equal(e.policyHash, policyHash) {
		e.logger.Warn("policy hash doesn't match original one. skipping updating")
		return nil
	}
	e.policyHash = hash
	e.query = query

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
