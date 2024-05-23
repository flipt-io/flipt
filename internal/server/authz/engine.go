package authz

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/open-policy-agent/opa/rego"
	"github.com/open-policy-agent/opa/storage"
	"github.com/open-policy-agent/opa/storage/inmem"
	"go.flipt.io/flipt/internal/containers"
	"go.uber.org/zap"
)

var (
	_ Verifier = &Engine{}

	defaultPollDuration = 30 * time.Second

	// ErrNotModified is returned from a source when the data has not
	// been modified, identified based on the provided hash value
	ErrNotModified = errors.New("not modified")
)

type Verifier interface {
	IsAllowed(ctx context.Context, input map[string]any) (bool, error)
}

type CachedSource[T any] interface {
	Get(_ context.Context, hash []byte) (T, []byte, error)
}

type PolicySource CachedSource[[]byte]

type DataSource CachedSource[map[string]any]

type Engine struct {
	logger *zap.Logger

	mu    sync.RWMutex
	query rego.PreparedEvalQuery
	store storage.Store

	policySource PolicySource
	policyHash   []byte

	dataSource DataSource
	dataHash   []byte

	pollDuration time.Duration
}

func WithDataSource(source DataSource) containers.Option[Engine] {
	return func(e *Engine) {
		e.dataSource = source
	}
}

func WithPollDuration(dur time.Duration) containers.Option[Engine] {
	return func(e *Engine) {
		e.pollDuration = dur
	}
}

func NewEngine(ctx context.Context, logger *zap.Logger, source PolicySource, opts ...containers.Option[Engine]) (*Engine, error) {
	engine := &Engine{
		logger:       logger,
		policySource: source,
		store:        inmem.New(),
		pollDuration: defaultPollDuration,
	}

	containers.ApplyAll(engine, opts...)

	err := engine.update(ctx, storage.AddOp)
	if err != nil {
		return nil, err
	}

	go engine.pollData(ctx)

	return engine, nil
}

func (e *Engine) IsAllowed(ctx context.Context, input map[string]interface{}) (bool, error) {
	e.mu.RLock()
	defer e.mu.RUnlock()

	results, err := e.query.Eval(ctx, rego.EvalInput(input))
	if err != nil {
		return false, err
	}

	if len(results) == 0 {
		return false, nil
	}

	return results[0].Expressions[0].Value.(bool), nil
}

func (e *Engine) pollData(ctx context.Context) {
	ticker := time.NewTicker(e.pollDuration)
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := e.update(ctx, storage.ReplaceOp); err != nil {
				e.logger.Error("updating policy and data", zap.Error(err))
			}
		}
	}
}

func (e *Engine) update(ctx context.Context, op storage.PatchOp) error {
	if err := e.updateData(ctx, op); err != nil {
		return err
	}

	return e.updatePolicy(ctx)
}

func (e *Engine) updatePolicy(ctx context.Context) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	policy, hash, err := e.policySource.Get(ctx, e.policyHash)
	if err != nil {
		if errors.Is(err, ErrNotModified) {
			return nil
		}

		return fmt.Errorf("getting policy definition: %w", err)
	}

	e.policyHash = hash

	r := rego.New(
		rego.Query("data.authz.v1.allow"),
		rego.Module("policy.rego", string(policy)),
		rego.Store(e.store),
	)

	e.query, err = r.PrepareForEval(ctx)
	if err != nil {
		return fmt.Errorf("preparing policy: %w", err)
	}

	return nil
}

func (e *Engine) updateData(ctx context.Context, op storage.PatchOp) (err error) {
	if e.dataSource == nil {
		return nil
	}

	data, hash, err := e.dataSource.Get(ctx, e.dataHash)
	if err != nil {
		if errors.Is(err, ErrNotModified) {
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
