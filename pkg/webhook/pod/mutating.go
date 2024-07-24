package pod

import (
	"context"

	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"

	"vacant.sh/vmanager/pkg/webhook_cache"
)

type Mutating struct {
	Cache webhook_cache.Interface
}

// Check if Mutating implements necessary func.
var _ admission.Handler = &Mutating{}

func (m *Mutating) Handle(ctx context.Context, req admission.Request) admission.Response {
	return nil
}
