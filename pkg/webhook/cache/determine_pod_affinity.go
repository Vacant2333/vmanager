package cache

import (
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/klog/v2"

	podaffinity "vacant.sh/vmanager/pkg/definitions/pod-affinity"
	"vacant.sh/vmanager/pkg/webhook/cache/apis"
	"vacant.sh/vmanager/pkg/webhook/cache/utils"
)

// DetermineNewPodAffinityPreference determines what NodeAffinity should be applied to a new Pod
// to meet our requirements, based on data collected from the Cache.
func (wc *WebhookCache) DetermineNewPodAffinityPreference(pod *corev1.Pod) podaffinity.PodAffinitySettingName {
	podSourceWorkloadType, podSourceWorkloadKey := utils.GetPodSourceWorkloadTypeAndKey(pod)

	if podSourceWorkloadType == "" {
		// Return if we cant get the workload type.
		klog.V(3).Infof("Pod %s/%s has no workload type", pod.Namespace, pod.Name)
		return podaffinity.PodAffinityUnset
	}

	if podSourceWorkloadKey == nil {
		// Return if we cant get the source workload key.
		klog.V(3).Infof("Pod source workload key is nil for Pod %s/%s ", pod.Namespace, pod.Name)
		return podaffinity.PodAffinityUnset
	}

	wc.mutex.Lock()
	defer wc.mutex.Unlock()

	switch podSourceWorkloadType {
	case "ReplicaSet":
		return wc.determineNewPodAffinityPreferenceForReplicaSet(*podSourceWorkloadKey)
	case "StatefulSet":
		return wc.determineNewPodAffinityPreferenceForStatefulSet(*podSourceWorkloadKey)
	default:
		return podaffinity.PodAffinityUnset
	}
}

// require mutex locked.
func (wc *WebhookCache) determineNewPodAffinityPreferenceForReplicaSet(replicaSetKey types.NamespacedName) podaffinity.PodAffinitySettingName {
	// First, we need to obtain the cached ReplicaSet, and through it, acquire the associated DeploymentKey.
	replicaSet, ok := wc.replicaSets[replicaSetKey]
	if !ok {
		klog.V(3).Infof("Cant find ReplicaSet in cache by key %v.", replicaSetKey)
		return podaffinity.PodAffinityUnset
	}
	deploymentKey := utils.GetReplicaSetSourceDeploymentKey(replicaSet)
	if deploymentKey == nil {
		klog.V(3).Infof("Cant find ReplicaSet %v source deployment.", replicaSetKey)
		return podaffinity.PodAffinityUnset
	}

	// Then, we need to obtain the Deployment associated with the ReplicaSet,
	// as well as the SchedulingInfo object associated with the ReplicaSet.
	deploymentInfo, ok := wc.deployments[*deploymentKey]
	if !ok {
		klog.Errorf("Cant find Deployment %v in cache, ReplicaSet key %v.", *deploymentKey, replicaSetKey)
		return podaffinity.PodAffinityUnset
	}
	replicaSetSchedulingInfo, ok := wc.replicaSetWorkloadSchedulingInfo[replicaSetKey]
	if !ok {
		klog.Errorf("Cant find ReplicaSet %v scheduling info in cache.", replicaSetKey)
		return podaffinity.PodAffinityUnset
	}

	// Finally, we have collected all the necessary information required to determine the Affinity.
	return wc.determineNewPodAffinityPreference(deploymentInfo.OptimizeSchedulingSetting, replicaSetSchedulingInfo)

}

// require mutex locked.
func (wc *WebhookCache) determineNewPodAffinityPreferenceForStatefulSet(statefulSetKey types.NamespacedName) podaffinity.PodAffinitySettingName {
	statefulSetInfo, ok := wc.statefulSets[statefulSetKey]
	if !ok {
		klog.Errorf("Cant find StatefulSet %v in cache.", statefulSetKey)
		return podaffinity.PodAffinityUnset
	}

	statefulSetSchedulingInfo, ok := wc.statefulSetWorkloadSchedulingInfo[statefulSetKey]
	if !ok {
		klog.Errorf("Cant find StatefulSet %v scheduling info in cache.", statefulSetKey)
		return podaffinity.PodAffinityUnset
	}

	return wc.determineNewPodAffinityPreference(statefulSetInfo.OptimizeSchedulingSetting, statefulSetSchedulingInfo)
}

// require mutex locked.
func (wc *WebhookCache) determineNewPodAffinityPreference(schedulingSetting *apis.OptimizeSchedulingSetting,
	wsi *apis.WorkloadSchedulingInfo) podaffinity.PodAffinitySettingName {

	if schedulingSetting == nil || wsi == nil {
		// This should never happen.
		return podaffinity.PodAffinityUnset
	}

	// Now, we know the target numbers for on-demand and spot Replicas,
	// and we also know how many existing Pods have been marked as on-demand or spot.
	// We simply need to make a straightforward judgment based on this information.

	// If the number of currently created Pods that have been marked as on-demand has not reached the target,
	// then return pod_affinity.PodAffinityOnDemand directly.
	if wsi.OnDemandReplicaCount < schedulingSetting.TargetOnDemandNum {
		return podaffinity.PodAffinityOnDemand
	}

	// If the number of on-demand replicas has been satisfied, under normal circumstances,
	// the remaining Pods should all be assigned to spot.
	if wsi.SpotReplicaCount < schedulingSetting.TargetOnSpotNum {
		return podaffinity.PodAffinitySpot
	}

	return podaffinity.PodAffinityUnset
}
