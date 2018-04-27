// Code generated by client-gen. DO NOT EDIT.

package v1alpha2

import (
	v1alpha2 "github.com/dpacierpnik/istio-0.7-eds-issue/pkg/apis/config.istio.io/v1alpha2"
	scheme "github.com/dpacierpnik/istio-0.7-eds-issue/pkg/clients/config.istio.io/clientset/versioned/scheme"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	rest "k8s.io/client-go/rest"
)

// RulesGetter has a method to return a RuleInterface.
// A group's client should implement this interface.
type RulesGetter interface {
	Rules(namespace string) RuleInterface
}

// RuleInterface has methods to work with Rule resources.
type RuleInterface interface {
	Create(*v1alpha2.Rule) (*v1alpha2.Rule, error)
	Update(*v1alpha2.Rule) (*v1alpha2.Rule, error)
	Delete(name string, options *v1.DeleteOptions) error
	DeleteCollection(options *v1.DeleteOptions, listOptions v1.ListOptions) error
	Get(name string, options v1.GetOptions) (*v1alpha2.Rule, error)
	List(opts v1.ListOptions) (*v1alpha2.RuleList, error)
	Watch(opts v1.ListOptions) (watch.Interface, error)
	Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *v1alpha2.Rule, err error)
	RuleExpansion
}

// rules implements RuleInterface
type rules struct {
	client rest.Interface
	ns     string
}

// newRules returns a Rules
func newRules(c *ConfigV1alpha2Client, namespace string) *rules {
	return &rules{
		client: c.RESTClient(),
		ns:     namespace,
	}
}

// Get takes name of the rule, and returns the corresponding rule object, and an error if there is any.
func (c *rules) Get(name string, options v1.GetOptions) (result *v1alpha2.Rule, err error) {
	result = &v1alpha2.Rule{}
	err = c.client.Get().
		Namespace(c.ns).
		Resource("rules").
		Name(name).
		VersionedParams(&options, scheme.ParameterCodec).
		Do().
		Into(result)
	return
}

// List takes label and field selectors, and returns the list of Rules that match those selectors.
func (c *rules) List(opts v1.ListOptions) (result *v1alpha2.RuleList, err error) {
	result = &v1alpha2.RuleList{}
	err = c.client.Get().
		Namespace(c.ns).
		Resource("rules").
		VersionedParams(&opts, scheme.ParameterCodec).
		Do().
		Into(result)
	return
}

// Watch returns a watch.Interface that watches the requested rules.
func (c *rules) Watch(opts v1.ListOptions) (watch.Interface, error) {
	opts.Watch = true
	return c.client.Get().
		Namespace(c.ns).
		Resource("rules").
		VersionedParams(&opts, scheme.ParameterCodec).
		Watch()
}

// Create takes the representation of a rule and creates it.  Returns the server's representation of the rule, and an error, if there is any.
func (c *rules) Create(rule *v1alpha2.Rule) (result *v1alpha2.Rule, err error) {
	result = &v1alpha2.Rule{}
	err = c.client.Post().
		Namespace(c.ns).
		Resource("rules").
		Body(rule).
		Do().
		Into(result)
	return
}

// Update takes the representation of a rule and updates it. Returns the server's representation of the rule, and an error, if there is any.
func (c *rules) Update(rule *v1alpha2.Rule) (result *v1alpha2.Rule, err error) {
	result = &v1alpha2.Rule{}
	err = c.client.Put().
		Namespace(c.ns).
		Resource("rules").
		Name(rule.Name).
		Body(rule).
		Do().
		Into(result)
	return
}

// Delete takes name of the rule and deletes it. Returns an error if one occurs.
func (c *rules) Delete(name string, options *v1.DeleteOptions) error {
	return c.client.Delete().
		Namespace(c.ns).
		Resource("rules").
		Name(name).
		Body(options).
		Do().
		Error()
}

// DeleteCollection deletes a collection of objects.
func (c *rules) DeleteCollection(options *v1.DeleteOptions, listOptions v1.ListOptions) error {
	return c.client.Delete().
		Namespace(c.ns).
		Resource("rules").
		VersionedParams(&listOptions, scheme.ParameterCodec).
		Body(options).
		Do().
		Error()
}

// Patch applies the patch and returns the patched rule.
func (c *rules) Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *v1alpha2.Rule, err error) {
	result = &v1alpha2.Rule{}
	err = c.client.Patch(pt).
		Namespace(c.ns).
		Resource("rules").
		SubResource(subresources...).
		Name(name).
		Body(data).
		Do().
		Into(result)
	return
}
