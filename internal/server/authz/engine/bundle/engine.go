package bundle

import (
	"context"
	"strings"

	"github.com/open-policy-agent/opa/sdk"
	"github.com/open-policy-agent/opa/storage/inmem"
	"go.flipt.io/flipt/internal/server/authz"
	"go.uber.org/zap"
)

var _ authz.Verifier = (*Engine)(nil)

type Engine struct {
	opa    *sdk.OPA
	logger *zap.Logger
}

const opaConfig = `
services:
  flipt:
    url: http://localhost:9001/
  s3:
    url: https://my-example-opa-bucket.s3.eu-north-1.amazonaws.com
    credentials:
      s3_signing:
        environment_credentials: {}

bundles:
  flipt:
    service: flipt
    resource: bundle.tar.gz
`

func NewEngine(ctx context.Context, logger *zap.Logger) (*Engine, error) {
	opa, err := sdk.New(ctx, sdk.Options{
		Config: strings.NewReader(opaConfig),
		Store:  inmem.New(),
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
