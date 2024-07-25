package pod

import (
	"context"
	"fmt"
	"net/http"

	admissionv1 "k8s.io/api/admission/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/klog/v2"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"

	podaffinity "vacant.sh/vmanager/pkg/definitions/pod-affinity"
)

type Validating struct {
	Decoder admission.Decoder
}

// Check if Validating implements necessary func.
var _ admission.Handler = &Validating{}

func (v *Validating) Handle(_ context.Context, req admission.Request) admission.Response {
	if req.Operation != admissionv1.Update {
		// This should never happen, we only care the UPDATE operation in validating.
		return admission.Allowed("")
	}

	var oldPod, newPod *corev1.Pod
	if err := v.Decoder.DecodeRaw(req.OldObject, oldPod); err != nil || oldPod == nil {
		return admission.Errored(http.StatusBadRequest, err)
	}
	if err := v.Decoder.DecodeRaw(req.Object, newPod); err != nil || newPod == nil {
		return admission.Errored(http.StatusBadRequest, err)
	}

	klog.V(3).Infof("Validating pod %s/%s %s", oldPod.Namespace, oldPod.Name, req.Operation)

	// We only need to ensure that users cannot update our label PodAffinityLabelKey.
	if oldPod.Labels[podaffinity.PodAffinityLabelKey] != newPod.Labels[podaffinity.PodAffinityLabelKey] {
		return admission.Denied(fmt.Sprintf("The label %s is not allowed to update its value", podaffinity.PodAffinityLabelKey))
	}
	return admission.Allowed("")
}
