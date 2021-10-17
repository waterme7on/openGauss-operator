package main

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"

	opengaussv1 "github.com/waterme7on/openGauss-operator/pkg/apis/opengausscontroller/v1"
	clientset "github.com/waterme7on/openGauss-operator/pkg/generated/clientset/versioned"
	ogscheme "github.com/waterme7on/openGauss-operator/pkg/generated/clientset/versioned/scheme"
	informers "github.com/waterme7on/openGauss-operator/pkg/generated/informers/externalversions/opengausscontroller/v1"
	listers "github.com/waterme7on/openGauss-operator/pkg/generated/listers/opengausscontroller/v1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/dynamic"
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

const controllerAgentName = "openGauss-operator"

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
	MessageResourceSynced = "OpenGauss synced successfully"
)

type Controller struct {
	// kubeclientset is a standard kubernetes clientset
	kubeClientset kubernetes.Interface
	// openGaussClientset is a clientset generated for OpenGauss Objects
	openGaussClientset clientset.Interface
	dynamicClient      dynamic.Interface

	// openGauss controller manage service, configmap and statefulset of OpenGauss object
	// thus needing listers of according resources
	openGaussLister   listers.OpenGaussLister
	openGaussSynced   cache.InformerSynced
	deploymentLister  appslisters.DeploymentLister
	deploymentSynced  cache.InformerSynced
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
	dynamicClient dynamic.Interface,
	deploymentInformer appsinformers.DeploymentInformer,
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
		dynamicClient:      dynamicClient,
		openGaussClientset: openGaussClientset,
		openGaussLister:    openGaussInformer.Lister(),
		openGaussSynced:    openGaussInformer.Informer().HasSynced,
		deploymentLister: 	deploymentInformer.Lister(),
		deploymentSynced: 	deploymentInformer.Informer().HasSynced,
		statefulsetLister:  statefulsetInformer.Lister(),
		statefulsetSynced:  statefulsetInformer.Informer().HasSynced,
		serviceLister:      serviceInformer.Lister(),
		serviceSynced:      serviceInformer.Informer().HasSynced,
		configMapLister:    configmapInformer.Lister(),
		configMapSynced:    configmapInformer.Informer().HasSynced,
		workqueue:          workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "OpenGausses"),
		recorder:           recorder,
	}

	klog.Infoln("Setting up event handlers")
	// Set up event handler for OpenGauss
	openGaussInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: controller.enqueueOpenGauss,
		UpdateFunc: func(old, new interface{}) {
			controller.enqueueOpenGauss(new)
		},
	})

	deploymentInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: controller.handleObjects,
		UpdateFunc: func(old, new interface{}) {
			newDepl := new.(*appsv1.Deployment)
			oldDepl := old.(*appsv1.Deployment)
			if newDepl.ResourceVersion == oldDepl.ResourceVersion {
				// Periodic resync will send update events for all known Deployments.
				// Two different versions of the same Deployment will always have different RVs.
				return
			}
			controller.handleObjects(new)
		},
		DeleteFunc: controller.handleObjects,
	})

	statefulsetInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: controller.handleObjects,
		UpdateFunc: func(old, new interface{}) {
			newSts := new.(*appsv1.StatefulSet)
			oldSts := old.(*appsv1.StatefulSet)
			if newSts.ResourceVersion == oldSts.APIVersion {
				return
			}
			controller.handleObjects(new)
		},
		DeleteFunc: controller.handleObjects,
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
	klog.Infoln("Starting openGauss controller")

	// wait for the caches to be synced before starting workers
	klog.Infoln("Syncing informers' caches")
	if ok := cache.WaitForCacheSync(stopCh, c.deploymentSynced, c.statefulsetSynced, c.serviceSynced, c.configMapSynced, c.openGaussSynced); !ok {
		return fmt.Errorf("failed to wait for caches to sync")
	}

	// starting workers
	klog.Infoln("Starting workers")
	// Launch workers to process OpenGauss Resources
	for i := 0; i < threadiness; i++ {
		go wait.Until(c.runWorker, time.Second, stopCh)
	}

	klog.Infoln("Started workers")
	<-stopCh
	klog.Infoln("Shutting down workers")

	return nil
}

// runWorker is a long-running function that will continually call the
// processNextWorkItem function in order to read and process a message on the
// workqueue.
func (c *Controller) runWorker() {
	for c.processNextWorkItem() {
		time.Sleep(time.Second * 5)
	}
}

