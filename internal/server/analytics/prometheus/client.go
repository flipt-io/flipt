package prometheus

import (
	"context"
	"fmt"
	"math"
	"net/http"
	"slices"
	"strings"
	"time"

	"github.com/prometheus/client_golang/api"
	promapi "github.com/prometheus/client_golang/api/prometheus/v1"
	"github.com/prometheus/common/model"
	"go.flipt.io/flipt/internal/config"
	panalytics "go.flipt.io/flipt/internal/server/analytics"
	"go.uber.org/zap"
)

type PrometheusClient interface {
	QueryRange(ctx context.Context, query string, r promapi.Range, opts ...promapi.Option) (model.Value, promapi.Warnings, error)
}

type client struct {
	promClient PrometheusClient
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
		`sum(increase(flipt_evaluations_requests_total{flipt_environment="%s", flipt_namespace="%s", flipt_flag="%s"}[%dm])) or vector(0)`,
		req.EnvironmentKey,
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

// GetBatchFlagEvaluationsCount retrieves evaluation counts for multiple flags in a single query
func (c *client) GetBatchFlagEvaluationsCount(ctx context.Context, req *panalytics.BatchFlagEvaluationsCountRequest) (map[string]panalytics.FlagEvaluationData, error) {
	// Create a regex pattern matching any of the requested flags
	flagRegex := strings.Join(req.FlagKeys, "|")

	query := fmt.Sprintf(
		`sum(increase(flipt_evaluations_requests_total{flipt_environment="%s", flipt_namespace="%s", flipt_flag=~"%s"}[%dm])) by (flipt_flag) or vector(0)`,
		req.EnvironmentKey,
		req.NamespaceKey,
		flagRegex,
		req.StepMinutes,
	)

	r := promapi.Range{
		Start: req.From.UTC(),
		End:   req.To.Add(time.Duration(req.StepMinutes) * time.Minute).UTC(),
		Step:  time.Duration(req.StepMinutes) * time.Minute,
	}

	data, warnings, err := c.promClient.QueryRange(ctx, query, r)
	if err != nil {
		return nil, err
	}

	if len(warnings) > 0 {
		c.logger.Warn("prometheus query returned warnings", zap.Strings("warnings", warnings))
	}

	m, ok := data.(model.Matrix)
	if !ok {
		return nil, fmt.Errorf("unexpected data type returned from prometheus")
	}

	// Extract base timestamps from the empty key series (vector(0))
	var baseTimestamps []string

	// First, look for the series with empty flag key (from vector(0))
	for _, series := range m {
		flagKey := string(series.Metric["flipt_flag"])
		if flagKey == "" {
			baseTimestamps = make([]string, len(series.Values))
			for i, v := range series.Values {
				baseTimestamps[i] = time.UnixMilli(int64(v.Timestamp)).UTC().Format(time.RFC3339)
			}
			break
		}
	}

	// If we didn't find a base series, use timestamps from the first series
	if len(baseTimestamps) == 0 && len(m) > 0 {
		series := m[0]
		baseTimestamps = make([]string, len(series.Values))
		for i, v := range series.Values {
			baseTimestamps[i] = time.UnixMilli(int64(v.Timestamp)).UTC().Format(time.RFC3339)
		}
	}

	// Map to store results by flag key
	result := make(map[string]panalytics.FlagEvaluationData)

	// Initialize result map with zero values for all timestamps for each requested flag
	for _, flagKey := range req.FlagKeys {
		values := make([]float32, len(baseTimestamps))
		// All values start as 0
		result[flagKey] = panalytics.FlagEvaluationData{
			Timestamps: slices.Clone(baseTimestamps),
			Values:     values,
		}
	}

	// Process each time series (one per flag)
	for _, series := range m {
		flagKey := string(series.Metric["flipt_flag"])

		// Skip series with empty flag keys (from vector(0))
		if flagKey == "" {
			continue
		}

		// Skip series for flags we didn't request
		if !slices.Contains(req.FlagKeys, flagKey) {
			continue
		}

		// Create a map of timestamp to value for this flag
		timestampToValue := make(map[string]float32)
		for _, v := range series.Values {
			ts := time.UnixMilli(int64(v.Timestamp)).UTC().Format(time.RFC3339)
			timestampToValue[ts] = float32(math.Round(float64(v.Value)))
		}

		// Update the values in the result using the consistent timeline
		flagData := result[flagKey]
		for i, ts := range flagData.Timestamps {
			if value, exists := timestampToValue[ts]; exists {
				flagData.Values[i] = value
			}
			// If the timestamp doesn't exist in this series, it keeps the zero value
		}

		result[flagKey] = flagData
	}

	// Apply limit if specified
	if req.Limit > 0 && len(baseTimestamps) > req.Limit {
		for flagKey, data := range result {
			factor := len(data.Timestamps) / req.Limit
			compressedValues := make([]float32, req.Limit)
			compressedTimestamps := make([]string, req.Limit)

			for i := 0; i < req.Limit; i++ {
				start := i * factor
				end := (i + 1) * factor
				if end > len(data.Values) {
					end = len(data.Values)
				}

				sum := float32(0)
				for j := start; j < end; j++ {
					sum += data.Values[j]
				}

				// Avoid division by zero
				if end > start {
					compressedValues[i] = sum / float32(end-start)
				}
				compressedTimestamps[i] = data.Timestamps[start]
			}

			result[flagKey] = panalytics.FlagEvaluationData{
				Timestamps: compressedTimestamps,
				Values:     compressedValues,
			}
		}
	}

	return result, nil
}

func (c *client) String() string {
	return "prometheus"
}

type roundTripFunc func(r *http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) {
	return f(r)
}
