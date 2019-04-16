package metrics

import "github.com/prometheus/client_golang/prometheus"

type DeprecatableCollector interface {
	prometheus.Collector
	GetDeprecatedVersion() *Version
	MarkDeprecated()
}
