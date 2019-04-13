package metrics

import (
    "time"
    "github.com/prometheus/client_golang/prometheus"
)

type SummaryOpts struct {
    Namespace string
    Subsystem string
    Name      string
    Help string
    ConstLabels prometheus.Labels
    Objectives map[float64]float64
    MaxAge time.Duration
    AgeBuckets uint32
    BufCap uint32
    DeprecatedVersion *Version
}

func (c SummaryOpts) toPromSummaryOpts() prometheus.SummaryOpts {
    return prometheus.SummaryOpts{
        Namespace:   c.Namespace,
        Subsystem:   c.Subsystem,
        Name:        c.Name,
        Help:        c.Help,
        ConstLabels: c.ConstLabels,
        Objectives: c.Objectives,
        MaxAge: c.MaxAge,
        AgeBuckets: c.AgeBuckets,
        BufCap: c.BufCap,
    }
}

type KubeSummary struct {
    PromSummary prometheus.Summary
    Version *Version
}

func (h KubeSummary) Describe(ch chan<- *prometheus.Desc) {
    h.PromSummary.Describe(ch)
}

func (h KubeSummary) Collect(ch chan<- prometheus.Metric) {
    h.PromSummary.Collect(ch)
}

func (h KubeSummary) Observe(v float64) {
    h.PromSummary.Observe(v)
}

type SummaryVec struct {
    vec              *prometheus.SummaryVec
    DeprecatedVersion *Version
}

func NewSummary(opts SummaryOpts) KubeSummary {
    s := prometheus.NewSummary(opts.toPromSummaryOpts())
    return KubeSummary{s, opts.DeprecatedVersion}
}

func NewSummaryVec(opts SummaryOpts, labels []string) *SummaryVec {
    vec := prometheus.NewSummaryVec(opts.toPromSummaryOpts(), labels)
    return &SummaryVec{vec, opts.DeprecatedVersion}
}

func (h *SummaryVec) GetMetricWithLabelValues(lvs ...string) (prometheus.Observer, error) {
    return h.vec.GetMetricWithLabelValues(lvs...)
}

func (h *SummaryVec) GetMetricWith(labels prometheus.Labels) (prometheus.Observer, error) {
    return h.vec.GetMetricWith(labels)
}

func (h *SummaryVec) With(labels prometheus.Labels) prometheus.Observer {
    return h.vec.With(labels)
}

func (h *SummaryVec) CurryWith(labels prometheus.Labels) (prometheus.ObserverVec, error) {
    return h.vec.CurryWith(labels)
}

func (h *SummaryVec) MustCurryWith(labels prometheus.Labels) prometheus.ObserverVec {
    vec, err := h.CurryWith(labels)
    if err != nil {
        panic(err)
    }
    return vec
}

// Describe implements Collector. It will send exactly one Desc to the provided
// channel.
func (h *SummaryVec) Describe(ch chan<- *prometheus.Desc) {
    h.vec.Describe(ch)
}

// Collect implements Collector.
func (h *SummaryVec) Collect(ch chan<- prometheus.Metric) {
    h.vec.Collect(ch)
}

// Reset deletes all metrics in this vector.
func (h *SummaryVec) Reset() {
    h.vec.Reset()
}