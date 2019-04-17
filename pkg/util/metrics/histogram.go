package metrics

import (
	"fmt"
	"github.com/blang/semver"
	"github.com/prometheus/client_golang/prometheus"
	"sync"
)

type HistogramOpts struct {
	Namespace         string
	Subsystem         string
	Name              string
	Help              string
	ConstLabels       prometheus.Labels
	Buckets           []float64
	DeprecatedVersion *semver.Version
	deprecateOnce     sync.Once
}

// Modify help description on the metric description.
func (o *HistogramOpts) MarkDeprecated() {
	o.deprecateOnce.Do(func() {
		o.Help = fmt.Sprintf("(Deprecated since %v) %v", o.DeprecatedVersion, o.Help)
	})
}

// convenience function to allow easy transformation to the prometheus
// counterpart. This will do more once we have a proper label abstraction
func (o HistogramOpts) toPromHistogramOpts() prometheus.HistogramOpts {
	return prometheus.HistogramOpts{
		Namespace:   o.Namespace,
		Subsystem:   o.Subsystem,
		Name:        o.Name,
		Help:        o.Help,
		ConstLabels: o.ConstLabels,
		Buckets:     o.Buckets,
	}
}

// This is our wrapper function for prometheus counters
// we store the options the metric was defined with in order
// to defer initialization until actual metric registration.
type KubeHistogram struct {
	prometheus.Histogram
	*HistogramOpts
	registerable
}

// NewHistogram returns an object which is Histogram-like. However, nothing
// will be measured until the histogram is registered somewhere.
func NewHistogram(opts HistogramOpts) *KubeHistogram {
	h := &KubeHistogram{
		Histogram:     noop,
		HistogramOpts: &opts,
		registerable:  registerable{},
	}
	// store a reference to ourselves so that we can defer registration
	h.init(h)
	return h
}

// GetDeprecatedVersion, InitializeMetric, InitializeDeprecatedMetric are required to
// satisfy the KubeCollector interface.

// GetDeprecatedVersion returns a pointer to the Version or nil
func (h *KubeHistogram) GetDeprecatedVersion() *semver.Version {
	return h.HistogramOpts.DeprecatedVersion
}

// InitializeMetric invokes the actual prometheus.Histogram object instantiation
// and stores a reference to it
func (h *KubeHistogram) InitializeMetric() {
	h.Histogram = prometheus.NewHistogram(h.HistogramOpts.toPromHistogramOpts())
}

// InitializeMetric invokes the actual prometheus.Histogram object instantiation
// but modifies the Help description prior to object instantiation.
func (h *KubeHistogram) InitializeDeprecatedMetric() {
	h.HistogramOpts.MarkDeprecated()
	h.InitializeMetric()
}

// Observe satisfies the prometheus.Observer interface. This call delegates to
// the underlying histogram.
func (h *KubeHistogram) Observe(v float64) {
	h.Histogram.Observe(v)
}

// Describe and Collect satisfy the prometheus.Collector interface

// Describe delegates to the wrapped prometheus.Histogram
func (h *KubeHistogram) Describe(ch chan<- *prometheus.Desc) {
	h.Histogram.Describe(ch)
}

// Collect delegates to the wrapped prometheus.Histogram
func (h *KubeHistogram) Collect(m chan<- prometheus.Metric) {
	h.Histogram.Collect(m)
}

type HistogramVec struct {
	*prometheus.HistogramVec
	*HistogramOpts
	registerable
	originalLabels []string
}

func NewHistogramVec(opts HistogramOpts, labels []string) *HistogramVec {
	v := &HistogramVec{
		HistogramOpts:  &opts,
		originalLabels: labels,
		registerable:   registerable{},
	}
	v.init(v)
	return v
}

// functions for HistogramVec
func (v *HistogramVec) GetDeprecatedVersion() *semver.Version {
	return v.HistogramOpts.DeprecatedVersion
}

func (v *HistogramVec) InitializeMetric() {
	v.HistogramVec = prometheus.NewHistogramVec(v.HistogramOpts.toPromHistogramOpts(), v.originalLabels)
}

func (v *HistogramVec) InitializeDeprecatedMetric() {
	v.HistogramOpts.MarkDeprecated()
	v.InitializeMetric()
}

// There is a problem with the underlying Prometheus method call here. Prometheus behavior
// actually results in the creation of a new metric if a metric with the unique
// label values is not found in the underlying stored metricMap.
// For reference: https://github.com/prometheus/client_golang/blob/master/prometheus/counter.go#L148-L177

// This means if we opt to disable a metric by NOT registering it, then we also need to ensure that
// we do not create new metrics when this is invoked, otherwise disabling a metric which
// causes a memory leak would still continue to leak memory (since we would be continuing to make
// arbitrary numbers of metrics for each unique label combo).
func (v *HistogramVec) GetMetricWithLabelValues(lvs ...string) (prometheus.Observer, error) {
	if !v.IsRegistered() {
		return noop, nil
	}
	return v.HistogramVec.GetMetricWithLabelValues(lvs...)
}

func (v *HistogramVec) GetMetricWith(labels prometheus.Labels) (prometheus.Observer, error) {
	if !v.IsRegistered() {
		return noop, nil
	}
	return v.HistogramVec.GetMetricWith(labels)
}

func (v *HistogramVec) WithLabelValues(lvs ...string) prometheus.Observer {
	if !v.IsRegistered() {
		return noop
	}
	return v.HistogramVec.WithLabelValues(lvs...)
}

func (v *HistogramVec) With(labels prometheus.Labels) prometheus.Observer {
	if !v.IsRegistered() {
		return noop
	}
	return v.HistogramVec.With(labels)
}

func (v *HistogramVec) CurryWith(labels prometheus.Labels) (prometheus.ObserverVec, error) {
	if !v.IsRegistered() {
		return noopObserverVec, nil
	}
	vec, err := v.HistogramVec.CurryWith(labels)
	if vec != nil {
		return vec, err
	}
	return nil, err
}

func (v *HistogramVec) MustCurryWith(labels prometheus.Labels) prometheus.ObserverVec {
	if !v.IsRegistered() {
		return noopObserverVec
	}
	vec, err := v.CurryWith(labels)
	if err != nil {
		panic(err)
	}
	return vec
}

// Describe implements Collector. It will send exactly one Desc to the provided
// channel.
func (v *HistogramVec) Describe(ch chan<- *prometheus.Desc) {
	v.HistogramVec.Describe(ch)
}

// Collect implements Collector.
func (v *HistogramVec) Collect(ch chan<- prometheus.Metric) {
	v.HistogramVec.Collect(ch)
}

// Reset deletes all metrics in this vector.
func (v *HistogramVec) Reset() {
	v.HistogramVec.Reset()
}
