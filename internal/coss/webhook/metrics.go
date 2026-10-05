// Flipt Commercial Open Source Feature
// This file contains functionality that is licensed under the Flipt Fair Core License (FCL).
// You may NOT use, modify, or distribute this file or its contents without a valid paid license.
// For details: https://github.com/flipt-io/flipt/blob/v2/LICENSE

package webhook

import (
	"context"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"go.flipt.io/flipt/internal/otel/metrics"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
)

const (
	namespace = "flipt"
	subsystem = "incoming_webhook"
)

// result is the outcome of a webhook request, as reported by the requests
// counter.
type result string

const (
	resultAccepted     result = "accepted"
	resultIgnored      result = "ignored"
	resultUnauthorized result = "unauthorized"
	resultNotFound     result = "not_found"
)

// unknownLabel is the environment and scm label value for requests that don't
// name a configured environment. It keeps label cardinality bounded against
// arbitrary request paths.
const unknownLabel = "unknown"

var (
	attrEnvironment = attribute.Key("environment")
	attrSCM         = attribute.Key("scm")
	attrResult      = attribute.Key("result")
)

type receiverMetrics struct {
	requestsTotal    metric.Int64Counter
	fetchErrorsTotal metric.Int64Counter
	syncDuration     metric.Float64Histogram
}

func newReceiverMetrics() receiverMetrics {
	return receiverMetrics{
		requestsTotal: metrics.MustInt64().Counter(
			prometheus.BuildFQName(namespace, subsystem, "requests_total"),
			metric.WithDescription("The total number of incoming webhook requests by result"),
		),
		fetchErrorsTotal: metrics.MustInt64().Counter(
			prometheus.BuildFQName(namespace, subsystem, "fetch_errors_total"),
			metric.WithDescription("The total number of webhook-triggered fetches that failed"),
		),
		// No unit is set: the Prometheus exporter would otherwise append a
		// _seconds suffix to the documented metric name.
		syncDuration: metrics.MustFloat64().Histogram(
			prometheus.BuildFQName(namespace, subsystem, "sync_duration"),
			metric.WithDescription("The time from receiving an accepted webhook to its fetch completing in seconds"),
			metric.WithExplicitBucketBoundaries(0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 10, 30, 60),
		),
	}
}

func (m receiverMetrics) recordRequest(ctx context.Context, environment, scm string, res result) {
	m.requestsTotal.Add(ctx, 1, metric.WithAttributes(
		attrEnvironment.String(environment),
		attrSCM.String(scm),
		attrResult.String(string(res)),
	))
}

func (m receiverMetrics) recordFetchError(ctx context.Context, environment, scm string) {
	m.fetchErrorsTotal.Add(ctx, 1, metric.WithAttributes(
		attrEnvironment.String(environment),
		attrSCM.String(scm),
	))
}

func (m receiverMetrics) recordSync(ctx context.Context, environment, scm string, d time.Duration) {
	m.syncDuration.Record(ctx, d.Seconds(), metric.WithAttributes(
		attrEnvironment.String(environment),
		attrSCM.String(scm),
	))
}
