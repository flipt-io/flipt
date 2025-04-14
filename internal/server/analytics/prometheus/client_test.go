package prometheus

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
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
	_, _, err = s.GetFlagEvaluationsCount(context.Background(), &panalytics.FlagEvaluationsCountRequest{})
	require.NoError(t, err)

	cfg.Analytics.Storage.Prometheus.URL = "\t"
	_, err = New(logger, cfg)
	require.Error(t, err)
}

func TestGetFlagEvaluationsCount(t *testing.T) {
	ctx := context.Background()
	from := time.Now().Add(-time.Hour).UTC()
	to := time.Now().UTC()
	mock := newMockPrometheusClient(t)

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
		).Return(data, promapi.Warnings{"test Warnings"}, nil)

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
		).Return(data, promapi.Warnings{"test Warnings"}, nil)

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
		).Return(&model.Scalar{}, nil, nil)

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
		).Return(&model.Matrix{}, nil, nil)

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
		).Return(nil, nil, errors.New("failed to query"))

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
