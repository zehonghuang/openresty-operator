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

// ServerBlockSpec defines the desired state of ServerBlock
type ServerBlockSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Foo is an example field of ServerBlock. Edit serverblock_types.go to remove/update
	Listen       string    `json:"listen"`              // 如 "80"、"443 ssl"
	AccessLog    string    `json:"accessLog,omitempty"` // 如 /var/log/nginx/xx.log main
	ErrorLog     string    `json:"errorLog,omitempty"`  // 如 /var/log/nginx/xx.log warn
	Headers      []NginxKV `json:"headers,omitempty"`   // add_header
	LocationRefs []string  `json:"locationRefs"`        // 引用多个 Location CRD（生成 include）
	Extra        []string  `json:"extra,omitempty"`     // 兜底扩展字段（如 error_page）
}

// ServerBlockStatus defines the observed state of ServerBlock
type ServerBlockStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	Ready  bool   `json:"ready"`
	Reason string `json:"reason,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// ServerBlock is the Schema for the serverblocks API
type ServerBlock struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ServerBlockSpec   `json:"spec,omitempty"`
	Status ServerBlockStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// ServerBlockList contains a list of ServerBlock
type ServerBlockList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ServerBlock `json:"items"`
}

func init() {
	SchemeBuilder.Register(&ServerBlock{}, &ServerBlockList{})
}
