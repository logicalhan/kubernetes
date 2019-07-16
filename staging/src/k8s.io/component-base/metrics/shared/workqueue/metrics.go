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

package prometheus

import (
	"fmt"
	"k8s.io/client-go/util/workqueue"
	"k8s.io/klog"
	"sync"

	"github.com/prometheus/client_golang/prometheus"
	k8smetrics "k8s.io/component-base/metrics"
)

// Package prometheus sets the workqueue DefaultMetricsFactory to produce
// prometheus metrics. To use this package, you just have to import it.

// Metrics subsystem and keys used by the workqueue.
const (
	WorkQueueSubsystem         = "workqueue"
	DepthKey                   = "depth"
	AddsKey                    = "adds_total"
	QueueLatencyKey            = "queue_duration_seconds"
	WorkDurationKey            = "work_duration_seconds"
	UnfinishedWorkKey          = "unfinished_work_seconds"
	LongestRunningProcessorKey = "longest_running_processor_seconds"
	RetriesKey                 = "retries_total"
)

var (
	depth = k8smetrics.NewGaugeVec(&k8smetrics.GaugeOpts{
		Subsystem: WorkQueueSubsystem,
		Name:      DepthKey,
		Help:      "Current depth of workqueue",
	}, []string{"name"})

	adds = k8smetrics.NewCounterVec(&k8smetrics.CounterOpts{
		Subsystem: WorkQueueSubsystem,
		Name:      AddsKey,
		Help:      "Total number of adds handled by workqueue",
	}, []string{"name"})

	latency = k8smetrics.NewHistogramVec(&k8smetrics.HistogramOpts{
		Subsystem: WorkQueueSubsystem,
		Name:      QueueLatencyKey,
		Help:      "How long in seconds an item stays in workqueue before being requested.",
		Buckets:   prometheus.ExponentialBuckets(10e-9, 10, 10),
	}, []string{"name"})

	workDuration = k8smetrics.NewHistogramVec(&k8smetrics.HistogramOpts{
		Subsystem: WorkQueueSubsystem,
		Name:      WorkDurationKey,
		Help:      "How long in seconds processing an item from workqueue takes.",
		Buckets:   prometheus.ExponentialBuckets(10e-9, 10, 10),
	}, []string{"name"})
	unfinished = k8smetrics.NewGaugeVec(&k8smetrics.GaugeOpts{
		Subsystem: WorkQueueSubsystem,
		Name:      UnfinishedWorkKey,
		Help: "How many seconds of work has done that " +
			"is in progress and hasn't been observed by work_duration. Large " +
			"values indicate stuck threads. One can deduce the number of stuck " +
			"threads by observing the rate at which this increases.",
	}, []string{"name"})
	longestRunningProcessor = k8smetrics.NewGaugeVec(&k8smetrics.GaugeOpts{
		Subsystem: WorkQueueSubsystem,
		Name:      LongestRunningProcessorKey,
		Help: "How many seconds has the longest running " +
			"processor for workqueue been running.",
	}, []string{"name"})
	retries = k8smetrics.NewCounterVec(&k8smetrics.CounterOpts{
		Subsystem: WorkQueueSubsystem,
		Name:      RetriesKey,
		Help:      "Total number of retries handled by workqueue",
	}, []string{"name"})

	metrics = []k8smetrics.Registerable{
		depth, adds, latency, workDuration, unfinished, longestRunningProcessor, retries,
	}
	// TODO: remove this in 1.17. These metrics are broken but they are also deprecated
	// let's actually get them to work correctly for one release before we yank them completely.
	mtx       sync.RWMutex // Protects metrics.
	gauges    = make(map[string]k8smetrics.GaugeMetric)
	counters  = make(map[string]workqueue.CounterMetric)
	summaries = make(map[string]workqueue.SummaryMetric)
)

type prometheusMetricsProvider struct {
	registry k8smetrics.KubeRegistry
}

func RegisterMetrics(registry k8smetrics.KubeRegistry) {
	for _, m := range metrics {
		registry.MustRegister(m)
	}
}

func (prometheusMetricsProvider) NewDepthMetric(name string) workqueue.GaugeMetric {
	return depth.WithLabelValues(name)
}

func (prometheusMetricsProvider) NewAddsMetric(name string) workqueue.CounterMetric {
	return adds.WithLabelValues(name)
}

func (prometheusMetricsProvider) NewLatencyMetric(name string) workqueue.HistogramMetric {
	return latency.WithLabelValues(name)
}

func (prometheusMetricsProvider) NewWorkDurationMetric(name string) workqueue.HistogramMetric {
	return workDuration.WithLabelValues(name)
}

func (prometheusMetricsProvider) NewUnfinishedWorkSecondsMetric(name string) workqueue.SettableGaugeMetric {
	return unfinished.WithLabelValues(name)
}

