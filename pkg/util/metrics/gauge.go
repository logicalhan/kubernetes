package metrics

import (
	"fmt"
	"github.com/blang/semver"
	"github.com/prometheus/client_golang/prometheus"
)

type GaugeOpts Opts

// Modify help description on the metric description.
func (o *GaugeOpts) MarkDeprecated() {
	o.deprecateOnce.Do(func() {
		o.Help = fmt.Sprintf("(Deprecated since %v) %v", o.DeprecatedVersion, o.Help)
	})
}

// convenience function to allow easy transformation to the prometheus
// counterpart. This will do more logic once we have a proper label abstraction
func (o GaugeOpts) toPromGaugeOpts() prometheus.GaugeOpts {
	return prometheus.GaugeOpts{
		Namespace:   o.Namespace,
		Subsystem:   o.Subsystem,
		Name:        o.Name,
		Help:        o.Help,
		ConstLabels: o.ConstLabels,
	}
}

// This is our wrapper function for Prometheus gauges.
// We store the options the metric was defined with in order
// to defer initialization until actual metric registration.
type KubeGauge struct {
	prometheus.Gauge
	*GaugeOpts
	registerable
}

// NewGauge returns an object which is Gauge-like. However, nothing
// will be measured until the gauge is registered somewhere.
func NewGauge(opts GaugeOpts) *KubeGauge {
	kg := &KubeGauge{
		Gauge:        noop,
		GaugeOpts:    &opts,
		registerable: registerable{},
	}
	kg.init(kg)
	return kg
}

// GetDeprecatedVersion, InitializeMetric, InitializeDeprecatedMetric are required to
// satisfy the KubeCollector interface.

// GetDeprecatedVersion returns a pointer to the Version or nil
func (g *KubeGauge) GetDeprecatedVersion() *semver.Version {
	return g.GaugeOpts.DeprecatedVersion
}

// InitializeMetric invokes the actual prometheus.Gauge object instantiation
// and stores a reference to it
func (g *KubeGauge) InitializeMetric() {
	g.Gauge = prometheus.NewGauge(g.GaugeOpts.toPromGaugeOpts())
}

// InitializeMetric invokes the actual prometheus.Gauge object instantiation
// but modifies the Help description prior to object instantiation.
func (g *KubeGauge) InitializeDeprecatedMetric() {
	g.GaugeOpts.MarkDeprecated()
	g.InitializeMetric()
}

// Inc,Add,Set,Sub,SetToCurrentTime satisfy the prometheus.Gauge interface
func (g *KubeGauge) Inc() {
	g.Gauge.Inc()
}

func (g *KubeGauge) Add(v float64) {
	g.Gauge.Add(v)
}

func (g *KubeGauge) Set(v float64) {
	g.Gauge.Set(v)
}

func (g *KubeGauge) Sub(v float64) {
	g.Gauge.Sub(v)
}

func (g *KubeGauge) SetToCurrentTime() {
	g.Gauge.SetToCurrentTime()
}

// Describe and Collect satisfy the prometheus.Collector interface

// Describe delegates to the wrapped prometheus.Gauge
func (g *KubeGauge) Describe(ch chan<- *prometheus.Desc) {
	g.Gauge.Describe(ch)
}

// Collect delegates to the wrapped prometheus.Gauge
func (g *KubeGauge) Collect(m chan<- prometheus.Metric) {
	g.Gauge.Collect(m)
}

type GaugeVec struct {
	*prometheus.GaugeVec
	*GaugeOpts
	registerable
	originalLabels []string
}

func NewGaugeVec(opts GaugeOpts, labels []string) *GaugeVec {
	cv := &GaugeVec{
		GaugeOpts:      &opts,
		originalLabels: labels,
		registerable:   registerable{},
	}
	cv.init(cv)
	return cv
}

// functions for GaugeVec
func (v *GaugeVec) GetDeprecatedVersion() *semver.Version {
	return v.GaugeOpts.DeprecatedVersion
}

func (v *GaugeVec) InitializeMetric() {
	v.GaugeVec = prometheus.NewGaugeVec(v.GaugeOpts.toPromGaugeOpts(), v.originalLabels)
}

func (v *GaugeVec) InitializeDeprecatedMetric() {
	v.GaugeOpts.MarkDeprecated()
	v.InitializeMetric()
}

// There is a problem with the underlying Prometheus method call here. Prometheus behavior
// actually results in the creation of a new metric if a metric with the unique
// label values is not found in the underlying stored metricMap.
// For reference: https://github.com/prometheus/client_golang/blob/master/prometheus/Gauge.go#L148-L177

// This means if we opt to disable a metric by NOT registering it, then we also need to ensure that
// we do not create new metrics when this is invoked, otherwise disabling a metric which
// causes a memory leak would still continue to leak memory (since we would be continuing to make
// arbitrary numbers of metrics for each unique label combo).
func (v *GaugeVec) GetMetricWithLabelValues(lvs ...string) (prometheus.Gauge, error) {
	if !v.IsRegistered() {
		return noop, nil
	}
	return v.GaugeVec.GetMetricWithLabelValues(lvs...)
}

func (v *GaugeVec) GetMetricWith(labels prometheus.Labels) (prometheus.Gauge, error) {
	if !v.IsRegistered() {
		return noop, nil
	}
	return v.GaugeVec.GetMetricWith(labels)
}

func (v *GaugeVec) WithLabelValues(lvs ...string) prometheus.Gauge {
	if !v.IsRegistered() {
		return noop
	}
	return v.GaugeVec.WithLabelValues(lvs...)
}

func (v *GaugeVec) With(labels prometheus.Labels) prometheus.Gauge {
	if !v.IsRegistered() {
		return noop
	}
	return v.GaugeVec.With(labels)
}

func (v *GaugeVec) CurryWith(labels prometheus.Labels) (*prometheus.GaugeVec, error) {
	if !v.IsRegistered() {
		return noopGaugeVec, nil
	}
	vec, err := v.GaugeVec.CurryWith(labels)
	if vec != nil {
		return vec, err
	}
	return nil, err
}

func (v *GaugeVec) MustCurryWith(labels prometheus.Labels) *prometheus.GaugeVec {
	if !v.IsRegistered() {
		return noopGaugeVec
	}
	vec, err := v.CurryWith(labels)
	if err != nil {
		panic(err)
	}
	return vec
}

// Describe implements Collector. It will send exactly one Desc to the provided
// channel.
func (v *GaugeVec) Describe(ch chan<- *prometheus.Desc) {
	v.GaugeVec.Describe(ch)
}

// Collect implements Collector.
func (v *GaugeVec) Collect(ch chan<- prometheus.Metric) {
	v.GaugeVec.Collect(ch)
}

// Reset deletes all metrics in this vector.
func (v *GaugeVec) Reset() {
	v.GaugeVec.Reset()
}
