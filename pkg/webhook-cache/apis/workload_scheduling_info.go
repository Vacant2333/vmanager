package apis

import (
	"k8s.io/apimachinery/pkg/types"

	pod_affinity "vacant.sh/vmanager/pkg/definitions/pod-affinity"
)

// WorkloadSchedulingInfo stores whether the current pods of the workload have node affinity settings,
// and what those settings are.
type WorkloadSchedulingInfo struct {
	OnDemandReplicaCount int32
	SpotReplicaCount     int32

	Pods map[types.NamespacedName]pod_affinity.PodAffinitySettingName
}

func NewWorkloadSchedulingInfo() *WorkloadSchedulingInfo {
	return &WorkloadSchedulingInfo{
		Pods: make(map[types.NamespacedName]pod_affinity.PodAffinitySettingName),
	}
}
