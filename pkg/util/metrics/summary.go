package metrics

import (
	"github.com/blang/semver"
	"github.com/prometheus/client_golang/prometheus"
)



// This is our wrapper for prometheus.Summary metrics.
// We store the options the metric was defined with in order
// to defer initialization until the metric is actually registered.
//
// DEPRECATED: as per the metrics overhaul KEP
type kubeSummary struct {
	KubeObserver
	*SummaryOpts
	lazyMetric
	selfCollector
}

// NewSummary returns an object which is Summary-like. However, nothing
// will be measured until the summary is registered somewhere.
//
// DEPRECATED: as per the metrics overhaul KEP
func NewSummary(opts *SummaryOpts) *kubeSummary {
	// todo: handle defaulting better
	if opts.StabilityLevel == "" {
		opts.StabilityLevel = ALPHA
	}
	s := &kubeSummary{
		SummaryOpts:        opts,
		lazyMetric: lazyMetric{},
	}
	s.setPrometheusSummary(noopMetric{})
	s.lazyInit(s)
	return s
}

// setPrometheusSummary sets the underlying KubeGauge object, i.e. the thing that does the measurement.
func (h *kubeSummary) setPrometheusSummary(summary prometheus.Summary) {
	h.KubeObserver = summary
	h.initSelfCollection(summary)
}

// DeprecatedVersion returns a pointer to the Version or nil
func (s *kubeSummary) DeprecatedVersion() *semver.Version {
	return s.SummaryOpts.DeprecatedVersion
}

// initializeMetric invokes the actual prometheus.Summary object instantiation
// and stores a reference to it
func (s *kubeSummary) initializeMetric() {
	s.SummaryOpts.annotateStabilityLevel()
	// this actually creates the underlying prometheus gauge.
	s.setPrometheusSummary(prometheus.NewSummary(s.SummaryOpts.toPromSummaryOpts()))
}

// initializeDeprecatedMetric invokes the actual prometheus.Summary object instantiation
// but modifies the Help description prior to object instantiation.
func (s *kubeSummary) initializeDeprecatedMetric() {
	s.SummaryOpts.markDeprecated()
	s.initializeMetric()
}

// DEPRECATED: as per the metrics overhaul KEP
type kubeSummaryVec struct {
	*prometheus.SummaryVec
	*SummaryOpts
	lazyMetric
	originalLabels []string
}

// DEPRECATED: as per the metrics overhaul KEP
func NewSummaryVec(opts *SummaryOpts, labels []string) *kubeSummaryVec {
	// todo: handle defaulting better
	if opts.StabilityLevel == "" {
		opts.StabilityLevel = ALPHA
	}
	v := &kubeSummaryVec{
		SummaryOpts:          opts,
		originalLabels:       labels,
		lazyMetric: lazyMetric{},
	}
	v.lazyInit(v)
	return v
}

func (v *kubeSummaryVec) DeprecatedVersion() *semver.Version {
	return v.SummaryOpts.DeprecatedVersion
}

func (v *kubeSummaryVec) initializeMetric() {
	v.SummaryOpts.annotateStabilityLevel()
	v.SummaryVec = prometheus.NewSummaryVec(v.SummaryOpts.toPromSummaryOpts(), v.originalLabels)
}

func (v *kubeSummaryVec) initializeDeprecatedMetric() {
	v.SummaryOpts.markDeprecated()
	v.initializeMetric()
}

func (v *kubeSummaryVec) WithLabelValues(lvs ...string) prometheus.Observer {
	if !v.IsCreated() {
		return noop
	}
	return v.SummaryVec.WithLabelValues(lvs...)
}

func (v *kubeSummaryVec) With(labels prometheus.Labels) prometheus.Observer {
	if !v.IsCreated() {
		return noop
	}
	return v.SummaryVec.With(labels)
}