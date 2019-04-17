package metrics

import (
    "fmt"
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
func (c GaugeOpts) toPromGaugeOpts() prometheus.GaugeOpts {
    return prometheus.GaugeOpts{
        Namespace:   c.Namespace,
        Subsystem:   c.Subsystem,
        Name:        c.Name,
        Help:        c.Help,
        ConstLabels: c.ConstLabels,
    }
}

type KubeGauge struct {
    prometheus.Gauge
    *GaugeOpts
    registerable
}

func NewGauge(opts GaugeOpts) *KubeGauge {
    kg := &KubeGauge{
        Gauge: noop,
        GaugeOpts: &opts,
        registerable: registerable{},
    }
    kg.init(kg)
    return kg
}

type GaugeVec struct {
    *prometheus.GaugeVec
    *GaugeOpts
    registerable
    originalLabels []string
}

func NewGaugeVec(opts GaugeOpts, labels []string) *GaugeVec {
    cv := &GaugeVec{
        GaugeOpts:    &opts,
        originalLabels: labels,
        registerable:   registerable{},
    }
    cv.init(cv)
    return cv
}

// functions for KubeGauge
func (g *KubeGauge) GetDeprecatedVersion() *Version {
    return g.GaugeOpts.DeprecatedVersion
}

func (g *KubeGauge) InitializeMetric() {
    g.Gauge = prometheus.NewGauge(g.GaugeOpts.toPromGaugeOpts())
}

func (g *KubeGauge) InitializeDeprecatedMetric() {
    g.GaugeOpts.MarkDeprecated()
    g.InitializeMetric()
}

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

func (g *KubeGauge) Describe(ch chan<- *prometheus.Desc) {
    g.Gauge.Describe(ch)
}

func (g *KubeGauge) Collect(m chan<- prometheus.Metric) {
    g.Gauge.Collect(m)
}


// functions for GaugeVec
func (v *GaugeVec) GetDeprecatedVersion() *Version {
    return v.GaugeOpts.DeprecatedVersion
}

func (v *GaugeVec) InitializeMetric() {
    v.GaugeVec = prometheus.NewGaugeVec(v.GaugeOpts.toPromGaugeOpts(), v.originalLabels)
}

func (v *GaugeVec) InitializeDeprecatedMetric() {
    v.GaugeOpts.MarkDeprecated()
    v.InitializeMetric()
}

// todo:        There is a problem with the underlying method call here. Prometheus behavior
// todo(cont):  here actually results in the creation of a new metric if a metric with the unique
// todo(cont):  label values is not found in the underlying stored metricMap.
// todo(cont):  For reference: https://github.com/prometheus/client_golang/blob/master/prometheus/counter.go#L148-L177

// todo(cont):  This means if we opt to disable a metric by NOT registering it, then we would also
// todo(cont):  need to ensure that this no-opts and does not create a new metric, otherwise disabling
// todo(cont):  a metric which causes a memory leak would still continue to leak memory.
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
        return &noopGaugeVec, nil
    }
    vec, err := v.GaugeVec.CurryWith(labels)
    if vec != nil {
        return vec, err
    }
    return nil, err
}

func (v *GaugeVec) MustCurryWith(labels prometheus.Labels) *prometheus.GaugeVec {
    if !v.IsRegistered() {
        return &noopGaugeVec
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
