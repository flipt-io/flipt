package prometheus

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"regexp"
	"strconv"
	"strings"
	"testing"
	"time"

	promapi "github.com/prometheus/client_golang/api/prometheus/v1"
	"github.com/prometheus/common/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.flipt.io/flipt/internal/config"
	panalytics "go.flipt.io/flipt/internal/server/analytics"
	"go.uber.org/zap/zaptest"
)

func TestNew(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("x-app-key") != "1234" {
			t.Error("missing or invalid x-app-key header")
		}

		w.WriteHeader(http.StatusOK)
		_, err := w.Write([]byte(`{
       "status":"success",
       "data":{
          "resultType":"matrix",
          "result":[{"metric":{},"values":[[1732699504.975,"0"]]}]
       }}`))
		if err != nil {
			t.Fatalf("failed to write response: %v", err)
		}
	}))
	t.Cleanup(srv.Close)

	logger := zaptest.NewLogger(t)
	cfg := config.Default()
	cfg.Analytics.Storage.Prometheus.URL = srv.URL
	cfg.Analytics.Storage.Prometheus.Headers = map[string]string{
		"x-app-key": "1234",
	}
	s, err := New(logger, cfg)
	require.NoError(t, err)
	assert.Equal(t, "prometheus", s.String())
	_, _, err = s.GetFlagEvaluationsCount(t.Context(), &panalytics.FlagEvaluationsCountRequest{})
	require.NoError(t, err)

	cfg.Analytics.Storage.Prometheus.URL = "\t"
	_, err = New(logger, cfg)
	require.Error(t, err)
}

func TestNewWithSigV4(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify that the Authorization header contains AWS SigV4 signature
		auth := r.Header.Get("Authorization")
		assert.Contains(t, auth, "AWS4-HMAC-SHA256")

		w.WriteHeader(http.StatusOK)
		_, err := w.Write([]byte(`{
       "status":"success",
       "data":{
          "resultType":"matrix",
          "result":[{"metric":{},"values":[[1732699504.975,"0"]]}]
       }}`))
		if err != nil {
			t.Fatalf("failed to write response: %v", err)
		}
	}))
	t.Cleanup(srv.Close)

	logger := zaptest.NewLogger(t)
	cfg := config.Default()
	cfg.Analytics.Storage.Prometheus.URL = srv.URL
	cfg.Analytics.Storage.Prometheus.SigV4 = config.PrometheusSigV4Config{
		Enabled:     true,
		Region:      "us-east-1",
		AccessKey:   "AKIAIOSFODNN7EXAMPLE",
		SecretKey:   "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY",
		ServiceName: "aps",
	}

	s, err := New(logger, cfg)
	require.NoError(t, err)
	assert.Equal(t, "prometheus", s.String())

	// Verify the client can make requests with SigV4 signing
	_, _, err = s.GetFlagEvaluationsCount(t.Context(), &panalytics.FlagEvaluationsCountRequest{})
	require.NoError(t, err)
}

