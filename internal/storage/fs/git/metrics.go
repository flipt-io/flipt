package git

import (
	"context"
	"sync/atomic"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"go.flipt.io/flipt/internal/metrics"
	"go.opentelemetry.io/otel/metric"
)

const (
	namespace = "flipt"
	subsystem = "git_sync"
)

var (
	// syncLatency is a histogram for git sync operation duration.
	syncLatency = metrics.MustFloat64().
			Histogram(
			prometheus.BuildFQName(namespace, subsystem, "latency"),
			metric.WithDescription("The duration of git sync operations in seconds"),
			metric.WithUnit("s"),
		)

	// syncFlags is a counter for the number of flags during sync.
	syncFlags = metrics.MustInt64().
			Counter(
			prometheus.BuildFQName(namespace, subsystem, "flags"),
			metric.WithDescription("The number of flags fetched during git sync"),
		)

	// syncErrors is a counter for failed git sync operations.
	syncErrors = metrics.MustInt64().
			Counter(
			prometheus.BuildFQName(namespace, subsystem, "errors"),
			metric.WithDescription("The number of errors git sync operations"),
		)

	_ = metrics.MustInt64().
		ObservableGauge(
			prometheus.BuildFQName(namespace, subsystem, "last_time"),
			getLastSyncTime,
			metric.WithDescription("The unix timestamp of the last git sync operation"),
		)

	// internal storage for last sync time value
	lastSyncTimeValue atomic.Int64
)

// observeSync records a complete git sync operation with all relevant metrics.
func observeSync(ctx context.Context, duration time.Duration, flagsFetched int64, success bool) {
	// Always record duration and update last sync time
	syncLatency.Record(ctx, duration.Seconds())
	setLastSyncTime(time.Now().UTC())

	syncFlags.Add(ctx, flagsFetched)
	if !success {
		syncErrors.Add(ctx, 1)
	}
}

// setLastSyncTime updates the last sync time value that will be reported by the observable gauge.
func setLastSyncTime(ts time.Time) {
	lastSyncTimeValue.Store(ts.Unix())
}

// getLastSyncTime returns the current last sync time value.
func getLastSyncTime() int64 {
	return lastSyncTimeValue.Load()
}
