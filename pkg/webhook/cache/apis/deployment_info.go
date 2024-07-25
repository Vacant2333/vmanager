package apis

import (
	appsv1 "k8s.io/api/apps/v1"
)

type DeploymentInfo struct {
	Deployment *appsv1.Deployment
	*OptimizeSchedulingSetting
}

func NewDeploymentInfo(deployment *appsv1.Deployment) *DeploymentInfo {
	return &DeploymentInfo{
		Deployment:                deployment,
		OptimizeSchedulingSetting: NewOptimizeSchedulingSettingFromLabels(deployment.Labels, int(*deployment.Spec.Replicas), "Deployment"),
	}
}
