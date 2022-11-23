package metrics

import (
	"log"

	"go.opentelemetry.io/otel/exporters/prometheus"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/metric/global"
	"go.opentelemetry.io/otel/metric/instrument"
	"go.opentelemetry.io/otel/metric/instrument/asyncint64"
	"go.opentelemetry.io/otel/metric/instrument/syncint64"
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

// MustSyncInt64 returns a syncint64 instrument provider based on the global Meter.
// The returns provider panics instead of returning an error when it cannot build
// a required counter, upDownCounter or histogram.
func MustSyncInt64() MustSyncInt64InstrumentProvider {
	return mustSyncInt64InstrumentProvider{}
}

// MustSyncInt64InstrumentProvider is a syncint64.InstrumentProvider which panics
// if it cannot successfully build the requestd counter, upDownCounter or histogram.
type MustSyncInt64InstrumentProvider interface {
	// Counter creates an instrument for recording increasing values.
	Counter(name string, opts ...instrument.Option) syncint64.Counter
	// UpDownCounter creates an instrument for recording changes of a value.
	UpDownCounter(name string, opts ...instrument.Option) syncint64.UpDownCounter
	// Histogram creates an instrument for recording a distribution of values.
	Histogram(name string, opts ...instrument.Option) syncint64.Histogram
}

type mustSyncInt64InstrumentProvider struct{}

// Counter creates an instrument for recording increasing values.
func (m mustSyncInt64InstrumentProvider) Counter(name string, opts ...instrument.Option) syncint64.Counter {
	counter, err := Meter.SyncInt64().Counter(name, opts...)
	if err != nil {
		panic(err)
	}

	return counter
}

// UpDownCounter creates an instrument for recording changes of a value.
func (m mustSyncInt64InstrumentProvider) UpDownCounter(name string, opts ...instrument.Option) syncint64.UpDownCounter {
	counter, err := Meter.SyncInt64().UpDownCounter(name, opts...)
	if err != nil {
		panic(err)
	}

	return counter
}

// Histogram creates an instrument for recording a distribution of values.
func (m mustSyncInt64InstrumentProvider) Histogram(name string, opts ...instrument.Option) syncint64.Histogram {
	hist, err := Meter.SyncInt64().Histogram(name, opts...)
	if err != nil {
		panic(err)
	}

	return hist
}

// MustAsyncInt64 returns a syncint64 instrument provider based on the global Meter.
// The returns provider panics instead of returning an error when it cannot build
// a required counter, upDownCounter or histogram.
func MustAsyncInt64() MustAsyncInt64InstrumentProvider {
	return mustAsyncInt64InstrumentProvider{}
}

// MustAsyncInt64InstrumentProvider is a syncint64.InstrumentProvider which panics
// if it cannot successfully build the requestd counter, upDownCounter or histogram.
type MustAsyncInt64InstrumentProvider interface {
	// Counter creates an instrument for recording increasing values.
	Counter(name string, opts ...instrument.Option) asyncint64.Counter
	// UpDownCounter creates an instrument for recording changes of a value.
	UpDownCounter(name string, opts ...instrument.Option) asyncint64.UpDownCounter
	// Histogram creates an instrument for recording a distribution of values.
	Gauge(name string, opts ...instrument.Option) asyncint64.Gauge
}

type mustAsyncInt64InstrumentProvider struct{}

// Counter creates an instrument for recording increasing values.
func (m mustAsyncInt64InstrumentProvider) Counter(name string, opts ...instrument.Option) asyncint64.Counter {
	counter, err := Meter.AsyncInt64().Counter(name, opts...)
	if err != nil {
		panic(err)
	}

	return counter
}

// UpDownCounter creates an instrument for recording changes of a value.
func (m mustAsyncInt64InstrumentProvider) UpDownCounter(name string, opts ...instrument.Option) asyncint64.UpDownCounter {
	counter, err := Meter.AsyncInt64().UpDownCounter(name, opts...)
	if err != nil {
		panic(err)
	}

	return counter
}

// Gauge creates an instrument for recording a distribution of values.
func (m mustAsyncInt64InstrumentProvider) Gauge(name string, opts ...instrument.Option) asyncint64.Gauge {
	gauge, err := Meter.AsyncInt64().Gauge(name, opts...)
	if err != nil {
		panic(err)
	}

	return gauge
}
