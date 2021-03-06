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

// Note: the example only works with the code within the same release/branch.
package main

import (
	"context"
	"fmt"
	"time"

	crd "opengauss-operator/src/crd"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	// "k8s.io/client-go/1.5/rest"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
)

func main() {
	// creates the in-cluster config
	config, err := rest.InClusterConfig()
	if err != nil {
		panic(err.Error())
	}

	// creates the clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}

	ogClient := crd.OpengaussClient{
		Client:    clientset,
		Namespace: "",
		CrdName:   "opengauss",
		Host:      config.Host,
	}

	for {
		fmt.Printf("Host:%s, APIPath: %s\n", config.Host, config.APIPath)
		// fmt.Printf("type of clientset.CoreV1().RESTClient(): %T\n", clientset.CoreV1().RESTClient()) // *rest.RESTClient

		// get pods in all the namespaces by omitting namespace
		// Or specify namespace to get pods in particular namespace
		pods, err := clientset.CoreV1().Pods("").List(context.TODO(), metav1.ListOptions{})
		if err != nil {
			panic(err.Error())
		}
		fmt.Printf("There are %d pods in the cluster\n", len(pods.Items))

		var timeout time.Duration
		opt := metav1.ListOptions{}
		req := clientset.CoreV1().RESTClient().Get()
		fmt.Printf("Origin Url: %s\n", req.URL())
		req.RequestURI(config.Host + "/apis/ljt.do/v1")
		fmt.Printf("New Url: %s\n", req.URL())
		// TRY: Successfully get all opengauss cluster info
		req.Namespace("default").Resource("opengauss").VersionedParams(&opt, scheme.ParameterCodec).Timeout(timeout)
		fmt.Printf("New Url: %s\n", req.URL())

		// res = req.Do(context.TODO())

		res, err := ogClient.Controller().List(context.TODO(), metav1.ListOptions{})

		fmt.Printf("%T\n", res)
		raw, _ := res.Raw()
		fmt.Printf("Contetn: %s\n", raw)

		time.Sleep(5 * time.Second)
	}
}

// clientset->.CoreV1->.RESTClient() rest.Interface
// example:
// c.client rest.Interface
// c.client.Get().
// 	Namespace(c.ns).
// 	Resource("endpoints").
// 	Name(name).
// 	VersionedParams(&options, scheme.ParameterCodec).
// 	Do(ctx).
// 	Into(result)

// Do formats and executes the request. Returns a Result object for easy response
// processing.
//
// Error type:
//  * If the server responds with a status: *errors.StatusError or *errors.UnexpectedObjectError
//  * http.Client.Do errors are returned directly.
// func (r *Request) Do(ctx context.Context) Result {
// 	var result Result
// 	err := r.request(ctx, func(req *http.Request, resp *http.Response) {
// 		result = r.transformResponse(resp, req)
// 	})
// 	if err != nil {
// 		return Result{err: err}
// 	}
// 	return result
// }

// Result contains the result of calling Request.Do().
// type Result struct {
// 	body        []byte
// 	warnings    []net.WarningHeader
// 	contentType string
// 	err         error
// 	statusCode  int
// 	decoder runtime.Decoder			https://github.com/kubernetes/client-go/blob/6085ad09f2ca53a788354582d28e3a797727fb13/rest/request.go#L1309
// }

// for {
// 	// get opengauss resources

// }
