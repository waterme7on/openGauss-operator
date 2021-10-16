package main

import (
	"context"
	"fmt"
	"time"

	scalerv1 "github.com/waterme7on/openGauss-operator/pkg/apis/autoscaler/v1"
	opengaussv1 "github.com/waterme7on/openGauss-operator/pkg/apis/opengausscontroller/v1"
	clientset "github.com/waterme7on/openGauss-operator/pkg/generated/clientset/versioned"
	"github.com/waterme7on/openGauss-operator/pkg/generated/clientset/versioned/scheme"
	ogscheme "github.com/waterme7on/openGauss-operator/pkg/generated/clientset/versioned/scheme"
	informers "github.com/waterme7on/openGauss-operator/pkg/generated/informers/externalversions/autoscaler/v1"
	listers "github.com/waterme7on/openGauss-operator/pkg/generated/listers/autoscaler/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	typedcorev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	cache "k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/record"
	"k8s.io/client-go/util/workqueue"
	"k8s.io/klog/v2"
)

const autoscalerControllerAgentName = "openGauss-autoscaler-controller"

const (
	// SuccessSynced is used as part of the Event 'reason' when a AutoScaler is synced
	ScalerSuccessSynced = "Synced"
	// ErrResourceExists is used as part of the Event 'reason' when a AutoScaler fails
	// to sync due to a Deployment of the same name already existing.
	ScalerErrResourceExists = "ErrResourceExists"

	// Messages
	//
	// MessageResourceExists is the message used for Events when a resource
	// fails to sync due to a Deployment already existing
	ScalerMessageResourceExists = "Resource %q already exists and is not managed by OpenGauss"
	// MessageResourceSynced is the message used for an Event fired when a OpenGauss
	// is synced successfully
	ScalerMessageResourceSynced = "AutoScaler synced successfully"
)

type AutoScalerController struct {
	// kubeclientset is a standard kubernetes clientset
	kubeClientset kubernetes.Interface
	// openGaussClientset is a clientset generated for OpenGauss Objects
	openGaussClientset clientset.Interface

	autoScalerLister listers.AutoScalerLister
	autoScalerSynced cache.InformerSynced

	// workqueue is a rate limited work queue. This is used to queue work to be
	// processed instead of performing it as soon as a change happens. This
	// means we can ensure we only process a fixed amount of resources at a
	// time, and makes it easy to ensure we are never processing the same item
	// simultaneously in two different workers.
	workqueue workqueue.RateLimitingInterface
	// recorder is an event recorder for recording Event resources to the
	// Kubernetes API.
	recorder record.EventRecorder
}

func NewAutoScalerController(
	kubeclientset kubernetes.Interface,
	openGaussClientset clientset.Interface,
	autoScalerInformer informers.AutoScalerInformer) *AutoScalerController {

	// Create new event broadcaster
	// add autoscaler type to default kubernetes scheme
	utilruntime.Must(ogscheme.AddToScheme(scheme.Scheme))
	klog.V(4).Info("Creating event broadcaster for auto scaler")
	eventBroadCaster := record.NewBroadcaster()
	eventBroadCaster.StartStructuredLogging(0)
	// start sending events received from the specified broadcaster the sink
	eventBroadCaster.StartRecordingToSink(&typedcorev1.EventSinkImpl{Interface: kubeclientset.CoreV1().Events("")})
	// event recorder records event in the given source
	recorder := eventBroadCaster.NewRecorder(scheme.Scheme, corev1.EventSource{Component: autoscalerControllerAgentName})

	autoScalerController := &AutoScalerController{
		kubeClientset:      kubeclientset,
		openGaussClientset: openGaussClientset,
		autoScalerLister:   autoScalerInformer.Lister(),
		autoScalerSynced:   autoScalerInformer.Informer().HasSynced,
		workqueue:          workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "AutoScalers"),
		recorder:           recorder,
	}

	klog.Infoln("Setting up event handlers")
	autoScalerInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: autoScalerController.enqueueAutoScaler,
		UpdateFunc: func(old, new interface{}) {
			if old != new {
				autoScalerController.enqueueAutoScaler(new)
			}
		},
	})
	return autoScalerController
}

// add AutoScaler resources in the workqueue and validate
func (c *AutoScalerController) enqueueAutoScaler(obj interface{}) {
	var key string
	var err error
	if key, err = cache.MetaNamespaceKeyFunc(obj); err != nil {
		utilruntime.HandleError(err)
		return
	}
	c.workqueue.Add(key)
}

// Run will set up the event handlers for types monitored.
// It will block until stopCh is closed, at which point it will shutdown the workqueue and
// wait for workers to finish processing their current work items.
func (c *AutoScalerController) Run(threadiness int, stopCh <-chan struct{}) error {
	defer utilruntime.HandleCrash()
	defer c.workqueue.ShutDown()

	// start the informer factories to begin populating the informer caches
	klog.Infoln("Starting openGauss-AutoScaler controller")

	// wait for the caches to be synced before starting workers
	klog.Infoln("Syncing AutoScaler informers' caches")
	if ok := cache.WaitForCacheSync(stopCh, c.autoScalerSynced); !ok {
		return fmt.Errorf("failed to wait for AutoScaler caches to sync")
	}

	// starting workers
	klog.Infoln("Starting AutoScaler workers")
	// Launch workers to process OpenGauss Resources
	for i := 0; i < threadiness; i++ {
		go wait.Until(c.runWorker, time.Second, stopCh)
	}

	klog.Infoln("Started AutoScaler workers")
	<-stopCh
	klog.Infoln("Shutting down AutoScaler workers")

	return nil
}

