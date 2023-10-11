/*
Copyright 2016 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package featuregate

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/blang/semver/v4"
	"github.com/spf13/pflag"

	"k8s.io/apimachinery/pkg/util/naming"
	featuremetrics "k8s.io/component-base/metrics/prometheus/feature"
	"k8s.io/component-base/version"
	"k8s.io/klog/v2"
)

type Feature string

const (
	flagName = "feature-gates"

	// allAlphaGate is a global toggle for alpha features. Per-feature key
	// values override the default set by allAlphaGate. Examples:
	//   AllAlpha=false,NewFeature=true  will result in newFeature=true
	//   AllAlpha=true,NewFeature=false  will result in newFeature=false
	allAlphaGate Feature = "AllAlpha"

	// allBetaGate is a global toggle for beta features. Per-feature key
	// values override the default set by allBetaGate. Examples:
	//   AllBeta=false,NewFeature=true  will result in NewFeature=true
	//   AllBeta=true,NewFeature=false  will result in NewFeature=false
	allBetaGate Feature = "AllBeta"
)

var (
	// The generic features.
	defaultFeatures = map[Feature]FeatureSpec{
		allAlphaGate: {Default: false, PreRelease: Alpha},
		allBetaGate:  {Default: false, PreRelease: Beta},
	}

	// Special handling for a few gates.
	specialFeatures = map[Feature]func(known map[Feature]FeatureSpec, enabled map[Feature]bool, val bool, cVer *semver.Version){
		allAlphaGate: setUnsetAlphaGates,
		allBetaGate:  setUnsetBetaGates,
	}

	ErrMajorAndMinorOnly = errors.New("version string must only contain major and minor")
)

type FeatureSpec struct {
	// Default is the default enablement state for the feature
	Default bool
	// DefaultEnabledVersion is the Kubernetes version that this feature is default enabled.
	DefaultEnabledVersion *string
	// LockToDefault indicates that the feature is locked to its default and cannot be changed
	LockToDefault bool
	// LockToDefaultVersion indicates from which version the feature is locked to its default and cannot be changed
	LockToDefaultVersion *string
	// PreRelease indicates the current maturity level of the feature
	PreRelease prerelease
	// PromotionVersionMap indicates the k8s version this feature was promoted
	PromotionVersionMap PromotionVersionMapping
	// DeprecatedVersion indicates the k8s version this feature was promoted
	DeprecatedVersion *string
}

type PromotionVersionMapping map[prerelease]string

func (fs *FeatureSpec) lockToDefaultAt(compatibilityVer *semver.Version) bool {
	if compatibilityVer == nil {
		return fs.LockToDefault
	}
	if fs.LockToDefaultVersion != nil {
		lockToDefaultVer, err := deriveVersion(*fs.LockToDefaultVersion)
		if err != nil {
			return false
		}
		return compatibilityVer.LT(*lockToDefaultVer)
	}
	return false
}

func (fs *FeatureSpec) defaultAt(compatibilityVer *semver.Version) bool {
	if compatibilityVer == nil {
		return fs.Default
	}
	if fs.LockToDefaultVersion != nil {
		lockToDefaultVer, err := deriveVersion(*fs.LockToDefaultVersion)
		if err != nil {
			return false
		}
		// this means our default value is valid
		if compatibilityVer.GTE(*lockToDefaultVer) {
			return fs.Default
		}
	}
	if fs.DefaultEnabledVersion != nil {
		defaultEnabledVer, err := deriveVersion(*fs.DefaultEnabledVersion)
		if err != nil {
			return false
		}
		if compatibilityVer.GTE(*defaultEnabledVer) {
			return true
		}
	}
	return false
}

func (fs *FeatureSpec) prereleaseAt(compatibilityVer *semver.Version) prerelease {
	if compatibilityVer == nil {
		return fs.PreRelease
	}
	versions := []semver.Version{}
	for _, rawVer := range fs.PromotionVersionMap {
		pv, err := deriveVersion(rawVer)
		if err == nil {
			versions = append(versions, *pv)
		}
	}
	semver.Sort(versions)
	if len(versions) == 0 {
		return preAlpha
	} else if len(versions) == 1 {
		if versions[0].GT(*compatibilityVer) {
			return preAlpha
		} else {
			return Alpha
		}
	} else if len(versions) == 2 {
		if versions[1].GT(*compatibilityVer) {
			return Alpha
		} else {
			return Beta
		}
	}
	if versions[2].GT(*compatibilityVer) {
		return Beta
	} else {
		return GA
	}
}

type prerelease string

const (
	// unexported because this stage is for internal use only
	preAlpha = prerelease("PRE-ALPHA")
	// Values for PreRelease.
	Alpha = prerelease("ALPHA")
	Beta  = prerelease("BETA")
	GA    = prerelease("")

	// Deprecated
	Deprecated = prerelease("DEPRECATED")
)

// FeatureGate indicates whether a given feature is enabled or not
type FeatureGate interface {
	// Enabled returns true if the key is enabled.
	Enabled(key Feature) bool
	// KnownFeatures returns a slice of strings describing the FeatureGate's known features.
	KnownFeatures() []string
	// DeepCopy returns a deep copy of the FeatureGate object, such that gates can be
	// set on the copy without mutating the original. This is useful for validating
	// config against potential feature gate changes before committing those changes.
	DeepCopy() MutableFeatureGate
}

// MutableFeatureGate parses and stores flag gates for known features from
// a string like feature1=true,feature2=false,...
type MutableFeatureGate interface {
	FeatureGate

	// AddFlag adds a flag for setting global feature gates to the specified FlagSet.
	AddFlag(fs *pflag.FlagSet)
	// Set parses and stores flag gates for known features
	// from a string like feature1=true,feature2=false,...
	Set(value string) error
	// SetFromMap stores flag gates for known features from a map[string]bool or returns an error
	SetFromMap(m map[string]bool) error
	// Add adds features to the featureGate.
	Add(features map[Feature]FeatureSpec) error
	// GetAll returns a copy of the map of known feature names to feature specs.
	GetAll() map[Feature]FeatureSpec
	// AddMetrics adds feature enablement metrics
	AddMetrics()
}

// featureGate implements FeatureGate as well as pflag.Value for flag parsing.
type featureGate struct {
	featureGateName string

	special map[Feature]func(map[Feature]FeatureSpec, map[Feature]bool, bool, *semver.Version)

	// lock guards writes to known, enabled, and reads/writes of closed
	lock sync.Mutex
	// known holds a map[Feature]FeatureSpec
	known *atomic.Value
	// enabled holds a map[Feature]bool
	enabled *atomic.Value
	// closed is set to true when AddFlag is called, and prevents subsequent calls to Add
	closed bool

	compatibilityVersion *semver.Version

	binaryVersion semver.Version
}

func setUnsetAlphaGates(known map[Feature]FeatureSpec, enabled map[Feature]bool, val bool, cVer *semver.Version) {
	for k, v := range known {
		if v.prereleaseAt(cVer) == Alpha {
			if _, found := enabled[k]; !found {
				enabled[k] = val
			}
		}
	}
}

func setUnsetBetaGates(known map[Feature]FeatureSpec, enabled map[Feature]bool, val bool, cVer *semver.Version) {
	for k, v := range known {
		if v.prereleaseAt(cVer) == Beta {
			if _, found := enabled[k]; !found {
				enabled[k] = val
			}
		}
	}
}

// Set, String, and Type implement pflag.Value
var _ pflag.Value = &featureGate{}

// internalPackages are packages that ignored when creating a name for featureGates. These packages are in the common
// call chains, so they'd be unhelpful as names.
var internalPackages = []string{"k8s.io/component-base/featuregate/feature_gate.go"}

func NewFeatureGateForTest(binaryVersion string) *featureGate {
	return newFeatureGateWithBinaryVersion(binaryVersion)
}

func newFeatureGateWithBinaryVersion(binaryVersion string) *featureGate {
	known := map[Feature]FeatureSpec{}
	for k, v := range defaultFeatures {
		known[k] = v
	}

	knownValue := &atomic.Value{}
	knownValue.Store(known)

	enabled := map[Feature]bool{}
	enabledValue := &atomic.Value{}
	enabledValue.Store(enabled)

	f := &featureGate{
		featureGateName: naming.GetNameFromCallsite(internalPackages...),
		known:           knownValue,
		special:         specialFeatures,
		enabled:         enabledValue,
	}
	bVer, err := deriveVersion(binaryVersion)
	if err != nil {
		panic("no binary version detected, can't initialize feature flags")
	}
	f.binaryVersion = *bVer
	return f
}

func NewFeatureGate() *featureGate {
	return newFeatureGateWithBinaryVersion(fmt.Sprintf("%s.%s", version.Get().Major, version.Get().Minor))
}

// Set parses a string of the form "key1=value1,key2=value2,..." into a
// map[string]bool of known keys or returns an error.
func (f *featureGate) Set(value string) error {
	m := make(map[string]bool)
	for _, s := range strings.Split(value, ",") {
		if len(s) == 0 {
			continue
		}
		arr := strings.SplitN(s, "=", 2)
		k := strings.TrimSpace(arr[0])
		if len(arr) != 2 {
			return fmt.Errorf("missing bool value for %s", k)
		}
		v := strings.TrimSpace(arr[1])
		boolValue, err := strconv.ParseBool(v)
		//println("1", k, v)
		if err != nil {
			return fmt.Errorf("invalid value of %s=%s, err: %v", k, v, err)
		}
		m[k] = boolValue
	}
	return f.SetFromMap(m)
}

// SetFromMap stores flag gates for known features from a map[string]bool or returns an error
func (f *featureGate) SetFromMap(m map[string]bool) error {
	f.lock.Lock()
	defer f.lock.Unlock()

	// Copy existing state
	known := map[Feature]FeatureSpec{}
	for k, v := range f.known.Load().(map[Feature]FeatureSpec) {
		known[k] = v
	}
	enabled := map[Feature]bool{}
	for k, v := range f.enabled.Load().(map[Feature]bool) {
		enabled[k] = v
	}

	for k, v := range m {
		key := Feature(k)
		featureSpec, ok := known[key]
		if !ok {
			return fmt.Errorf("unrecognized feature gate: %s", k)
		}
		if featureSpec.lockToDefaultAt(f.compatibilityVersion) && featureSpec.defaultAt(f.compatibilityVersion) != v {
			println("ahahahahahaha")
			return fmt.Errorf("cannot set feature gate %v to %v, feature is locked to %v", k, v, featureSpec.Default)
		}
		enabled[key] = v
		// Handle "special" features like "all alpha gates"
		if fn, found := f.special[key]; found {
			fn(known, enabled, v, f.compatibilityVersion)
		}

		if featureSpec.prereleaseAt(f.compatibilityVersion) == Deprecated {
			klog.Warningf("Setting deprecated feature gate %s=%t. It will be removed in a future release.", k, v)
		} else if featureSpec.prereleaseAt(f.compatibilityVersion) == GA {
			klog.Warningf("Setting GA feature gate %s=%t. It will be removed in a future release.", k, v)
		}
	}

	// Persist changes
	f.known.Store(known)
	f.enabled.Store(enabled)

	klog.V(1).Infof("feature gates: %v", f.enabled)
	return nil
}

// String returns a string containing all enabled feature gates, formatted as "key1=value1,key2=value2,...".
func (f *featureGate) String() string {
	pairs := []string{}
	for k, v := range f.enabled.Load().(map[Feature]bool) {
		pairs = append(pairs, fmt.Sprintf("%s=%t", k, v))
	}
	sort.Strings(pairs)
	return strings.Join(pairs, ",")
}

func (f *featureGate) Type() string {
	return "mapStringBool"
}

// Add adds features to the featureGate.
func (f *featureGate) Add(features map[Feature]FeatureSpec) error {
	f.lock.Lock()
	defer f.lock.Unlock()

	if f.closed {
		return fmt.Errorf("cannot add a feature gate after adding it to the flag set")
	}

	// Copy existing state
	known := map[Feature]FeatureSpec{}
	for k, v := range f.known.Load().(map[Feature]FeatureSpec) {
		known[k] = v
	}

	for name, spec := range features {
		if existingSpec, found := known[name]; found {
			if reflect.DeepEqual(existingSpec, spec) {
				continue
			}
			return fmt.Errorf("feature gate %q with different spec already exists: %v", name, existingSpec)
		}
		known[name] = spec
	}

	// Persist updated state
	f.known.Store(known)

	return nil
}

// GetAll returns a copy of the map of known feature names to feature specs.
func (f *featureGate) GetAll() map[Feature]FeatureSpec {
	retval := map[Feature]FeatureSpec{}
	for k, v := range f.known.Load().(map[Feature]FeatureSpec) {
		retval[k] = v
	}
	return retval
}

func (f *featureGate) SetCompatibilityVersion(v string) error {
	if len(strings.Split(v, ".")) != 2 {
		return ErrMajorAndMinorOnly
	}
	withPatch := fmt.Sprintf("%s.0", v)
	compatibilityVersion := semver.MustParse(withPatch)
	f.compatibilityVersion = &compatibilityVersion
	return nil
}

// Enabled returns true if the key is enabled.  If the key is not known, this call will panic.
func (f *featureGate) Enabled(key Feature) bool {
	// fallback to default behavior, since we don't have compatibility version set
	if v, ok := f.enabled.Load().(map[Feature]bool)[key]; ok {
		return v
	}
	if v, ok := f.known.Load().(map[Feature]FeatureSpec)[key]; ok {
		dVer := v.DeprecatedVersion
		if dVer == nil {
			return v.defaultAt(f.compatibilityVersion)
		}
		parsedDVer, err := deriveVersion(*dVer)
		if err != nil {
			return v.defaultAt(f.compatibilityVersion)
		}
		if f.compatibilityVersion != nil {
			if parsedDVer.LTE(*f.compatibilityVersion) {
				return false
			}
			return v.defaultAt(f.compatibilityVersion)
		}
		return false
	}
	panic(fmt.Errorf("feature %q is not registered in FeatureGate %q", key, f.featureGateName))
}

func deriveVersion(ver string) (*semver.Version, error) {
	if len(strings.Split(ver, ".")) != 2 {
		return nil, ErrMajorAndMinorOnly
	}
	withPatch := fmt.Sprintf("%s.0", ver)
	compatibilityVersion := semver.MustParse(withPatch)
	return &compatibilityVersion, nil
}

// AddFlag adds a flag for setting global feature gates to the specified FlagSet.
func (f *featureGate) AddFlag(fs *pflag.FlagSet) {
	f.lock.Lock()
	// TODO(mtaufen): Shouldn't we just close it on the first Set/SetFromMap instead?
	// Not all components expose a feature gates flag using this AddFlag method, and
	// in the future, all components will completely stop exposing a feature gates flag,
	// in favor of componentconfig.
	f.closed = true
	f.lock.Unlock()

	known := f.KnownFeatures()
	fs.Var(f, flagName, ""+
		"A set of key=value pairs that describe feature gates for alpha/experimental features. "+
		"Options are:\n"+strings.Join(known, "\n"))
}

func (f *featureGate) AddMetrics() {
	for feature, featureSpec := range f.GetAll() {
		featuremetrics.RecordFeatureInfo(context.Background(), string(feature), string(featureSpec.PreRelease), f.Enabled(feature))
	}
}

// KnownFeatures returns a slice of strings describing the FeatureGate's known features.
// Deprecated and GA features are hidden from the list.
func (f *featureGate) KnownFeatures() []string {
	var known []string
	for k, v := range f.known.Load().(map[Feature]FeatureSpec) {
		if v.PreRelease == GA || v.PreRelease == Deprecated {
			continue
		}
		known = append(known, fmt.Sprintf("%s=true|false (%s - default=%t)", k, v.PreRelease, v.Default))
	}
	sort.Strings(known)
	return known
}

// DeepCopy returns a deep copy of the FeatureGate object, such that gates can be
// set on the copy without mutating the original. This is useful for validating
// config against potential feature gate changes before committing those changes.
func (f *featureGate) DeepCopy() MutableFeatureGate {
	// Copy existing state.
	known := map[Feature]FeatureSpec{}
	for k, v := range f.known.Load().(map[Feature]FeatureSpec) {
		known[k] = v
	}
	enabled := map[Feature]bool{}
	for k, v := range f.enabled.Load().(map[Feature]bool) {
		enabled[k] = v
	}

	// Store copied state in new atomics.
	knownValue := &atomic.Value{}
	knownValue.Store(known)
	enabledValue := &atomic.Value{}
	enabledValue.Store(enabled)

	// Construct a new featureGate around the copied state.
	// Note that specialFeatures is treated as immutable by convention,
	// and we maintain the value of f.closed across the copy.
	return &featureGate{
		special: specialFeatures,
		known:   knownValue,
		enabled: enabledValue,
		closed:  f.closed,
	}
}
