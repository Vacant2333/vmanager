package apis

import appsv1 "k8s.io/api/apps/v1"

type StatefulSetInfo struct {
	StatefulSet *appsv1.StatefulSet
	*OptimizeSchedulingInfo
}

func NewStatefulSetInfo(statefulSet *appsv1.StatefulSet) *StatefulSetInfo {
	return &StatefulSetInfo{
		StatefulSet:            statefulSet,
		OptimizeSchedulingInfo: NewOptimizeSchedulingInfoFromLabels(statefulSet.Labels, int(*statefulSet.Spec.Replicas)),
	}
}
