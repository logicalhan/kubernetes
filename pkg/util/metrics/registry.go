package metrics

import (
	"fmt"
	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
	"k8s.io/klog"
)

var (
	DefaultGlobalRegistry                         = NewKubeRegistry()
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
			switch c.(type) {
			case *KubeCounter:
				originalOpts := c.(*KubeCounter).originalOpts
				newOpts := CounterOpts{
					Namespace: originalOpts.Namespace,
					Name: originalOpts.Name,
					Subsystem: originalOpts.Subsystem,
					ConstLabels: originalOpts.ConstLabels,
					Help: fmt.Sprintf("(Deprecated since %v) %v", c.GetDeprecatedVersion(), originalOpts.Help),
					DeprecatedVersion: c.GetDeprecatedVersion(),
				}
				newCounter := NewCounter(newOpts)
				metrics = append(metrics, newCounter)

			default: // TODO: handle other cases
				metrics = append(metrics, c)
			}
			continue
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
