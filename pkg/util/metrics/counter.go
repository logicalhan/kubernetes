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

func NewCounter(opts CounterOpts) KubeMetric {
	c := prometheus.NewCounter(opts.toPromCounterOpts())
	return KubeMetric{c, opts.DeprecatedVersion}
}

type CounterVec struct {
	cVec              *prometheus.CounterVec
	DeprecatedVersion *Version
}

func NewCounterVec(opts CounterOpts, labels []string) *CounterVec {
	cVec := prometheus.NewCounterVec(opts.toPromCounterOpts(), labels)
	return &CounterVec{cVec: cVec, DeprecatedVersion: opts.DeprecatedVersion}
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
	return v.cVec.GetMetricWithLabelValues(lvs...)
}

// todo: let's not return a promoetheus counter here
func (v *CounterVec) GetMetricWith(labels prometheus.Labels) (prometheus.Counter, error) {
	return v.cVec.GetMetricWith(labels)
}

// todo: let's not return a promoetheus counter here
func (v *CounterVec) WithLabelValues(lvs ...string) prometheus.Counter {
	return v.cVec.WithLabelValues(lvs...)
}

// todo: let's not return a promoetheus counter here
func (v *CounterVec) With(labels prometheus.Labels) prometheus.Counter {
	return v.cVec.With(labels)
}

func (v *CounterVec) CurryWith(labels prometheus.Labels) (*CounterVec, error) {
	vec, err := v.cVec.CurryWith(labels)
	if vec != nil {
		return &CounterVec{vec, v.DeprecatedVersion}, err
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

// Describe implements Collector. It will send exactly one Desc to the provided
// channel.
func (v *CounterVec) Describe(ch chan<- *prometheus.Desc) {
	v.cVec.Describe(ch)
}

// Collect implements Collector.
func (v *CounterVec) Collect(ch chan<- prometheus.Metric) {
	v.cVec.Collect(ch)
}

// Reset deletes all metrics in this vector.
func (m *CounterVec) Reset() {
	m.cVec.Reset()
}
