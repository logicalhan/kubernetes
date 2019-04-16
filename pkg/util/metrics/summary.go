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
    prometheus.Summary
    deprecatedVersion *Version
}

func (h KubeSummary) GetDeprecatedVersion() *Version {
    return h.deprecatedVersion
}

func (h KubeSummary) Describe(ch chan<- *prometheus.Desc) {
    h.Summary.Describe(ch)
}

func (h KubeSummary) Collect(ch chan<- prometheus.Metric) {
    h.Summary.Collect(ch)
}

func (h KubeSummary) Observe(v float64) {
    h.Summary.Observe(v)
}

type SummaryVec struct {
    *prometheus.SummaryVec
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

func (h *SummaryVec) GetMetricWithLabelValues(lvs ...string) (DeprecatableObserver, error) {
    o, e := h.SummaryVec.GetMetricWithLabelValues(lvs...)
    return DeprecatableObserver{o, h.DeprecatedVersion}, e
}

func (h *SummaryVec) GetMetricWith(labels prometheus.Labels) (DeprecatableObserver, error) {
    o, e := h.SummaryVec.GetMetricWith(labels)
    return DeprecatableObserver{o, h.DeprecatedVersion}, e
}

func (h *SummaryVec) With(labels prometheus.Labels) DeprecatableObserver {
    return DeprecatableObserver{h.SummaryVec.With(labels), h.DeprecatedVersion}
}

func (h *SummaryVec) CurryWith(labels prometheus.Labels) (DeprecatableObserverVec, error) {
    ov, e := h.SummaryVec.CurryWith(labels)
    return DeprecatableObserverVec{ov, h.DeprecatedVersion}, e
}

func (h *SummaryVec) MustCurryWith(labels prometheus.Labels) DeprecatableObserverVec {
    vec, err := h.CurryWith(labels)
    if err != nil {
        panic(err)
    }
    return vec
}

// Describe implements Collector. It will send exactly one Desc to the provided
// channel.
func (h *SummaryVec) Describe(ch chan<- *prometheus.Desc) {
    h.SummaryVec.Describe(ch)
}

// Collect implements Collector.
func (h *SummaryVec) Collect(ch chan<- prometheus.Metric) {
    h.SummaryVec.Collect(ch)
}

// Reset deletes all metrics in this vector.
func (h *SummaryVec) Reset() {
    h.SummaryVec.Reset()
}