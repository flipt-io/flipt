package prometheus

import (
	"context"
	"fmt"
	"math"
	"net/http"
	"slices"
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
		RoundTripper: roundTripFunc(func(r *http.Request) (*http.Response, error) {
			for k, v := range cfg.Analytics.Storage.Prometheus.Headers {
				r.Header.Set(k, v)
			}
			return api.DefaultRoundTripper.RoundTrip(r)
		}),
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
	if !ok || len(m) == 0 {
		return nil, nil, fmt.Errorf("unexpected data type returned from prometheus")
	}

	v := m[len(m)-1]
	var (
		timestamps = make([]string, len(v.Values))
		values     = make([]float32, len(v.Values))
	)

	for i, vv := range v.Values {
		timestamps[i] = time.UnixMilli(int64(vv.Timestamp)).UTC().Format(time.RFC3339)
		values[i] = float32(math.Round(float64(vv.Value)))
	}

	// Prometheus returns one data series for our query with vector(0).
	// Victoria Metrics returns one or two data series: one for the actual data and one for
	// the vector(0). The actual data has only points with non-zero values or may absent at all
	// if there is no data. The vector(0) has only points with zero values. We need to combine it
	for i := len(m) - 2; i >= 0; i-- {
		for _, vv := range m[i].Values {
			t := time.UnixMilli(int64(vv.Timestamp)).UTC().Format(time.RFC3339)
			if i := slices.Index(timestamps, t); i != -1 {
				values[i] += float32(math.Round(float64(vv.Value)))
			}
		}
	}

	return timestamps, values, nil
}

func (c *client) String() string {
	return "prometheus"
}

type roundTripFunc func(r *http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) {
	return f(r)
}
