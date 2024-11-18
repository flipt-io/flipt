package prometheus

import (
	"context"
	"fmt"
	"math"
	"time"

	"github.com/prometheus/client_golang/api"
	promapi "github.com/prometheus/client_golang/api/prometheus/v1"
	"github.com/prometheus/common/model"
	"go.flipt.io/flipt/internal/config"
	panalytics "go.flipt.io/flipt/internal/server/analytics"
	"go.uber.org/zap"
)

type prometheusClient interface {
	QueryRange(ctx context.Context, query string, r promapi.Range, opts ...promapi.Option) (model.Value, promapi.Warnings, error)
}

type client struct {
	promClient prometheusClient
	logger     *zap.Logger
}

func New(logger *zap.Logger, cfg *config.Config) (*client, error) {
	apiClient, err := api.NewClient(api.Config{
		Address: cfg.Analytics.Storage.Prometheus.URL,
	})
	if err != nil {
		return nil, err
	}
	promClient := promapi.NewAPI(apiClient)
	return &client{promClient: promClient, logger: logger}, nil
}

func (c *client) GetFlagEvaluationsCount(ctx context.Context, req *panalytics.FlagEvaluationsCountRequest) ([]string, []float32, error) {
	query := fmt.Sprintf(
		`sum(increase(flipt_evaluations_requests_total{namespace="%s", flag="%s"}[%dm])) or vector(0)`,
		req.NamespaceKey,
		req.FlagKey,
		req.StepMinutes,
	)
	r := promapi.Range{
		Start: req.From.UTC(),
		End:   req.To.Add(time.Duration(req.StepMinutes) * time.Minute).UTC(),
		Step:  time.Duration(req.StepMinutes) * time.Minute,
	}
	data, warnings, err := c.promClient.QueryRange(ctx, query, r)
	if err != nil {
		return nil, nil, err
	}

	if len(warnings) > 0 {
		c.logger.Warn("prometheus query returned warnings", zap.Strings("warnings", warnings))
	}

	m, ok := data.(model.Matrix)
	if !ok || len(m) != 1 {
		return nil, nil, fmt.Errorf("unexpected data type returned from prometheus")
	}

	v := m[0]
	var (
		timestamps = make([]string, len(v.Values))
		values     = make([]float32, len(v.Values))
	)

	for i, vv := range v.Values {
		timestamps[i] = time.UnixMilli(int64(vv.Timestamp)).Format(time.DateTime)
		values[i] = float32(math.Round(float64(vv.Value)))
	}
	return timestamps, values, nil
}

func (c *client) String() string {
	return "prometheus"
}