func TestGetFlagEvaluationsCount(t *testing.T) {
	ctx := t.Context()
	from := time.Now().Add(-time.Hour).UTC()
	to := time.Now().UTC()
	mock := NewMockPrometheusClient(t)

	logger := zaptest.NewLogger(t)
	t.Run("success", func(t *testing.T) {
		data := model.Matrix{
			{
				Values: []model.SamplePair{
					{Timestamp: model.Time(from.UnixMilli()), Value: 1},
					{Timestamp: model.Time(to.UnixMilli()), Value: 2},
				},
			},
		}
		mock.EXPECT().QueryRange(
			ctx,
			`sum(increase(flipt_evaluations_requests_total{flipt_environment="default", flipt_namespace="bar", flipt_flag="foo"}[1m])) or vector(0)`,
			promapi.Range{
				Start: from,
				End:   to.Add(time.Minute),
				Step:  time.Minute,
			},
		).Return(data, promapi.Warnings{"test Warnings"}, nil, nil)

		client := &client{
			logger:     logger,
			promClient: mock,
		}

		labels, values, err := client.GetFlagEvaluationsCount(ctx, &panalytics.FlagEvaluationsCountRequest{
			EnvironmentKey: "default",
			NamespaceKey:   "bar",
			FlagKey:        "foo",
			From:           from,
			To:             to,
			StepMinutes:    1,
		})

		require.NoError(t, err)
		assert.Len(t, labels, 2)
		assert.Len(t, values, 2)
	})

	t.Run("success with victoria metrics", func(t *testing.T) {
		data := model.Matrix{
			{
				Values: []model.SamplePair{
					{Timestamp: model.Time(from.Add(1 * time.Minute).UnixMilli()), Value: 1},
					{Timestamp: model.Time(from.Add(4 * time.Minute).UnixMilli()), Value: 4},
				},
			},
			{
				Values: []model.SamplePair{
					{Timestamp: model.Time(from.Add(1 * time.Minute).UnixMilli()), Value: 0},
					{Timestamp: model.Time(from.Add(2 * time.Minute).UnixMilli()), Value: 0},
					{Timestamp: model.Time(from.Add(3 * time.Minute).UnixMilli()), Value: 0},
					{Timestamp: model.Time(from.Add(4 * time.Minute).UnixMilli()), Value: 0},
				},
			},
		}
		mock.EXPECT().QueryRange(
			ctx,
			`sum(increase(flipt_evaluations_requests_total{flipt_environment="default", flipt_namespace="victoria", flipt_flag="metrics"}[1m])) or vector(0)`,
			promapi.Range{
				Start: from,
				End:   to.Add(time.Minute),
				Step:  time.Minute,
			},
		).Return(data, promapi.Warnings{"test Warnings"}, nil, nil)

		client := &client{
			logger:     logger,
			promClient: mock,
		}

		labels, values, err := client.GetFlagEvaluationsCount(ctx, &panalytics.FlagEvaluationsCountRequest{
			EnvironmentKey: "default",
			NamespaceKey:   "victoria",
			FlagKey:        "metrics",
			From:           from,
			To:             to,
			StepMinutes:    1,
		})

		require.NoError(t, err)
		assert.Len(t, labels, 4)
		assert.Len(t, values, 4)
		assert.Equal(t, []float32{1, 0, 0, 4}, values)
	})

	t.Run("no data type", func(t *testing.T) {
		mock.EXPECT().QueryRange(
			ctx,
			`sum(increase(flipt_evaluations_requests_total{flipt_environment="default", flipt_namespace="no-data", flipt_flag="foo"}[1m])) or vector(0)`,
			promapi.Range{
				Start: from,
				End:   to.Add(time.Minute),
				Step:  time.Minute,
			},
		).Return(&model.Scalar{}, nil, nil, nil)

		client := &client{
			logger:     logger,
			promClient: mock,
		}

		_, _, err := client.GetFlagEvaluationsCount(ctx, &panalytics.FlagEvaluationsCountRequest{
			EnvironmentKey: "default",
			NamespaceKey:   "no-data",
			FlagKey:        "foo",
			From:           from,
			To:             to,
			StepMinutes:    1,
		})

		require.Error(t, err)
	})

	t.Run("wrong data type", func(t *testing.T) {
		mock.EXPECT().QueryRange(
			ctx,
			`sum(increase(flipt_evaluations_requests_total{flipt_environment="default", flipt_namespace="wrong-data", flipt_flag="foo"}[1m])) or vector(0)`,
			promapi.Range{
				Start: from,
				End:   to.Add(time.Minute),
				Step:  time.Minute,
			},
		).Return(&model.Matrix{}, nil, nil, nil)

		client := &client{
			logger:     logger,
			promClient: mock,
		}

		_, _, err := client.GetFlagEvaluationsCount(ctx, &panalytics.FlagEvaluationsCountRequest{
			EnvironmentKey: "default",
			NamespaceKey:   "wrong-data",
			FlagKey:        "foo",
			From:           from,
			To:             to,
			StepMinutes:    1,
		})

		require.Error(t, err)
	})

	t.Run("error", func(t *testing.T) {
		mock.EXPECT().QueryRange(
			ctx,
			`sum(increase(flipt_evaluations_requests_total{flipt_environment="default", flipt_namespace="error", flipt_flag="foo"}[5m])) or vector(0)`,
			promapi.Range{
				Start: from,
				End:   to.Add(5 * time.Minute),
				Step:  5 * time.Minute,
			},
		).Return(nil, nil, nil, errors.New("failed to query"))

		client := &client{
			logger:     logger,
			promClient: mock,
		}

		_, _, err := client.GetFlagEvaluationsCount(ctx, &panalytics.FlagEvaluationsCountRequest{
			EnvironmentKey: "default",
			NamespaceKey:   "error",
			FlagKey:        "foo",
			From:           from,
			To:             to,
			StepMinutes:    5,
		})

		require.Error(t, err)
	})
}

