package cache

import (
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/klog/v2"

	podaffinity "vacant.sh/vmanager/pkg/definitions/pod-affinity"
	"vacant.sh/vmanager/pkg/webhook/cache/apis"
	"vacant.sh/vmanager/pkg/webhook/cache/utils"
)

func (wc *WebhookCache) addReplicaSet(obj interface{}) {
	replicaSet := convertToReplicaSet(obj)

	if replicaSet == nil {
		return
	}

	wc.mutex.Lock()
	defer wc.mutex.Unlock()

	replicaSetKey := types.NamespacedName{
		Namespace: replicaSet.Namespace,
		Name:      replicaSet.Name,
	}

	wc.replicaSets[replicaSetKey] = replicaSet
	if wc.replicaSetWorkloadSchedulingInfo[replicaSetKey] == nil {
		wc.replicaSetWorkloadSchedulingInfo[replicaSetKey] = apis.NewWorkloadSchedulingInfo()
	}

	klog.V(5).Infof("Added ReplicaSet %s/%s", replicaSet.Namespace, replicaSet.Name)
}

func (wc *WebhookCache) deleteReplicaSet(obj interface{}) {
	replicaSet := convertToReplicaSet(obj)

	if replicaSet == nil {
		return
	}

	wc.mutex.Lock()
	defer wc.mutex.Unlock()

	delete(wc.replicaSets, types.NamespacedName{
		Namespace: replicaSet.Namespace,
		Name:      replicaSet.Name,
	})
}

func (wc *WebhookCache) updateReplicaSet(oldObj, newObj interface{}) {
	wc.deleteReplicaSet(oldObj)
	wc.addReplicaSet(newObj)
}

func (wc *WebhookCache) addDeployment(obj interface{}) {
	deployment := convertToDeployment(obj)

	if deployment == nil {
		return
	}

	wc.mutex.Lock()
	defer wc.mutex.Unlock()

	wc.deployments[types.NamespacedName{
		Namespace: deployment.Namespace,
		Name:      deployment.Name,
	}] = apis.NewDeploymentInfo(deployment)

	klog.V(5).Infof("Added DeploymentInfo %s/%s", deployment.Namespace, deployment.Name)
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
	wc.deleteDeployment(oldObj)
	wc.addDeployment(newObj)
}

func (wc *WebhookCache) addStatefulSet(obj interface{}) {
	statefulSet := convertToStatefulSet(obj)

	if statefulSet == nil {
		return
	}

	wc.mutex.Lock()
	defer wc.mutex.Unlock()

	statefulSetKey := types.NamespacedName{
		Namespace: statefulSet.Namespace,
		Name:      statefulSet.Name,
	}

	wc.statefulSets[statefulSetKey] = apis.NewStatefulSetInfo(statefulSet)
	if wc.statefulSetWorkloadSchedulingInfo[statefulSetKey] == nil {
		wc.statefulSetWorkloadSchedulingInfo[statefulSetKey] = apis.NewWorkloadSchedulingInfo()
	}

	klog.V(5).Infof("Added StatefulSetInfo %s/%s", statefulSet.Namespace, statefulSet.Name)
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
	wc.deleteStatefulSet(oldObj)
	wc.addStatefulSet(newObj)
}

