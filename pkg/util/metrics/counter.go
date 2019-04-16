package metrics

import (
	"github.com/prometheus/client_golang/prometheus"

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
	originalOpts prometheus.CounterOpts
	deprecatedVersion *Version
}

func (c *KubeCounter) GetDeprecatedVersion() *Version {
	return c.deprecatedVersion
}

func (c *KubeCounter) Inc() {
	c.Counter.Inc()
}

func (c *KubeCounter) Add(v float64) {
	c.Counter.Add(v)
}

func (c *KubeCounter) Describe(ch chan<- *prometheus.Desc) {
	c.Counter.Describe(ch)
}

func (c *KubeCounter) Collect(m chan<- prometheus.Metric) {
	c.Counter.Collect(m)
}

func NewCounter(opts CounterOpts) *KubeCounter {
	c := prometheus.NewCounter(opts.toPromCounterOpts())
	return &KubeCounter{c, opts.toPromCounterOpts(), opts.DeprecatedVersion}
}

type CounterVec struct {
	*prometheus.CounterVec
	originalOpts prometheus.CounterOpts
	DeprecatedVersion *Version
}

func NewCounterVec(opts CounterOpts, labels []string) *CounterVec {
	vec := prometheus.NewCounterVec(opts.toPromCounterOpts(), labels)
	return &CounterVec{CounterVec: vec, originalOpts: opts.toPromCounterOpts(), DeprecatedVersion: opts.DeprecatedVersion}
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
	return &KubeCounter{c, v.originalOpts, v.DeprecatedVersion}, e
}

func (v *CounterVec) WithLabelValues(lvs ...string) *KubeCounter {
	return &KubeCounter{v.CounterVec.WithLabelValues(lvs...), v.originalOpts, v.DeprecatedVersion}
}

func (v *CounterVec) With(labels prometheus.Labels) *KubeCounter {
	return &KubeCounter{v.CounterVec.With(labels), v.originalOpts, v.DeprecatedVersion}
}

func (v *CounterVec) CurryWith(labels prometheus.Labels) (*CounterVec, error) {
	vec, err := v.CounterVec.CurryWith(labels)
	if vec != nil {
		return &CounterVec{vec, v.originalOpts, v.DeprecatedVersion}, err
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
func (m *CounterVec) GetDeprecatedVersion() *Version {
	return m.DeprecatedVersion
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
func (m *CounterVec) Reset() {
	m.CounterVec.Reset()
}
