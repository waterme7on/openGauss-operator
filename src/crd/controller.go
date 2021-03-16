package crd

import (
	"context"
	"fmt"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	rest "k8s.io/client-go/rest"
)

// OpengaussClient is client for operator
type OpengaussClient struct {
	Client     *kubernetes.Clientset
	CrdName    string
	Namespace  string
	Host       string
	ApiVersion string
}

// OpengaussManager defines interface for opengauss controller manager
type OpengaussManager interface {
	ControllerGetter
}

// Controller returns a controller
func (c *OpengaussClient) Controller() ControllerInterface {
	return newController(c)
}

func newController(c *OpengaussClient) *controller {
	return &controller{
		client:     c.Client.CoreV1().RESTClient(),
		ns:         &c.Namespace,
		crdName:    &c.CrdName,
		host:       &c.Host,
		decoder:    OpenGaussDecoder{},
		apiVersion: &c.ApiVersion,
	}
}

// ControllerGetter returns Controller
type ControllerGetter interface {
	Controller() ControllerInterface
}

// ControllerInterface defines interface for opengauss controller
type ControllerInterface interface {
	List(ctx context.Context, opts metav1.ListOptions) (*OpenGaussList, error)
	Get(ctx context.Context, name string, opts metav1.ListOptions) (*OpenGauss, error)
	Create(ctx context.Context, og *OpenGaussConfiguration, opts metav1.CreateOptions) (result *OpenGauss, err error)
}

// opengauss controller
type controller struct {
	decoder    OpenGaussDecoder
	client     rest.Interface
	ns         *string
	crdName    *string
	host       *string
	apiVersion *string
}

// List returns runtime object OpenGaussList
func (c *controller) List(ctx context.Context, opts metav1.ListOptions) (obj *OpenGaussList, err error) {
	req := c.client.Get()
	req.RequestURI(*c.host + "/apis/" + *c.apiVersion)
	if opts.TimeoutSeconds != nil {
		timeout := time.Duration(*opts.TimeoutSeconds) * time.Second
		req.Timeout(timeout)
	}
	res := req.Namespace(*c.ns).
		Resource(*c.crdName).
		VersionedParams(&opts, scheme.ParameterCodec).
		Do(ctx)
	fmt.Printf("[log] request URL: %s\n", req.URL())
	raw, _ := res.Raw()
	s := string(raw)
	obj = &OpenGaussList{
		Config: c.decoder.DecodeList(&s),
	}
	return
}

// Get returns runtime object OpenGauss
func (c *controller) Get(ctx context.Context, name string, opts metav1.ListOptions) (obj *OpenGauss, err error) {
	req := c.client.Get()
	req.RequestURI(*c.host + "/apis/" + *c.apiVersion)
	if opts.TimeoutSeconds != nil {
		timeout := time.Duration(*opts.TimeoutSeconds) * time.Second
		req.Timeout(timeout)
	}
	res := req.Namespace(*c.ns).
		Resource(*c.crdName).
		VersionedParams(&opts, scheme.ParameterCodec).
		Do(ctx)
	fmt.Printf("[log] request URL: %s\n", req.URL())
	raw, _ := res.Raw()
	s := string(raw)
	obj = &OpenGauss{
		Config: c.decoder.Decode(&s),
	}
	return
}

// Create adds a opengauss cluster in the kubernetes cluster, including statefulset of master and replicas
func (c *controller) Create(ctx context.Context, og *OpenGaussConfiguration, opts metav1.CreateOptions) (result *OpenGauss, err error) {
	result = &OpenGauss{}
	req := c.client.Post()
	req.RequestURI(*c.host + "/apis/" + *c.apiVersion)
	res := req.Namespace(*c.ns).
		Resource(*c.crdName).
		VersionedParams(&opts, scheme.ParameterCodec).
		Body(og).
		Do(ctx)
	fmt.Printf("[log] request URL: %s\n", req.URL())
	// TBD: haven't tested
	raw, _ := res.Raw()
	fmt.Println("[log] Raw: ", raw)
	return
}
