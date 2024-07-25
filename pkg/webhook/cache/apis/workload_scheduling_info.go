package apis

import (
	"k8s.io/apimachinery/pkg/types"

	podaffinity "vacant.sh/vmanager/pkg/definitions/pod-affinity"
)

// WorkloadSchedulingInfo stores whether the current pods of the workload have node affinity settings,
// and what those settings are.
type WorkloadSchedulingInfo struct {
	OnDemandReplicaCount int
	SpotReplicaCount     int

	Pods map[types.NamespacedName]podaffinity.PodAffinitySettingName
}

func NewWorkloadSchedulingInfo() *WorkloadSchedulingInfo {
	return &WorkloadSchedulingInfo{
		Pods: make(map[types.NamespacedName]podaffinity.PodAffinitySettingName),
	}
}