func TestGetBatchFlagEvaluationsCount(t *testing.T) {
	ctx := t.Context()
	from := time.Now().Add(-time.Hour).UTC()
	to := time.Now().UTC()
	mock := NewMockPrometheusClient(t)

	logger := zaptest.NewLogger(t)
	t.Run("success", func(t *testing.T) {
		data := model.Matrix{
			{
				Values: []model.SamplePair{
					{Timestamp: model.Time(from.UnixMilli()), Value: 1},
					{Timestamp: model.Time(to.UnixMilli()), Value: 2},
				},
			},
		}
		mock.EXPECT().QueryRange(
			ctx,
			`sum(increase(flipt_evaluations_requests_total{flipt_environment="default", flipt_namespace="bar", flipt_flag=~"foo|bar"}[1m])) by (flipt_flag) or vector(0)`,
			promapi.Range{
				Start: from,
				End:   to.Add(time.Minute),
				Step:  time.Minute,
			},
		).Return(data, promapi.Warnings{"test Warnings"}, nil, nil)

		client := &client{
			logger:     logger,
			promClient: mock,
		}

		results, err := client.GetBatchFlagEvaluationsCount(ctx, &panalytics.BatchFlagEvaluationsCountRequest{
			EnvironmentKey: "default",
			NamespaceKey:   "bar",
			FlagKeys:       []string{"foo", "bar"},
			From:           from,
			To:             to,
			StepMinutes:    1,
		})

		require.NoError(t, err)
		assert.NotNil(t, results)
		assert.Len(t, results, 2)
	})

	t.Run("success with multiple series", func(t *testing.T) {
		// Create a matrix with two series (for two different flags)
		data := model.Matrix{
			{
				Metric: model.Metric{
					"flipt_flag": "foo",
				},
				Values: []model.SamplePair{
					{Timestamp: model.Time(from.Add(1 * time.Minute).UnixMilli()), Value: 1},
					{Timestamp: model.Time(from.Add(2 * time.Minute).UnixMilli()), Value: 2},
				},
			},
			{
				Metric: model.Metric{
					"flipt_flag": "bar",
				},
				Values: []model.SamplePair{
					{Timestamp: model.Time(from.Add(1 * time.Minute).UnixMilli()), Value: 3},
					{Timestamp: model.Time(from.Add(2 * time.Minute).UnixMilli()), Value: 4},
				},
			},
		}
		mock.EXPECT().QueryRange(
			ctx,
			`sum(increase(flipt_evaluations_requests_total{flipt_environment="default", flipt_namespace="multi", flipt_flag=~"foo|bar"}[1m])) by (flipt_flag) or vector(0)`,
			promapi.Range{
				Start: from,
				End:   to.Add(time.Minute),
				Step:  time.Minute,
			},
		).Return(data, promapi.Warnings{"test Warnings"}, nil, nil)

		client := &client{
			logger:     logger,
			promClient: mock,
		}

		results, err := client.GetBatchFlagEvaluationsCount(ctx, &panalytics.BatchFlagEvaluationsCountRequest{
			EnvironmentKey: "default",
			NamespaceKey:   "multi",
			FlagKeys:       []string{"foo", "bar"},
			From:           from,
			To:             to,
			StepMinutes:    1,
		})

		require.NoError(t, err)
		assert.NotNil(t, results)
		assert.Len(t, results, 2)

		// Verify first flag results
		fooData, ok := results["foo"]
		assert.True(t, ok, "Data for 'foo' flag should exist")
		assert.Len(t, fooData.Timestamps, 2)
		assert.Len(t, fooData.Values, 2)
		assert.Equal(t, []float32{1, 2}, fooData.Values)

		// Verify second flag results
		barData, ok := results["bar"]
		assert.True(t, ok, "Data for 'bar' flag should exist")
		assert.Len(t, barData.Timestamps, 2)
		assert.Len(t, barData.Values, 2)
		assert.Equal(t, []float32{3, 4}, barData.Values)
	})

	t.Run("no data type", func(t *testing.T) {
		mock.EXPECT().QueryRange(
			ctx,
			`sum(increase(flipt_evaluations_requests_total{flipt_environment="default", flipt_namespace="no-data", flipt_flag=~"foo|bar"}[1m])) by (flipt_flag) or vector(0)`,
			promapi.Range{
				Start: from,
				End:   to.Add(time.Minute),
				Step:  time.Minute,
			},
		).Return(&model.Scalar{}, nil, nil, nil)

		client := &client{
			logger:     logger,
			promClient: mock,
		}

		_, err := client.GetBatchFlagEvaluationsCount(ctx, &panalytics.BatchFlagEvaluationsCountRequest{
			EnvironmentKey: "default",
			NamespaceKey:   "no-data",
			FlagKeys:       []string{"foo", "bar"},
			From:           from,
			To:             to,
			StepMinutes:    1,
		})

		require.Error(t, err)
	})

	t.Run("wrong data type", func(t *testing.T) {
		mock.EXPECT().QueryRange(
			ctx,
			`sum(increase(flipt_evaluations_requests_total{flipt_environment="default", flipt_namespace="wrong-data", flipt_flag=~"foo|bar"}[1m])) by (flipt_flag) or vector(0)`,
			promapi.Range{
				Start: from,
				End:   to.Add(time.Minute),
				Step:  time.Minute,
			},
		).Return(&model.Matrix{}, nil, nil, nil)

		client := &client{
			logger:     logger,
			promClient: mock,
		}

		_, err := client.GetBatchFlagEvaluationsCount(ctx, &panalytics.BatchFlagEvaluationsCountRequest{
			EnvironmentKey: "default",
			NamespaceKey:   "wrong-data",
			FlagKeys:       []string{"foo", "bar"},
			From:           from,
			To:             to,
			StepMinutes:    1,
		})

		require.Error(t, err)
	})

	t.Run("error", func(t *testing.T) {
		mock.EXPECT().QueryRange(
			ctx,
			`sum(increase(flipt_evaluations_requests_total{flipt_environment="default", flipt_namespace="error", flipt_flag=~"foo|bar"}[5m])) by (flipt_flag) or vector(0)`,
			promapi.Range{
				Start: from,
				End:   to.Add(5 * time.Minute),
				Step:  5 * time.Minute,
			},
		).Return(nil, nil, nil, errors.New("failed to query"))

		client := &client{
			logger:     logger,
			promClient: mock,
		}

		_, err := client.GetBatchFlagEvaluationsCount(ctx, &panalytics.BatchFlagEvaluationsCountRequest{
			EnvironmentKey: "default",
			NamespaceKey:   "error",
			FlagKeys:       []string{"foo", "bar"},
			From:           from,
			To:             to,
			StepMinutes:    5,
		})

		require.Error(t, err)
	})
}

