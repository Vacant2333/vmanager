package pod_affinity

import (
	corev1 "k8s.io/api/core/v1"

	node_type "vacant.sh/vmanager/pkg/definitions/node-type"
)

// PodAffinitySettingName indicates that a pod has been designated as either PodAffinityOnDemand or PodAffinitySpot.
// If it is not set, it is considered PodAffinityUnset.
type PodAffinitySettingName string

const (
	PodAffinityLabelKey = "vacant.io/affinity"

	PodAffinityOnDemand PodAffinitySettingName = "on-demand"
	PodAffinitySpot     PodAffinitySettingName = "spot"
	PodAffinityUnset    PodAffinitySettingName = "unset"
)

var PodPreferSpotAffinity = corev1.PreferredSchedulingTerm{
	Weight: 10,
	Preference: corev1.NodeSelectorTerm{
		MatchExpressions: []corev1.NodeSelectorRequirement{
			{
				Key:      node_type.NodeTypeLabelKey,
				Operator: corev1.NodeSelectorOpIn,
				Values:   []string{string(node_type.NodeTypeSpot)},
			},
		},
	},
}

var PodRequireOnDemandAffinity = corev1.NodeSelectorTerm{
	MatchExpressions: []corev1.NodeSelectorRequirement{
		{
			Key:      node_type.NodeTypeLabelKey,
			Operator: corev1.NodeSelectorOpIn,
			Values:   []string{string(node_type.NodeTypeOnDemand)},
		},
	},
}
