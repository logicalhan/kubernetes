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

package clientgo

import (
	k8smetrics "k8s.io/component-base/metrics"
	"k8s.io/component-base/metrics/prometheus/clientgo/leaderelection"
	"k8s.io/component-base/metrics/prometheus/restclient"
	"k8s.io/component-base/metrics/prometheus/workqueue"
)

// RegisterMetrics registers all the dependency injected client-go metrics to a KubeRegistry
func RegisterMetrics(registry k8smetrics.KubeRegistry) {
	leaderelection.RegisterMetric(registry)
	workqueue.RegisterMetrics(registry)
	restclient.RegisterMetrics(registry)
}
