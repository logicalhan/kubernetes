package testgates

import (
	informerfeatures "k8s.io/client-go/informers/features"
	"k8s.io/component-base/featuregate"
	"k8s.io/gate2/host"
	"k8s.io/gate2/provider"
)

// Feature definition
const (
	MyTestFeature1 featuregate.Feature = "MyTestFeature1"
	MyTestFeature2 featuregate.Feature = "MyTestFeature2"
)

// Command-line parsing feature gate
var (
	DefaultMutableFeatureGate featuregate.MutableFeatureGate = featuregate.NewFeatureGate()
	DefaultFeatureGate        featuregate.FeatureGate        = DefaultMutableFeatureGate
)

// Operation Definitions
var (
	Feature1Ops = provider.Operations{}
	Feature2Ops = provider.Operations{}
)

const ClassName = "testfeatures"

// Assemble providers
var Providers = provider.NewClass(ClassName, DefaultFeatureGate).
	AddProvider(MyTestFeature1, Feature1Ops).
	AddProvider(MyTestFeature2, Feature2Ops).
	List()

// Construct a Host
func NewHost(hostname, version string) *host.Host {
	client := kubernetes.NewInClusterConfig()
	config := host.HostConfig{
		Remote:   client.Features(),
		Informer: informerfeatures.New(),
		Local:    Providers,
		Self: host.HostInfo{
			Class:   ClassName,
			Name:    hostname,
			Version: version,
		},
	}

	h, err := host.NewHost(config)
	if err != nil {
		panic("error constructing feature host")
	}
	return h
}

// In real code, every host runs in its own process (eg kube-apiserver); for
// test code we will run them all in the same process.
type hostSet map[host.HostInfo]*host.Host

func (hs hostSet) Add(h *host.Host) {
	if _, ok := hs[h.Config.Self]; ok {
		panic("host already exists!")
	}
	hs[h.Config.Self] = h
	hs.announce()
}

func (hs hostSet) Remove(h *host.Host) {
	delete(hs, h.Config.Self)
	hs.announce()
}

func (hs hostSet) announce() {
	// In real code, each process needs some way of knowing about the other
	// processes; we simulate that here.
	var remotes []host.RemoteHost
	for k := range hostSet {
		remotes = append(remotes, host.RemoteHost{k})
	}
	for _, h := range hostSet {
		h.OnRemoteHostChange(remotes)
	}
}

// The above are enough pieces to test the host/provider flow.
// TODO: write tests.
