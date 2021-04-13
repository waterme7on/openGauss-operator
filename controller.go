package main

import (
	"fmt"
	"time"

	clientset "github.com/waterme7on/openGauss-controller/pkg/generated/clientset"
	ogscheme "github.com/waterme7on/openGauss-controller/pkg/generated/clientset/versioned/scheme"
	informers "github.com/waterme7on/openGauss-controller/pkg/generated/informers/externalversions/openGaussController/v1"
	listers "github.com/waterme7on/openGauss-controller/pkg/generated/listers/openGaussController/v1"
	corev1 "k8s.io/api/core/v1"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/informers"
	appsinformers "k8s.io/client-go/informers/apps/v1"
	coreinformers "k8s.io/client-go/informers/core/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	typedcorev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	appslisters "k8s.io/client-go/listers/apps/v1"
	corelisters "k8s.io/client-go/listers/core/v1"
	cache "k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/record"
	"k8s.io/client-go/util/workqueue"
	"k8s.io/klog/v2"
)

const controllerAgentName = "openGauss-controller"

const (
	// SuccessSynced is used as part of the Event 'reason' when a Foo is synced
	SuccessSynced = "Synced"
	// ErrResourceExists is used as part of the Event 'reason' when a Foo fails
	// to sync due to a Deployment of the same name already existing.
	ErrResourceExists = "ErrResourceExists"

	// Messages
	//
	// MessageResourceExists is the message used for Events when a resource
	// fails to sync due to a Deployment already existing
	MessageResourceExists = "Resource %q already exists and is not managed by OpenGauss"
	// MessageResourceSynced is the message used for an Event fired when a OpenGauss
	// is synced successfully
	MessageResourceSynced = "Foo synced successfully"
)

type Controller struct {
	// kubeclientset is a standard kubernetes clientset
	kubeClientset kubernetes.Interface
	// openGaussClientset is a clientset generated for OpenGauss Objects
	openGaussClientset clientset.Interface

	// openGauss controller manage service, configmap and statefulset of OpenGauss object
	// thus needing listers of according resources
	openGaussLister   listers.OpenGaussLister
	openGaussSynced   cache.InformerSynced
	statefulsetLister appslisters.StatefulSetLister
	statefulsetSynced cache.InformerSynced
	serviceLister     corelisters.ServiceLister
	serviceSynced     cache.InformerSynced
	configMapLister   corelisters.ConfigMapLister
	configMapSynced   cache.InformerSynced

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

// NewController returns a new OpenGauss controller
func NewController(
	kubeclientset kubernetes.Interface,
	openGaussClientset clientset.Interface,
	statefulsetInformer appsinformers.StatefulSetInformer,
	serviceInformer coreinformers.ServiceInformer,
	configmapInformer coreinformers.ConfigMapInformer,
	openGaussInformer informers.OpenGaussInformer) *Controller {

	// Create new event broadcaster
	// Add OpenGauss controller types to the default kubernetes scheme
	// so events can be logged for OpenGauss controller types
	utilruntime.Must(ogscheme.AddToScheme(scheme.Scheme))
	klog.V(4).Info("Creating event Broadcaster")
	eventBroadCaster := record.NewBroadcaster()
	eventBroadCaster.StartStructuredLogging(0)
	// starts sending events received from the specified eventBroadcaster to the given sink
	// EventSink knows how to store events.
	eventBroadCaster.StartRecordingToSink(&typedcorev1.EventSinkImpl{Interface: kubeclientset.CoreV1().Events("")})
	// EventRecorder that records events with the given event source.
	recorder := eventBroadCaster.NewRecorder(scheme.Scheme, corev1.EventSource{Component: controllerAgentName})

	controller := &Controller{
		kubeClientset:      kubeclientset,
		openGaussClientset: openGaussClientset,
		openGaussLister:    openGaussInformer.Lister(),
		openGaussSynced:    openGaussInformer.Informer().HasSynced,
		statefulsetLister:  statefulsetInformer.LIster(),
		statefulsetSynced:  statefulsetInformer.Informer().HasSynced,
		serviceLister:      serviceInformer.Lister(),
		serviceSynced:      serviceInformer.Informer().HasSynced,
		configMapLister:    configmapInformer.Lister(),
		configMapSynced:    configmapInformer.Informer().HasSynced,
		workqueue:          workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "OpenGausses"),
		recorder:           recorder,
	}

	klog.Info("Setting up event handlers")
	// Set up event handler for OpenGauss
	openGaussInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: controller.enqueueOpenGauss,
		UpdateFunc: func(old, new interface{}) {
			controller.enqueueOpenGauss(new)
		},
	})

	return controller
}

// Run will set up the event handlers for types monitored.
// It will block until stopCh is closed, at which point it will shutdown the workqueue and
// wait for workers to finish processing their current work items.
func (c *Controller) Run(threadiness int, stopCh <-chan struct{}) error {
	defer utilruntime.HandleCrash()
	defer c.workqueue.ShutDown()

	// start the informer factories to begin populating the informer caches
	klog.Info("Starting openGauss controller")

	// wait for the caches to be synced before starting workers
	klog.Info("Syncing informers' caches")
	if ok := cache.WaitForCacheSync(stopCh, c.statefulsetSynced, c.serviceSynced, c.configMapSynced, c.openGaussSynced); !ok {
		return fmt.Errorf("failed to wait for caches to sync")
	}

	// starting workers
	klog.Info("Starting workers")
	// Launch workers to process OpenGauss Resources
	for i := 0; i < threadiness; i++ {
		go wait.Until(c.runWorker, time.Second, stopCh)
	}

	klog.Info("Started workers")
	<-stopCh
	klog.Info("Shutting down workers")

	return nil
}

// runWorker is a long-running function that will continually call the
// processNextWorkItem function in order to read and process a message on the
// workqueue.
func (c *Controller) runWorker() {
	for c.processNextWorkItem() {
	}
}

// processNextWorkItem will read a single work item from workqueue and attempt to process it by calling syncHandler
func (c *Controller) processNextWorkItem() bool {
	obj, shutdown := c.workqueue.Get()
	if shutdown {
		return false
	}
	klog.Info(obj)
	return true
}

// enqueueFoo takes a OpenGauss resource and converts it into a namespace/name
// string which is then put onto the work queue.
// This method should **not** be passed resources of any type other than OpenGauss.
func (c *Controller) enqueueOpenGauss(obj interface{}) {
	var key string
	var err error
	if key, err = cache.MetaNamespaceKeyFunc(obj); err != nil {
		utilruntime.HandleError(err)
		return
	}
	c.workqueue.Add(key)
}
