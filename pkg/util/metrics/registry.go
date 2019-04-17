package metrics

import (
	"github.com/blang/semver"
	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
	"k8s.io/klog"
)

var (
	DefaultGlobalRegistry = NewKubeRegistry(semver.MustParse("1.15.0"))
)

type PromRegistry interface {
	prometheus.Registerer
	prometheus.Gatherer
}

type KubeRegistry struct {
	registry PromRegistry
	version  semver.Version
}

func (kr *KubeRegistry) Register(collector KubeCollector) error {
	return kr.registry.Register(collector)
}

func (kr *KubeRegistry) MustRegister(cs ...KubeCollector) {
	metrics := make([]prometheus.Collector, 0, len(cs))
	for _, c := range cs {
		if c.GetDeprecatedVersion() != nil && c.GetDeprecatedVersion().LT(kr.version) {
			klog.Warningf("This metric has been deprecated for more than one release, hiding.")
			continue
		}

		if c.GetDeprecatedVersion() != nil && c.GetDeprecatedVersion().EQ(kr.version) {
			c.CreateMetric(IsDeprecated)
			metrics = append(metrics, c)
		} else {
			c.CreateMetric(NotDeprecated)
			metrics = append(metrics, c)
		}
	}
	kr.registry.MustRegister(metrics...)
}

func (kr *KubeRegistry) Unregister(collector KubeCollector) bool {
	return kr.registry.Unregister(collector)
}

func (kr *KubeRegistry) Gather() ([]*dto.MetricFamily, error) {
	return kr.registry.Gather()
}

// NewRegistry creates a new vanilla Registry without any Collectors
// pre-registered.
func NewKubeRegistry(version semver.Version) *KubeRegistry {

	return &KubeRegistry{
		prometheus.NewRegistry(),
		version,
	}
}
