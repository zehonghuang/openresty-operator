/*
Copyright 2023.

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
	corev1 "k8s.io/api/core/v1"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// UpstreamType defines how upstreams are resolved and rendered in OpenResty
type UpstreamType string

const (
	// UpstreamTypeAddress Address mode uses host:port entries rendered as standard Nginx upstream servers
	UpstreamTypeAddress UpstreamType = "Address"

	// UpstreamTypeFullURL FullURL mode uses complete URLs (e.g., https://foo.com/api), rendered into Lua logic
	UpstreamTypeFullURL UpstreamType = "FullURL"
)

// UpstreamServer defines a backend server with optional normalization logic.
type UpstreamServer struct {
	Address string `json:"address"`

	// NormalizeRequestRef refers to a reusable NormalizeRequest CRD
	NormalizeRequestRef *corev1.LocalObjectReference `json:"normalizeRequestRef,omitempty"`

	// NormalizeRequest defines rules for normalizing the request payload before proxying.
	// Each entry maps a target field name to either:
	// - a JSONPath string (e.g., "$.input.prompt") to extract from the original request object
	// - or a Lua script block { lua: "..." } that returns the desired value
	//+kubebuilder:pruning:PreserveUnknownFields
	NormalizeRequest map[string]apiextensionsv1.JSON `json:"normalizeRequest,omitempty"`

	// NormalizeResponse defines rules for normalizing the response payload before returning to the client.
	// Each entry maps a target field name to either:
	// - a JSONPath string (e.g., "$.data.content") to extract from the original response object
	// - or a Lua script block { lua: "..." } that returns the transformed value
	//+kubebuilder:pruning:PreserveUnknownFields
	NormalizeResponse map[string]apiextensionsv1.JSON `json:"normalizeResponse,omitempty"`
}

// UpstreamSpec defines the desired state of Upstream
type UpstreamSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Servers is a list of backend servers
	//+operator-sdk:csv:customresourcedefinitions:type=spec,displayName="Servers"
	Servers []UpstreamServer `json:"servers"`

	// +kubebuilder:default=Address
	Type UpstreamType `json:"type"`
}

type UpstreamServerStatus struct {
	// Address is the full address of the upstream server (e.g., "example.com:80")
	//+operator-sdk:csv:customresourcedefinitions:type=spec,displayName="Address"
	Address string `json:"address"`

	// Alive indicates whether the server is reachable and responsive
	//+operator-sdk:csv:customresourcedefinitions:type=spec,displayName="Alive"
	Alive bool `json:"alive"`
}

// UpstreamStatus defines the observed state of Upstream
type UpstreamStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	NginxConfig string                 `json:"nginxConfig,omitempty"`
	Servers     []UpstreamServerStatus `json:"servers,omitempty"`

	Ready   bool   `json:"ready"`             // 是否有效
	Version string `json:"version,omitempty"` // 对应 generation
	Reason  string `json:"reason,omitempty"`  // 可选：失败原因
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// Upstream is the Schema for the upstreams API
// +operator-sdk:csv:customresourcedefinitions:displayName="Upstream",resources={{ConfigMap,v1,upstream-cm}}
type Upstream struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   UpstreamSpec   `json:"spec,omitempty"`
	Status UpstreamStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// UpstreamList contains a list of Upstream
type UpstreamList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Upstream `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Upstream{}, &UpstreamList{})
}
