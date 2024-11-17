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
	"go.flipt.io/flipt/rpc/flipt/analytics"
	"go.uber.org/zap"
)

type client struct {
	promClient promapi.API
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

func (c *client) getStepFromDuration(duration time.Duration) int {
	switch {
	case duration >= 12*time.Hour:
		return 15
	case duration >= 4*time.Hour:
		return 5
	default:
		return 1
	}
}

func (c *client) GetFlagEvaluationsCount(ctx context.Context, req *analytics.GetFlagEvaluationsCountRequest) ([]string, []float32, error) {
	fromTime, err := time.Parse(time.DateTime, req.From)
	if err != nil {
		return nil, nil, err
	}

	toTime, err := time.Parse(time.DateTime, req.To)
	if err != nil {
		return nil, nil, err
	}

	step := c.getStepFromDuration(toTime.Sub(fromTime))
	query := fmt.Sprintf(
		`sum(increase(flipt_evaluations_requests_total{namespace="%s", flag="%s"}[%dm]))`,
		req.GetNamespaceKey(),
		req.GetFlagKey(),
		step,
	)
	toTime = toTime.Add(time.Duration(step) * time.Minute)
	r := promapi.Range{
		Start: fromTime,
		End:   toTime,
		Step:  time.Duration(step) * time.Minute,
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
