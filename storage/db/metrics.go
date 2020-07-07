package db

import (
	"database/sql"

	"github.com/prometheus/client_golang/prometheus"
)

const (
	namespace = "flipt"
	subsystem = "db"
)

// StatsGetter is an interface that gets sql.DBStats.
type StatsGetter interface {
	Stats() sql.DBStats
}

func NewMetricsCollector(d Driver, s StatsGetter) *MetricsCollector {
	labels := prometheus.Labels{"driver": d.String()}

	return &MetricsCollector{
		sg: s,
		maxOpenDesc: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, subsystem, "max_open_conn"),
			"Maximum number of open connections to the database.",
			nil,
			labels,
		),
		openDesc: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, subsystem, "open_conn"),
			"The number of established connections both in use and idle.",
			nil,
			labels,
		),
		inUseDesc: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, subsystem, "in_use_conn"),
			"The number of connections currently in use.",
			nil,
			labels,
		),
		idleDesc: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, subsystem, "idle_conn"),
			"The number of idle connections.",
			nil,
			labels,
		),
		waitedForDesc: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, subsystem, "waited_for_conn"),
			"The total number of connections waited for.",
			nil,
			labels,
		),
		blockedSecondsDesc: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, subsystem, "blocked_seconds_conn"),
			"The total time blocked waiting for a new connection.",
			nil,
			labels,
		),
		closedMaxIdleDesc: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, subsystem, "closed_max_idle_conn"),
			"The total number of connections closed due to SetMaxIdleConns.",
			nil,
			labels,
		),
		closedMaxLifetimeDesc: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, subsystem, "closed_max_lifetime_conn"),
			"The total number of connections closed due to SetConnMaxLifetime.",
			nil,
			labels,
		),
	}
}

type MetricsCollector struct {
	sg StatsGetter

	maxOpenDesc           *prometheus.Desc
	openDesc              *prometheus.Desc
	inUseDesc             *prometheus.Desc
	idleDesc              *prometheus.Desc
	waitedForDesc         *prometheus.Desc
	blockedSecondsDesc    *prometheus.Desc
	closedMaxIdleDesc     *prometheus.Desc
	closedMaxLifetimeDesc *prometheus.Desc
}

func (c *MetricsCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.maxOpenDesc
	ch <- c.openDesc
	ch <- c.inUseDesc
	ch <- c.idleDesc
	ch <- c.waitedForDesc
	ch <- c.blockedSecondsDesc
	ch <- c.closedMaxIdleDesc
	ch <- c.closedMaxLifetimeDesc
}

func (c *MetricsCollector) Collect(ch chan<- prometheus.Metric) {
	stats := c.sg.Stats()

	ch <- prometheus.MustNewConstMetric(
		c.maxOpenDesc,
		prometheus.GaugeValue,
		float64(stats.MaxOpenConnections),
	)
	ch <- prometheus.MustNewConstMetric(
		c.openDesc,
		prometheus.GaugeValue,
		float64(stats.OpenConnections),
	)
	ch <- prometheus.MustNewConstMetric(
		c.inUseDesc,
		prometheus.GaugeValue,
		float64(stats.InUse),
	)
	ch <- prometheus.MustNewConstMetric(
		c.idleDesc,
		prometheus.GaugeValue,
		float64(stats.Idle),
	)
	ch <- prometheus.MustNewConstMetric(
		c.waitedForDesc,
		prometheus.CounterValue,
		float64(stats.WaitCount),
	)
	ch <- prometheus.MustNewConstMetric(
		c.blockedSecondsDesc,
		prometheus.CounterValue,
		stats.WaitDuration.Seconds(),
	)
	ch <- prometheus.MustNewConstMetric(
		c.closedMaxIdleDesc,
		prometheus.CounterValue,
		float64(stats.MaxIdleClosed),
	)
	ch <- prometheus.MustNewConstMetric(
		c.closedMaxLifetimeDesc,
		prometheus.CounterValue,
		float64(stats.MaxLifetimeClosed),
	)
}
