package metrics

import (
	"context"
	"errors"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.flipt.io/flipt/internal/config"
	"go.opentelemetry.io/otel/metric"
)

func TestGetExporter(t *testing.T) {
	tests := []struct {
		name    string
		cfg     *config.MetricsConfig
		wantErr error
	}{
		{
			name: "prometheus exporter",
			cfg: &config.MetricsConfig{
				Exporter: config.MetricsPrometheus,
			},
		},
		{
			name: "otlp exporter",
			cfg: &config.MetricsConfig{
				Exporter: config.MetricsOTLP,
			},
		},
		{
			name: "unsupported exporter",
			cfg: &config.MetricsConfig{
				Exporter: "unsupported",
			},
			wantErr: errors.New("unsupported metrics exporter: unsupported"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset the singleton state for each test
			metricExpOnce = sync.Once{}
			metricExp = nil
			metricExpFunc = func(context.Context) error { return nil }
			metricExpErr = nil

			exp, expFunc, err := GetExporter(context.Background(), tt.cfg)
			if tt.wantErr != nil {
				require.EqualError(t, err, tt.wantErr.Error())
				assert.Nil(t, exp, "expected nil exporter when error occurs")
				return
			}

			require.NoError(t, err)
			assert.NotNil(t, exp)
			assert.NotNil(t, expFunc)

			t.Cleanup(func() {
				err := expFunc(context.Background())
				assert.NoError(t, err)
			})
		})
	}
}

func TestMustInt64Meter(t *testing.T) {
	m := MustInt64()

	t.Run("counter", func(t *testing.T) {
		counter := m.Counter("test_counter",
			metric.WithDescription("A test counter"),
			metric.WithUnit("1"),
		)
		assert.NotNil(t, counter)
	})

	t.Run("up_down_counter", func(t *testing.T) {
		counter := m.UpDownCounter("test_up_down_counter",
			metric.WithDescription("A test up/down counter"),
			metric.WithUnit("1"),
		)
		assert.NotNil(t, counter)
	})

	t.Run("histogram", func(t *testing.T) {
		histogram := m.Histogram("test_histogram",
			metric.WithDescription("A test histogram"),
			metric.WithUnit("ms"),
		)
		assert.NotNil(t, histogram)
	})
}

func TestMustFloat64Meter(t *testing.T) {
	m := MustFloat64()

	t.Run("counter", func(t *testing.T) {
		counter := m.Counter("test_counter",
			metric.WithDescription("A test counter"),
			metric.WithUnit("1"),
		)
		assert.NotNil(t, counter)
	})

	t.Run("up_down_counter", func(t *testing.T) {
		counter := m.UpDownCounter("test_up_down_counter",
			metric.WithDescription("A test up/down counter"),
			metric.WithUnit("1"),
		)
		assert.NotNil(t, counter)
	})

	t.Run("histogram", func(t *testing.T) {
		histogram := m.Histogram("test_histogram",
			metric.WithDescription("A test histogram"),
			metric.WithUnit("ms"),
		)
		assert.NotNil(t, histogram)
	})
}
