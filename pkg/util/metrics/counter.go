package metrics

import (
	"fmt"
	"github.com/blang/semver"
	"github.com/prometheus/client_golang/prometheus"
)

type CounterOpts Opts

// Modify help description on the metric description.
func (o *CounterOpts) MarkDeprecated() {
	o.deprecateOnce.Do(func() {
		o.Help = fmt.Sprintf("(Deprecated since %v) %v", o.DeprecatedVersion, o.Help)
	})
}

// convenience function to allow easy transformation to the prometheus
// counterpart. This will do more once we have a proper label abstraction
func (c CounterOpts) toPromCounterOpts() prometheus.CounterOpts {
	return prometheus.CounterOpts{
		Namespace:   c.Namespace,
		Subsystem:   c.Subsystem,
		Name:        c.Name,
		Help:        c.Help,
		ConstLabels: c.ConstLabels,
	}
}

// This is our wrapper function for prometheus counters
// we store the options the metric was defined with in order
// to defer initialization until actual metric registration.
type KubeCounter struct {
	prometheus.Counter
	*CounterOpts
	registerable
}

// NewCounter returns an object which is Counter-like. However, nothing
// will be measured until the counter is actually registered.
func NewCounter(opts CounterOpts) *KubeCounter {
	kc := &KubeCounter{
		Counter:      noop,
		CounterOpts:  &opts,
		registerable: registerable{},
	}
	// store a reference to ourselves so that we can defer registration
	kc.init(kc)
	return kc
}

// GetDeprecatedVersion, InitializeMetric, InitializeDeprecatedMetric are required to
// satisfy the KubeCollector interface.

// GetDeprecatedVersion returns a pointer to the Version or nil
func (c *KubeCounter) GetDeprecatedVersion() *semver.Version {
	return c.CounterOpts.DeprecatedVersion
}

// InitializeMetric invokes the actual prometheus.Counter object instantiation
// and stores a reference to it
func (c *KubeCounter) InitializeMetric() {
	c.Counter = prometheus.NewCounter(c.CounterOpts.toPromCounterOpts())
}

// InitializeMetric invokes the actual prometheus.Counter object instantiation
// but modifies the Help description prior to object instantiation.
func (c *KubeCounter) InitializeDeprecatedMetric() {
	c.CounterOpts.MarkDeprecated()
	c.InitializeMetric()
}

// Inc and Add satisfy the prometheus.Counter interface

// Inc delegates to the wrapped prometheus.Counter
func (c *KubeCounter) Inc() {
	c.Counter.Inc()
}

// Add delegates to the wrapped prometheus.Counter
func (c *KubeCounter) Add(v float64) {
	c.Counter.Add(v)
}

// Describe and Collect satisfy the prometheus.Collector interface

// Describe delegates to the wrapped prometheus.Counter
func (c *KubeCounter) Describe(ch chan<- *prometheus.Desc) {
	c.Counter.Describe(ch)
}

// Collect delegates to the wrapped prometheus.Counter
func (c *KubeCounter) Collect(m chan<- prometheus.Metric) {
	c.Counter.Collect(m)
}

// Wrappers for Counters with Labels below.
type CounterVec struct {
	*prometheus.CounterVec
	*CounterOpts
	registerable
	originalLabels []string
}

func NewCounterVec(opts CounterOpts, labels []string) *CounterVec {
	cv := &CounterVec{
		CounterVec:     nil,
		CounterOpts:    &opts,
		originalLabels: labels,
		registerable:   registerable{},
	}
	cv.init(cv)
	return cv
}

// GetDeprecatedVersion, InitializeMetric, InitializeDeprecatedMetric are required to
// satisfy the KubeCollector interface.
func (v *CounterVec) GetDeprecatedVersion() *semver.Version {
	return v.CounterOpts.DeprecatedVersion
}

func (v *CounterVec) InitializeMetric() {
	v.CounterVec = prometheus.NewCounterVec(v.CounterOpts.toPromCounterOpts(), v.originalLabels)
}

func (v *CounterVec) InitializeDeprecatedMetric() {
	v.CounterOpts.MarkDeprecated()
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
func (v *CounterVec) GetMetricWithLabelValues(lvs ...string) (prometheus.Counter, error) {
	if !v.IsRegistered() {
		return noop, nil
	}
	return v.CounterVec.GetMetricWithLabelValues(lvs...)
}

func (v *CounterVec) GetMetricWith(labels prometheus.Labels) (prometheus.Counter, error) {
	if !v.IsRegistered() {
		return noop, nil
	}
	return v.CounterVec.GetMetricWith(labels)
}

func (v *CounterVec) WithLabelValues(lvs ...string) prometheus.Counter {
	if !v.IsRegistered() {
		return noop
	}
	return v.CounterVec.WithLabelValues(lvs...)
}

func (v *CounterVec) With(labels prometheus.Labels) prometheus.Counter {
	if !v.IsRegistered() {
		return noop
	}
	return v.CounterVec.With(labels)
}

func (v *CounterVec) CurryWith(labels prometheus.Labels) (*prometheus.CounterVec, error) {
	if !v.IsRegistered() {
		return noopCounterVec, nil
	}
	vec, err := v.CounterVec.CurryWith(labels)
	if vec != nil {
		return vec, err
	}
	return nil, err
}

func (v *CounterVec) MustCurryWith(labels prometheus.Labels) *prometheus.CounterVec {
	if !v.IsRegistered() {
		return noopCounterVec
	}
	vec, err := v.CurryWith(labels)
	if err != nil {
		panic(err)
	}
	return vec
}

// Describe implements Collector. It will send exactly one Desc to the provided
// channel.
func (v *CounterVec) Describe(ch chan<- *prometheus.Desc) {
	v.CounterVec.Describe(ch)
}

// Collect implements Collector.
func (v *CounterVec) Collect(ch chan<- prometheus.Metric) {
	v.CounterVec.Collect(ch)
}

// Reset deletes all metrics in this vector.
func (v *CounterVec) Reset() {
	v.CounterVec.Reset()
}
