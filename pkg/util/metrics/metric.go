package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"sync"

	dto "github.com/prometheus/client_model/go"
)

type Deprecated bool

const (
	IsDeprecated  Deprecated = true
	NotDeprecated Deprecated = false
)

/**
 * This extends the prometheus.Collector interface so
 * that we can add additional functionality on top
 * of our metric registration process.
 */
type KubeCollector interface {
	prometheus.Collector
	Registerable
	GetDeprecatedVersion() *Version
	RegisterDeprecatedMetric()
	RegisterMetric()
}

type Registerable interface {
	InitializeMetric(Deprecated)
	IsRegistered() bool
}

type registerable struct {
	Registerable
	isRegistered bool
	registerOnce sync.Once
	self         KubeCollector
}

// Store a reference so that we can defer initialization of the metric
func (r *registerable) init(self KubeCollector) {
	r.self = self
}

func (r *registerable) IsRegistered() bool {
	return r.isRegistered
}

// Defer initialization of metric until we know if we actually need to
// register the thing.
func (r *registerable) InitializeMetric(isDeprecated Deprecated) {
	r.registerOnce.Do(func() {
		r.isRegistered = true
		if isDeprecated {
			r.self.RegisterDeprecatedMetric()
		} else {
			r.self.RegisterMetric()
		}
	})
}

// no-op vecs for convenience
var noopCounterVec = prometheus.CounterVec{}
var noopHistogramVec = prometheus.HistogramVec{}
var noopSummaryVec = prometheus.SummaryVec{}
var noopGaugeVec = prometheus.GaugeVec{}

// just use a convenience struct for all the no-ops
var noop = noopMetric{}

type noopMetric struct{}

func (noopMetric) Inc()                             {}
func (noopMetric) Add(float64)                      {}
func (noopMetric) Dec()                             {}
func (noopMetric) Set(float64)                      {}
func (noopMetric) Observe(float64)                  {}
func (noopMetric) Desc() *prometheus.Desc           { return nil }
func (noopMetric) Write(*dto.Metric) error          { return nil }
func (noopMetric) Describe(chan<- *prometheus.Desc) {}
func (noopMetric) Collect(chan<- prometheus.Metric) {}
