package bundle

import (
	"context"
	"strings"

	"github.com/open-policy-agent/contrib/logging/plugins/ozap"
	"github.com/open-policy-agent/opa/sdk"
	"github.com/open-policy-agent/opa/storage/inmem"
	"go.flipt.io/flipt/internal/config"
	"go.flipt.io/flipt/internal/server/authz"
	"go.uber.org/zap"
)

var _ authz.Verifier = (*Engine)(nil)

type Engine struct {
	opa    *sdk.OPA
	logger *zap.Logger
}

func NewEngine(ctx context.Context, logger *zap.Logger, cfg *config.Config) (*Engine, error) {
	var opaConfig string

	switch cfg.Authorization.Backend {
	case config.AuthorizationBackendObject:
		opaConfig = cfg.Authorization.Object.String()
	case config.AuthorizationBackendBundle:
		opaConfig = cfg.Authorization.Bundle.String()
	}

	level, err := zap.ParseAtomicLevel(cfg.Log.Level)
	if err != nil {
		return nil, err
	}

	opa, err := sdk.New(ctx, sdk.Options{
		Config: strings.NewReader(opaConfig),
		Store:  inmem.New(),
		Logger: ozap.Wrap(logger, &level),
	})
	if err != nil {
		return nil, err
	}

	return &Engine{
		logger: logger,
		opa:    opa,
	}, nil
}

func (e *Engine) IsAllowed(ctx context.Context, input map[string]interface{}) (bool, error) {
	e.logger.Debug("evaluating policy", zap.Any("input", input))
	dec, err := e.opa.Decision(ctx, sdk.DecisionOptions{
		Path:  "flipt/authz/v1/allow",
		Input: input,
	})

	if err != nil {
		return false, err
	}

	allow, ok := dec.Result.(bool)
	if !ok || !allow {
		return false, nil
	}

	return true, nil
}

func (e *Engine) Shutdown(ctx context.Context) error {
	e.opa.Stop(ctx)
	return nil
}
