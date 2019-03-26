/*
Copyright 2014 The Kubernetes Authors.

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

package routes

import (
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"io"
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	apimetrics "k8s.io/apiserver/pkg/endpoints/metrics"
	"k8s.io/apiserver/pkg/server/mux"
	etcdmetrics "k8s.io/apiserver/pkg/storage/etcd/metrics"
	clientmetrics "k8s.io/kubernetes/pkg/util/prometheusclientgo"      // load all the prometheus client-go plugins
	workqueuemetrics "k8s.io/kubernetes/pkg/util/workqueue/prometheus" // for workqueue metric registration
	versionmetrics "k8s.io/kubernetes/pkg/version/prometheus"          // for version metric registration
)

// DefaultMetrics installs the default prometheus metrics handler
type DefaultMetrics struct{}

// Install adds the DefaultMetrics handler
func (m DefaultMetrics) Install(c *mux.PathRecorderMux) {
	r := prometheus.NewRegistry()
	register(r)
	c.Handle("/metrics", promhttp.HandlerFor(r, promhttp.HandlerOpts{ErrorHandling: promhttp.ContinueOnError}))
}

// MetricsWithReset install the prometheus metrics handler extended with support for the DELETE method
// which resets the metrics.
type MetricsWithReset struct{}

// Install adds the MetricsWithReset handler
func (m MetricsWithReset) Install(c *mux.PathRecorderMux) {
	r := prometheus.NewRegistry()
	register(r)
	defaultMetricsHandler := promhttp.HandlerFor(r, promhttp.HandlerOpts{ErrorHandling: promhttp.ContinueOnError}).ServeHTTP
	c.HandleFunc("/metrics", func(w http.ResponseWriter, req *http.Request) {
		if req.Method == "DELETE" {
			apimetrics.Reset()
			etcdmetrics.Reset()
			io.WriteString(w, "metrics reset\n")
			return
		}
		defaultMetricsHandler(w, req)
	})
}

// register apiserver and etcd metrics
func register(registerer prometheus.Registerer) {
	apimetrics.Register(registerer)
	etcdmetrics.Register(registerer)
	versionmetrics.Register(registerer)
	workqueuemetrics.Register(registerer)
	clientmetrics.Register(registerer)

}
