package apis

import appsv1 "k8s.io/api/apps/v1"

type StatefulSetInfo struct {
	StatefulSet *appsv1.StatefulSet
	*OptimizeSchedulingSetting
}

func NewStatefulSetInfo(statefulSet *appsv1.StatefulSet) *StatefulSetInfo {
	return &StatefulSetInfo{
		StatefulSet:               statefulSet,
		OptimizeSchedulingSetting: NewOptimizeSchedulingSettingFromLabels(statefulSet.Labels, int(*statefulSet.Spec.Replicas), "StatefulSet"),
	}
}
