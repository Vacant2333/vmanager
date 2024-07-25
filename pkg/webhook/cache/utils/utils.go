package utils

import (
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"

	podaffinity "vacant.sh/vmanager/pkg/definitions/pod-affinity"
)

// GetPodSourceWorkloadTypeAndKey retrieves the type and name of the source workload of a Pod, returning either
// ReplicaSet or StatefulSet. In other cases, it returns "".
func GetPodSourceWorkloadTypeAndKey(pod *corev1.Pod) (string, *types.NamespacedName) {
	if pod == nil {
		return "", nil
	}

	for _, ownerRef := range pod.OwnerReferences {
		if ownerRef.Kind == "ReplicaSet" || ownerRef.Kind == "StatefulSet" {
			return ownerRef.Kind, &types.NamespacedName{
				Namespace: pod.Namespace,
				Name:      ownerRef.Name,
			}
		}
	}
	return "", nil
}

func GetReplicaSetSourceDeploymentKey(replicaSet *appsv1.ReplicaSet) *types.NamespacedName {
	if replicaSet == nil {
		return nil
	}

	for _, ownerRef := range replicaSet.OwnerReferences {
		if ownerRef.Kind == "Deployment" {
			return &types.NamespacedName{
				Namespace: replicaSet.Namespace,
				Name:      ownerRef.Name,
			}
		}
	}
	return nil
}

// GetPodAffinitySetting determines if the Pod has been modified by our webhook to have the required affinity
// based on a preset Label. If not, it returns "unset".
func GetPodAffinitySetting(pod *corev1.Pod) podaffinity.PodAffinitySettingName {
	if pod == nil || pod.Labels == nil || pod.Labels[podaffinity.PodAffinityLabelKey] == "" {
		return podaffinity.PodAffinityUnset
	}

	switch podaffinity.PodAffinitySettingName(pod.Labels[podaffinity.PodAffinityLabelKey]) {
	case podaffinity.PodAffinityOnDemand:
		return podaffinity.PodAffinityOnDemand
	case podaffinity.PodAffinitySpot:
		return podaffinity.PodAffinitySpot
	default:
		return podaffinity.PodAffinityUnset
	}
}
