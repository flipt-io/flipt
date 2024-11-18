package prometheus

import (
	"context"
	"errors"
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
	logger := zaptest.NewLogger(t)
	cfg := config.Default()
	cfg.Analytics.Storage.Prometheus.URL = "http://prometheus:9090"
	s, err := New(logger, cfg)
	require.NoError(t, err)
	assert.Equal(t, "prometheus", s.String())

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
			`sum(increase(flipt_evaluations_requests_total{namespace="bar", flag="foo"}[1m]))`,
			promapi.Range{
				Start: from,
				End:   to,
				Step:  time.Minute,
			},
		).Return(data, promapi.Warnings{"test Warnings"}, nil)

		client := &client{
			logger:     logger,
			promClient: mock,
		}

		labels, values, err := client.GetFlagEvaluationsCount(ctx, &panalytics.FlagEvaluationsCountRequest{
			NamespaceKey: "bar",
			FlagKey:      "foo",
			From:         from,
			To:           to,
			StepMinutes:  1,
		})

		require.NoError(t, err)
		assert.Len(t, labels, 2)
		assert.Len(t, values, 2)
	})

	t.Run("no data type", func(t *testing.T) {
		mock.EXPECT().QueryRange(
			ctx,
			`sum(increase(flipt_evaluations_requests_total{namespace="no-data", flag="foo"}[1m]))`,
			promapi.Range{
				Start: from,
				End:   to,
				Step:  time.Minute,
			},
		).Return(&model.Scalar{}, nil, nil)

		client := &client{
			logger:     logger,
			promClient: mock,
		}

		_, _, err := client.GetFlagEvaluationsCount(ctx, &panalytics.FlagEvaluationsCountRequest{
			NamespaceKey: "no-data",
			FlagKey:      "foo",
			From:         from,
			To:           to,
			StepMinutes:  1,
		})

		require.Error(t, err)
	})

	t.Run("wrong data type", func(t *testing.T) {
		mock.EXPECT().QueryRange(
			ctx,
			`sum(increase(flipt_evaluations_requests_total{namespace="wrong-data", flag="foo"}[1m]))`,
			promapi.Range{
				Start: from,
				End:   to,
				Step:  time.Minute,
			},
		).Return(&model.Matrix{}, nil, nil)

		client := &client{
			logger:     logger,
			promClient: mock,
		}

		_, _, err := client.GetFlagEvaluationsCount(ctx, &panalytics.FlagEvaluationsCountRequest{
			NamespaceKey: "wrong-data",
			FlagKey:      "foo",
			From:         from,
			To:           to,
			StepMinutes:  1,
		})

		require.Error(t, err)
	})

	t.Run("error", func(t *testing.T) {
		mock.EXPECT().QueryRange(
			ctx,
			`sum(increase(flipt_evaluations_requests_total{namespace="error", flag="foo"}[1m]))`,
			promapi.Range{
				Start: from,
				End:   to,
				Step:  time.Minute,
			},
		).Return(nil, nil, errors.New("failed to query"))

		client := &client{
			logger:     logger,
			promClient: mock,
		}

		_, _, err := client.GetFlagEvaluationsCount(ctx, &panalytics.FlagEvaluationsCountRequest{
			NamespaceKey: "error",
			FlagKey:      "foo",
			From:         from,
			To:           to,
			StepMinutes:  1,
		})

		require.Error(t, err)
	})
}
