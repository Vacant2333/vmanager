package pod

import (
	"context"
	"encoding/json"
	"net/http"

	admissionv1 "k8s.io/api/admission/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/klog/v2"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"

	nodetype "vacant.sh/vmanager/pkg/definitions/node-type"
	podaffinity "vacant.sh/vmanager/pkg/definitions/pod-affinity"
	"vacant.sh/vmanager/pkg/webhook/cache"
)

type Mutating struct {
	Decoder admission.Decoder
	Cache   cache.Interface
}

// Check if Mutating implements necessary func.
var _ admission.Handler = &Mutating{}

// PodPreferSpotAffinity the default configuration for let pod prefers to spot node.
var podPreferSpotAffinity = corev1.PreferredSchedulingTerm{
	Weight: 10,
	Preference: corev1.NodeSelectorTerm{
		MatchExpressions: []corev1.NodeSelectorRequirement{
			{
				Key:      nodetype.NodeTypeLabelKey,
				Operator: corev1.NodeSelectorOpIn,
				Values:   []string{string(nodetype.NodeTypeSpot)},
			},
		},
	},
}

// PodRequireOnDemandAffinity the default configuration for restrict a pod must on on-demand node.
var podRequireOnDemandAffinity = corev1.NodeSelectorTerm{
	MatchExpressions: []corev1.NodeSelectorRequirement{
		{
			Key:      nodetype.NodeTypeLabelKey,
			Operator: corev1.NodeSelectorOpIn,
			Values:   []string{string(nodetype.NodeTypeOnDemand)},
		},
	},
}

func (m *Mutating) Handle(_ context.Context, req admission.Request) admission.Response {
	if req.Operation != admissionv1.Create {
		// This should never happen, we only care the CREATE operation in validating.
		return admission.Allowed("")
	}

	// We selectively apply affinity and the corresponding affinity labels to the new Pod.
	pod := &corev1.Pod{}
	if err := m.Decoder.Decode(req, pod); err != nil {
		return admission.Errored(http.StatusBadRequest, err)
	}

	targetAffinitySettingName := m.Cache.DetermineNewPodAffinityPreference(pod)

	klog.V(3).Infof("Determine new pod %s/%s affinity setting %s", pod.Namespace, pod.GenerateName, targetAffinitySettingName)

	if targetAffinitySettingName == podaffinity.PodAffinityUnset {
		return admission.Allowed("")
	}

	// Prepare the label and affinity struct, then patch the pod.
	pod.Labels[podaffinity.PodAffinityLabelKey] = string(targetAffinitySettingName)

	if pod.Spec.Affinity == nil {
		pod.Spec.Affinity = &corev1.Affinity{}
	}
	if pod.Spec.Affinity.NodeAffinity == nil {
		pod.Spec.Affinity.NodeAffinity = &corev1.NodeAffinity{}
	}

	nodeAffinity := pod.Spec.Affinity.NodeAffinity

	if targetAffinitySettingName == podaffinity.PodAffinityOnDemand {
		if nodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution == nil {
			nodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution = &corev1.NodeSelector{
				NodeSelectorTerms: []corev1.NodeSelectorTerm{
					podRequireOnDemandAffinity,
				},
			}
		} else {
			require := nodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution
			require.NodeSelectorTerms = append(require.NodeSelectorTerms, podRequireOnDemandAffinity)
		}
	}
	if targetAffinitySettingName == podaffinity.PodAffinitySpot {
		if nodeAffinity.PreferredDuringSchedulingIgnoredDuringExecution == nil {
			nodeAffinity.PreferredDuringSchedulingIgnoredDuringExecution = []corev1.PreferredSchedulingTerm{
				podPreferSpotAffinity,
			}
		} else {
			nodeAffinity.PreferredDuringSchedulingIgnoredDuringExecution = append(
				nodeAffinity.PreferredDuringSchedulingIgnoredDuringExecution, podPreferSpotAffinity)
		}
	}

	marshaledBytes, err := json.Marshal(pod)
	if err != nil {
		return admission.Errored(http.StatusInternalServerError, err)
	}

	return admission.PatchResponseFromRaw(req.Object.Raw, marshaledBytes)
}
