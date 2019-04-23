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
	"github.com/prometheus/client_golang/prometheus"

	dto "github.com/prometheus/client_model/go"
)

// This file contains a series of interfaces which we explicitly define for
// integrating with prometheus. We redefine the interfaces explicitly here
// so that we can prevent breakage if methods are ever added to prometheus
// variants of them.

/**
 * Collector defines a subset of prometheus.Collector interface methods
 */
type Collector interface {
	Describe(chan<- *prometheus.Desc)
	Collect(chan<- prometheus.Metric)
}

/**
 * Metric defines a subset of prometheus.Metric interface methods
 */
type Metric interface {
	Desc() *prometheus.Desc
	Write(*dto.Metric) error
}

// Counter is a Metric that represents a single numerical value that only ever
// goes up. That implies that it cannot be used to count items whose number can
// also go down, e.g. the number of currently running goroutines. Those
// "counters" are represented by Gauges.
//
// This interface defines a subset of the interface provided by prometheus.Counter
type KubeCounter interface {
	Inc()
	Add(float64)
}

type KubeCounterVec interface {
	WithLabelValues(...string) KubeCounter
	With(prometheus.Labels) KubeCounter
}

type PromRegistry interface {
	Register(prometheus.Collector) error
	MustRegister(...prometheus.Collector)
	Unregister(prometheus.Collector) bool
	Gather() ([]*dto.MetricFamily, error)
}
