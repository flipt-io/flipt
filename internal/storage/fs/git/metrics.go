package git

import (
	"context"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"go.flipt.io/flipt/internal/metrics"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
)

const (
	namespace = "flipt"
	subsystem = "git_sync"

	nanosPerSecond = 1e9
)

var (
	// Duration is a histogram for git sync operation duration.
	Duration = metrics.MustFloat64().
			Histogram(
			prometheus.BuildFQName(namespace, subsystem, "duration_seconds"),
			metric.WithDescription("The duration of git sync operations in seconds"),
			metric.WithUnit("s"),
		)

	// FlagsFetched is a counter for the number of flags fetched during sync.
	FlagsFetched = metrics.MustInt64().
			Counter(
			prometheus.BuildFQName(namespace, subsystem, "flags_fetched"),
			metric.WithDescription("The number of flags fetched during git sync"),
		)

	// Success is a counter for successful git sync operations.
	Success = metrics.MustInt64().
		Counter(
			prometheus.BuildFQName(namespace, subsystem, "success"),
			metric.WithDescription("The number of successful git sync operations"),
		)

	// Failure is a counter for failed git sync operations.
	Failure = metrics.MustInt64().
		Counter(
			prometheus.BuildFQName(namespace, subsystem, "error"),
			metric.WithDescription("The number of failed git sync operations"),
		)

	// LastTime will be initialized when InitMetrics is called
	LastTime metric.Int64ObservableGauge

	// internal storage for last sync time value
	lastSyncTimeValue int64
	lastSyncTimeMu    sync.RWMutex

	AttributeSyncType = attribute.Key("sync_type")
)

func init() {
	// Create ObservableGauge for last sync time and register callback
	LastTime = metrics.MustInt64ObservableGauge(
		prometheus.BuildFQName(namespace, subsystem, "last_time_unix"),
		metric.WithDescription("The unix timestamp of the last git sync operation"),
	)

	metrics.MustRegisterCallback(
		func(ctx context.Context, observer metric.Observer) error {
			lastSyncTimeMu.RLock()
			value := lastSyncTimeValue
			lastSyncTimeMu.RUnlock()
			observer.ObserveInt64(LastTime, value/nanosPerSecond)
			return nil
		},
		LastTime,
	)
}

func ObserveSuccess(ctx context.Context, typ string) {
	Success.Add(ctx, 1, metric.WithAttributeSet(
		attribute.NewSet(AttributeSyncType.String(typ)),
	))
}

// ObserveFailure records a failed git sync operation with the specified type.
func ObserveFailure(ctx context.Context, typ string) {
	Failure.Add(ctx, 1, metric.WithAttributeSet(
		attribute.NewSet(AttributeSyncType.String(typ)),
	))
}

// ObserveFlagsFetched records the number of flags fetched during sync.
func ObserveFlagsFetched(ctx context.Context, count int64, typ string) {
	FlagsFetched.Add(ctx, count, metric.WithAttributeSet(
		attribute.NewSet(AttributeSyncType.String(typ)),
	))
}

// ObserveDuration records the duration of a git sync operation.
func ObserveDuration(ctx context.Context, duration float64, typ string) {
	Duration.Record(ctx, duration, metric.WithAttributeSet(
		attribute.NewSet(AttributeSyncType.String(typ)),
	))
}

// ObserveSync records a complete git sync operation with all relevant metrics.
func ObserveSync(ctx context.Context, duration float64, flagsFetched int64, success bool, syncType string) {
	// Always record duration and update last sync time
	ObserveDuration(ctx, duration, syncType)
	setLastSyncTime(time.Now().UnixNano())

	if flagsFetched > 0 {
		ObserveFlagsFetched(ctx, flagsFetched, syncType)
	}

	if success {
		ObserveSuccess(ctx, syncType)
	} else {
		ObserveFailure(ctx, syncType)
	}
}

// setLastSyncTime updates the last sync time value that will be reported by the observable gauge.
func setLastSyncTime(ts int64) {
	lastSyncTimeMu.Lock()
	defer lastSyncTimeMu.Unlock()
	lastSyncTimeValue = ts
}

// GetLastSyncTime returns the current last sync time value.
func GetLastSyncTime() int64 {
	lastSyncTimeMu.RLock()
	defer lastSyncTimeMu.RUnlock()
	return lastSyncTimeValue
}