func (wc *WebhookCache) addPod(obj interface{}) {
	pod := convertToPod(obj)

	if pod == nil {
		return
	}

	podKey := types.NamespacedName{Name: pod.Name, Namespace: pod.Namespace}
	podSourceWorkloadType, podSourceWorkloadKey := utils.GetPodSourceWorkloadTypeAndKey(pod)

	// In the first step, we need to identify the source workload type of the Pod, such as StatefulSet and ReplicaSet
	// (which may not necessarily have a corresponding Deployment).
	if podSourceWorkloadType == "" {
		// Return if we cant get the workload type.
		klog.V(3).Infof("Pod %s/%s has no workload type", pod.Namespace, pod.Name)
		return
	}

	// The second step is to get the Key of the source workload.
	if podSourceWorkloadKey == nil {
		// Return if we cant get the source workload key.
		klog.V(3).Infof("Pod source workload key is nil for Pod %s/%s ", pod.Namespace, pod.Name)
		return
	}

	// The third step is to determine whether the Pod has been marked with our required Affinity by the Webhook,
	// which will be indicated in the label pod_affinity.PodAffinityLabelKey.
	podAffinitySetting := utils.GetPodAffinitySetting(pod)

	wc.mutex.Lock()
	defer wc.mutex.Unlock()

	// The fourth step is to retrieve the WorkloadSchedulingInfo of this Pod from the Cache,
	// which includes the status of the workload we are concerned about,
	// such as the number of existing Pods and the number of Pods marked with Affinity.
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
		wsi = wc.statefulSetWorkloadSchedulingInfo[*podSourceWorkloadKey]
	default:
		// Return if it's not ReplicaSet or StatefulSet.
		return
	}

	// The fifth step is to maintain the consistency of the WorkloadSchedulingInfo data.
	wsi.Pods[podKey] = podAffinitySetting
	switch podAffinitySetting {
	case podaffinity.PodAffinityOnDemand:
		wsi.OnDemandReplicaCount++
	case podaffinity.PodAffinitySpot:
		wsi.SpotReplicaCount++
	}
}

func (wc *WebhookCache) deletePod(obj interface{}) {
	pod := convertToPod(obj)

	if pod == nil {
		return
	}

	podKey := types.NamespacedName{Name: pod.Name, Namespace: pod.Namespace}
	podSourceWorkloadType, podSourceWorkloadKey := utils.GetPodSourceWorkloadTypeAndKey(pod)

	if podSourceWorkloadType == "" {
		// Return if we cant get the workload type.
		klog.V(3).Infof("Pod %s/%s has no workload type", pod.Namespace, pod.Name)
		return
	}

	if podSourceWorkloadKey == nil {
		// Return if we cant get the source workload key.
		klog.V(3).Infof("Pod source workload key is nil for Pod %s/%s ", pod.Namespace, pod.Name)
		return
	}

	wc.mutex.Lock()
	defer wc.mutex.Unlock()

	var wsi *apis.WorkloadSchedulingInfo

	switch podSourceWorkloadType {
	case "ReplicaSet":
		wsi = wc.replicaSetWorkloadSchedulingInfo[*podSourceWorkloadKey]
	case "StatefulSet":
		wsi = wc.statefulSetWorkloadSchedulingInfo[*podSourceWorkloadKey]
	default:
		// Return if it's not ReplicaSet or StatefulSet.
		return
	}

	if wsi.Pods[podKey] == "" {
		klog.Errorf("Cant find pod affinity setting for Pod %s/%s, workload type: %v, workload key: %v ",
			pod.Namespace, pod.Name, podSourceWorkloadType, podSourceWorkloadKey)
		return
	}

	// In deletePod, the initial steps are the same as in addPod.
	// The difference here is that the type of the Pod’s Affinity is directly obtained from the Cache,
	// and then it is used to maintain the data of our WorkloadSchedulingInfo.
	switch wsi.Pods[podKey] {
	case podaffinity.PodAffinityOnDemand:
		wsi.OnDemandReplicaCount--
	case podaffinity.PodAffinitySpot:
		wsi.SpotReplicaCount--
	}
	delete(wsi.Pods, podKey)

	// If all the Pods of a certain workload have been deleted, then we can choose to clear the Cache.
	if len(wsi.Pods) == 0 {
		switch podSourceWorkloadType {
		case "ReplicaSet":
			delete(wc.replicaSetWorkloadSchedulingInfo, *podSourceWorkloadKey)
		case "StatefulSet":
			delete(wc.statefulSetWorkloadSchedulingInfo, *podSourceWorkloadKey)
		}
	}
}

func (wc *WebhookCache) updatePod(oldObj, newObj interface{}) {
	wc.deletePod(oldObj)
	wc.addPod(newObj)
}
