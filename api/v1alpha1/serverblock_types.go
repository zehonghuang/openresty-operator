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

	// Listen specifies the address and port that this server block listens on (e.g., "80", "443 ssl")
	Listen string `json:"listen"`

	// AccessLog specifies the path and format of the access log (e.g., "/var/log/nginx/access.log main")
	AccessLog string `json:"accessLog,omitempty"`

	// ErrorLog specifies the path and log level of the error log (e.g., "/var/log/nginx/error.log warn")
	ErrorLog string `json:"errorLog,omitempty"`

	// Headers defines additional headers to include using the `add_header` directive
	Headers []NginxKV `json:"headers,omitempty"`

	// LocationRefs is a list of referenced Location resource names included in this server block
	LocationRefs []string `json:"locationRefs"`

	// Extra contains raw Nginx directives for advanced configuration (e.g., custom error_page rules)
	Extra []string `json:"extra,omitempty"`
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
