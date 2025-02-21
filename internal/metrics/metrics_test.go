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
			name: "otlp http exporter",
			cfg: &config.MetricsConfig{
				Exporter: config.MetricsOTLP,
				OTLP: config.OTLPMetricsConfig{
					Endpoint: "http://localhost:4317",
					Headers:  map[string]string{"key": "value"},
				},
			},
		},
		{
			name: "otlp https exporter",
			cfg: &config.MetricsConfig{
				Exporter: config.MetricsOTLP,
				OTLP: config.OTLPMetricsConfig{
					Endpoint: "https://localhost:4317",
					Headers:  map[string]string{"key": "value"},
				},
			},
		},
		{
			name: "otlp grpc exporter",
			cfg: &config.MetricsConfig{
				Exporter: config.MetricsOTLP,
				OTLP: config.OTLPMetricsConfig{
					Endpoint: "grpc://localhost:4317",
					Headers:  map[string]string{"key": "value"},
				},
			},
		},
		{
			name: "otlp default grpc exporter",
			cfg: &config.MetricsConfig{
				Exporter: config.MetricsOTLP,
				OTLP: config.OTLPMetricsConfig{
					Endpoint: "localhost:4317",
					Headers:  map[string]string{"key": "value"},
				},
			},
		},
		{
			name: "invalid otlp endpoint",
			cfg: &config.MetricsConfig{
				Exporter: config.MetricsOTLP,
				OTLP: config.OTLPMetricsConfig{
					Endpoint: "://invalid",
				},
			},
			wantErr: errors.New("parsing otlp endpoint: parse \"://invalid\": missing protocol scheme"),
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

func TestGetResources(t *testing.T) {
	ctx := context.Background()

	resources, err := GetResources(ctx)
	require.NoError(t, err)
	assert.NotNil(t, resources)

	// Verify that the resource has the required attributes
	attrs := resources.Attributes()
	assert.NotEmpty(t, attrs)

	// Check for required service name
	var hasServiceName bool
	for _, attr := range attrs {
		if attr.Key == "service.name" {
			hasServiceName = true
			assert.Equal(t, "flipt", attr.Value.AsString())
			break
		}
	}
	assert.True(t, hasServiceName, "service.name attribute not found")
}
