package metrics

import (
	"fmt"
	"github.com/blang/semver"
	"github.com/prometheus/client_golang/prometheus"
	"sync"
	"time"
)

type SummaryOpts struct {
	Namespace         string
	Subsystem         string
	Name              string
	Help              string
	ConstLabels       prometheus.Labels
	Objectives        map[float64]float64
	MaxAge            time.Duration
	AgeBuckets        uint32
	BufCap            uint32
	DeprecatedVersion *semver.Version
	deprecateOnce     sync.Once
}

// Modify help description on the metric description.
func (o *SummaryOpts) MarkDeprecated() {
	o.deprecateOnce.Do(func() {
		o.Help = fmt.Sprintf("(Deprecated since %v) %v", o.DeprecatedVersion, o.Help)
	})
}

// convenience function to allow easy transformation to the prometheus
// counterpart. This will do more once we have a proper label abstraction
func (o SummaryOpts) toPromSummaryOpts() prometheus.SummaryOpts {
	return prometheus.SummaryOpts{
		Namespace:   o.Namespace,
		Subsystem:   o.Subsystem,
		Name:        o.Name,
		Help:        o.Help,
		ConstLabels: o.ConstLabels,
		Objectives:  o.Objectives,
		MaxAge:      o.MaxAge,
		AgeBuckets:  o.AgeBuckets,
		BufCap:      o.BufCap,
	}
}

// This is our wrapper for prometheus.Summary metrics.
// We store the options the metric was defined with in order
// to defer initialization until the metric is actually registered.
type KubeSummary struct {
	prometheus.Summary
	*SummaryOpts
	registerable
}

// NewSummary returns an object which is Summary-like. However, nothing
// will be measured until the summary is registered somewhere.
func NewSummary(opts SummaryOpts) *KubeSummary {
	kc := &KubeSummary{
		Summary:      noop,
		SummaryOpts:  &opts,
		registerable: registerable{},
	}
	// store a reference to ourselves so that we can defer registration
	kc.init(kc)
	return kc
}

// GetDeprecatedVersion, InitializeMetric, InitializeDeprecatedMetric are required to
// satisfy the KubeCollector interface.

// GetDeprecatedVersion returns a pointer to the Version or nil
func (s *KubeSummary) GetDeprecatedVersion() *semver.Version {
	return s.SummaryOpts.DeprecatedVersion
}

// InitializeMetric invokes the actual prometheus.Summary object instantiation
// and stores a reference to it
func (s *KubeSummary) InitializeMetric() {
	s.Summary = prometheus.NewSummary(s.SummaryOpts.toPromSummaryOpts())
}

// InitializeMetric invokes the actual prometheus.Summary object instantiation
// but modifies the Help description prior to object instantiation.
func (s *KubeSummary) InitializeDeprecatedMetric() {
	s.SummaryOpts.MarkDeprecated()
	s.InitializeMetric()
}

// Observe satisfies the prometheus.Observer interface. This call delegates to
// the underlying summary.
func (s *KubeSummary) Observe(v float64) {
	s.Summary.Observe(v)
}

// Describe and Collect satisfy the prometheus.Collector interface

// Describe delegates to the wrapped prometheus.Summary
func (s *KubeSummary) Describe(ch chan<- *prometheus.Desc) {
	s.Summary.Describe(ch)
}

// Collect delegates to the wrapped prometheus.Summary
func (s *KubeSummary) Collect(m chan<- prometheus.Metric) {
	s.Summary.Collect(m)
}

type SummaryVec struct {
	*prometheus.SummaryVec
	*SummaryOpts
	registerable
	originalLabels []string
}

func NewSummaryVec(opts SummaryOpts, labels []string) *SummaryVec {
	v := &SummaryVec{
		SummaryOpts:    &opts,
		originalLabels: labels,
		registerable:   registerable{},
	}
	v.init(v)
	return v
}

func (v *SummaryVec) GetDeprecatedVersion() *semver.Version {
	return v.SummaryOpts.DeprecatedVersion
}

func (v *SummaryVec) InitializeMetric() {
	v.SummaryVec = prometheus.NewSummaryVec(v.SummaryOpts.toPromSummaryOpts(), v.originalLabels)
}

func (v *SummaryVec) InitializeDeprecatedMetric() {
	v.SummaryOpts.MarkDeprecated()
	v.InitializeMetric()
}

// There is a problem with the underlying Prometheus method call here. Prometheus behavior
// actually results in the creation of a new metric if a metric with the unique
// label values is not found in the underlying stored metricMap.
// For reference: https://github.com/prometheus/client_golang/blob/master/prometheus/counter.go#L148-L177

// This means if we opt to disable a metric by NOT registering it, then we also need to ensure that
// we do not create new metrics when this is invoked, otherwise disabling a metric which
// causes a memory leak would still continue to leak memory (since we would be continuing to make
// arbitrary numbers of metrics for each unique label combo).
func (v *SummaryVec) GetMetricWithLabelValues(lvs ...string) (prometheus.Observer, error) {
	if !v.IsRegistered() {
		return noop, nil
	}
	return v.SummaryVec.GetMetricWithLabelValues(lvs...)
}

func (v *SummaryVec) GetMetricWith(labels prometheus.Labels) (prometheus.Observer, error) {
	if !v.IsRegistered() {
		return noop, nil
	}
	return v.SummaryVec.GetMetricWith(labels)
}

func (v *SummaryVec) WithLabelValues(lvs ...string) prometheus.Observer {
	if !v.IsRegistered() {
		return noop
	}
	return v.SummaryVec.WithLabelValues(lvs...)
}

func (v *SummaryVec) With(labels prometheus.Labels) prometheus.Observer {
	if !v.IsRegistered() {
		return noop
	}
	return v.SummaryVec.With(labels)
}

func (v *SummaryVec) CurryWith(labels prometheus.Labels) (prometheus.ObserverVec, error) {
	if !v.IsRegistered() {
		return noopObserverVec, nil
	}
	vec, err := v.SummaryVec.CurryWith(labels)
	if vec != nil {
		return vec, err
	}
	return nil, err
}

func (v *SummaryVec) MustCurryWith(labels prometheus.Labels) prometheus.ObserverVec {
	if !v.IsRegistered() {
		return noopObserverVec
	}
	vec, err := v.CurryWith(labels)
	if err != nil {
		panic(err)
	}
	return vec
}

// Describe implements Collector. It will send exactly one Desc to the provided
// channel.
func (v *SummaryVec) Describe(ch chan<- *prometheus.Desc) {
	v.SummaryVec.Describe(ch)
}

// Collect implements Collector.
func (v *SummaryVec) Collect(ch chan<- prometheus.Metric) {
	v.SummaryVec.Collect(ch)
}

// Reset deletes all metrics in this vector.
func (v *SummaryVec) Reset() {
	v.SummaryVec.Reset()
}
