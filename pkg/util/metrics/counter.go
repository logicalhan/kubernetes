package metrics

import (
	"fmt"
	"github.com/prometheus/client_golang/prometheus"
	"sync"
)

type CounterOpts Opts

// convenience function to allow easy transformation to the prometheus
// counterpart. This will do more once we have a proper label abstraction
func (c CounterOpts) toPromCounterOpts() prometheus.CounterOpts {
	return prometheus.CounterOpts{
		Namespace:   c.Namespace,
		Subsystem:   c.Subsystem,
		Name:        c.Name,
		Help:        c.Help,
		ConstLabels: c.ConstLabels}
}

type KubeCounter struct {
	prometheus.Counter
	CounterOpts
	isDeprecated         bool
	initializeDeprecated sync.Once
	deprecatedCounter    prometheus.Counter
}

func getDeprecatedCounterOpts(originalOpts CounterOpts) CounterOpts {
	return CounterOpts{
		Namespace:         originalOpts.Namespace,
		Name:              originalOpts.Name,
		Subsystem:         originalOpts.Subsystem,
		ConstLabels:       originalOpts.ConstLabels,
		Help:              fmt.Sprintf("(Deprecated since %v) %v", originalOpts.DeprecatedVersion, originalOpts.Help),
		DeprecatedVersion: originalOpts.DeprecatedVersion,
	}
}

func (c *KubeCounter) GetMetric() prometheus.Counter {
	if c.isDeprecated {
		return c.deprecatedCounter
	}
	return c.Counter
}

func (c *KubeCounter) GetDeprecatedVersion() *Version {
	return c.CounterOpts.DeprecatedVersion
}

func (c *KubeCounter) MarkDeprecated() {
	c.initializeDeprecated.Do(func() {
		c.isDeprecated = true
		c.deprecatedCounter = prometheus.NewCounter(getDeprecatedCounterOpts(c.CounterOpts).toPromCounterOpts())
	})
}

func (c *KubeCounter) Inc() {
	c.GetMetric().Inc()
}

func (c *KubeCounter) Add(v float64) {
	c.GetMetric().Add(v)
}

func (c *KubeCounter) Describe(ch chan<- *prometheus.Desc) {
	c.GetMetric().Describe(ch)
}

func (c *KubeCounter) Collect(m chan<- prometheus.Metric) {
	c.GetMetric().Collect(m)
}

func NewCounter(opts CounterOpts) *KubeCounter {
	c := prometheus.NewCounter(opts.toPromCounterOpts())
	return &KubeCounter{Counter: c, CounterOpts: opts}
}

type CounterVec struct {
	*prometheus.CounterVec
	CounterOpts
	originalLabels       []string
	isDeprecated         bool
	initializeDeprecated sync.Once
	deprecatedCounterVec *prometheus.CounterVec
}

func NewCounterVec(opts CounterOpts, labels []string) *CounterVec {
	vec := prometheus.NewCounterVec(opts.toPromCounterOpts(), labels)
	return &CounterVec{
		CounterVec:     vec,
		CounterOpts:    opts,
		originalLabels: labels,
	}
}

func (v *CounterVec) MarkDeprecated() {
	v.initializeDeprecated.Do(func() {
		v.isDeprecated = true
		newOpts := getDeprecatedCounterOpts(v.CounterOpts)
		v.deprecatedCounterVec = prometheus.NewCounterVec(newOpts.toPromCounterOpts(), v.originalLabels)
	})
	v.isDeprecated = true
}

func (v *CounterVec) GetMetric() *prometheus.CounterVec {
	if v.isDeprecated {
		return v.deprecatedCounterVec
	}
	return v.CounterVec
}

// todo:        There is a problem with the underlying method call here. Prometheus behavior
// todo(cont):  here actually results in the creation of a new metric if a metric with the unique
// todo(cont):  label values is not found in the underlying stored metricMap.
// todo(cont):  For reference: https://github.com/prometheus/client_golang/blob/master/prometheus/counter.go#L148-L177

// todo(cont):  This means if we opt to disable a metric by NOT registering it, then we would also
// todo(cont):  need to ensure that this no-opts and does not create a new metric, otherwise disabling
// todo(cont):  a metric which causes a memory leak would still continue to leak memory.
func (v *CounterVec) GetMetricWithLabelValues(lvs ...string) (prometheus.Counter, error) {
	/*
		    Do something like:
			if v.isRegistered {
				return v.cVec.GetMetricWithLabelValues(lvs ...)
			} else {
				return noOptCounter, nil
			}
	*/
	return v.GetMetric().GetMetricWithLabelValues(lvs...)
}

func (v *CounterVec) GetMetricWith(labels prometheus.Labels) (*KubeCounter, error) {
	c, e := v.GetMetric().GetMetricWith(labels)
	return &KubeCounter{Counter: c, CounterOpts: v.CounterOpts}, e
}

func (v *CounterVec) WithLabelValues(lvs ...string) *KubeCounter {
	return &KubeCounter{Counter: v.GetMetric().WithLabelValues(lvs...), CounterOpts: v.CounterOpts}
}

func (v *CounterVec) With(labels prometheus.Labels) *KubeCounter {
	return &KubeCounter{Counter: v.GetMetric().With(labels), CounterOpts: v.CounterOpts}
}

func (v *CounterVec) CurryWith(labels prometheus.Labels) (*CounterVec, error) {
	vec, err := v.GetMetric().CurryWith(labels)
	if vec != nil {
		return &CounterVec{CounterVec: vec, CounterOpts: v.CounterOpts, originalLabels: v.originalLabels}, err
	}
	return nil, err
}

func (v *CounterVec) MustCurryWith(labels prometheus.Labels) *CounterVec {
	vec, err := v.CurryWith(labels)
	if err != nil {
		panic(err)
	}
	return vec
}

// Reset deletes all metrics in this vector.
func (v *CounterVec) GetDeprecatedVersion() *Version {
	return v.CounterOpts.DeprecatedVersion
}

// Describe implements Collector. It will send exactly one Desc to the provided
// channel.
func (v *CounterVec) Describe(ch chan<- *prometheus.Desc) {
	v.GetMetric().Describe(ch)
}

// Collect implements Collector.
func (v *CounterVec) Collect(ch chan<- prometheus.Metric) {
	v.GetMetric().Collect(ch)
}

// Reset deletes all metrics in this vector.
func (v *CounterVec) Reset() {
	v.GetMetric().Reset()
}
