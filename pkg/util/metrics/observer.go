package metrics

import "github.com/prometheus/client_golang/prometheus"


type DeprecatableObserver struct {
	prometheus.Observer
	DeprecatedVersion *Version
}

type DeprecatableObserverVec struct {
	prometheus.ObserverVec
	DeprecatedVersion *Version
}
