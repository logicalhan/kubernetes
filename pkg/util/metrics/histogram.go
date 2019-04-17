package metrics

import (
	"fmt"
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
	DeprecatedVersion *Version
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

// functions for KubeCounter
func (h *KubeHistogram) GetDeprecatedVersion() *Version {
	return h.HistogramOpts.DeprecatedVersion
}

func (h *KubeHistogram) InitializeMetric() {
	h.Histogram = prometheus.NewHistogram(h.HistogramOpts.toPromHistogramOpts())
}

func (h *KubeHistogram) InitializeDeprecatedMetric() {
	h.HistogramOpts.MarkDeprecated()
	h.InitializeMetric()
}

func (h *KubeHistogram) Observe(v float64) {
	h.Histogram.Observe(v)
}

func (h *KubeHistogram) Describe(ch chan<- *prometheus.Desc) {
	h.Histogram.Describe(ch)
}

func (h *KubeHistogram) Collect(m chan<- prometheus.Metric) {
	h.Histogram.Collect(m)
}

// functions for HistogramVec
func (v *HistogramVec) GetDeprecatedVersion() *Version {
	return v.HistogramOpts.DeprecatedVersion
}

func (v *HistogramVec) InitializeMetric() {
	v.HistogramVec = prometheus.NewHistogramVec(v.HistogramOpts.toPromHistogramOpts(), v.originalLabels)
}

func (v *HistogramVec) InitializeDeprecatedMetric() {
	v.HistogramOpts.MarkDeprecated()
	v.InitializeMetric()
}

// todo:        There is a problem with the underlying method call here. Prometheus behavior
// todo(cont):  here actually results in the creation of a new metric if a metric with the unique
// todo(cont):  label values is not found in the underlying stored metricMap.
// todo(cont):  For reference: https://github.com/prometheus/client_golang/blob/master/prometheus/counter.go#L148-L177

// todo(cont):  This means if we opt to disable a metric by NOT registering it, then we would also
// todo(cont):  need to ensure that this no-opts and does not create a new metric, otherwise disabling
// todo(cont):  a metric which causes a memory leak would still continue to leak memory.
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
