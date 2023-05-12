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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// UpstreamSpec defines the desired state of Upstream
type UpstreamSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Foo is an example field of Upstream. Edit upstream_types.go to remove/update
	Servers []string `json:"servers"`
}

type UpstreamServerStatus struct {
	Address string `json:"address"` // 如 "example.com:80"
	Alive   bool   `json:"alive"`
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
