package metrics

import (
	"github.com/blang/semver"
	"github.com/prometheus/client_golang/prometheus"
)

// This is our wrapper function for prometheus counters
// we store the options the metric was defined with in order
// to defer initialization until actual metric registration.
type kubeHistogram struct {
	KubeObserver
	*HistogramOpts
	lazyMetric
	selfCollector
}

// NewHistogram returns an object which is Histogram-like. However, nothing
// will be measured until the histogram is registered somewhere.
func NewHistogram(opts *HistogramOpts) *kubeHistogram {
	// todo: handle defaulting better
	if opts.StabilityLevel == "" {
		opts.StabilityLevel = ALPHA
	}
	h := &kubeHistogram{
		HistogramOpts:        opts,
		lazyMetric: lazyMetric{},
	}
	h.setPrometheusHistogram(noopMetric{})
	h.lazyInit(h)
	return h
}

// setPrometheusHistogram sets the underlying KubeGauge object, i.e. the thing that does the measurement.
func (h *kubeHistogram) setPrometheusHistogram(histogram prometheus.Histogram) {
	h.KubeObserver = histogram
	h.initSelfCollection(histogram)
}

// DeprecatedVersion returns a pointer to the Version or nil
func (h *kubeHistogram) DeprecatedVersion() *semver.Version {
	return h.HistogramOpts.DeprecatedVersion
}

// InitializeMetric invokes the actual prometheus.Histogram object instantiation
// and stores a reference to it
func (h *kubeHistogram) initializeMetric() {
	h.HistogramOpts.annotateStabilityLevel()
	// this actually creates the underlying prometheus gauge.
	h.setPrometheusHistogram(prometheus.NewHistogram(h.HistogramOpts.toPromHistogramOpts()))
}

// InitializeMetric invokes the actual prometheus.Histogram object instantiation
// but modifies the Help description prior to object instantiation.
func (h *kubeHistogram) initializeDeprecatedMetric() {
	h.HistogramOpts.markDeprecated()
	h.initializeMetric()
}

type kubeHistogramVec struct {
	*prometheus.HistogramVec
	*HistogramOpts
	lazyMetric
	originalLabels []string
}

func NewHistogramVec(opts *HistogramOpts, labels []string) *kubeHistogramVec {
	// todo: handle defaulting better
	if opts.StabilityLevel == "" {
		opts.StabilityLevel = ALPHA
	}
	v := &kubeHistogramVec{
		HistogramVec: noopHistogramVec,
		HistogramOpts:        opts,
		originalLabels:       labels,
		lazyMetric:     lazyMetric{},
	}
	v.lazyInit(v)
	return v
}

// functions for kubeHistogramVec
func (v *kubeHistogramVec) DeprecatedVersion() *semver.Version {
	return v.HistogramOpts.DeprecatedVersion
}

func (v *kubeHistogramVec) initializeMetric() {
	v.HistogramOpts.annotateStabilityLevel()
	v.HistogramVec = prometheus.NewHistogramVec(v.HistogramOpts.toPromHistogramOpts(), v.originalLabels)
}

func (v *kubeHistogramVec) initializeDeprecatedMetric() {
	v.HistogramOpts.markDeprecated()
	v.initializeMetric()
}


func (v *kubeHistogramVec) WithLabelValues(lvs ...string) prometheus.Observer {
	if !v.IsCreated() {
		return noop
	}
	return v.HistogramVec.WithLabelValues(lvs...)
}

func (v *kubeHistogramVec) With(labels prometheus.Labels) prometheus.Observer {
	if !v.IsCreated() {
		return noop
	}
	return v.HistogramVec.With(labels)
}
