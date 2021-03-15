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
	Client    *kubernetes.Clientset
	CrdName   string
	Namespace string
	Host      string
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
		client:  c.Client.CoreV1().RESTClient(),
		ns:      &c.Namespace,
		crdName: &c.CrdName,
		host:    &c.Host,
	}
}

// ControllerGetter returns Controller
type ControllerGetter interface {
	Controller() ControllerInterface
}

// ControllerInterface defines interface for opengauss controller
type ControllerInterface interface {
	List(ctx context.Context, opts metav1.ListOptions) (rest.Result, error)
}

// opengauss controller
type controller struct {
	client  rest.Interface
	ns      *string
	crdName *string
	host    *string
}

func (c *controller) List(ctx context.Context, opts metav1.ListOptions) (res rest.Result, err error) {
	var timeout time.Duration
	if opts.TimeoutSeconds != nil {
		timeout = time.Duration(*opts.TimeoutSeconds) * time.Second
	}
	req := c.client.Get()
	req.RequestURI(*c.host + "/apis/ljt.do/v1")
	res = req.Namespace(*c.ns).
		Resource(*c.crdName).
		VersionedParams(&opts, scheme.ParameterCodec).
		Timeout(timeout).
		Do(ctx)
	fmt.Printf("[log] Actual request URL: %s\n", req.URL())
	return
}
