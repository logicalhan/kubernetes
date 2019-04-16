package metrics

import (
    "fmt"
    "github.com/prometheus/client_golang/prometheus"
)


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

func getDeprecatedHistogramOpts(originalOpts HistogramOpts) HistogramOpts {
    return HistogramOpts{
        Namespace: originalOpts.Namespace,
        Name: originalOpts.Name,
        Subsystem: originalOpts.Subsystem,
        ConstLabels: originalOpts.ConstLabels,
        Buckets:        originalOpts.Buckets,
        Help: fmt.Sprintf("(Deprecated since %v) %v", originalOpts.DeprecatedVersion, originalOpts.Help),
        DeprecatedVersion: originalOpts.DeprecatedVersion,
    }
}
type KubeHistogram struct {
    prometheus.Histogram
    HistogramOpts
    isDeprecated bool
}

func (h *KubeHistogram) MarkDeprecated() {
    h.isDeprecated = true
}

func (h *KubeHistogram) GetDeprecatedVersion() *Version {
    return h.HistogramOpts.DeprecatedVersion
}

func (h *KubeHistogram) GetDeprecatedMetric() DeprecatableCollector {
    return NewHistogram(getDeprecatedHistogramOpts(h.HistogramOpts))
}


func (h *KubeHistogram) Describe(ch chan<- *prometheus.Desc) {
    h.Histogram.Describe(ch)
}

func (h *KubeHistogram) Collect(ch chan<- prometheus.Metric) {
    h.Histogram.Collect(ch)
}

func (h *KubeHistogram) Observe(v float64) {
    h.Histogram.Observe(v)
}

type HistogramVec struct {
    *prometheus.HistogramVec
    HistogramOpts
    originalLabels []string
    isDeprecated bool
}

func NewHistogram(opts HistogramOpts) *KubeHistogram {
    h := prometheus.NewHistogram(opts.toPromHistogramOpts())
    return &KubeHistogram{Histogram: h, HistogramOpts: opts}
}

func NewHistogramVec(opts HistogramOpts, labels []string) *HistogramVec {
    hVec := prometheus.NewHistogramVec(opts.toPromHistogramOpts(), labels)
    return &HistogramVec{HistogramVec: hVec, HistogramOpts: opts, originalLabels: labels}
}
func (h *HistogramVec) MarkDeprecated() {
    h.isDeprecated = true
}
func (h *HistogramVec) GetDeprecatedMetric() DeprecatableCollector {
    newOpts := getDeprecatedHistogramOpts(h.HistogramOpts)
    return NewHistogramVec(newOpts, h.originalLabels)
}

func (h *HistogramVec) GetDeprecatedVersion() *Version {
    return h.HistogramOpts.DeprecatedVersion
}


func (h *HistogramVec) GetMetricWithLabelValues(lvs ...string) (DeprecatableObserver, error) {
    o, e := h.HistogramVec.GetMetricWithLabelValues(lvs...)
    return DeprecatableObserver{o, h.HistogramOpts.DeprecatedVersion}, e
}

func (h *HistogramVec) GetMetricWith(labels prometheus.Labels) (DeprecatableObserver, error) {
    o, e := h.HistogramVec.GetMetricWith(labels)
    return DeprecatableObserver{o, h.HistogramOpts.DeprecatedVersion}, e
}

func (h *HistogramVec) With(labels prometheus.Labels) DeprecatableObserver {
    return DeprecatableObserver{h.HistogramVec.With(labels), h.HistogramOpts.DeprecatedVersion}
}

func (h *HistogramVec) CurryWith(labels prometheus.Labels) (DeprecatableObserverVec, error) {
    ov, e := h.HistogramVec.CurryWith(labels)
    return DeprecatableObserverVec{ov, h.HistogramOpts.DeprecatedVersion}, e
}

func (h *HistogramVec) MustCurryWith(labels prometheus.Labels) DeprecatableObserverVec {
    vec, err := h.CurryWith(labels)
    if err != nil {
        panic(err)
    }
    return vec
}

// Describe implements Collector. It will send exactly one Desc to the provided
// channel.
func (h *HistogramVec) Describe(ch chan<- *prometheus.Desc) {
    h.HistogramVec.Describe(ch)
}

// Collect implements Collector.
func (h *HistogramVec) Collect(ch chan<- prometheus.Metric) {
    h.HistogramVec.Collect(ch)
}

// Reset deletes all metrics in this vector.
func (h *HistogramVec) Reset() {
    h.HistogramVec.Reset()
}
