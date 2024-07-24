package deployment

import (
	"context"
	"net/http"

	admissionv1 "k8s.io/api/admission/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/klog/v2"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"

	optimize_scheduling "vacant.sh/vmanager/pkg/definitions/optimize-scheduling"
)

type Validating struct {
	Decoder admission.Decoder
}

// Check if Validating implements necessary func.
var _ admission.Handler = &Validating{}

func (v *Validating) Handle(_ context.Context, req admission.Request) admission.Response {
	if req.Operation != admissionv1.Update && req.Operation != admissionv1.Create {
		// This should never happen, we only care the CREATE and UPDATE operation.
		return admission.Allowed("")
	}

	// Parse the uncertain type resource object, we don't need care the oldObject here.
	obj := &unstructured.Unstructured{}
	if err := v.Decoder.DecodeRaw(req.Object, obj); err != nil {
		return admission.Errored(http.StatusBadRequest, err)
	}

	klog.V(3).Infof("Validating Deployment %s/%s", obj.GetNamespace(), obj.GetName())

	errList := optimize_scheduling.ValidateOptimizeSchedulingConfiguration(obj.GetLabels())
	if len(errList) > 0 {
		return admission.Denied(errList.ToAggregate().Error())
	}

	return admission.Allowed("")
}