// queryRecorder records the query passed to QueryRange.
type queryRecorder struct {
	query string
	data  model.Value
}

func (q *queryRecorder) QueryRange(_ context.Context, query string, _ promapi.Range, _ ...promapi.Option) (model.Value, promapi.Warnings, promapi.Infos, error) {
	q.query = query
	return q.data, nil, nil, nil
}

// parseSelector reads the label matchers of the first selector in query,
// decoding each value as a PromQL string literal, and returns them with the
// rest of the query after the closing brace.
func parseSelector(t *testing.T, query string) (map[string]string, string) {
	t.Helper()

	_, rest, ok := strings.Cut(query, "{")
	require.True(t, ok, "query has no selector: %s", query)

	matchers := map[string]string{}
	for {
		name, after, ok := strings.Cut(rest, "=")
		require.True(t, ok, "malformed matcher in: %s", rest)
		op := "="
		if strings.HasPrefix(after, "~") {
			op, after = "=~", after[1:]
		}

		literal, err := strconv.QuotedPrefix(after)
		require.NoError(t, err)
		value, err := strconv.Unquote(literal)
		require.NoError(t, err)
		matchers[strings.TrimSpace(name)+op] = value

		rest = after[len(literal):]
		if next, ok := strings.CutPrefix(rest, ", "); ok {
			rest = next
			continue
		}

		rest, ok = strings.CutPrefix(rest, "}")
		require.True(t, ok, "selector not closed after matchers: %s", rest)
		return matchers, rest
	}
}