// processNextWorkItem will read a single work item from workqueue and attempt to process it by calling syncHandler
func (c *Controller) processNextWorkItem() bool {
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

// syncHandler compares the actual state with the desired and attempt to coverge the two.
// It then updates the status of OpenGauss
func (c *Controller) syncHandler(key string) error {
	// Convert the namespace/name into a distinct namespace and name
	namespace, name, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		utilruntime.HandleError(fmt.Errorf("invalid resource key: %s", key))
		return nil
	}

	// Get the openGauss resource with the namespace and name
	og, err := c.openGaussLister.OpenGausses(namespace).Get(name)

	if err != nil {
		// The openGauss object may not exist.
		if errors.IsNotFound(err) {
			utilruntime.HandleError(fmt.Errorf("openGauss '%s' in work queue no longer exists", key))
			return nil
		}
		return err
	}
	klog.Info("Syncing status of OpenGauss ", og.Name)

	// 第一阶段不同于sample-controller，在object.go中新定义的函数可以创建或者获得
	// 1. check if all components are deployed, includes service, configmap, master and worker statefulsets
	// create or get pvc
	var pvc *corev1.PersistentVolumeClaim
	pvcConfig := NewPersistentVolumeClaim(og)
	pvc, err = c.createOrGetPVC(og.Namespace, pvcConfig)
	if err != nil {
		return err
	}

	// create or update master configmap
	masterConfigMap, masterConfigMapRes := NewMasterConfigMap(og)
	err = c.createOrUpdateConfigMap(og.Namespace, masterConfigMap, masterConfigMapRes)
	if err != nil {
		return err
	}

	// create or get master statefulset
	var masterStatefulset *appsv1.StatefulSet
	masterStsConfig := NewMasterStatefulsets(og)
	masterStatefulset, err = c.createOrGetStatefulset(og.Namespace, masterStsConfig)
	// If an error occurs during Get/Create, we'll requeue the item so we can
	// attempt processing again later. This could have been caused by a
	// temporary network failure, or any other transient reason.
	if err != nil {
		return err
	}

	// create or update replica configmap
	replicaConfigMap, relicaConfigMapRes := NewReplicaConfigMap(og)
	err = c.createOrUpdateConfigMap(og.Namespace, replicaConfigMap, relicaConfigMapRes)
	if err != nil {
		return err
	}

	// create or get replica statefulset
	var replicasStatefulset *appsv1.StatefulSet
	replicaStsConfig := NewReplicaStatefulsets(og)
	replicasStatefulset, err = c.createOrGetStatefulset(og.Namespace, replicaStsConfig)
	// If an error occurs during Get/Create, we'll requeue the item so we can
	// attempt processing again later. This could have been caused by a
	// temporary network failure, or any other transient reason.
	if err != nil {
		return err
	}

	// create or get mycat deployment
	mycatDeployConfig := NewMycatDeployment(og)
	mycatDeployment, err := c.createOrGetDeployment(og.Namespace, mycatDeployConfig)
	
	if err != nil {
		return err
	}

	// 2. check if all components are controlled by opengauss
	// checked if statefulsets are controlled by this og resource
	if !v1.IsControlledBy(masterStatefulset, og) {
		msg := fmt.Sprintf(MessageResourceExists, masterStatefulset.Name)
		c.recorder.Event(og, corev1.EventTypeWarning, ErrResourceExists, msg)
		return fmt.Errorf(msg)
	}
	if !v1.IsControlledBy(replicasStatefulset, og) {
		msg := fmt.Sprintf(MessageResourceExists, replicasStatefulset.Name)
		c.recorder.Event(og, corev1.EventTypeWarning, ErrResourceExists, msg)
		return fmt.Errorf(msg)
	}
	if !v1.IsControlledBy(mycatDeployment, og) {
		msg := fmt.Sprintf(MessageResourceExists, mycatDeployment.Name)
		c.recorder.Event(og, corev1.EventTypeWarning, ErrResourceExists, msg)
		return fmt.Errorf(msg)
	}

	// 3. check if the status of all components satisfy
	// checked if replicas number are correct
	if *og.Spec.OpenGauss.Master.Replicas != (*masterStatefulset.Spec.Replicas) ||
		*og.Spec.OpenGauss.Worker.Replicas != (*replicasStatefulset.Spec.Replicas) {
		// update configmap
		masterConfigMap, masterConfigMapRes := NewMasterConfigMap(og)
		err = c.createOrUpdateConfigMap(og.Namespace, masterConfigMap, masterConfigMapRes)
		replicaConfigMap, replicaConfigMapRes := NewReplicaConfigMap(og)
		err = c.createOrUpdateConfigMap(og.Namespace, replicaConfigMap, replicaConfigMapRes)
		// update statefulset
		klog.V(4).Infof("OpenGauss '%s' specified master replicas: %d, master statefulset Replicas %d", name, *og.Spec.OpenGauss.Master.Replicas, *masterStatefulset.Spec.Replicas)
		masterStatefulset, err = c.kubeClientset.AppsV1().StatefulSets(og.Namespace).Update(context.TODO(), NewMasterStatefulsets(og), v1.UpdateOptions{})
		klog.V(4).Infof("OpenGauss '%s' specified master replicas: %d, master statefulset Replicas %d", name, *og.Spec.OpenGauss.Worker.Replicas, *replicasStatefulset.Spec.Replicas)
		replicasStatefulset, err = c.kubeClientset.AppsV1().StatefulSets(og.Namespace).Update(context.TODO(), NewReplicaStatefulsets(og), v1.UpdateOptions{})
	}
	// checked if persistent volume claims are correct
	if *og.Spec.Resources.Requests.Storage() != *pvc.Spec.Resources.Requests.Storage() {
		klog.V(4).Infof("Update OpenGauss pvc storage")
		pvc, err = c.kubeClientset.CoreV1().PersistentVolumeClaims(og.Namespace).Update(context.TODO(), NewPersistentVolumeClaim(og), v1.UpdateOptions{})
	}
	if err != nil {
		return err
	}

	// check if mycat deployment is correct
	
	if *og.Spec.OpenGauss.Mycat.Replicas != *mycatDeployment.Spec.Replicas {
		klog.V(4).Infof("Openguass %s mycat deployments, expected replicas: %d, actual replicas: %d", og.Name, *og.Spec.OpenGauss.Mycat.Replicas, *mycatDeployment.Spec.Replicas)
		mycatDeployment, err = c.kubeClientset.AppsV1().Deployments(og.Namespace).Update(context.TODO(), NewMycatDeployment(og), v1.UpdateOptions{})
	}
	if err != nil {
		return err
	}

	// finally update opengauss resource status
	err = c.updateOpenGaussStatus(og, masterStatefulset, replicasStatefulset, mycatDeployment, pvc)
	if err != nil {
		return err
	}

	// record normal event
	c.recorder.Event(og, corev1.EventTypeNormal, SuccessSynced, MessageResourceSynced)
	return nil
}

