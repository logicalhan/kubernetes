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
    vec              *prometheus.GaugeVec
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

func (g *GaugeVec) GetMetricWithLabelValues(lvs ...string) (prometheus.Gauge, error) {
    return g.vec.GetMetricWithLabelValues(lvs...)
}

func (g *GaugeVec) GetMetricWith(labels prometheus.Labels) (prometheus.Gauge, error) {
    return g.vec.GetMetricWith(labels)
}

func (g *GaugeVec) With(labels prometheus.Labels) prometheus.Gauge {
    return g.vec.With(labels)
}

func (g *GaugeVec) CurryWith(labels prometheus.Labels) (*GaugeVec, error) {
    vec, err := g.vec.CurryWith(labels)
    if vec != nil {
        return &GaugeVec{vec, g.DeprecatedVersion}, err
    }
    return nil, err
}

func (g *GaugeVec) MustCurryWith(labels prometheus.Labels) *GaugeVec {
    vec, err := g.CurryWith(labels)
    if err != nil {
        panic(err)
    }
    return vec
}

// Describe implements Collector. It will send exactly one Desc to the provided
// channel.
func (g *GaugeVec) Describe(ch chan<- *prometheus.Desc) {
    g.vec.Describe(ch)
}

// Collect implements Collector.
func (g *GaugeVec) Collect(ch chan<- prometheus.Metric) {
    g.vec.Collect(ch)
}

// Reset deletes all metrics in this vector.
func (g *GaugeVec) Reset() {
    g.vec.Reset()
}
