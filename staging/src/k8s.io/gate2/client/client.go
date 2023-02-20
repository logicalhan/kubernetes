package client

import (
	api "k8s.io/api/features/v1alpha1"
	"k8s.io/client-go/informers"
	informerfeatures "k8s.io/client-go/informers/features"
	"k8s.io/client-go/kubernetes"
	featureclient "k8s.io/client-go/kubernetes/typed/features/v1alpha1"
	"k8s.io/component-base/featuregate"
)

// FeatureGateClient may be NewFeatureGateClient()ed and used immediately, but
// calls won't return until Init() is called and it is able to read features
// from the control plane.
//
// TODO: decide what to do when values change
// TODO: report back uses
// TODO: this should be a drop-in replacement for featuregate.FeatureGate
type FeatureGateClient struct {
	class string

	factoryStopch chan<- struct{}
	informer      informerfeatures.FeatureInformer
	client        featureclient.FeaturesV1alpha1Interface
	handler       cached.ResourceEventHandlerRegistration

	initialized sync.WaitGroup

	lock    sync.Mutex
	checked map[featuregate.Feature]usageInfo
}

type usageInfo struct {
	checked                 bool
	requiresRestartToChange bool
	latchedValue            bool // only latched when requiresRestart is true
}

func NewFeatureGateClient(class string) *FeatureGateClient {
	c := &FeatureGateClient{
		class:   class,
		checked: map[featuregate.Feature]usageInfo{},
	}
	c.initialized.Add(1)
	return c
}

// Waits until inititialization and returns whether or not `f` is an enabled
// feature.
func (c *FeatureGateClient) Enabled(f featuregate.Feature) bool {
	return c.enabled(f, true)
}

// Waits until inititialization and returns whether or not `f` is an enabled
// feature. The "Stateless" suffix indicate that this feature is checked at
// every use and may change dynamically. If the feature requires initialization
// or internal state such as caching, you should call the ordinary "Enabled"
// version of this function.
//
// The name is chosen so that it's safe to drop FeatureGateClient in where the
// existing featuregate.FeatureGate object is used (so that existing calls
// don't accidentally declare they are safe for dynamic changing).
func (c *FeatureGateClient) EnabledStateless(f featuregate.Feature) bool {
	return c.enabled(f, false)
}

func (c *FeatureGateClient) enabled(f featuregate.Feature, requireRestart bool) bool {
	c.initialized.Wait()

	c.lock.Lock()
	defer c.lock.Unlock()

	// If this has previously been checked by something that can't change
	// without a restart, we have to return whatever value that got
	// forever.
	if prev, ok := c.checked[f]; ok {
		if prev.requiresRestartToChange {
			return prev.latchedValue
		}
	}

	// Check the current (cached) value.
	var current bool
	if cachedF, err := c.informer.Lister().Get(string(f)); err == nil {
		current = cachedF.Status.State == api.FeatureStateOn ||
			cachedF.Status.State == api.FeatureStateTurningOn
	} else {
		current = false
	}

	if prev, ok := c.checked[f]; ok {
		if !prev.requiresRestartToChange && requireRestart {

			prev.latchedValue = current
			prev.requiresRestartToChange = true
			c.checked[f] = prev
		}
	} else {
		c.checked[f] = usageInfo{
			checked:                 true,
			requiresRestartToChange: requireRestart,
			latchedValue:            current,
		}
	}

	return current
}

func (c *FeatureGateClient) Init(client kubernetes.Interface) error {
	defer c.initialized.Done()

	stopCh := make(chan struct{})
	c.factoryStopCh = stopCh

	// make our own factory to limit to the namespace for this class (TODO is this right?)
	factory := informers.NewSharedInformerFactoryWithOptions(client, 0, WithNamespace(c.class))
	c.informer = factory.Features().V1alpha1().Features()
	c.client = client.Feature().V1alpha1()
	c.handler = c.informer.Informer().AddEventHandler(cache.ResourceHandlerFuncs{})
	go factory.Start(stopCh)

	cache.WaitForNamedCacheSync("feature_client", wait.NeverStop, c.informer.HasSynced())

	// Start workers to write back usage when needed.

}