// update opengauss status
func (c *Controller) updateOpenGaussStatus(
	og *opengaussv1.OpenGauss,
	masterStatefulset *appsv1.StatefulSet,
	replicasStatefulset *appsv1.StatefulSet,
	deployment *appsv1.Deployment,
	pvc *corev1.PersistentVolumeClaim) error {
	var err error
	ogCopy := og.DeepCopy()
	if ogCopy.Status == nil {
		ogCopy.Status = &opengaussv1.OpenGaussStatus{}
	}
	ogCopy.Status.MasterStatefulset = masterStatefulset.Name
	ogCopy.Status.ReplicasStatefulset = replicasStatefulset.Name
	ogCopy.Status.ReadyMaster = (strconv.Itoa(int(masterStatefulset.Status.ReadyReplicas)))
	ogCopy.Status.ReadyReplicas = (strconv.Itoa(int(replicasStatefulset.Status.ReadyReplicas)))
	ogCopy.Status.ReadyMycat = (strconv.Itoa(int(deployment.Status.ReadyReplicas)))
	ogCopy.Status.PersistentVolumeClaimName = pvc.Name
	if (masterStatefulset.Status.ReadyReplicas) == *ogCopy.Spec.OpenGauss.Master.Replicas &&
	(replicasStatefulset.Status.ReadyReplicas) == *ogCopy.Spec.OpenGauss.Worker.Replicas &&
	deployment.Status.ReadyReplicas == 1{
		ogCopy.Status.OpenGaussStatus = "READY"
	}
	ogCopy, err = c.openGaussClientset.ControllerV1().OpenGausses(ogCopy.Namespace).UpdateStatus(context.TODO(), ogCopy, v1.UpdateOptions{})
	if err != nil {
		klog.Infoln("Failed to update opengauss status:", ogCopy.Name, " error:", err)
	}
	return err
}

