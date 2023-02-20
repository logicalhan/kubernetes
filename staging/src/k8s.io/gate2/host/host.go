package host

import (
	"context"

	api "k8s.io/api/features/v1alpha1"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	featureinformers "k8s.io/client-go/informers/features/v1alpha1"
	featureclient "k8s.io/client-go/kubernetes/typed/features/v1alpha1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"
	"k8s.io/klog/v2"

	"k8s.io/component-base/featuregate"
)

// HostConfig contains the parameters needed to construct a feature host.
type HostConfig struct {
	Remote   featureclient.FeaturesV1alpha1Interface
	Informer featureinformers.FeaturesInformer
	Local    []Provider
	Self     HostInfo
}

// TODO: combine this with the api types
type FeatureState int

const (
	FeatureOff FeatureState = iota
	FeatureOn
	FeatureEnabling
	FeatureDisabling
)

type FeatureSettability int

const (
	Immutable FeatureSettability = iota
	Dynamic
)

type FeatureLocalState struct {
	Name         string
	Class        string
	Settability  FeatureSettability
	DefaultValue FeatureState
	Stability

	// This is used if Settability == Immutable; otherwise, this is used as
	// the initial state if there is no dynamic setting yet.
	ConfiguredValue FeatureState
}

// Provider is the interface feature providers supply for each feature.
type Provider interface {
	FeatureName() string
	FeatureClass() string
	Settings() FeatureLocalState

	// All operations must be idempotent, they may get called multiple times.
	PreVersionChange(oldVersion, newVersion string) error
	PostVersionChange(oldVersion, newVersion string) error
	StateChange(desired FeatureState) (result FeatureState)

	// Exercise the feature so it can be automatically tested.
	Exercise() error
}

// HostEvents is the set of things that can happen that Host needs to know
// about.
type HostEvents interface {
	OnRemoteHostChange(remotes []RemoteHost)
}

// HostInfo identifies a Host, either the local one or its remote siblings.
type HostInfo struct {
	// Class is the kind of host, e.g. "kube-apiserver". Each class of host
	// handles only its own gates.
	Class string
	// Name is the identity of a particular binary, e.g. "kube-apiserver
	// running on myhost123".
	Name string
	// Version is the version of this host.
	Version string
}

type RemoteHost struct {
	HostInfo
}

type Host struct {
	Config HostConfig

	handler cache.ResourceEventHandlerRegistration
	queue   workqueue.RateLimitingInterface

	featureToProvider map[string]Provider

	lock  sync.Mutex
	state combinedState
}

// NewHost produces a new feature host controller. The caller is responsible
// for sending HostEvents at the appropriate times.
func NewHost(cfg HostConfig) (*Host, error) {
	h := &Host{
		Config:            cfg,
		queue:             workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "featurehost"),
		featureToProvider: map[string]Provider{},
	}

	handler, err := h.Config.Informer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    h.onAddFeature,
		UpdateFunc: h.onUpdateFeature,
		DeleteFunc: h.onDeleteFeature,
	})

	if err != nil {
		return nil, err
	}

	h.handler = handler

	h.enqueueAllLocal()

	return h, nil
}

func (h *Host) enqueueAllLocal() {
	for _, p := range h.Config.Local {
		if p.FeatureClass() != h.Config.Self.Class {
			continue
		}
		h.featureToProvider[p.FeatureName()] = p
		h.queue.Add(p.FeatureName())
	}
}

func (h *Host) onAddFeature(obj interface{}) {
	f := obj.(*api.Feature)
	if f.Status.Class != h.Config.Self.Class {
		return
	}
	h.queue.Add(f.Status.Name)
}

func (h *Host) onUpdateFeature(oldObj, newObj interface{}) {
	f := newObj.(*api.Feature)
	if f.Status.Class != h.Config.Self.Class {
		return
	}
	h.queue.Add(f.Status.Name)
}

func (h *Host) onDeleteFeature(obj interface{}) {
	f, ok := newObj.(*api.Feature)
	if !ok {
		tombstone, ok := obj.(cache.DeletedFinalStateUnknown)
		if !ok {
		}
		f, ok = tombstone.Obj.(*api.Feature)
		if !ok {
		}
	}
	if f.Status.Class != h.Config.Self.Class {
		return
	}
	h.queue.Add(f.Status.Name)
}

// Run runs the ServiceAccountsController blocks until receiving signal from stopCh.
func (h *Host) Run(ctx context.Context, workers int) {
	defer utilruntime.HandleCrash()
	defer h.queue.ShutDown()
	defer h.Config.Informer.Informer().RemoveEventHandler(h.handler)

	klog.Infof("Starting feature host controller")

	defer klog.Infof("Shutting down feature host controller")
	if !cache.WaitForNamedCacheSync("featurehost", ctx.Done(), h.handler.HasSynced) {
		return
	}
	for i := 0; i < workers; i++ {
		go wait.UntilWithContext(ctx, h.runWorker, time.Second)
	}
	<-ctx.Done()
}

func (h *Host) runWorker(ctx context.Context) {
	for h.processNextWorkItem(ctx) {
	}
}

// processNextWorkItem deals with one key off the queue.  It returns false when it's time to quit.
func (h *Host) processNextWorkItem(ctx context.Context) bool {
	key, quit := h.queue.Get()
	if quit {
		return false
	}
	defer h.queue.Done(key)
	err := h.syncFeature(ctx, key.(string))
	if err == nil {
		h.queue.Forget(key)
		return true
	}
	utilruntime.HandleError(fmt.Errorf("syncing feature %q failed: %v", key, err))
	h.queue.AddRateLimited(key)
	return true
}

// OnRemoteHostChange should be called with the complete list of sibling hosts
// whenever that list changes.
func (h *Host) OnRemoteHostChange(remotes []RemoteHost) {
	h.changeHostState(h.summarizeRemotes(remotes))
}

func (h *Host) observeClusterVersion(version string) {
	h.changeClusterState(clusterState{latchedVersion: version})
}
