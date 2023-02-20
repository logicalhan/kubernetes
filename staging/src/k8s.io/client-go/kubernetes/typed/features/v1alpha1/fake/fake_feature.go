/*
Copyright The Kubernetes Authors.

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

// Code generated by client-gen. DO NOT EDIT.

package fake

import (
	"context"
	json "encoding/json"
	"fmt"

	v1alpha1 "k8s.io/api/features/v1alpha1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	labels "k8s.io/apimachinery/pkg/labels"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	featuresv1alpha1 "k8s.io/client-go/applyconfigurations/features/v1alpha1"
	testing "k8s.io/client-go/testing"
)

// FakeFeatures implements FeatureInterface
type FakeFeatures struct {
	Fake *FakeFeaturesV1alpha1
}

var featuresResource = v1alpha1.SchemeGroupVersion.WithResource("features")

var featuresKind = v1alpha1.SchemeGroupVersion.WithKind("Feature")

// Get takes name of the feature, and returns the corresponding feature object, and an error if there is any.
func (c *FakeFeatures) Get(ctx context.Context, name string, options v1.GetOptions) (result *v1alpha1.Feature, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewRootGetAction(featuresResource, name), &v1alpha1.Feature{})
	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.Feature), err
}

// List takes label and field selectors, and returns the list of Features that match those selectors.
func (c *FakeFeatures) List(ctx context.Context, opts v1.ListOptions) (result *v1alpha1.FeatureList, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewRootListAction(featuresResource, featuresKind, opts), &v1alpha1.FeatureList{})
	if obj == nil {
		return nil, err
	}

	label, _, _ := testing.ExtractFromListOptions(opts)
	if label == nil {
		label = labels.Everything()
	}
	list := &v1alpha1.FeatureList{ListMeta: obj.(*v1alpha1.FeatureList).ListMeta}
	for _, item := range obj.(*v1alpha1.FeatureList).Items {
		if label.Matches(labels.Set(item.Labels)) {
			list.Items = append(list.Items, item)
		}
	}
	return list, err
}

// Watch returns a watch.Interface that watches the requested features.
func (c *FakeFeatures) Watch(ctx context.Context, opts v1.ListOptions) (watch.Interface, error) {
	return c.Fake.
		InvokesWatch(testing.NewRootWatchAction(featuresResource, opts))
}

// Create takes the representation of a feature and creates it.  Returns the server's representation of the feature, and an error, if there is any.
func (c *FakeFeatures) Create(ctx context.Context, feature *v1alpha1.Feature, opts v1.CreateOptions) (result *v1alpha1.Feature, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewRootCreateAction(featuresResource, feature), &v1alpha1.Feature{})
	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.Feature), err
}

// Update takes the representation of a feature and updates it. Returns the server's representation of the feature, and an error, if there is any.
func (c *FakeFeatures) Update(ctx context.Context, feature *v1alpha1.Feature, opts v1.UpdateOptions) (result *v1alpha1.Feature, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewRootUpdateAction(featuresResource, feature), &v1alpha1.Feature{})
	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.Feature), err
}

// UpdateStatus was generated because the type contains a Status member.
// Add a +genclient:noStatus comment above the type to avoid generating UpdateStatus().
func (c *FakeFeatures) UpdateStatus(ctx context.Context, feature *v1alpha1.Feature, opts v1.UpdateOptions) (*v1alpha1.Feature, error) {
	obj, err := c.Fake.
		Invokes(testing.NewRootUpdateSubresourceAction(featuresResource, "status", feature), &v1alpha1.Feature{})
	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.Feature), err
}

// Delete takes name of the feature and deletes it. Returns an error if one occurs.
func (c *FakeFeatures) Delete(ctx context.Context, name string, opts v1.DeleteOptions) error {
	_, err := c.Fake.
		Invokes(testing.NewRootDeleteActionWithOptions(featuresResource, name, opts), &v1alpha1.Feature{})
	return err
}

// DeleteCollection deletes a collection of objects.
func (c *FakeFeatures) DeleteCollection(ctx context.Context, opts v1.DeleteOptions, listOpts v1.ListOptions) error {
	action := testing.NewRootDeleteCollectionAction(featuresResource, listOpts)

	_, err := c.Fake.Invokes(action, &v1alpha1.FeatureList{})
	return err
}

// Patch applies the patch and returns the patched feature.
func (c *FakeFeatures) Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts v1.PatchOptions, subresources ...string) (result *v1alpha1.Feature, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewRootPatchSubresourceAction(featuresResource, name, pt, data, subresources...), &v1alpha1.Feature{})
	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.Feature), err
}

// Apply takes the given apply declarative configuration, applies it and returns the applied feature.
func (c *FakeFeatures) Apply(ctx context.Context, feature *featuresv1alpha1.FeatureApplyConfiguration, opts v1.ApplyOptions) (result *v1alpha1.Feature, err error) {
	if feature == nil {
		return nil, fmt.Errorf("feature provided to Apply must not be nil")
	}
	data, err := json.Marshal(feature)
	if err != nil {
		return nil, err
	}
	name := feature.Name
	if name == nil {
		return nil, fmt.Errorf("feature.Name must be provided to Apply")
	}
	obj, err := c.Fake.
		Invokes(testing.NewRootPatchSubresourceAction(featuresResource, *name, types.ApplyPatchType, data), &v1alpha1.Feature{})
	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.Feature), err
}

// ApplyStatus was generated because the type contains a Status member.
// Add a +genclient:noStatus comment above the type to avoid generating ApplyStatus().
func (c *FakeFeatures) ApplyStatus(ctx context.Context, feature *featuresv1alpha1.FeatureApplyConfiguration, opts v1.ApplyOptions) (result *v1alpha1.Feature, err error) {
	if feature == nil {
		return nil, fmt.Errorf("feature provided to Apply must not be nil")
	}
	data, err := json.Marshal(feature)
	if err != nil {
		return nil, err
	}
	name := feature.Name
	if name == nil {
		return nil, fmt.Errorf("feature.Name must be provided to Apply")
	}
	obj, err := c.Fake.
		Invokes(testing.NewRootPatchSubresourceAction(featuresResource, *name, types.ApplyPatchType, data, "status"), &v1alpha1.Feature{})
	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.Feature), err
}
