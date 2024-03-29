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

package v1

import (
	"context"
	"time"

	v1 "github.com/waterme7on/openGauss-operator/pkg/apis/autoscaler/v1"
	scheme "github.com/waterme7on/openGauss-operator/pkg/generated/clientset/versioned/scheme"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	rest "k8s.io/client-go/rest"
)

// AutoScalersGetter has a method to return a AutoScalerInterface.
// A group's client should implement this interface.
type AutoScalersGetter interface {
	AutoScalers(namespace string) AutoScalerInterface
}

// AutoScalerInterface has methods to work with AutoScaler resources.
type AutoScalerInterface interface {
	Create(ctx context.Context, autoScaler *v1.AutoScaler, opts metav1.CreateOptions) (*v1.AutoScaler, error)
	Update(ctx context.Context, autoScaler *v1.AutoScaler, opts metav1.UpdateOptions) (*v1.AutoScaler, error)
	Delete(ctx context.Context, name string, opts metav1.DeleteOptions) error
	DeleteCollection(ctx context.Context, opts metav1.DeleteOptions, listOpts metav1.ListOptions) error
	Get(ctx context.Context, name string, opts metav1.GetOptions) (*v1.AutoScaler, error)
	List(ctx context.Context, opts metav1.ListOptions) (*v1.AutoScalerList, error)
	Watch(ctx context.Context, opts metav1.ListOptions) (watch.Interface, error)
	Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts metav1.PatchOptions, subresources ...string) (result *v1.AutoScaler, err error)
	AutoScalerExpansion
}

// autoScalers implements AutoScalerInterface
type autoScalers struct {
	client rest.Interface
	ns     string
}

// newAutoScalers returns a AutoScalers
func newAutoScalers(c *ScalerV1Client, namespace string) *autoScalers {
	return &autoScalers{
		client: c.RESTClient(),
		ns:     namespace,
	}
}

// Get takes name of the autoScaler, and returns the corresponding autoScaler object, and an error if there is any.
func (c *autoScalers) Get(ctx context.Context, name string, options metav1.GetOptions) (result *v1.AutoScaler, err error) {
	result = &v1.AutoScaler{}
	err = c.client.Get().
		Namespace(c.ns).
		Resource("autoscalers").
		Name(name).
		VersionedParams(&options, scheme.ParameterCodec).
		Do(ctx).
		Into(result)
	return
}

// List takes label and field selectors, and returns the list of AutoScalers that match those selectors.
func (c *autoScalers) List(ctx context.Context, opts metav1.ListOptions) (result *v1.AutoScalerList, err error) {
	var timeout time.Duration
	if opts.TimeoutSeconds != nil {
		timeout = time.Duration(*opts.TimeoutSeconds) * time.Second
	}
	result = &v1.AutoScalerList{}
	err = c.client.Get().
		Namespace(c.ns).
		Resource("autoscalers").
		VersionedParams(&opts, scheme.ParameterCodec).
		Timeout(timeout).
		Do(ctx).
		Into(result)
	return
}

// Watch returns a watch.Interface that watches the requested autoScalers.
func (c *autoScalers) Watch(ctx context.Context, opts metav1.ListOptions) (watch.Interface, error) {
	var timeout time.Duration
	if opts.TimeoutSeconds != nil {
		timeout = time.Duration(*opts.TimeoutSeconds) * time.Second
	}
	opts.Watch = true
	return c.client.Get().
		Namespace(c.ns).
		Resource("autoscalers").
		VersionedParams(&opts, scheme.ParameterCodec).
		Timeout(timeout).
		Watch(ctx)
}

// Create takes the representation of a autoScaler and creates it.  Returns the server's representation of the autoScaler, and an error, if there is any.
func (c *autoScalers) Create(ctx context.Context, autoScaler *v1.AutoScaler, opts metav1.CreateOptions) (result *v1.AutoScaler, err error) {
	result = &v1.AutoScaler{}
	err = c.client.Post().
		Namespace(c.ns).
		Resource("autoscalers").
		VersionedParams(&opts, scheme.ParameterCodec).
		Body(autoScaler).
		Do(ctx).
		Into(result)
	return
}

// Update takes the representation of a autoScaler and updates it. Returns the server's representation of the autoScaler, and an error, if there is any.
func (c *autoScalers) Update(ctx context.Context, autoScaler *v1.AutoScaler, opts metav1.UpdateOptions) (result *v1.AutoScaler, err error) {
	result = &v1.AutoScaler{}
	err = c.client.Put().
		Namespace(c.ns).
		Resource("autoscalers").
		Name(autoScaler.Name).
		VersionedParams(&opts, scheme.ParameterCodec).
		Body(autoScaler).
		Do(ctx).
		Into(result)
	return
}

// Delete takes name of the autoScaler and deletes it. Returns an error if one occurs.
func (c *autoScalers) Delete(ctx context.Context, name string, opts metav1.DeleteOptions) error {
	return c.client.Delete().
		Namespace(c.ns).
		Resource("autoscalers").
		Name(name).
		Body(&opts).
		Do(ctx).
		Error()
}

// DeleteCollection deletes a collection of objects.
func (c *autoScalers) DeleteCollection(ctx context.Context, opts metav1.DeleteOptions, listOpts metav1.ListOptions) error {
	var timeout time.Duration
	if listOpts.TimeoutSeconds != nil {
		timeout = time.Duration(*listOpts.TimeoutSeconds) * time.Second
	}
	return c.client.Delete().
		Namespace(c.ns).
		Resource("autoscalers").
		VersionedParams(&listOpts, scheme.ParameterCodec).
		Timeout(timeout).
		Body(&opts).
		Do(ctx).
		Error()
}

// Patch applies the patch and returns the patched autoScaler.
func (c *autoScalers) Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts metav1.PatchOptions, subresources ...string) (result *v1.AutoScaler, err error) {
	result = &v1.AutoScaler{}
	err = c.client.Patch(pt).
		Namespace(c.ns).
		Resource("autoscalers").
		Name(name).
		SubResource(subresources...).
		VersionedParams(&opts, scheme.ParameterCodec).
		Body(data).
		Do(ctx).
		Into(result)
	return
}
