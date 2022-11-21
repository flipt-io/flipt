package cache

import (
	"context"

	"github.com/prometheus/client_golang/prometheus"
	"go.flipt.io/flipt/internal/metrics"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric/instrument"
	"go.opentelemetry.io/otel/metric/instrument/syncint64"
)

const (
	namespace = "flipt"
	subsystem = "cache"
)

var (
	// Hit is a counter for cache hits.
	Hit = metrics.MustSyncInt64().
		Counter(
			prometheus.BuildFQName(namespace, subsystem, "hit"),
			instrument.WithDescription("The number of cache hits"),
		)
		// Miss is a counter for cache misses.
	Miss = metrics.MustSyncInt64().
		Counter(
			prometheus.BuildFQName(namespace, subsystem, "miss"),
			instrument.WithDescription("The number of cache misses"),
		)
		// Error is a counter for cache errors.
	Error = metrics.MustSyncInt64().
		Counter(
			prometheus.BuildFQName(namespace, subsystem, "error"),
			instrument.WithDescription("The number of times an error occurred reading or writing to the cache"),
		)
)

// Observe adds one to the provided counter and records the
// cache type attribute supplied by typ.
func Observe(ctx context.Context, typ string, counter syncint64.Counter) {
	counter.Add(ctx, 1, attribute.Key("cache").String(typ))
}