// createOrUpdatePVC creates or get pvc of opengauss
func (c *Controller) createOrGetPVC(ns string, config *corev1.PersistentVolumeClaim) (pvc *corev1.PersistentVolumeClaim, err error) {
	// get pvc
	klog.V(4).Infoln("try to get pvc for opengauss:", config.Name)
	pvc, err = c.kubeClientset.CoreV1().PersistentVolumeClaims(ns).Get(context.TODO(), config.Name, v1.GetOptions{})
	if err != nil {
		// (try to) create pvc
		klog.V(4).Infoln("try to create pvc for opengauss:", config.Name)
		pvc, err = c.kubeClientset.CoreV1().PersistentVolumeClaims(ns).Create(context.TODO(), config, v1.CreateOptions{})
	}
	return
}

// createOrGetStatefulset creates or get statefulset of opengauss
func (c *Controller) createOrGetStatefulset(ns string, config *appsv1.StatefulSet) (sts *appsv1.StatefulSet, err error) {
	// get pvc
	klog.V(4).Infoln("try to get statefulset for opengauss:", config.Name)
	sts, err = c.kubeClientset.AppsV1().StatefulSets(ns).Get(context.TODO(), config.Name, v1.GetOptions{})
	if err != nil {
		// (try to) create pvc
		klog.V(4).Infoln("try to create statefulset for opengauss:", config.Name)
		sts, err = c.kubeClientset.AppsV1().StatefulSets(ns).Create(context.TODO(), config, v1.CreateOptions{})
	}
	if err != nil {
		klog.V(4).Infoln(config.Spec)
	}
	return
}

// createOrGetDeployment creates or get deployment of mycat
func (c *Controller) createOrGetDeployment(ns string, config *appsv1.Deployment) (deployment *appsv1.Deployment, err error){
	// get deployment
	klog.V(4).Infoln("try to get deployment for opengauss:", config.Name)
	deployment, err = c.deploymentLister.Deployments(ns).Get(config.Name)
	if err != nil {
		klog.V(4).Infoln("try to create deployment for opengauss:", config.Name)
		deployment, err = c.kubeClientset.AppsV1().Deployments(ns).Create(context.TODO(), config, v1.CreateOptions{})
	}
	if err != nil {
		klog.V(4).Infoln(config.Spec)
	}
	return
}

// createOrUpdateConfigMap creates or update configmap for opengauss
func (c *Controller) createOrUpdateConfigMap(ns string, cm *unstructured.Unstructured, cmRes schema.GroupVersionResource) error {
	klog.V(4).Infoln("try to create configmap:", cm.GetName())
	_, err := c.dynamicClient.Resource(cmRes).Namespace(ns).Create(context.TODO(), cm, v1.CreateOptions{})
	if err != nil {
		klog.V(4).Infoln("failed to create, try to update configmap:", cm.GetName())
		_, err = c.dynamicClient.Resource(cmRes).Namespace(ns).Update(context.TODO(), cm, v1.UpdateOptions{})
	}
	if err != nil {
		klog.Infoln("failed to create or update configmap:", cm.GetName())
	}
	return err
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

// handdleObjects will take any resource implementing metav1.Object and attempt
// to find the opengauss resource that owns it.
// It does this by looking at the objects metadata.ownerReferences field for an appropriate OwnerReference
// It then enqueues that opengauss resource to be processed.
// If the resource does not have a ownerReference, it will be skipped.
func (c *Controller) handleObjects(obj interface{}) {
	var object v1.Object
	var ok bool
	if object, ok = obj.(v1.Object); !ok {
		tombstone, ok := obj.(cache.DeletedFinalStateUnknown)
		if !ok {
			utilruntime.HandleError(fmt.Errorf("error decoding object, invalid type"))
			return
		}
		object, ok = tombstone.Obj.(v1.Object)
		if !ok {
			utilruntime.HandleError(fmt.Errorf("error decoding object tombstone, invalid type"))
			return
		}
		klog.V(4).Infof("Recovered deleted object '%s' from tombstone", object.GetName())
	}
	klog.V(4).Infof("Processing object: %s", object.GetName())
	if ownerRef := v1.GetControllerOf(object); ownerRef != nil {
		// If this object is not owned by a Foo, we should not do anything more
		// with it.
		if ownerRef.Kind != "OpenGauss" {
			return
		}

		og, err := c.openGaussLister.OpenGausses(object.GetNamespace()).Get(ownerRef.Name)
		if err != nil {
			klog.V(4).Infof("ignoring orphaned object '%s' of og '%s'", object.GetSelfLink(), ownerRef.Name)
			return
		}

		c.enqueueOpenGauss(og)
		return
	}
}
