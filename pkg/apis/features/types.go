/*
Copyright 2019 The Kubernetes Authors.

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

package features

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +genclient
// +genclient:nonNamespaced
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// Feature represents a single feature that can be turned on or off.
type Feature struct {
	metav1.TypeMeta
	// The name is <status.Class>/<status.Name>.
	metav1.ObjectMeta

	// `spec` contains the user-settable attributes of the feature.
	Spec FeatureSpec

	// `status` declares the facts about this feature.
	Status FeatureStatus
}

type FeatureSpec struct {
	// `desired` is the desired state of the feature, if it differs from
	// `status.default`. This field may be set by users.
	// +optional
	Desired *FeatureEnablement
}

type FeatureEnablement string

const (
	FeatureEnablementEnabled  FeatureEnablement = "Enabled"
	FeatureEnablementDisabled FeatureEnablement = "Disabled"
)

type FeatureStatus struct {
	// `class` is the class of feature. "cluster.kubernetes.io" indicates
	// the feature is about the host cluster. Third parties may use a
	// domain name if they wish to reuse this system for their own
	// canarying.
	Class string

	// `name` is the name of the feature.
	Name string

	// `stability` declares the stability of this feature in the current
	// installed version.
	Stability StabilityLevel

	// `default` indicates whether the system thinks the field should be
	// enabled by default.
	Default FeatureEnablement

	// `state` declares the current state of the feature ("On", "Off", or a
	// transitional state).
	State FeatureState
}

type StabilityLevel string

const (
	// Indicates that the feature doesn't exist at the current version
	// (i.e., there has been a downgrade).
	StabilityLevelUnavailable StabilityLevel = "Unavailable"

	// Indicates that the feature is available at alpha quality.
	StabilityLevelAlpha StabilityLevel = "Alpha"

	// Indicates that the feature is available at beta quality.
	StabilityLevelBeta StabilityLevel = "Beta"

	// Indicates that the feature is available at GA (stable / Generally
	// Available) quality.
	StabilityLevelGA StabilityLevel = "GA"

	// Indicates that the feature will be removed and permenantly turned
	// off in a future version. (I.e., we decided not to keep the feature.)
	StabilityLevelDeprecated StabilityLevel = "Deprecated"

	// Indicates that the feature will be removed and permenantly turned
	// on in a future version. (I.e., the feature is finished and is no longer optional.)
	StabilityLevelUniversal StabilityLevel = "Universal"
)

type FeatureState string

const (
	// The feature is on.
	FeatureStateOn FeatureState = "On"
	// The feature is off.
	FeatureStateOff FeatureState = "Off"
	// The system is turning the feature on.
	FeatureStateTurningOn FeatureState = "TurningOn"
	// The system is turning the feature off.
	FeatureStateTurningOff FeatureState = "TurningOff"
	// The system is e.g. migrating data between versions.
	FeatureStateMigrating FeatureState = "Migrating"
)

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// A list of Features.
type FeatureList struct {
	metav1.TypeMeta
	// Standard list metadata.
	// More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#metadata
	// +optional
	metav1.ListMeta
	// Items holds a list of StorageVersion
	Items []Feature
}
