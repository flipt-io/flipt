package git

import (
	"context"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"go.flipt.io/flipt/internal/containers"
	"go.flipt.io/flipt/internal/otel/metrics"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
)

const (
	namespace = "flipt"
	subsystem = "git"
)

var (
	attrRemote    = attribute.Key("remote")
	attrBranch    = attribute.Key("branch")
	attrStatus    = attribute.Key("status")
	attrErrorType = attribute.Key("error_type")
	attrOperation = attribute.Key("operation")
)

type repoMetrics struct {
	attrs []attribute.KeyValue
	set   attribute.Set

	viewsTotal      metric.Int64Counter
	viewErrorsTotal metric.Int64Counter
	viewLatency     metric.Float64Histogram

	updatesTotal      metric.Int64Counter
	updateErrorsTotal metric.Int64Counter
	updateLatency     metric.Float64Histogram

	pollErrorsTotal       metric.Int64Counter
	updateSubsErrorsTotal metric.Int64Counter

	branchesTotal       metric.Int64UpDownCounter
	branchesErrorsTotal metric.Int64Counter

	// Sync operation metrics
	syncTotal       metric.Int64Counter
	syncErrorsTotal metric.Int64Counter
	syncDuration    metric.Float64Histogram

	// Change tracking metrics
	filesChanged metric.Int64Counter
}

func withRemote(name string) containers.Option[repoMetrics] {
	return func(rm *repoMetrics) {
		rm.attrs = append(rm.attrs, attrRemote.String(name))
	}
}

func newRepoMetrics(opts ...containers.Option[repoMetrics]) repoMetrics {
	m := repoMetrics{
		viewsTotal: metrics.MustInt64().Counter(
			prometheus.BuildFQName(namespace, subsystem, "views_total"),
			metric.WithDescription("The total number of attempted repository reads"),
		),
		viewErrorsTotal: metrics.MustInt64().Counter(
			prometheus.BuildFQName(namespace, subsystem, "view_errors_total"),
			metric.WithDescription("The total number of errors reading from repository"),
		),
		viewLatency: metrics.MustFloat64().Histogram(
			prometheus.BuildFQName(namespace, subsystem, "view_latency"),
			metric.WithDescription("The latency of repository reads in milliseconds"),
			metric.WithUnit("ms"),
		),
		updatesTotal: metrics.MustInt64().Counter(
			prometheus.BuildFQName(namespace, subsystem, "updates_total"),
			metric.WithDescription("The total number of attempted repository writes"),
		),
		updateErrorsTotal: metrics.MustInt64().Counter(
			prometheus.BuildFQName(namespace, subsystem, "update_errors_total"),
			metric.WithDescription("The total number of errors writing (and pushing) to the repository"),
		),
		updateLatency: metrics.MustFloat64().Histogram(
			prometheus.BuildFQName(namespace, subsystem, "update_latency"),
			metric.WithDescription("The latency of repository writes (and pushes) in milliseconds"),
			metric.WithUnit("ms"),
		),
		pollErrorsTotal: metrics.MustInt64().Counter(
			prometheus.BuildFQName(namespace, subsystem, "poll_errors_total"),
			metric.WithDescription("The total number of errors observed during polling"),
		),
		updateSubsErrorsTotal: metrics.MustInt64().Counter(
			prometheus.BuildFQName(namespace, subsystem, "update_subscribers_errors_total"),
			metric.WithDescription("The total number of errors observed updating repository subscribers (e.g. snapshot builds)"),
		),
		branchesTotal: metrics.MustInt64().UpDownCounter(
			prometheus.BuildFQName(namespace, subsystem, "branches_total"),
			metric.WithDescription("The total number of branches"),
		),
		branchesErrorsTotal: metrics.MustInt64().Counter(
			prometheus.BuildFQName(namespace, subsystem, "branches_errors_total"),
			metric.WithDescription("The total number of errors observed creating or deleting branches"),
		),
		syncTotal: metrics.MustInt64().Counter(
			prometheus.BuildFQName(namespace, subsystem, "sync_total"),
			metric.WithDescription("The total number of sync operations (fetch)"),
		),
		syncErrorsTotal: metrics.MustInt64().Counter(
			prometheus.BuildFQName(namespace, subsystem, "sync_errors_total"),
			metric.WithDescription("The total number of sync operation errors"),
		),
		syncDuration: metrics.MustFloat64().Histogram(
			prometheus.BuildFQName(namespace, subsystem, "sync_duration"),
			metric.WithDescription("The duration of sync operations in milliseconds"),
			metric.WithUnit("ms"),
		),
		filesChanged: metrics.MustInt64().Counter(
			prometheus.BuildFQName(namespace, subsystem, "files_changed_total"),
			metric.WithDescription("The total number of files changed during sync operations"),
		),
	}

	containers.ApplyAll(&m, opts...)

	m.set = attribute.NewSet(m.attrs...)

	return m
}

