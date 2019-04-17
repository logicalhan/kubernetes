package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"sync"
)

type Opts struct {
	Namespace         string
	Subsystem         string
	Name              string
	Help              string
	ConstLabels       prometheus.Labels
	DeprecatedVersion *Version
	deprecateOnce     sync.Once
}
