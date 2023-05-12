package webhookserver

import (
	"openresty-operator/internal/webhookserver/validating"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

func SetupWebhook(mgr ctrl.Manager) {
	decoder := admission.NewDecoder(mgr.GetScheme())

	hook := &validating.LocationValidator{}
	_ = hook.InjectDecoder(decoder)

	mgr.GetWebhookServer().Register("/validate-location", &admission.Webhook{
		Handler: hook,
	})
}
