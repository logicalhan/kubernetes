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
	isDeprecated bool
	initializeDeprecated sync.Once
	deprecatedCounter prometheus.Counter
}

func getDeprecatedCounterOpts(originalOpts CounterOpts) CounterOpts {
	return CounterOpts{
		Namespace: originalOpts.Namespace,
		Name: originalOpts.Name,
		Subsystem: originalOpts.Subsystem,
		ConstLabels: originalOpts.ConstLabels,
		Help: fmt.Sprintf("(Deprecated since %v) %v", originalOpts.DeprecatedVersion, originalOpts.Help),
		DeprecatedVersion: originalOpts.DeprecatedVersion,
	}
}

func (c *KubeCounter) GetDeprecatedMetric() prometheus.Counter {
	c.initializeDeprecated.Do(func() {
		c.deprecatedCounter = prometheus.NewCounter(getDeprecatedCounterOpts(c.CounterOpts).toPromCounterOpts())
	})
	return c.deprecatedCounter
}

func (c *KubeCounter) GetDeprecatedVersion() *Version {
	return c.CounterOpts.DeprecatedVersion
}

func (c *KubeCounter) MarkDeprecated() {
	c.isDeprecated = true
}

func (c *KubeCounter) Inc() {
	if c.isDeprecated {
		c.GetDeprecatedMetric().Inc()
	} else {
		c.Counter.Inc()
	}

}

func (c *KubeCounter) Add(v float64) {
	if c.isDeprecated {
		c.GetDeprecatedMetric().Add(v)
	} else {
		c.Counter.Add(v)
	}
}

func (c *KubeCounter) Describe(ch chan<- *prometheus.Desc) {
	if c.isDeprecated {
		c.GetDeprecatedMetric().Describe(ch)
	} else {
		c.Counter.Describe(ch)
	}
}

func (c *KubeCounter) Collect(m chan<- prometheus.Metric) {
	if c.isDeprecated {
		c.GetDeprecatedMetric().Collect(m)
	} else {
		c.Counter.Collect(m)
	}
}

func NewCounter(opts CounterOpts) *KubeCounter {
	c := prometheus.NewCounter(opts.toPromCounterOpts())
	return &KubeCounter{Counter: c, CounterOpts: opts}
}

type CounterVec struct {
	*prometheus.CounterVec
	CounterOpts
	originalLabels []string
	isDeprecated bool
}

func NewCounterVec(opts CounterOpts, labels []string) *CounterVec {
	vec := prometheus.NewCounterVec(opts.toPromCounterOpts(), labels)
	return &CounterVec{
		CounterVec: vec,
		CounterOpts: opts,
		originalLabels: labels,
	}
}

func (v *CounterVec) MarkDeprecated() {
	v.isDeprecated = true
}
func (v *CounterVec) GetDeprecatedMetric() DeprecatableCollector {
	newOpts := getDeprecatedCounterOpts(v.CounterOpts)
	return NewCounterVec(newOpts, v.originalLabels)
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
	return v.CounterVec.GetMetricWithLabelValues(lvs...)
}

func (v *CounterVec) GetMetricWith(labels prometheus.Labels) (*KubeCounter, error) {
	c, e := v.CounterVec.GetMetricWith(labels)
	return &KubeCounter{Counter: c, CounterOpts: v.CounterOpts}, e
}

func (v *CounterVec) WithLabelValues(lvs ...string) *KubeCounter {
	return &KubeCounter{Counter: v.CounterVec.WithLabelValues(lvs...), CounterOpts: v.CounterOpts}
}

func (v *CounterVec) With(labels prometheus.Labels) *KubeCounter {
	return &KubeCounter{Counter: v.CounterVec.With(labels), CounterOpts: v.CounterOpts}
}

func (v *CounterVec) CurryWith(labels prometheus.Labels) (*CounterVec, error) {
	vec, err := v.CounterVec.CurryWith(labels)
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
