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

package validation

import (
	"strings"
	"testing"

	"k8s.io/apimachinery/pkg/util/validation/field"
	"k8s.io/kubernetes/pkg/apis/features"
)

func TestValidateFeature(t *testing.T) {
	cases := []struct {
		f           features.Feature
		expectedErr string
	}{
		{
			ssv:         features.Feature{},
			expectedErr: "apiServerID: Invalid value",
		},
	}

	for _, tc := range cases {
		err := ValidateFeature(tc.f).ToAggregate()
		if err == nil && len(tc.expectedErr) == 0 {
			continue
		}
		if err != nil && len(tc.expectedErr) == 0 {
			t.Errorf("unexpected error %v", err)
			continue
		}
		if err == nil && len(tc.expectedErr) != 0 {
			t.Errorf("unexpected empty error")
			continue
		}
		if !strings.Contains(err.Error(), tc.expectedErr) {
			t.Errorf("expected error to contain %s, got %s", tc.expectedErr, err)
		}
	}
}
