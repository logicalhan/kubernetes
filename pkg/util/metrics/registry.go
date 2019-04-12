package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
)

var (
	defaultRegistry                         = NewKubeRegistry()
	DefaultRegisterer prometheus.Registerer = defaultRegistry
	DefaultGatherer   prometheus.Gatherer   = defaultRegistry
)

type PromRegistry interface {
	prometheus.Registerer
	prometheus.Gatherer
}

type KubeRegistry struct {
	registry PromRegistry
	version  *Version
}

func (kr *KubeRegistry) Register(collector prometheus.Collector) error {
	return kr.registry.Register(collector)
}

func (kr *KubeRegistry) MustRegister(cs ...prometheus.Collector) {
	kr.registry.MustRegister(cs...)
}

func (kr *KubeRegistry) Unregister(collector prometheus.Collector) bool {
	return kr.registry.Unregister(collector)
}

func (r *KubeRegistry) Gather() ([]*dto.MetricFamily, error) {
	return r.registry.Gather()
}

// NewRegistry creates a new vanilla Registry without any Collectors
// pre-registered.
func NewKubeRegistry() *KubeRegistry {
	// todo: hardcode version for now
	v, err := parse("1.15.0", true)
	if err != nil {
		panic(err)
	}
	return &KubeRegistry{
		prometheus.NewRegistry(),
		v,
	}
}