// runWorker is a long-running function that will continually call the
// processNextWorkItem function in order to read and process a message on the
// workqueue.
func (c *AutoScalerController) runWorker() {
	for c.processNextWorkItem() {
		time.Sleep(time.Second * 5)
	}
}

// process autoscaler objects
// compare status to the desired status, then update it
func (c *AutoScalerController) processNextWorkItem() bool {
	obj, shutdown := c.workqueue.Get()
	if shutdown {
		return false
	}

	// wrap this block in a func so we can defer c.workqueue.Done
	err := func(obj interface{}) error {
		// call Done here so that workqueue knows that the item have been processed
		defer c.workqueue.Done(obj)
		// We expect strings to come off the workqueue. These are of the form namespace/name.
		// We do this as the delayed nature of the workqueue means the items in the informer cache may actually be
		// more up to date that when the item was initially put onto the workqueue.
		var key string
		var ok bool
		if key, ok = obj.(string); !ok {
			// As the item in the workqueue is actually invalid, we call
			// Forget here else we'd go into a loop of attempting to
			// process a work item that is invalid.
			c.workqueue.Forget(obj)
			utilruntime.HandleError(fmt.Errorf("expected string in workqueue but got %#v", obj))
			return nil
		}
		// run syncHandler, passing the string  "namespace/name" of opengauss to be synced
		// TODO: syncHandler
		// here simply print out the object
		if err := c.syncHandler(key); err != nil {
			// Put the item back on the workqueue to handle any transient errors.
			c.workqueue.AddRateLimited(key)
			return fmt.Errorf("error syncing '%s': %s, requeuing", key, err.Error())
		}

		// if no error occurs, we Forget the items as it has been processed successfully
		c.workqueue.Forget(obj)
		klog.Infoln("Successfully synced", key)
		return nil
	}(obj)

	if err != nil {
		utilruntime.HandleError(err)
		return true
	}
	return true
}

// handle key of form "namespace/name" from workqueue
func (c *AutoScalerController) syncHandler(key string) error {
	// Convert the namespace/name into a distinct namespace and name
	namespace, name, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		utilruntime.HandleError(fmt.Errorf("invalid resource key: %s", key))
		return nil
	}

	// Get the openGauss resource with the namespace and name
	scaler, err := c.autoScalerLister.AutoScalers(namespace).Get(name)

	if err != nil {
		// The openGauss object may not exist.
		if errors.IsNotFound(err) {
			utilruntime.HandleError(fmt.Errorf("AutoScaler '%s' in work queue no longer exists", key))
			return nil
		}
		return err
	}
	klog.Info("Syncing status of OpenGauss AutoScaler ", scaler.Name)
	// check if Cluster info are correct
	og, err := c.openGaussClientset.ControllerV1().OpenGausses(scaler.Spec.Cluster.Namespace).Get(context.TODO(), scaler.Spec.Cluster.Name, v1.GetOptions{})
	if err != nil {
		klog.V(4).Infof("OpenGauss Cluster information error : scaler %s , 'cluster information %s'", key, scaler.Spec.Cluster)
		return err
	}
	klog.Info("Update autoscaler ", scaler.Name)

	err = c.createOrUpdateAutoScaler(og, scaler)
	if err != nil {
		return err
	}
	err = c.updateAutoScalerStatus(scaler)
	return nil
}

// createOrUpdateScaler creates or updates auto-scaler
func (c *AutoScalerController) createOrUpdateAutoScaler(og *opengaussv1.OpenGauss, autoscaler *scalerv1.AutoScaler) error {
	if autoscaler.Spec.Master == nil && autoscaler.Spec.Worker == nil {
		return nil
	}
	if autoscaler.Spec.Master != nil {
		masterHpaConfig := NewHorizontalPodAutoscaler(og, autoscaler, Master)
		_, err := c.kubeClientset.AutoscalingV2beta2().HorizontalPodAutoscalers(og.Namespace).Get(context.TODO(), masterHpaConfig.Name, v1.GetOptions{})
		if err != nil {
			// not exist, try to create
			_, err = c.kubeClientset.AutoscalingV2beta2().HorizontalPodAutoscalers(og.Namespace).Create(context.TODO(), masterHpaConfig, v1.CreateOptions{})
		} else {
			// exist, try to update
			_, err = c.kubeClientset.AutoscalingV2beta2().HorizontalPodAutoscalers(og.Namespace).Update(context.TODO(), masterHpaConfig, v1.UpdateOptions{})
		}
		if err != nil {
			klog.V(4).Infof("fail to create or update hpa for master: %s", og.Name)
			return err
		}
	}
	if autoscaler.Spec.Worker != nil {
		workerHpaConfig := NewHorizontalPodAutoscaler(og, autoscaler, Replicas)
		_, err := c.kubeClientset.AutoscalingV2beta2().HorizontalPodAutoscalers(og.Namespace).Get(context.TODO(), workerHpaConfig.Name, v1.GetOptions{})
		if err != nil {
			// not exist, try to create
			_, err = c.kubeClientset.AutoscalingV2beta2().HorizontalPodAutoscalers(og.Namespace).Create(context.TODO(), workerHpaConfig, v1.CreateOptions{})
		} else {
			// exist, try to update
			_, err = c.kubeClientset.AutoscalingV2beta2().HorizontalPodAutoscalers(og.Namespace).Update(context.TODO(), workerHpaConfig, v1.UpdateOptions{})
		}
		if err != nil {
			klog.V(4).Infof("fail to create or update hpa for worker: %s", og.Name)
			return err
		}
	}
	return nil
}

// updateAutoScalerStatus
// TODO
func (c *AutoScalerController) updateAutoScalerStatus(autoscaler *scalerv1.AutoScaler) error {
	return nil
}
