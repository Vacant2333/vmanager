package cache

import (
	"sync"

	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/informers"
	informerappsv1 "k8s.io/client-go/informers/apps/v1"
	informercorev1 "k8s.io/client-go/informers/core/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	"k8s.io/klog/v2"

	"vacant.sh/vmanager/pkg/webhook/cache/apis"
)

// Check if the Cache implements necessary func.
var _ Interface = &WebhookCache{}

type WebhookCache struct {
	mutex sync.Mutex

	kubeClient      kubernetes.Interface
	informerFactory informers.SharedInformerFactory

	replicaSetInformer informerappsv1.ReplicaSetInformer
	replicaSets        map[types.NamespacedName]*appsv1.ReplicaSet

	deploymentInformer informerappsv1.DeploymentInformer
	deployments        map[types.NamespacedName]*apis.DeploymentInfo

	statefulSetInformer informerappsv1.StatefulSetInformer
	statefulSets        map[types.NamespacedName]*apis.StatefulSetInfo

	podInformer                       informercorev1.PodInformer
	replicaSetWorkloadSchedulingInfo  map[types.NamespacedName]*apis.WorkloadSchedulingInfo
	statefulSetWorkloadSchedulingInfo map[types.NamespacedName]*apis.WorkloadSchedulingInfo
}

func NewWebhookCache(kubeConfig *rest.Config) (Interface, error) {
	kubeClient, err := kubernetes.NewForConfig(kubeConfig)
	if err != nil {
		return nil, err
	}

	wc := &WebhookCache{
		kubeClient:      kubeClient,
		informerFactory: informers.NewSharedInformerFactory(kubeClient, 0),

		replicaSets:  map[types.NamespacedName]*appsv1.ReplicaSet{},
		deployments:  map[types.NamespacedName]*apis.DeploymentInfo{},
		statefulSets: map[types.NamespacedName]*apis.StatefulSetInfo{},

		replicaSetWorkloadSchedulingInfo:  map[types.NamespacedName]*apis.WorkloadSchedulingInfo{},
		statefulSetWorkloadSchedulingInfo: map[types.NamespacedName]*apis.WorkloadSchedulingInfo{},
	}

	wc.deploymentInformer = wc.informerFactory.Apps().V1().Deployments()
	_, err = wc.deploymentInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    wc.addDeployment,
		UpdateFunc: wc.updateDeployment,
		DeleteFunc: wc.deleteDeployment,
	})
	if err != nil {
		return nil, err
	}

	wc.statefulSetInformer = wc.informerFactory.Apps().V1().StatefulSets()
	_, err = wc.statefulSetInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    wc.addStatefulSet,
		UpdateFunc: wc.updateStatefulSet,
		DeleteFunc: wc.deleteStatefulSet,
	})
	if err != nil {
		return nil, err
	}

	wc.podInformer = wc.informerFactory.Core().V1().Pods()
	_, err = wc.podInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    wc.addPod,
		UpdateFunc: wc.updatePod,
		DeleteFunc: wc.deletePod,
	})

	wc.replicaSetInformer = wc.informerFactory.Apps().V1().ReplicaSets()
	_, err = wc.replicaSetInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    wc.addReplicaSet,
		UpdateFunc: wc.updateReplicaSet,
		DeleteFunc: wc.deleteReplicaSet,
	})

	return wc, nil
}

func (wc *WebhookCache) Run(stopCh <-chan struct{}) {
	// Start the informerFactory, wait for cache sync.
	wc.informerFactory.Start(stopCh)

	for informerType, ok := range wc.informerFactory.WaitForCacheSync(stopCh) {
		if !ok {
			klog.Errorf("Cache failed to sync: %v", informerType)
		}
	}

	klog.V(2).Info("WebhookCache start to run.")
}
