package metrics

import (
	"log"

	"go.opentelemetry.io/otel/exporters/prometheus"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/metric/global"
	"go.opentelemetry.io/otel/metric/instrument"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
)

// Meter is the default Flipt-wide otel metric Meter.
var Meter metric.Meter

func init() {
	// exporter registers itself on the prom client DefaultRegistrar
	exporter, err := prometheus.New()
	if err != nil {
		log.Fatal(err)
	}

	provider := sdkmetric.NewMeterProvider(sdkmetric.WithReader(exporter))
	global.SetMeterProvider(provider)

	Meter = provider.Meter("github.com/flipt-io/flipt")
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
	// ounter returns a new instrument identified by name and configured
	// with options. The instrument is used to synchronously record increasing
	// int64 measurements during a computational operation.
	Counter(name string, options ...instrument.Int64Option) instrument.Int64Counter
	// UpDownCounter returns a new instrument identified by name and
	// configured with options. The instrument is used to synchronously record
	// int64 measurements during a computational operation.
	UpDownCounter(name string, options ...instrument.Int64Option) instrument.Int64UpDownCounter
	// Histogram returns a new instrument identified by name and
	// configured with options. The instrument is used to synchronously record
	// the distribution of int64 measurements during a computational operation.
	Histogram(name string, options ...instrument.Int64Option) instrument.Int64Histogram
}

type mustInt64Meter struct{}

// Counter creates an instrument for recording increasing values.
func (m mustInt64Meter) Counter(name string, opts ...instrument.Int64Option) instrument.Int64Counter {
	counter, err := Meter.Int64Counter(name, opts...)
	if err != nil {
		panic(err)
	}

	return counter
}

// UpDownCounter creates an instrument for recording changes of a value.
func (m mustInt64Meter) UpDownCounter(name string, opts ...instrument.Int64Option) instrument.Int64UpDownCounter {
	counter, err := Meter.Int64UpDownCounter(name, opts...)
	if err != nil {
		panic(err)
	}

	return counter
}

// Histogram creates an instrument for recording a distribution of values.
func (m mustInt64Meter) Histogram(name string, opts ...instrument.Int64Option) instrument.Int64Histogram {
	hist, err := Meter.Int64Histogram(name, opts...)
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
	// ounter returns a new instrument identified by name and configured
	// with options. The instrument is used to synchronously record increasing
	// float64 measurements during a computational operation.
	Counter(name string, options ...instrument.Float64Option) instrument.Float64Counter
	// UpDownCounter returns a new instrument identified by name and
	// configured with options. The instrument is used to synchronously record
	// float64 measurements during a computational operation.
	UpDownCounter(name string, options ...instrument.Float64Option) instrument.Float64UpDownCounter
	// Histogram returns a new instrument identified by name and
	// configured with options. The instrument is used to synchronously record
	// the distribution of float64 measurements during a computational operation.
	Histogram(name string, options ...instrument.Float64Option) instrument.Float64Histogram
}

type mustFloat64Meter struct{}

// Counter creates an instrument for recording increasing values.
func (m mustFloat64Meter) Counter(name string, opts ...instrument.Float64Option) instrument.Float64Counter {
	counter, err := Meter.Float64Counter(name, opts...)
	if err != nil {
		panic(err)
	}

	return counter
}

// UpDownCounter creates an instrument for recording changes of a value.
func (m mustFloat64Meter) UpDownCounter(name string, opts ...instrument.Float64Option) instrument.Float64UpDownCounter {
	counter, err := Meter.Float64UpDownCounter(name, opts...)
	if err != nil {
		panic(err)
	}

	return counter
}

// Histogram creates an instrument for recording a distribution of values.
func (m mustFloat64Meter) Histogram(name string, opts ...instrument.Float64Option) instrument.Float64Histogram {
	hist, err := Meter.Float64Histogram(name, opts...)
	if err != nil {
		panic(err)
	}

	return hist
}
