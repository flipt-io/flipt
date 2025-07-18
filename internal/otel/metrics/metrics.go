package metrics

import (
	"context"
	"fmt"
	"sync"

	"go.flipt.io/flipt/internal/config"
	"go.opentelemetry.io/contrib/exporters/autoexport"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/prometheus"
	"go.opentelemetry.io/otel/metric"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
)

// This is memoized in the OTEL library to avoid creating multiple instances of the same exporter.
func meter() metric.Meter {
	return otel.Meter("go.flipt.io/flipt/v2")
}

// MustInt64 returns an instrument provider based on the global Meter.
// The returns provider panics instead of returning an error when it cannot build
// a required counter, upDownCounter or histogram.
func MustInt64() MustInt64Meter {
	return mustInt64Meter{}
}

// MustInt64Meter is a meter/Meter which panics if it cannot successfully build the
// requestd counter, upDownCounter or histogram.
type MustInt64Meter interface {
	// Counter returns a new instrument identified by name and configured
	// with options. The instrument is used to synchronously record increasing
	// int64 measurements during a computational operation.
	Counter(name string, options ...metric.Int64CounterOption) metric.Int64Counter
	// UpDownCounter returns a new instrument identified by name and
	// configured with options. The instrument is used to synchronously record
	// int64 measurements during a computational operation.
	UpDownCounter(name string, options ...metric.Int64UpDownCounterOption) metric.Int64UpDownCounter
	// Histogram returns a new instrument identified by name and
	// configured with options. The instrument is used to synchronously record
	// the distribution of int64 measurements during a computational operation.
	Histogram(name string, options ...metric.Int64HistogramOption) metric.Int64Histogram
}

type mustInt64Meter struct{}

// Counter creates an instrument for recording increasing values.
func (m mustInt64Meter) Counter(name string, opts ...metric.Int64CounterOption) metric.Int64Counter {
	counter, err := meter().Int64Counter(name, opts...)
	if err != nil {
		panic(err)
	}

	return counter
}

// UpDownCounter creates an instrument for recording changes of a value.
func (m mustInt64Meter) UpDownCounter(name string, opts ...metric.Int64UpDownCounterOption) metric.Int64UpDownCounter {
	counter, err := meter().Int64UpDownCounter(name, opts...)
	if err != nil {
		panic(err)
	}

	return counter
}

// Histogram creates an instrument for recording a distribution of values.
func (m mustInt64Meter) Histogram(name string, opts ...metric.Int64HistogramOption) metric.Int64Histogram {
	hist, err := meter().Int64Histogram(name, opts...)
	if err != nil {
		panic(err)
	}

	return hist
}

// MustFloat64 returns an instrument provider based on the global Meter.
// The returns provider panics instead of returning an error when it cannot build
// a required counter, upDownCounter or histogram.
func MustFloat64() MustFloat64Meter {
	return mustFloat64Meter{}
}

// MustFloat64Meter is a meter/Meter which panics if it cannot successfully build the
// requestd counter, upDownCounter or histogram.
type MustFloat64Meter interface {
	// Counter returns a new instrument identified by name and configured
	// with options. The instrument is used to synchronously record increasing
	// float64 measurements during a computational operation.
	Counter(name string, options ...metric.Float64CounterOption) metric.Float64Counter
	// UpDownCounter returns a new instrument identified by name and
	// configured with options. The instrument is used to synchronously record
	// float64 measurements during a computational operation.
	UpDownCounter(name string, options ...metric.Float64UpDownCounterOption) metric.Float64UpDownCounter
	// Histogram returns a new instrument identified by name and
	// configured with options. The instrument is used to synchronously record
	// the distribution of float64 measurements during a computational operation.
	Histogram(name string, options ...metric.Float64HistogramOption) metric.Float64Histogram
}

type mustFloat64Meter struct{}

// Counter creates an instrument for recording increasing values.
func (m mustFloat64Meter) Counter(name string, opts ...metric.Float64CounterOption) metric.Float64Counter {
	counter, err := meter().Float64Counter(name, opts...)
	if err != nil {
		panic(err)
	}

	return counter
}

// UpDownCounter creates an instrument for recording changes of a value.
func (m mustFloat64Meter) UpDownCounter(name string, opts ...metric.Float64UpDownCounterOption) metric.Float64UpDownCounter {
	counter, err := meter().Float64UpDownCounter(name, opts...)
	if err != nil {
		panic(err)
	}

	return counter
}

// Histogram creates an instrument for recording a distribution of values.
func (m mustFloat64Meter) Histogram(name string, opts ...metric.Float64HistogramOption) metric.Float64Histogram {
	hist, err := meter().Float64Histogram(name, opts...)
	if err != nil {
		panic(err)
	}

	return hist
}

var (
	metricExpOnce sync.Once
	metricExp     sdkmetric.Reader
	metricExpFunc func(context.Context) error = func(context.Context) error { return nil }
	metricExpErr  error
)

func GetExporter(ctx context.Context, cfg *config.MetricsConfig) (sdkmetric.Reader, func(context.Context) error, error) {
	metricExpOnce.Do(func() {
		switch cfg.Exporter {
		case config.MetricsPrometheus:
			// exporter registers itself on the prom client DefaultRegistrar
			metricExp, metricExpErr = prometheus.New()
			if metricExpErr != nil {
				return
			}

		case config.MetricsOTLP:
			metricExp, metricExpErr = autoexport.NewMetricReader(ctx)
			if metricExpErr != nil {
				return
			}

			metricExpFunc = func(ctx context.Context) error {
				return metricExp.Shutdown(ctx)
			}
		default:
			metricExpErr = fmt.Errorf("unsupported metrics exporter: %s", cfg.Exporter)
			return
		}
	})

	return metricExp, metricExpFunc, metricExpErr
}
