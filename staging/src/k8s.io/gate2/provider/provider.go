package provider

import (
	"k8s.io/component-base/featuregate"
	"k8s.io/gate2/host"
)

func NewClass(class string, commandLineGate featuregate.FeatureGate) *Class {
	return &Class{
		class: class,
		gate:  commandLineGate,
	}
}

type Class struct {
	class string
	gate  featuregate.FeatureGate

	providers []host.Provider
}

func tri[T any](b bool, whenTrue T, whenFalse T) T {
	if b {
		return whenTrue
	}
	return whenFalse
}

func (c *Class) AddProvider(f featuregate.Feature, ops Operations) *Class {
	on := gate.Enabled(f)
	// TODO: extend featuregate to provide these.
	dyn := gate.MayBeSetDynamically(f)
	def := gate.DefaultValue(f)

	fls := host.FeatureLocalState{
		Name:            string(f),
		Class:           c.class,
		Settability:     tri(dyn, host.Dynamic, host.Immutable),
		DefaultValue:    tri(def, host.FeatureOn, host.FeatureOff),
		ConfiguredValue: tri(on, host.FeatureOn, host.FeatureOff),
	}

	c.Providers = append(c.Providers, &Provider{
		localState: fls,
		operations: ops,
	})
	return c
}

func (c *Class) List() []Provider {
	return c.providers
}

// Operations collects the functions that feature authors may implement.
type Operations struct {
	// All operations must be idempotent, they may get called multiple times.
	// TODO: these could take a while; implement a progress system.
	PreVersionChangeFunc  func(oldVersion, newVersion string) error
	PostVersionChangeFunc func(oldVersion, newVersion string) error
	StateChangeFunc       func(desired FeatureState) (result FeatureState)

	ExerciseFunc func() error
}

// Make it easy to implement a feature provider.
type Provider struct {
	localState host.FeatureLocalState
	operations Operations
}

func (p *Provider) FeatureName() string {
	return p.localState.Name
}
func (p *Provider) FeatureClass() string {
	return p.localState.Class
}
func (p *Provider) Settings() FeatureLocalState {
	return p.localState
}
func (p *Provider) PreVersionChange(oldVersion, newVersion string) error {
	if p.operations.PreVersionChangeFunc != nil {
		return p.operations.PreVersionChangeFunc(oldVersion, newVersion)
	}
	return nil
}
func (p *Provider) PostVersionChange(oldVersion, newVersion string) error {
	if p.operations.PostVersionChangeFunc != nil {
		return p.operations.PostVersionChangeFunc(oldVersion, newVersion)
	}
	return nil
}
func (p *Provider) StateChange(desired FeatureState) (result FeatureState) {
	if p.operations.StateChangeFunc != nil {
		return p.operations.StateChangeFunc(desired)
	}
	return desired
}
func (p *Provider) ExerciseFeature() error {
	if p.ExerciseFunc != nil {
		return p.ExerciseFunc()
	}
	return errors.New("unimplemented")
}
