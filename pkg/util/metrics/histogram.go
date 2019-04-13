package metrics

import "github.com/prometheus/client_golang/prometheus"

type HistogramOpts struct {
    Namespace string
    Subsystem string
    Name      string
    Help string
    ConstLabels prometheus.Labels
    Buckets []float64
    DeprecatedVersion *Version
}

func (c HistogramOpts) toPromHistogramOpts() prometheus.HistogramOpts {
    return prometheus.HistogramOpts{
        Namespace:   c.Namespace,
        Subsystem:   c.Subsystem,
        Name:        c.Name,
        Help:        c.Help,
        ConstLabels: c.ConstLabels,
        Buckets:        c.Buckets,
    }
}

type KubeHistogram struct {
    PromHistogram prometheus.Histogram
    Version *Version
}

type HistogramVec struct {
    vec              *prometheus.HistogramVec
    DeprecatedVersion *Version
}

func NewHistogram(opts HistogramOpts) KubeHistogram {
    h := prometheus.NewHistogram(opts.toPromHistogramOpts())
    return KubeHistogram{h, opts.DeprecatedVersion}
}

func NewHistogramVec(opts HistogramOpts, labels []string) *HistogramVec {
    hVec := prometheus.NewHistogramVec(opts.toPromHistogramOpts(), labels)
    return &HistogramVec{hVec, opts.DeprecatedVersion}
}

func (h *HistogramVec) GetMetricWithLabelValues(lvs ...string) (prometheus.Observer, error) {
    return h.vec.GetMetricWithLabelValues(lvs...)
}

func (h *HistogramVec) GetMetricWith(labels prometheus.Labels) (prometheus.Observer, error) {
    return h.vec.GetMetricWith(labels)
}

func (h *HistogramVec) With(labels prometheus.Labels) prometheus.Observer {
    return h.vec.With(labels)
}

func (h *HistogramVec) CurryWith(labels prometheus.Labels) (prometheus.ObserverVec, error) {
    return h.vec.CurryWith(labels)
}

func (h *HistogramVec) MustCurryWith(labels prometheus.Labels) prometheus.ObserverVec {
    vec, err := h.CurryWith(labels)
    if err != nil {
        panic(err)
    }
    return vec
}

// Describe implements Collector. It will send exactly one Desc to the provided
// channel.
func (h *HistogramVec) Describe(ch chan<- *prometheus.Desc) {
    h.vec.Describe(ch)
}

// Collect implements Collector.
func (h *HistogramVec) Collect(ch chan<- prometheus.Metric) {
    h.vec.Collect(ch)
}

// Reset deletes all metrics in this vector.
func (h *HistogramVec) Reset() {
    h.vec.Reset()
}
