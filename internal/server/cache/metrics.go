package cache

import (
	"github.com/prometheus/client_golang/prometheus"
)

type Stats struct {
	MissTotal  uint64
	HitTotal   uint64
	ErrorTotal uint64
}

// statsGetter is an interface that gets cache.Stats.
type statsGetter interface {
	Stats() Stats
}

const (
	namespace = "flipt"
	subsystem = "cache"
)

func RegisterMetrics(c Cacher) {
	labels := prometheus.Labels{"cache": c.String()}

	collector := &metricsCollector{
		sg: c,
		hitTotalDesc: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, subsystem, "hit_total"),
			"The number of cache hits",
			nil,
			labels,
		),
		missTotalDec: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, subsystem, "miss_total"),
			"The number of cache misses",
			nil,
			labels,
		),
		errorTotalDesc: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, subsystem, "error_total"),
			"The number of times an error occurred reading or writing to the cache",
			nil,
			labels,
		),
	}

	prometheus.Unregister(collector)
	prometheus.MustRegister(collector)
}

type metricsCollector struct {
	sg statsGetter

	hitTotalDesc   *prometheus.Desc
	missTotalDec   *prometheus.Desc
	errorTotalDesc *prometheus.Desc
}

func (c *metricsCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.hitTotalDesc
	ch <- c.missTotalDec
	ch <- c.errorTotalDesc
}

func (c *metricsCollector) Collect(ch chan<- prometheus.Metric) {
	stats := c.sg.Stats()

	ch <- prometheus.MustNewConstMetric(
		c.hitTotalDesc,
		prometheus.CounterValue,
		float64(stats.HitTotal),
	)
	ch <- prometheus.MustNewConstMetric(
		c.missTotalDec,
		prometheus.CounterValue,
		float64(stats.MissTotal),
	)
	ch <- prometheus.MustNewConstMetric(
		c.errorTotalDesc,
		prometheus.CounterValue,
		float64(stats.ErrorTotal),
	)
}
