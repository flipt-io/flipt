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
	attrRemote = attribute.Key("remote")
	attrBranch = attribute.Key("branch")
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
