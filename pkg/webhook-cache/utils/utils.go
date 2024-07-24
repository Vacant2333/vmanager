package utils

import (
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"

	pod_affinity "vacant.sh/vmanager/pkg/definitions/pod-affinity"
)

// GetPodSourceWorkloadType retrieves the type of the source workload of a Pod, returning either
// ReplicaSet or StatefulSet. In other cases, it returns "".
func GetPodSourceWorkloadType(pod *corev1.Pod) string {
	if pod == nil {
		return ""
	}

	for _, ownerRef := range pod.OwnerReferences {
		if ownerRef.Kind == "ReplicaSet" || ownerRef.Kind == "StatefulSet" {
			return ownerRef.Kind
		}
	}

	return ""
}

// GetPodSourceResourceKey retrieves the name of the source workload resource through the value of the
// Pod's Label["app"]. It is important to note that for a ReplicaSet,
// there may not necessarily be a corresponding Deployment.
func GetPodSourceResourceKey(pod *corev1.Pod) *types.NamespacedName {
	if pod == nil || pod.Labels == nil || pod.Labels["app"] == "" {
		return nil
	}

	return &types.NamespacedName{
		Namespace: pod.Namespace,
		Name:      pod.Labels["app"],
	}
}

// GetPodAffinitySetting determines if the Pod has been modified by our webhook to have the required affinity
// based on a preset Label. If not, it returns "unset".
func GetPodAffinitySetting(pod *corev1.Pod) pod_affinity.PodAffinitySettingName {
	if pod == nil || pod.Labels == nil || pod.Labels[pod_affinity.PodAffinityLabelKey] == "" {
		return pod_affinity.PodAffinityUnset
	}

	switch pod_affinity.PodAffinitySettingName(pod.Labels[pod_affinity.PodAffinityLabelKey]) {
	case pod_affinity.PodAffinityOnDemand:
		return pod_affinity.PodAffinityOnDemand
	case pod_affinity.PodAffinitySpot:
		return pod_affinity.PodAffinitySpot
	default:
		return pod_affinity.PodAffinityUnset
	}
}
