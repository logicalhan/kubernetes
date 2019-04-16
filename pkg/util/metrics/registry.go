package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
	"k8s.io/klog"
)

var (
	DefaultGlobalRegistry                         = NewKubeRegistry(MustParseGeneric("1.15.0"))
)

type PromRegistry interface {
	prometheus.Registerer
	prometheus.Gatherer
}

type KubeRegistry struct {
	registry PromRegistry
	version  *Version
}

func (kr *KubeRegistry) Register(collector DeprecatableCollector) error {
	return kr.registry.Register(collector)
}

func (kr *KubeRegistry) MustRegister(cs ...DeprecatableCollector) {
	metrics := make([]prometheus.Collector, 0)
	for _, c := range cs {
		if c.GetDeprecatedVersion() != nil && c.GetDeprecatedVersion().compareInternal(kr.version) < 0 {
			klog.Warningf("This metric has been deprecated for more than one release, hiding.")
			continue
		} else if c.GetDeprecatedVersion() != nil && c.GetDeprecatedVersion().compareInternal(kr.version) == 0 {
			c.MarkDeprecated()
			metrics = append(metrics, c)
		} else {
			metrics = append(metrics, c)
		}
	}
	kr.registry.MustRegister(metrics...)
}

func (kr *KubeRegistry) Unregister(collector DeprecatableCollector) bool {
	return kr.registry.Unregister(collector)
}

func (r *KubeRegistry) Gather() ([]*dto.MetricFamily, error) {
	return r.registry.Gather()
}

// NewRegistry creates a new vanilla Registry without any Collectors
// pre-registered.
func NewKubeRegistry(version *Version) *KubeRegistry {

	return &KubeRegistry{
		prometheus.NewRegistry(),
		version,
	}
}
