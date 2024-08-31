package bundle

import (
	"context"
	"os"
	"strings"

	"github.com/open-policy-agent/contrib/logging/plugins/ozap"
	"github.com/open-policy-agent/opa/sdk"
	"github.com/open-policy-agent/opa/storage/inmem"
	"go.flipt.io/flipt/internal/config"
	"go.flipt.io/flipt/internal/server/authz"
	_ "go.flipt.io/flipt/internal/server/authz/engine/ext"
	"go.uber.org/zap"
)

var _ authz.Verifier = (*Engine)(nil)

type cleanupFunc func()

type Engine struct {
	opa          *sdk.OPA
	logger       *zap.Logger
	cleanupFuncs []cleanupFunc
}

func NewEngine(ctx context.Context, logger *zap.Logger, cfg *config.Config) (*Engine, error) {
	var (
		authConfig   = cfg.Authorization
		opaConfig    string
		cleanupFuncs []cleanupFunc
	)

	switch authConfig.Backend {
	case config.AuthorizationBackendObject:
		opaConfig = authConfig.Object.String()

		switch authConfig.Object.Type { //nolint
		case config.S3ObjectAuthorizationBackendType:
			// set AWS_REGION env var if not set and region is specified
			// this is a nicety as the OPA env credentials provider requires this env var
			// to be set, but we don't want the user to have to supply it twice if they already have it in the config
			if authConfig.Object.S3.Region != "" && os.Getenv("AWS_REGION") == "" {
				os.Setenv("AWS_REGION", authConfig.Object.S3.Region)
				cleanupFuncs = append(cleanupFuncs, func() {
					os.Unsetenv("AWS_REGION")
				})
			}
		}
	case config.AuthorizationBackendBundle:
		opaConfig = authConfig.Bundle.String()
	}

	level := zap.NewAtomicLevelAt(logger.Level())

	opa, err := sdk.New(ctx, sdk.Options{
		Config: strings.NewReader(opaConfig),
		Store:  inmem.New(),
		Logger: ozap.Wrap(logger, &level),
	})
	if err != nil {
		return nil, err
	}

	return &Engine{
		logger:       logger,
		opa:          opa,
		cleanupFuncs: cleanupFuncs,
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

	allow, _ := dec.Result.(bool)
	return allow, nil
}

func (e *Engine) Shutdown(ctx context.Context) error {
	e.opa.Stop(ctx)
	for _, cleanup := range e.cleanupFuncs {
		cleanup()
	}
	return nil
}
