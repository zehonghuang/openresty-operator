package validating

import (
	"context"
	"fmt"
	"net/http"
	webv1alpha1 "openresty-operator/api/v1alpha1"
	"openresty-operator/internal/utils"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"

	"sigs.k8s.io/controller-runtime/pkg/client"
)

type LocationValidator struct {
	Client  client.Client
	Decoder admission.Decoder
}

func (v *LocationValidator) Handle(ctx context.Context, req admission.Request) admission.Response {
	var loc webv1alpha1.Location
	if err := v.Decoder.Decode(req, &loc); err != nil {
		return admission.Errored(http.StatusBadRequest, err)
	}

	pathSet := make(map[string]struct{})
	var invalidPaths []string
	var duplicatePaths []string

	for _, entry := range loc.Spec.Entries {
		valid, reason := utils.ValidateLocationPath(entry.Path)
		if !valid {
			invalidPaths = append(invalidPaths, fmt.Sprintf("%s (%s)", entry.Path, reason))
		}
		if _, exists := pathSet[entry.Path]; exists {
			duplicatePaths = append(duplicatePaths, entry.Path)
		}
		pathSet[entry.Path] = struct{}{}
	}

	if len(invalidPaths)+len(duplicatePaths) > 0 {
		msg := fmt.Sprintf("Invalid paths: %v; Duplicates: %v", invalidPaths, duplicatePaths)
		return admission.Denied(msg)
	}

	return admission.Allowed("Location is valid")
}

func (v *LocationValidator) InjectDecoder(d admission.Decoder) error {
	v.Decoder = d
	return nil
}
