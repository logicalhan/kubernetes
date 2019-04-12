package metrics

import "github.com/prometheus/client_golang/prometheus"

type Opts struct {
    Namespace         string
    Subsystem         string
    Name              string
    Help              string
    ConstLabels       prometheus.Labels // todo: don't refer to prometheus specifically in our external API
    DeprecatedVersion *Version
}