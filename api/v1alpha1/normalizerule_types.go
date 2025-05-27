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

// NormalizeRuleSpec defines the desired transformation rules for requests and responses.
type NormalizeRuleSpec struct {
	// Request defines how to normalize the incoming request before proxying.
	// Includes field mapping, static headers, and secrets-based headers.
	//+kubebuilder:pruning:PreserveUnknownFields
	Request *RequestSpec `json:"request,omitempty"`

	// Response defines how to normalize the upstream response before returning to the client.
	// Each entry maps a target field name to either:
	// - a JSONPath string to extract from the response
	// - or a Lua script block { lua: "..." } that returns the transformed value
	//+kubebuilder:pruning:PreserveUnknownFields
	Response map[string]apiextensionsv1.JSON `json:"response,omitempty"`
}

// RequestSpec defines how to construct the outbound request sent to the upstream.
type RequestSpec struct {
	// Body rewrites the request body using field mappings or Lua logic.
	// Each entry maps a target field name to a JSONPath string or a Lua script block.
	Body map[string]apiextensionsv1.JSON `json:"body,omitempty"`

	// Query appends or overrides query parameters in the upstream request URL.
	// Each entry maps a query key to either:
	// - a JSONPath string extracted from the request body
	// - a Lua script block { lua: "..." } returning the query value
	// - or a static string { value: "..." } representing a constant value
	//+kubebuilder:pruning:PreserveUnknownFields
	Query map[string]apiextensionsv1.JSON `json:"query,omitempty"`

	// QueryFromSecret injects query parameters whose values are loaded from Kubernetes Secrets.
	// Each entry defines the query key, target secret name, and key inside the secret.
	QueryFromSecret []ValueFromSecret `json:"queryFromSecret,omitempty"`

	// Headers injects static HTTP headers into the outbound request.
	Headers []NginxKV `json:"headers,omitempty"`

	// HeadersFromSecret injects sensitive HTTP headers whose values are loaded from Kubernetes Secrets.
	HeadersFromSecret []ValueFromSecret `json:"headersFromSecret,omitempty"`
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
