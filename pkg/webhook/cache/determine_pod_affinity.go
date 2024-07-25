package cache

import (
	"time"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/klog/v2"

	podaffinity "vacant.sh/vmanager/pkg/definitions/pod-affinity"
	"vacant.sh/vmanager/pkg/webhook/cache/apis"
	"vacant.sh/vmanager/pkg/webhook/cache/utils"
)

// todo: Consider updating the cache when returning the Patch.
// Otherwise, if the pods are created too quickly and the cache has not synchronized, such as with three pods,
// and the strategy requires a majority-in-on-demand, there might be a situation where all three pods are on-demand.

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

	// In some cases, we need to attempt a retry and wait for the cache to synchronize the deployment.
	// For example, when a deployment has just been created, the pods might have already been generated,
	// but the replica set has not yet been synchronized to our cache.
	var result podaffinity.PodAffinitySettingName

	err := wait.ExponentialBackoff(wait.Backoff{
		Duration: 100 * time.Millisecond,
		Factor:   1.0,
		Jitter:   0.1,
		Steps:    5,
	}, func() (bool, error) {
		var needRetry bool

		switch podSourceWorkloadType {
		case "ReplicaSet":
			result, needRetry = wc.determineNewPodAffinityPreferenceForReplicaSet(*podSourceWorkloadKey)
		case "StatefulSet":
			result, needRetry = wc.determineNewPodAffinityPreferenceForStatefulSet(*podSourceWorkloadKey)
		}

		return !needRetry, nil
	})

	if err != nil {
		klog.Errorf("Retry determine new pod affinity for Pod %s/%s failed: %v, return unset.", pod.Namespace, pod.Name, err)
		return podaffinity.PodAffinityUnset
	}
	return result
}

func (wc *WebhookCache) determineNewPodAffinityPreferenceForReplicaSet(replicaSetKey types.NamespacedName) (podaffinity.PodAffinitySettingName, bool) {
	wc.mutex.Lock()
	defer wc.mutex.Unlock()

	// First, we need to obtain the cached ReplicaSet, and through it, acquire the associated DeploymentKey.
	replicaSet, ok := wc.replicaSets[replicaSetKey]
	if !ok {
		klog.V(3).Infof("Cant find ReplicaSet in cache by key %v, wait for cache sync.", replicaSetKey)
		return "", true
	}
	deploymentKey := utils.GetReplicaSetSourceDeploymentKey(replicaSet)
	if deploymentKey == nil {
		// The ReplicaSet have no parent Deployment resource.
		klog.V(3).Infof("Cant find ReplicaSet %v source deployment.", replicaSetKey)
		return podaffinity.PodAffinityUnset, false
	}

	// Then, we need to obtain the Deployment associated with the ReplicaSet,
	// as well as the SchedulingInfo object associated with the ReplicaSet.
	deploymentInfo, ok := wc.deployments[*deploymentKey]
	if !ok {
		klog.Infof("Cant find Deployment %v in cache, ReplicaSet key %v, wait for cache sync.", *deploymentKey, replicaSetKey)
		return "", true
	}
	replicaSetSchedulingInfo, ok := wc.replicaSetWorkloadSchedulingInfo[replicaSetKey]
	if !ok {
		// This should never happen.
		klog.Errorf("Cant find ReplicaSet %v scheduling info in cache.", replicaSetKey)
		return podaffinity.PodAffinityUnset, false
	}

	// Finally, we have collected all the necessary information required to determine the Affinity.
	return wc.determineNewPodAffinityPreference(deploymentInfo.OptimizeSchedulingSetting, replicaSetSchedulingInfo), false
}

func (wc *WebhookCache) determineNewPodAffinityPreferenceForStatefulSet(statefulSetKey types.NamespacedName) (podaffinity.PodAffinitySettingName, bool) {
	wc.mutex.Lock()
	defer wc.mutex.Unlock()

	statefulSetInfo, ok := wc.statefulSets[statefulSetKey]
	if !ok {
		klog.Errorf("Cant find StatefulSet %v in cache, wait for cache sync.", statefulSetKey)
		return "", true
	}

	statefulSetSchedulingInfo, ok := wc.statefulSetWorkloadSchedulingInfo[statefulSetKey]
	if !ok {
		// This should never happen.
		klog.Errorf("Cant find StatefulSet %v scheduling info in cache.", statefulSetKey)
		return podaffinity.PodAffinityUnset, false
	}

	return wc.determineNewPodAffinityPreference(statefulSetInfo.OptimizeSchedulingSetting, statefulSetSchedulingInfo), false
}

// require mutex locked.
func (wc *WebhookCache) determineNewPodAffinityPreference(schedulingSetting *apis.OptimizeSchedulingSetting,
	wsi *apis.WorkloadSchedulingInfo) podaffinity.PodAffinitySettingName {

	if schedulingSetting == nil || wsi == nil {
		// This should never happen.
		klog.Errorf("Workload shcuedlingSetting or workloadSchedulingInfo is nil.")
		return podaffinity.PodAffinityUnset
	}

	if !schedulingSetting.Enable {
		klog.V(3).Infof("Wokrload didnt enable optimize scheduling, return unset.")
		return podaffinity.PodAffinityUnset
	}

	klog.V(3).Infof("Determining new pod affinity, strategy: %s, target-on-demand: %d target-spot: %d, "+
		"available-on-demand: %d, available-spot: %d", schedulingSetting.Strategy, schedulingSetting.TargetOnDemandNum,
		schedulingSetting.TargetOnSpotNum, wsi.OnDemandReplicaCount, wsi.SpotReplicaCount)

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