func (r repoMetrics) recordView(ctx context.Context, branch string) func(error) {
	attrs := metric.WithAttributes(append(r.attrs, attrBranch.String(branch))...)
	r.viewsTotal.Add(ctx, 1, attrs)

	start := time.Now().UTC()
	return func(err error) {
		r.viewLatency.Record(ctx, float64(time.Since(start).Milliseconds()), attrs)
		if err != nil {
			r.viewErrorsTotal.Add(ctx, 1, attrs)
		}
	}
}

func (r repoMetrics) recordUpdate(ctx context.Context, branch string) func(error) {
	attrs := metric.WithAttributes(append(r.attrs, attrBranch.String(branch))...)
	r.updatesTotal.Add(ctx, 1, attrs)

	start := time.Now().UTC()
	return func(err error) {
		r.updateLatency.Record(ctx, float64(time.Since(start).Milliseconds()), attrs)
		if err != nil {
			r.updateErrorsTotal.Add(ctx, 1, attrs)
		}
	}
}

func (r repoMetrics) recordPollError(ctx context.Context) {
	r.pollErrorsTotal.Add(ctx, 1, metric.WithAttributeSet(r.set))
}

func (r repoMetrics) recordUpdateSubsError(ctx context.Context) {
	r.updateSubsErrorsTotal.Add(ctx, 1, metric.WithAttributeSet(r.set))
}

func (r repoMetrics) recordBranchCreated(ctx context.Context) func(error) {
	r.branchesTotal.Add(ctx, 1, metric.WithAttributeSet(r.set))

	return func(err error) {
		if err != nil {
			r.branchesErrorsTotal.Add(ctx, 1, metric.WithAttributeSet(r.set))
		}
	}
}

func (r repoMetrics) recordBranchDeleted(ctx context.Context) func(error) {
	r.branchesTotal.Add(ctx, -1, metric.WithAttributeSet(r.set))

	return func(err error) {
		if err != nil {
			r.branchesErrorsTotal.Add(ctx, 1, metric.WithAttributeSet(r.set))
		}
	}
}

func (r repoMetrics) recordSyncStart(ctx context.Context, branch string) func(error) {
	start := time.Now().UTC()
	return func(err error) {
		duration := float64(time.Since(start).Milliseconds())
		branchAttrs := metric.WithAttributes(append(r.attrs, attrBranch.String(branch))...)
		r.syncDuration.Record(ctx, duration, branchAttrs)

		if err != nil {
			statusAttrs := metric.WithAttributes(append(r.attrs,
				attrBranch.String(branch),
				attrStatus.String("failure"))...)
			r.syncTotal.Add(ctx, 1, statusAttrs)
		} else {
			statusAttrs := metric.WithAttributes(append(r.attrs,
				attrBranch.String(branch),
				attrStatus.String("success"))...)
			r.syncTotal.Add(ctx, 1, statusAttrs)
		}
	}
}

func (r repoMetrics) recordSyncError(ctx context.Context, branch, errorType string) {
	attrs := metric.WithAttributes(append(r.attrs,
		attrBranch.String(branch),
		attrErrorType.String(errorType))...)
	r.syncErrorsTotal.Add(ctx, 1, attrs)
}

func (r repoMetrics) recordFilesChanged(ctx context.Context, branch, operation string, count int) {
	if count > 0 {
		attrs := metric.WithAttributes(append(r.attrs,
			attrBranch.String(branch),
			attrOperation.String(operation))...)
		r.filesChanged.Add(ctx, int64(count), attrs)
	}
}
