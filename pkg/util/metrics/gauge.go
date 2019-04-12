package metrics

import (
    "github.com/prometheus/client_golang/prometheus"
)

// GaugeOpts is an alias for Opts. See there for doc comments.
type GaugeOpts Opts

// convenience function to allow easy transformation to the prometheus
// counterpart. This will do more once we have a proper label abstraction
func (c GaugeOpts) toPromGaugeOpts() prometheus.GaugeOpts {
    return prometheus.GaugeOpts{
        Namespace:   c.Namespace,
        Subsystem:   c.Subsystem,
        Name:        c.Name,
        Help:        c.Help,
        ConstLabels: c.ConstLabels}
}

type KubeGauge struct {
    PromGauge prometheus.Gauge
    Version *Version
}

type GaugeVec struct {
    gVec              *prometheus.GaugeVec
    DeprecatedVersion *Version
}


func NewGauge(opts GaugeOpts) KubeGauge {
    g := prometheus.NewGauge(opts.toPromGaugeOpts())
    return KubeGauge{g, opts.DeprecatedVersion}
}

func NewGaugeVec(opts GaugeOpts, labels []string) *GaugeVec {
    gVec := prometheus.NewGaugeVec(opts.toPromGaugeOpts(), labels)
    return &GaugeVec{gVec, opts.DeprecatedVersion}
}

func (v *GaugeVec) GetMetricWithLabelValues(lvs ...string) (prometheus.Gauge, error) {
    return v.gVec.GetMetricWithLabelValues(lvs...)
}

func (v *GaugeVec) GetMetricWith(labels prometheus.Labels) (prometheus.Gauge, error) {
    return v.gVec.GetMetricWith(labels)
}

func (v *GaugeVec) With(labels prometheus.Labels) prometheus.Gauge {
    return v.gVec.With(labels)
}

func (v *GaugeVec) CurryWith(labels prometheus.Labels) (*GaugeVec, error) {
    vec, err := v.gVec.CurryWith(labels)
    if vec != nil {
        return &GaugeVec{vec, v.DeprecatedVersion}, err
    }
    return nil, err
}

func (v *GaugeVec) MustCurryWith(labels prometheus.Labels) *GaugeVec {
    vec, err := v.CurryWith(labels)
    if err != nil {
        panic(err)
    }
    return vec
}

// Describe implements Collector. It will send exactly one Desc to the provided
// channel.
func (v *GaugeVec) Describe(ch chan<- *prometheus.Desc) {
    v.gVec.Describe(ch)
}

// Collect implements Collector.
func (v *GaugeVec) Collect(ch chan<- prometheus.Metric) {
    v.gVec.Collect(ch)
}

// Reset deletes all metrics in this vector.
func (m *GaugeVec) Reset() {
    m.gVec.Reset()
}
