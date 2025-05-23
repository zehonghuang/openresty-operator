/*
Copyright 2025.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package v1alpha1

import (
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// NormalizeRuleSpec defines the desired state of NormalizeRule
type NormalizeRuleSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Request defines rules for normalizing the request payload before proxying.
	// Each entry maps a target field name to either:
	// - a JSONPath string (e.g., "$.input.prompt") to extract from the original request object
	// - or a Lua script block { lua: "..." } that returns the desired value
	//+kubebuilder:pruning:PreserveUnknownFields
	Request map[string]apiextensionsv1.JSON `json:"request,omitempty"`

	// Response defines rules for normalizing the response payload before returning to the client.
	// Each entry maps a target field name to either:
	// - a JSONPath string (e.g., "$.data.content") to extract from the original response object
	// - or a Lua script block { lua: "..." } that returns the transformed value
	//+kubebuilder:pruning:PreserveUnknownFields
	Response map[string]apiextensionsv1.JSON `json:"response,omitempty"`
}

// NormalizeRuleStatus defines the observed state of NormalizeRule
type NormalizeRuleStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	Ready   bool   `json:"ready"`
	Version string `json:"version,omitempty"` // 对应 generation
	Reason  string `json:"reason,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// NormalizeRule is the Schema for the normalizerules API.
//
// This rule is only applied when the corresponding Upstream is of type `FullURL`.
// It defines how request/response payloads should be transformed before proxying.
type NormalizeRule struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   NormalizeRuleSpec   `json:"spec,omitempty"`
	Status NormalizeRuleStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// NormalizeRuleList contains a list of NormalizeRule
type NormalizeRuleList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []NormalizeRule `json:"items"`
}

func init() {
	SchemeBuilder.Register(&NormalizeRule{}, &NormalizeRuleList{})
}
