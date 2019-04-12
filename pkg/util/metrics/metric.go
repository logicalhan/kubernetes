package metrics

import "github.com/prometheus/client_golang/prometheus"

type InnerMetric interface {
	prometheus.Collector
	prometheus.Metric
}

type KubeMetric struct {
	Metric  InnerMetric
	Version *Version
}
