/*
Copyright 2019 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package metrics

import (
	"github.com/blang/semver"
	"github.com/prometheus/client_golang/prometheus"
)

// kubeGauge is our internal representation for our wrapping struct around prometheus
// gauges. kubeGauge implements both KubeCollector and KubeGauge.
type kubeGauge struct {
	KubeGauge
	*GaugeOpts
	lazyMetric
	selfCollector
}

// NewGauge returns an object which satisfies the KubeCollector and KubeGauge interfaces.
// However, the object returned will not measure anything unless the collector is first
// registered, since the metric is lazily instantiated.
func NewGauge(opts *GaugeOpts) *kubeGauge {
	// todo: handle defaulting better
	if opts.StabilityLevel == "" {
		opts.StabilityLevel = ALPHA
	}
	kc := &kubeGauge{
		GaugeOpts: opts,
		lazyMetric:  lazyMetric{},
	}
	kc.setPrometheusGauge(noop)
	kc.lazyInit(kc)
	return kc
}

// setPrometheusGauge sets the underlying KubeGauge object, i.e. the thing that does the measurement.
func (c *kubeGauge) setPrometheusGauge(gauge prometheus.Gauge) {
	c.KubeGauge = gauge
	c.initSelfCollection(gauge)
}

// DeprecatedVersion returns a pointer to the Version or nil
func (c *kubeGauge) DeprecatedVersion() *semver.Version {
	return c.GaugeOpts.DeprecatedVersion
}

// initializeMetric invocation creates the actual underlying Gauge. Until this method is called
// the underlying gauge is a no-op.
func (c *kubeGauge) initializeMetric() {
	c.GaugeOpts.annotateStabilityLevel()
	// this actually creates the underlying prometheus gauge.
	c.setPrometheusGauge(prometheus.NewGauge(c.GaugeOpts.toPromGaugeOpts()))
}

// initializeDeprecatedMetric invocation creates the actual (but deprecated) Gauge. Until this method
// is called the underlying gauge is a no-op.
func (c *kubeGauge) initializeDeprecatedMetric() {
	c.GaugeOpts.markDeprecated()
	c.initializeMetric()
}

// kubeGaugeVec is the internal representation of our wrapping struct around prometheus
// gaugeVecs. kubeGaugeVec implements both KubeCollector and KubeGaugeVec.
type kubeGaugeVec struct {
	*prometheus.GaugeVec
	*GaugeOpts
	lazyMetric
	originalLabels []string
}

// NewGaugeVec returns an object which satisfies the KubeCollector and KubeGaugeVec interfaces.
// However, the object returned will not measure anything unless the collector is first
// registered, since the metric is lazily instantiated.
func NewGaugeVec(opts *GaugeOpts, labels []string) *kubeGaugeVec {
	// todo: handle defaulting better
	if opts.StabilityLevel == "" {
		opts.StabilityLevel = ALPHA
	}
	cv := &kubeGaugeVec{
		GaugeVec:     noopGaugeVec,
		GaugeOpts:    opts,
		originalLabels: labels,
		lazyMetric:     lazyMetric{},
	}
	cv.lazyInit(cv)
	return cv
}

// DeprecatedVersion returns a pointer to the Version or nil
func (v *kubeGaugeVec) DeprecatedVersion() *semver.Version {
	return v.GaugeOpts.DeprecatedVersion
}

// initializeMetric invocation creates the actual underlying GaugeVec. Until this method is called
// the underlying gaugeVec is a no-op.
func (v *kubeGaugeVec) initializeMetric() {
	v.GaugeOpts.annotateStabilityLevel()
	v.GaugeVec = prometheus.NewGaugeVec(v.GaugeOpts.toPromGaugeOpts(), v.originalLabels)
}

// initializeDeprecatedMetric invocation creates the actual (but deprecated) GaugeVec. Until this method is called
// the underlying gaugeVec is a no-op.
func (v *kubeGaugeVec) initializeDeprecatedMetric() {
	v.GaugeOpts.markDeprecated()
	v.initializeMetric()
}

// Default Prometheus behavior actually results in the creation of a new metric
// if a metric with the unique label values is not found in the underlying stored metricMap.
// This means  that if this function is called but the underlying metric is not registered
// (which means it will never be exposed externally nor consumed), the metric will exist in memory
// for perpetuity (i.e. throughout application lifecycle).
//
// For reference: https://github.com/prometheus/client_golang/blob/v0.9.2/prometheus/gauge.go#L190-L208
//
// This method returns a no-op metric if the metric is not actually created/registered, avoiding that
// memory leak.
func (v *kubeGaugeVec) WithLabelValues(lvs ...string) KubeGauge {
	if !v.IsCreated() {
		return noop // return no-op gauge
	}
	return v.GaugeVec.WithLabelValues(lvs...)
}

func (v *kubeGaugeVec) With(labels prometheus.Labels) KubeGauge {
	if !v.IsCreated() {
		return noop // return no-op gauge
	}
	return v.GaugeVec.With(labels)
}
