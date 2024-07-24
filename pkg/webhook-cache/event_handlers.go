package webhook_cache

import (
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/klog/v2"

	pod_affinity "vacant.sh/vmanager/pkg/definitions/pod-affinity"
	"vacant.sh/vmanager/pkg/webhook-cache/apis"

	"vacant.sh/vmanager/pkg/webhook-cache/utils"
)

func (wc *WebhookCache) addDeployment(obj interface{}) {
	deployment := convertToDeployment(obj)

	if deployment == nil {
		return
	}

	wc.mutex.Lock()
	defer wc.mutex.Unlock()

	deploymentInfo := apis.NewDeploymentInfo(deployment)

	wc.deployments[types.NamespacedName{
		Namespace: deployment.Namespace,
		Name:      deployment.Name,
	}] = deploymentInfo

	klog.V(5).Infof("Added DeploymentInfo %v", deploymentInfo)
}

func (wc *WebhookCache) deleteDeployment(obj interface{}) {
	deployment := convertToDeployment(obj)

	if deployment == nil {
		return
	}

	wc.mutex.Lock()
	defer wc.mutex.Unlock()

	delete(wc.deployments, types.NamespacedName{
		Namespace: deployment.Namespace,
		Name:      deployment.Name,
	})
}

func (wc *WebhookCache) updateDeployment(oldObj, newObj interface{}) {
	oldDeployment := convertToDeployment(oldObj)
	newDeployment := convertToDeployment(newObj)

	if oldDeployment == nil || newDeployment == nil {
		return
	}

	wc.deleteDeployment(oldDeployment)
	wc.addDeployment(newDeployment)
}

func (wc *WebhookCache) addStatefulSet(obj interface{}) {
	statefulSet := convertToStatefulSet(obj)

	if statefulSet == nil {
		return
	}

	wc.mutex.Lock()
	defer wc.mutex.Unlock()

	statefulSetInfo := apis.NewStatefulSetInfo(statefulSet)

	wc.statefulSets[types.NamespacedName{
		Namespace: statefulSet.Namespace,
		Name:      statefulSet.Name,
	}] = statefulSetInfo

	klog.V(5).Infof("Added StatefulSetInfo %v", statefulSetInfo)
}

func (wc *WebhookCache) deleteStatefulSet(obj interface{}) {
	statefulSet := convertToStatefulSet(obj)

	if statefulSet == nil {
		return
	}

	wc.mutex.Lock()
	defer wc.mutex.Unlock()

	delete(wc.statefulSets, types.NamespacedName{
		Namespace: statefulSet.Namespace,
		Name:      statefulSet.Name,
	})
}

func (wc *WebhookCache) updateStatefulSet(oldObj, newObj interface{}) {
	oldStatefulSet := convertToStatefulSet(oldObj)
	newStatefulSet := convertToStatefulSet(newObj)

	if oldStatefulSet == nil || newStatefulSet == nil {
		return
	}

	wc.deleteStatefulSet(oldStatefulSet)
	wc.addStatefulSet(newStatefulSet)
}

func (wc *WebhookCache) addPod(obj interface{}) {
	pod := convertToPod(obj)

	if pod == nil {
		return
	}

	podKey := types.NamespacedName{Name: pod.Name, Namespace: pod.Namespace}

	podSourceWorkloadType := utils.GetPodSourceWorkloadType(pod)
	if podSourceWorkloadType == "" {
		// Return if we cant get the workload type.
		klog.V(3).Infof("Pod %s/%s has no workload type", pod.Namespace, pod.Name)
		return
	}

	podSourceWorkloadKey := utils.GetPodSourceResourceKey(pod)
	if podSourceWorkloadKey == nil {
		// Return if we cant get the source workload key.
		klog.V(3).Infof("PodSourceWorkloadKey is nil for Pod %s/%s ", pod.Namespace, pod.Name)
		return
	}

	podAffinitySetting := utils.GetPodAffinitySetting(pod)

	wc.mutex.Lock()
	defer wc.mutex.Unlock()

	var wsi *apis.WorkloadSchedulingInfo

	switch podSourceWorkloadType {
	case "ReplicaSet":
		if wc.replicaSetWorkloadSchedulingInfo[*podSourceWorkloadKey] == nil {
			wc.replicaSetWorkloadSchedulingInfo[*podSourceWorkloadKey] = apis.NewWorkloadSchedulingInfo()
		}
		wsi = wc.replicaSetWorkloadSchedulingInfo[*podSourceWorkloadKey]
	case "StatefulSet":
		if wc.statefulSetWorkloadSchedulingInfo[*podSourceWorkloadKey] == nil {
			wc.statefulSetWorkloadSchedulingInfo[*podSourceWorkloadKey] = apis.NewWorkloadSchedulingInfo()
		}
		wsi = wc.replicaSetWorkloadSchedulingInfo[*podSourceWorkloadKey]
	default:
		// Return if it's not ReplicaSet or StatefulSet.
		return
	}

	wsi.Pods[podKey] = podAffinitySetting

	switch podAffinitySetting {
	case pod_affinity.PodAffinityOnDemand:
		wsi.OnDemandReplicaCount++
	case pod_affinity.PodAffinitySpot:
		wsi.SpotReplicaCount++
	}
}

func (wc *WebhookCache) deletePod(obj interface{}) {
	pod := convertToPod(obj)

	if pod == nil {
		return
	}

	podKey := types.NamespacedName{Name: pod.Name, Namespace: pod.Namespace}

	podSourceWorkloadType := utils.GetPodSourceWorkloadType(pod)
	if podSourceWorkloadType == "" {
		// Return if we cant get the workload type.
		klog.V(3).Infof("Pod %s/%s has no workload type", pod.Namespace, pod.Name)
		return
	}

	podSourceWorkloadKey := utils.GetPodSourceResourceKey(pod)
	if podSourceWorkloadKey == nil {
		// Return if we cant get the source workload key.
		klog.V(3).Infof("PodSourceWorkloadKey is nil for Pod %s/%s ", pod.Namespace, pod.Name)
		return
	}

	wc.mutex.Lock()
	defer wc.mutex.Unlock()

	var wsi *apis.WorkloadSchedulingInfo

	switch podSourceWorkloadType {
	case "ReplicaSet":
		wsi = wc.replicaSetWorkloadSchedulingInfo[*podSourceWorkloadKey]
	case "StatefulSet":
		wsi = wc.replicaSetWorkloadSchedulingInfo[*podSourceWorkloadKey]
	default:
		// Return if it's not ReplicaSet or StatefulSet.
		return
	}

	if wsi.Pods[podKey] == "" {
		klog.Errorf("Cant find pod affinity setting for Pod %s/%s, workload type: %v, workload key: %v ",
			pod.Namespace, pod.Name, podSourceWorkloadType, podSourceWorkloadKey)
		return
	}

	podAffinitySetting := wsi.Pods[podKey]

	switch podAffinitySetting {
	case pod_affinity.PodAffinityOnDemand:
		wsi.OnDemandReplicaCount--
	case pod_affinity.PodAffinitySpot:
		wsi.SpotReplicaCount--
	}

	delete(wsi.Pods, podKey)

	if len(wsi.Pods) == 0 {
		switch podSourceWorkloadType {
		case "ReplicaSet":
			delete(wc.replicaSetWorkloadSchedulingInfo, *podSourceWorkloadKey)
		case "StatefulSet":
			delete(wc.statefulSetWorkloadSchedulingInfo, *podSourceWorkloadKey)
		}
	}
}
