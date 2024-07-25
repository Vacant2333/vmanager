package pod_affinity

// PodAffinitySettingName indicates that a pod has been designated as either PodAffinityOnDemand or PodAffinitySpot.
// If it is not set, it is considered PodAffinityUnset.
type PodAffinitySettingName string

const (
	PodAffinityLabelKey = "vacant.sh/affinity"

	PodAffinityOnDemand PodAffinitySettingName = "on-demand"
	PodAffinitySpot     PodAffinitySettingName = "spot"
	PodAffinityUnset    PodAffinitySettingName = "unset"
)
