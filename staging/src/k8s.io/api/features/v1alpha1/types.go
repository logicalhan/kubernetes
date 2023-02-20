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

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +genclient
// +genclient:nonNamespaced
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// Feature represents a single feature that can be turned on or off.
type Feature struct {
	metav1.TypeMeta `json:",inline"`
	// The name is <status.Class>/<status.Name>.
	metav1.ObjectMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`

	// `spec` contains the user-settable attributes of the feature.
	Spec FeatureSpec `json:"spec" protobuf:"bytes,2,opt,name=spec"`

	// `status` declares the facts about this feature.
	Status FeatureStatus `json:"status" protobuf:"bytes,3,opt,name=status"`
}

type FeatureSpec struct {
	// `desired` is the desired state of the feature, if it differs from
	// `status.default`. This field may be set by users if the feature is
	// set to accept dynamic values.
	// +optional
	Desired *FeatureEnablement `json:"desired" protobuf:"bytes,1,opt,name=desired,casttype=FeatureEnablement"`
}

type FeatureEnablement string

const (
	FeatureEnablementEnabled  FeatureEnablement = "Enabled"
	FeatureEnablementDisabled FeatureEnablement = "Disabled"
)

type FeatureStatus struct {
	// `class` is the class of feature. "kube-system" indicates
	// the feature is about the host cluster. Third parties may use a
	// domain name if they wish to reuse this system for their own
	// canarying. `class` should match `metadata.namespace`.
	Class string `json:"class" protobuf:"bytes,1,opt,name=class"`

	// `name` is the name of the feature. It should match `metadata.name`.
	Name string `json:"name" protobuf:"bytes,2,opt,name=name"`

	// `stability` declares the stability of this feature in the current
	// installed version.
	Stability StabilityLevel `json:"stability" protobuf:"bytes,3,opt,name=stability,casttype=StabilityLevel"`

	// `version` declares the version of software currently providing this feature.
	Version string `json:"version"`

	// `default` indicates whether the system thinks the field should be
	// enabled by default.
	Default FeatureEnablement `json:"default" protobuf:"bytes,4,opt,name=default,casttype=FeatureEnablement"`

	// `state` declares the current state of the feature ("On", "Off", or a
	// transitional state).
	State FeatureState `json:"state" protobuf:"bytes,5,opt,name=state,casttype=FeatureState"`

	// `uses` is for clients to report their use of the feature. Clients
	// should report their use only if there is not already an entry
	// matching their condition; this keeps this field very low-qps no
	// matter how many clients there are. The server may occasionally clear
	// non-desired-state uses and wait for clients to add them back, as a
	// way of telling whether a state transition has completed or not. When
	// that happens, `useEvaluationTime` will be set to a time in the
	// future; clients have until then to record their use.
	Uses []FeatureUse `json:"uses"`

	// `useEvaluationTime`, if set, is set to a time in the future when the
	// server will evaluate the contents of `uses` and do something with
	// that information, such as complete a state transition.
	// +optional
	UseEvaluationTime *metav1.Time `json:"useEvaluationTime"`
}

// FeatureUse records facts about a single process's use of a feature.
type FeatureUse struct {
	// `reportTime` is the time at which this report is made.
	ReportTime metav1.Time `json:"reportTime"`

	// `version` is the version of the process making the report.
	Version string `json:"version"`

	// `enabled` is the local state of this feature for the process making
	// the report. It may differ from the current desired state. Clients
	// may not report transitional states.
	Enabled FeatureEnablement `json:"enabled"`
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
	// The feature is on. Relevant actors in the system know it is on.
	FeatureStateOn FeatureState = "On"
	// The feature is off. Relevant actors in the system know it is off.
	FeatureStateOff FeatureState = "Off"
	// The system is turning the feature on. New uses of the feature are
	// permitted. Some processes may have observed an "off" value on start
	// up and need to restart.
	FeatureStateTurningOn FeatureState = "TurningOn"
	// The system is turning the feature off. New uses of the feature are
	// NOT permitted, but existing uses have not been cleaned up, and
	// processes that observed the feature being on when they started may
	// still be running and need to be restarted.
	FeatureStateTurningOff FeatureState = "TurningOff"
)

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// A list of Features.
type FeatureList struct {
	metav1.TypeMeta `json:",inline"`
	// Standard list metadata.
	// More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#metadata
	// +optional
	metav1.ListMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`
	// Items holds a list of StorageVersion
	Items []Feature `json:"items" protobuf:"bytes,2,rep,name=items"`
}