func TestQueriesEscapeRequestKeys(t *testing.T) {
	// Each payload tries to close its label matcher and append a query over
	// any metric.
	const (
		envPayload  = `env"}`
		nsPayload   = `ns\`
		flagPayload = `x"}[1m])) or sum(rate(up{job=~".*`
	)

	from := time.Now().Add(-time.Hour).UTC()
	to := time.Now().UTC()
	data := model.Matrix{{Values: []model.SamplePair{{Timestamp: model.Time(from.UnixMilli()), Value: 1}}}}

	t.Run("single flag", func(t *testing.T) {
		recorder := &queryRecorder{data: data}
		c := &client{logger: zaptest.NewLogger(t), promClient: recorder}
		_, _, err := c.GetFlagEvaluationsCount(t.Context(), &panalytics.FlagEvaluationsCountRequest{
			EnvironmentKey: envPayload,
			NamespaceKey:   nsPayload,
			FlagKey:        flagPayload,
			From:           from,
			To:             to,
			StepMinutes:    1,
		})
		require.NoError(t, err)

		matchers, rest := parseSelector(t, recorder.query)
		assert.Equal(t, map[string]string{
			"flipt_environment=": envPayload,
			"flipt_namespace=":   nsPayload,
			"flipt_flag=":        flagPayload,
		}, matchers)
		assert.Equal(t, "[1m])) or vector(0)", rest)
	})

	t.Run("batch flags", func(t *testing.T) {
		recorder := &queryRecorder{data: model.Matrix{
			{Values: []model.SamplePair{{Timestamp: model.Time(from.UnixMilli()), Value: 0}}},
			{
				Metric: model.Metric{"flipt_flag": "a.b"},
				Values: []model.SamplePair{{Timestamp: model.Time(from.UnixMilli()), Value: 7}},
			},
		}}
		c := &client{logger: zaptest.NewLogger(t), promClient: recorder}
		result, err := c.GetBatchFlagEvaluationsCount(t.Context(), &panalytics.BatchFlagEvaluationsCountRequest{
			EnvironmentKey: envPayload,
			NamespaceKey:   nsPayload,
			FlagKeys:       []string{"a.b", flagPayload},
			From:           from,
			To:             to,
			StepMinutes:    1,
		})
		require.NoError(t, err)

		matchers, rest := parseSelector(t, recorder.query)
		assert.Equal(t, envPayload, matchers["flipt_environment="])
		assert.Equal(t, nsPayload, matchers["flipt_namespace="])
		assert.Equal(t, "[1m])) by (flipt_flag) or vector(0)", rest)

		// Prometheus regex matchers are fully anchored RE2, so the pattern
		// must match exactly the requested keys and nothing else.
		pattern := regexp.MustCompile("^(?:" + matchers["flipt_flag=~"] + ")$")
		assert.True(t, pattern.MatchString("a.b"))
		assert.True(t, pattern.MatchString(flagPayload))
		assert.False(t, pattern.MatchString("axb"), "regex metacharacters must be escaped")
		assert.False(t, pattern.MatchString("other"))

		// series are mapped back to the raw, unescaped keys
		assert.Equal(t, []float32{7}, result["a.b"].Values)
		assert.Equal(t, []float32{0}, result[flagPayload].Values)
	})

	t.Run("batch without flags", func(t *testing.T) {
		recorder := &queryRecorder{data: data}
		c := &client{logger: zaptest.NewLogger(t), promClient: recorder}
		result, err := c.GetBatchFlagEvaluationsCount(t.Context(), &panalytics.BatchFlagEvaluationsCountRequest{
			EnvironmentKey: "default",
			NamespaceKey:   "default",
			From:           from,
			To:             to,
			StepMinutes:    1,
		})
		require.NoError(t, err)
		assert.Empty(t, result)
		assert.Empty(t, recorder.query, "no query is sent without flags")
	})
}
