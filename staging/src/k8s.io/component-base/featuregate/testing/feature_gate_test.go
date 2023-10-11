/*
Copyright 2023 The Kubernetes Authors.

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

package testing

import (
	gotest "testing"

	"k8s.io/component-base/featuregate"
)

func TestCompatibilityVersion(t *gotest.T) {
	gate := featuregate.NewFeatureGateForTest("1.30")
	v1_24 := "1.24"
	v1_28 := "1.28"
	v1_29 := "1.29"

	gate.Add(map[featuregate.Feature]featuregate.FeatureSpec{
		"alpha_default_on_v1_29": {PreRelease: featuregate.Alpha, Default: true, DefaultEnabledVersion: &v1_29},
		"alpha_default_off":      {PreRelease: featuregate.Alpha, Default: false},

		"beta_default_on_v1_28_deprecated_from_v1_29": {PreRelease: featuregate.Beta, Default: true, DefaultEnabledVersion: &v1_28, DeprecatedVersion: &v1_29},
		"beta_default_off":                            {PreRelease: featuregate.Beta, Default: false},

		"stable_default_on_v1_24": {PreRelease: featuregate.GA, Default: true, DefaultEnabledVersion: &v1_24},
		"stable_default_off":      {PreRelease: featuregate.GA, Default: false},
	})
	testcases := []struct {
		compatibilityVersion string
		expectedMap          map[string]bool
	}{
		{
			compatibilityVersion: "1.24",
			expectedMap: map[string]bool{
				"alpha_default_on_v1_29":                      false,
				"alpha_default_off":                           false,
				"beta_default_on_v1_28_deprecated_from_v1_29": false,
				"beta_default_off":                            false,
				"stable_default_on_v1_24":                     true,
				"stable_default_off":                          false,
			},
		},
		{
			compatibilityVersion: "1.25",
			expectedMap: map[string]bool{
				"alpha_default_on_v1_29":                      false,
				"alpha_default_off":                           false,
				"beta_default_on_v1_28_deprecated_from_v1_29": false,
				"beta_default_off":                            false,
				"stable_default_on_v1_24":                     true,
				"stable_default_off":                          false,
			},
		},
		{
			compatibilityVersion: "1.26",
			expectedMap: map[string]bool{
				"alpha_default_on_v1_29":                      false,
				"alpha_default_off":                           false,
				"beta_default_on_v1_28_deprecated_from_v1_29": false,
				"beta_default_off":                            false,
				"stable_default_on_v1_24":                     true,
				"stable_default_off":                          false,
			},
		},
		{
			compatibilityVersion: "1.27",
			expectedMap: map[string]bool{
				"alpha_default_on_v1_29":                      false,
				"alpha_default_off":                           false,
				"beta_default_on_v1_28_deprecated_from_v1_29": false,
				"beta_default_off":                            false,
				"stable_default_on_v1_24":                     true,
				"stable_default_off":                          false,
			},
		},
		{
			compatibilityVersion: "1.28",
			expectedMap: map[string]bool{
				"alpha_default_on_v1_29":                      false,
				"alpha_default_off":                           false,
				"beta_default_on_v1_28_deprecated_from_v1_29": true,
				"beta_default_off":                            false,
				"stable_default_on_v1_24":                     true,
				"stable_default_off":                          false,
			},
		},
		{
			compatibilityVersion: "1.29",
			expectedMap: map[string]bool{
				"alpha_default_on_v1_29":                      true,
				"alpha_default_off":                           false,
				"beta_default_on_v1_28_deprecated_from_v1_29": false,
				"beta_default_off":                            false,
				"stable_default_on_v1_24":                     true,
				"stable_default_off":                          false,
			},
		},
	}
	for _, tc := range testcases {
		t.Run(tc.compatibilityVersion, func(t *gotest.T) {
			if err := gate.SetCompatibilityVersion(tc.compatibilityVersion); err != nil {
				t.Fatalf("unexpected err: %v", err)
			}
			for featureName, expectedEnablement := range tc.expectedMap {
				gotEnabled := gate.Enabled(featuregate.Feature(featureName))
				if gotEnabled != expectedEnablement {
					t.Errorf("got Enabled(%s)=%v, wanted %v", featureName, gotEnabled, expectedEnablement)
				}
			}
		})
	}
}

func stringPtr(v string) *string { return &v }

func TestSpecialGates(t *gotest.T) {
	gate := featuregate.NewFeatureGateForTest("1.29")
	gate.SetCompatibilityVersion("1.29")
	gate.Add(map[featuregate.Feature]featuregate.FeatureSpec{
		"alpha_default_on": {
			Default:               true,
			DefaultEnabledVersion: stringPtr("1.29"),
			PreRelease:            featuregate.Alpha,
			PromotionVersionMap: featuregate.PromotionVersionMapping{
				featuregate.Alpha: "1.28",
			},
		},
		"alpha_default_on_set_off": {
			PreRelease:            featuregate.Alpha,
			DefaultEnabledVersion: stringPtr("1.29"),
			Default:               true,
			PromotionVersionMap: featuregate.PromotionVersionMapping{
				featuregate.Alpha: "1.28",
			},
		},
		"alpha_default_off": {
			PreRelease: featuregate.Alpha,
			Default:    false,
			PromotionVersionMap: featuregate.PromotionVersionMapping{
				featuregate.Alpha: "1.28",
			},
		},
		"alpha_default_off_set_on": {
			PreRelease: featuregate.Alpha,
			Default:    false,
			PromotionVersionMap: featuregate.PromotionVersionMapping{
				featuregate.Alpha: "1.28",
			},
		},

		"beta_default_on": {
			PreRelease:            featuregate.Beta,
			Default:               true,
			DefaultEnabledVersion: stringPtr("1.28"),
			PromotionVersionMap: featuregate.PromotionVersionMapping{
				featuregate.Alpha: "1.27",
				featuregate.Beta:  "1.28",
			},
		},
		"beta_default_on_set_off": {
			PreRelease:            featuregate.Beta,
			Default:               true,
			DefaultEnabledVersion: stringPtr("1.28"),
			PromotionVersionMap: featuregate.PromotionVersionMapping{
				featuregate.Alpha: "1.27",
				featuregate.Beta:  "1.28",
			},
		},
		"beta_default_off": {
			PreRelease: featuregate.Beta,
			Default:    false,
			PromotionVersionMap: featuregate.PromotionVersionMapping{
				featuregate.Alpha: "1.27",
				featuregate.Beta:  "1.28",
			},
		},
		"beta_default_off_set_on": {
			PreRelease: featuregate.Beta,
			Default:    false,
			PromotionVersionMap: featuregate.PromotionVersionMapping{
				featuregate.Alpha: "1.27",
				featuregate.Beta:  "1.28",
			},
		},

		"stable_default_on": {
			PreRelease:            featuregate.GA,
			Default:               true,
			DefaultEnabledVersion: stringPtr("1.28"),
			PromotionVersionMap: featuregate.PromotionVersionMapping{
				featuregate.Alpha: "1.27",
				featuregate.Beta:  "1.28",
				featuregate.GA:    "1.29",
			},
		},
		"stable_default_on_set_off": {
			PreRelease:            featuregate.GA,
			Default:               true,
			DefaultEnabledVersion: stringPtr("1.28"),
			PromotionVersionMap: featuregate.PromotionVersionMapping{
				featuregate.Alpha: "1.27",
				featuregate.Beta:  "1.28",
				featuregate.GA:    "1.29",
			},
		},
		"stable_default_off": {
			PreRelease: featuregate.GA,
			Default:    false,
			PromotionVersionMap: featuregate.PromotionVersionMapping{
				featuregate.Alpha: "1.27",
				featuregate.Beta:  "1.28",
				featuregate.GA:    "1.29",
			},
		},
		"stable_default_off_set_on": {
			PreRelease: featuregate.GA,
			Default:    false,
			PromotionVersionMap: featuregate.PromotionVersionMapping{
				featuregate.Alpha: "1.27",
				featuregate.Beta:  "1.28",
				featuregate.GA:    "1.29",
			},
		},
	})
	gate.Set("alpha_default_on_set_off=false")
	gate.Set("beta_default_on_set_off=false")
	gate.Set("stable_default_on_set_off=false")
	gate.Set("alpha_default_off_set_on=true")
	gate.Set("beta_default_off_set_on=true")
	gate.Set("stable_default_off_set_on=true")

	before := map[featuregate.Feature]bool{
		"AllAlpha": false,
		"AllBeta":  false,

		"alpha_default_on":         true,
		"alpha_default_on_set_off": false,
		"alpha_default_off":        false,
		"alpha_default_off_set_on": true,

		"beta_default_on":         true,
		"beta_default_on_set_off": false,
		"beta_default_off":        false,
		"beta_default_off_set_on": true,

		"stable_default_on":         true,
		"stable_default_on_set_off": false,
		"stable_default_off":        false,
		"stable_default_off_set_on": true,
	}
	expect(t, gate, before)

	cleanupAlpha := SetFeatureGateDuringTest(t, gate, "AllAlpha", true)
	expect(t, gate, map[featuregate.Feature]bool{
		"AllAlpha": true,
		"AllBeta":  false,

		"alpha_default_on":         true,
		"alpha_default_on_set_off": true,
		"alpha_default_off":        true,
		"alpha_default_off_set_on": true,

		"beta_default_on":         true,
		"beta_default_on_set_off": false,
		"beta_default_off":        false,
		"beta_default_off_set_on": true,

		"stable_default_on":         true,
		"stable_default_on_set_off": false,
		"stable_default_off":        false,
		"stable_default_off_set_on": true,
	})

	cleanupBeta := SetFeatureGateDuringTest(t, gate, "AllBeta", true)
	expect(t, gate, map[featuregate.Feature]bool{
		"AllAlpha": true,
		"AllBeta":  true,

		"alpha_default_on":         true,
		"alpha_default_on_set_off": true,
		"alpha_default_off":        true,
		"alpha_default_off_set_on": true,

		"beta_default_on":         true,
		"beta_default_on_set_off": true,
		"beta_default_off":        true,
		"beta_default_off_set_on": true,

		"stable_default_on":         true,
		"stable_default_on_set_off": false,
		"stable_default_off":        false,
		"stable_default_off_set_on": true,
	})

	// run cleanups in reverse order like defer would
	cleanupBeta()
	cleanupAlpha()
	expect(t, gate, before)
}

func expect(t *gotest.T, gate featuregate.FeatureGate, expect map[featuregate.Feature]bool) {
	t.Helper()
	for k, v := range expect {
		if gate.Enabled(k) != v {
			t.Errorf("Expected %v=%v, got %v", k, v, gate.Enabled(k))
		}
	}
}