func (prometheusMetricsProvider) NewLongestRunningProcessorSecondsMetric(name string) workqueue.SettableGaugeMetric {
	return longestRunningProcessor.WithLabelValues(name)
}

func (prometheusMetricsProvider) NewRetriesMetric(name string) workqueue.CounterMetric {
	return retries.WithLabelValues(name)
}

func getOrCreateGaugeMetric(subsystem string, name string, help string, registry k8smetrics.KubeRegistry) k8smetrics.GaugeMetric {
	mtx.RLock()
	gm, ok := gauges[subsystem+"_"+name]
	mtx.RUnlock()
	if ok {
		return gm
	}
	gauge := k8smetrics.NewGauge(&k8smetrics.GaugeOpts{
		Subsystem: subsystem,
		Name:      name,
		Help:      fmt.Sprintf("(Deprecated) Current %s of workqueue: %s", name, subsystem),
	})
	mtx.Lock()
	gauges[subsystem+"_"+name] = gauge
	mtx.Unlock()
	if err := registry.Register(depth); err != nil {
		klog.Errorf("failed to register %s metric %v: %v", name, subsystem, err)
	}
	return gauge
}

func getOrCreateCounterMetric(subsystem string, name string, registry k8smetrics.KubeRegistry) workqueue.CounterMetric {
	mtx.RLock()
	c, ok := counters[subsystem+"_"+name]
	mtx.RUnlock()
	if ok {
		return c
	}
	counter := k8smetrics.NewCounter(&k8smetrics.CounterOpts{
		Subsystem: subsystem,
		Name:      name,
		Help:      fmt.Sprintf("(Deprecated) Total number of %s handled by workqueue: %s", name, subsystem),
	})
	mtx.Lock()
	counters[subsystem+"_"+name] = counter
	mtx.Unlock()
	if err := registry.Register(depth); err != nil {
		klog.Errorf("failed to register %s metric %v: %v", name, subsystem, err)
	}
	return counter
}

func getOrCreateSummaryMetric(subsystem string, name string, help string, registry k8smetrics.KubeRegistry) workqueue.SummaryMetric {
	mtx.RLock()
	s, ok := summaries[subsystem+"_"+name]
	mtx.RUnlock()
	if ok {
		return s
	}
	summary := k8smetrics.NewSummary(&k8smetrics.SummaryOpts{
		Subsystem: subsystem,
		Name:      name,
		Help:      help,
	})
	mtx.Lock()
	summaries[subsystem+"_"+name] = summary
	mtx.Unlock()
	if err := registry.Register(depth); err != nil {
		klog.Errorf("failed to register %s metric %v: %v", name, subsystem, err)
	}
	return summary
}

func (p *prometheusMetricsProvider) NewDeprecatedDepthMetric(name string) workqueue.GaugeMetric {
	return getOrCreateGaugeMetric(name, "depth", fmt.Sprintf("(Deprecated) Current depth of workqueue: %s", name), p.registry)
}

func (p *prometheusMetricsProvider) NewDeprecatedAddsMetric(name string) workqueue.CounterMetric {
	return getOrCreateCounterMetric(name, "adds", p.registry)
}

func (p *prometheusMetricsProvider) NewDeprecatedLatencyMetric(name string) workqueue.SummaryMetric {
	return getOrCreateSummaryMetric(name, "queue_latency", "(Deprecated) How long an item stays in workqueue"+name+" before being requested.", p.registry)
}

func (p *prometheusMetricsProvider) NewDeprecatedWorkDurationMetric(name string) workqueue.SummaryMetric {
	return getOrCreateSummaryMetric(name, "work_duration", "(Deprecated) How long processing an item from workqueue"+name+" takes.", p.registry)
}

func (p *prometheusMetricsProvider) NewDeprecatedUnfinishedWorkSecondsMetric(name string) workqueue.SettableGaugeMetric {
	help := "(Deprecated) How many seconds of work " + name + " has done that " +
		"is in progress and hasn't been observed by work_duration. Large " +
		"values indicate stuck threads. One can deduce the number of stuck " +
		"threads by observing the rate at which this increases."
	return getOrCreateGaugeMetric(name, "unfinished_work_seconds", help, p.registry)
}

func (p *prometheusMetricsProvider) NewDeprecatedLongestRunningProcessorMicrosecondsMetric(name string) workqueue.SettableGaugeMetric {
	help := "(Deprecated) How many microseconds has the longest running " +
		"processor for " + name + " been running."
	return getOrCreateGaugeMetric(name, "longest_running_processor_microseconds", help, p.registry)
}

func (p *prometheusMetricsProvider) NewDeprecatedRetriesMetric(name string) workqueue.CounterMetric {
	return getOrCreateCounterMetric(name, "retries", p.registry)
}
