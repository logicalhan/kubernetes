package metrics

import (
	"github.com/blang/semver"
	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
	"k8s.io/klog"
)

var (
	// todo: load the version dynamically at application boot.
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

// Register registers a collectable metric, but it uses a global registry.
func Register(c KubeCollector) error {
	return DefaultGlobalRegistry.Register(c)
}

// MustRegister works like Register but registers any number of
// Collectors and panics upon the first registration that causes an
// error.
func MustRegister(cs ...KubeCollector) {
	DefaultGlobalRegistry.MustRegister(cs...)
}

func (kr *KubeRegistry) Register(c KubeCollector) error {
	if c.GetDeprecatedVersion() != nil && c.GetDeprecatedVersion().LT(kr.version) {
		klog.Warningf("This metric has been deprecated for more than one release, hiding.")
	}

	if c.GetDeprecatedVersion() != nil && c.GetDeprecatedVersion().EQ(kr.version) {
		c.CreateMetric(IsDeprecated)
	} else {
		c.CreateMetric(NotDeprecated)
	}
	return kr.registry.Register(c)
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
