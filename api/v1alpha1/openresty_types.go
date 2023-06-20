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

// OpenRestySpec defines the desired state of OpenResty
type OpenRestySpec struct {
	// Replicas defines how many OpenResty pods to run
	Replicas *int32 `json:"replicas,omitempty"`

	// Image specifies the Docker image for OpenResty
	Image string `json:"image,omitempty"`

	// Http contains configuration for the HTTP block of the OpenResty instance
	Http *HttpBlock `json:"http"`

	// MetricsServer defines an optional Prometheus metrics endpoint
	MetricsServer *MetricsServer `json:"metrics,omitempty"`
}

type HttpBlock struct {
	// Include is a list of additional Nginx include files (e.g., mime.types)
	Include []string `json:"include,omitempty"`

	// LogFormat specifies the log_format directive in Nginx
	LogFormat string `json:"logFormat,omitempty"`

	// AccessLog specifies the path for access logs
	AccessLog string `json:"accessLog,omitempty"`

	// ErrorLog specifies the path for error logs
	ErrorLog string `json:"errorLog,omitempty"`

	// ClientMaxBodySize sets the client_max_body_size directive
	ClientMaxBodySize string `json:"clientMaxBodySize,omitempty"`

	// Gzip enables gzip compression in the HTTP block
	Gzip bool `json:"gzip,omitempty"`

	// Extra allows appending custom HTTP directives
	Extra []string `json:"extra,omitempty"`

	// ServerRefs lists referenced ServerBlock CR names
	ServerRefs []string `json:"serverRefs"`

	// UpstreamRefs lists referenced Upstream CR names
	UpstreamRefs []string `json:"upstreamRefs,omitempty"`
}

// MetricsServer defines an optional server to expose Prometheus metrics
type MetricsServer struct {
	// Enable controls whether the /metrics endpoint is exposed
	Enable bool `json:"enable,omitempty"`

	// Listen specifies the port to expose Prometheus metrics on (default: "8080")
	Listen string `json:"listen,omitempty"`

	// Path defines the HTTP path for Prometheus metrics (default: "/metrics")
	Path string `json:"path,omitempty"`
}

// OpenRestyStatus defines the observed state of OpenResty
type OpenRestyStatus struct {
	AvailableReplicas int32  `json:"availableReplicas,omitempty"`
	Ready             bool   `json:"ready"`
	Reason            string `json:"reason,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

type OpenResty struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   OpenRestySpec   `json:"spec,omitempty"`
	Status OpenRestyStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

type OpenRestyList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []OpenResty `json:"items"`
}

func init() {
	SchemeBuilder.Register(&OpenResty{}, &OpenRestyList{})
}
