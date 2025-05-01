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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// OpenRestySpec defines the desired state of OpenResty
type OpenRestySpec struct {
	// Replicas defines how many OpenResty pods to run
	//+operator-sdk:csv:customresourcedefinitions:type=spec,displayName="Replicas",xDescriptors="urn:alm:descriptor:com.tectonic.ui:podCount"
	Replicas *int32 `json:"replicas,omitempty"`

	// Image specifies the Docker image for OpenResty
	//+operator-sdk:csv:customresourcedefinitions:type=spec,displayName="Image",xDescriptors="urn:alm:descriptor:com.tectonic.ui:text"
	Image string `json:"image,omitempty"`

	Resources corev1.ResourceRequirements `json:"resources,omitempty"`

	// Http contains configuration for the HTTP block of the OpenResty instance
	//+operator-sdk:csv:customresourcedefinitions:type=spec,displayName="Http",xDescriptors="urn:alm:descriptor:com.tectonic.ui:object"
	Http *HttpBlock `json:"http"`

	// MetricsServer defines an optional Prometheus metrics endpoint
	//+operator-sdk:csv:customresourcedefinitions:type=spec,displayName="Metrics Server",xDescriptors="urn:alm:descriptor:com.tectonic.ui:object"
	MetricsServer *MetricsServer `json:"metrics,omitempty"`

	// +kubebuilder:default:={enable:true}
	// ServiceMonitor controls whether to automatically create a Prometheus ServiceMonitor for OpenResty metrics
	//+operator-sdk:csv:customresourcedefinitions:type=spec,displayName="Enable ServiceMonitor",xDescriptors="urn:alm:descriptor:com.tectonic.ui:booleanSwitch"
	ServiceMonitor *ServiceMonitor `json:"serviceMonitor,omitempty"`

	ReloadAgentEnv []corev1.EnvVar `json:"reloadAgentEnv,omitempty"`

	// NodeSelector defines node labels for pod assignment
	NodeSelector map[string]string `json:"nodeSelector,omitempty"`

	// Tolerations are applied to allow scheduling onto nodes with matching taints
	Tolerations []corev1.Toleration `json:"tolerations,omitempty"`

	// Affinity defines pod scheduling preferences
	Affinity *corev1.Affinity `json:"affinity,omitempty"`

	// PriorityClassName defines the pod priority class
	PriorityClassName string `json:"priorityClassName,omitempty"`

	// TerminationGracePeriodSeconds defines the duration in seconds the pod needs to terminate gracefully
	TerminationGracePeriodSeconds *int64 `json:"terminationGracePeriodSeconds,omitempty"`

	// +kubebuilder:default:={type:EmptyDir}
	LogVolume LogVolumeSpec `json:"logVolume,omitempty"`
}

type ServiceMonitor struct {

	// +kubebuilder:default=false
	Enable bool `json:"enable,omitempty"`

	Labels map[string]string `json:"labels,omitempty"`

	Annotations map[string]string `json:"annotations,omitempty"`
}

type LogVolumeSpec struct {
	// Type of volume mounted at /var/log/nginx. EmptyDir uses ephemeral storage (logs lost after pod deletion); PVC uses a PersistentVolumeClaim for persistent storage.
	Type LogVolumeType `json:"type,omitempty"`
	// Name of the PersistentVolumeClaim to use when type is PVC. Only required if type: PVC.
	PersistentVolumeClaim string `json:"persistentVolumeClaim,omitempty"`
}

type LogVolumeType string

const (
	LogVolumeTypeEmptyDir LogVolumeType = "EmptyDir"
	LogVolumeTypePVC      LogVolumeType = "PVC"
)

type HttpBlock struct {
	// Include is a list of additional Nginx include files (e.g., mime.types)
	//+operator-sdk:csv:customresourcedefinitions:type=spec,displayName="Include",xDescriptors="urn:alm:descriptor:com.tectonic.ui:array"
	Include []string `json:"include,omitempty"`

	// LogFormat specifies the log_format directive in Nginx
	//+operator-sdk:csv:customresourcedefinitions:type=spec,displayName="Log Format",xDescriptors="urn:alm:descriptor:com.tectonic.ui:text"
	LogFormat string `json:"logFormat,omitempty"`

	// AccessLog specifies the path for access logs
	//+operator-sdk:csv:customresourcedefinitions:type=spec,displayName="Access Log",xDescriptors="urn:alm:descriptor:com.tectonic.ui:text"
	AccessLog string `json:"accessLog,omitempty"`

	// ErrorLog specifies the path for error logs
	//+operator-sdk:csv:customresourcedefinitions:type=spec,displayName="Access Log",xDescriptors="urn:alm:descriptor:com.tectonic.ui:text"
	ErrorLog string `json:"errorLog,omitempty"`

	// ClientMaxBodySize sets the client_max_body_size directive
	//+operator-sdk:csv:customresourcedefinitions:type=spec,displayName="Client Max Body Size",xDescriptors="urn:alm:descriptor:com.tectonic.ui:text"
	ClientMaxBodySize string `json:"clientMaxBodySize,omitempty"`

	// Gzip enables gzip compression in the HTTP block
	//+operator-sdk:csv:customresourcedefinitions:type=spec,displayName="Gzip",xDescriptors="urn:alm:descriptor:com.tectonic.ui:text"
	Gzip bool `json:"gzip,omitempty"`

	// Extra allows appending custom HTTP directives
	//+operator-sdk:csv:customresourcedefinitions:type=spec,displayName="Gzip",xDescriptors="urn:alm:descriptor:com.tectonic.ui:array"
	Extra []string `json:"extra,omitempty"`

	// ServerRefs lists referenced ServerBlock CR names
	//+operator-sdk:csv:customresourcedefinitions:type=spec,displayName="ServerRefs",xDescriptors="urn:alm:descriptor:com.tectonic.ui:array"
	ServerRefs []string `json:"serverRefs"`

	// UpstreamRefs lists referenced Upstream CR names
	//+operator-sdk:csv:customresourcedefinitions:type=spec,displayName="UpstreamRefs",xDescriptors="urn:alm:descriptor:com.tectonic.ui:array"
	UpstreamRefs []string `json:"upstreamRefs,omitempty"`
}

// MetricsServer defines an optional server to expose Prometheus metrics
type MetricsServer struct {
	// Enable controls whether the /metrics endpoint is exposed
	//+operator-sdk:csv:customresourcedefinitions:type=spec,displayName="Enable",xDescriptors="urn:alm:descriptor:com.tectonic.ui:booleanSwitch"
	Enable bool `json:"enable,omitempty"`

	// +kubebuilder:default="9090"
	// Listen specifies the port to expose Prometheus metrics on (default: "8080")
	//+operator-sdk:csv:customresourcedefinitions:type=spec,displayName="Listen",xDescriptors="urn:alm:descriptor:com.tectonic.ui:text"
	Listen string `json:"listen,omitempty"`

	// +kubebuilder:default=/metrics
	// Path defines the HTTP path for Prometheus metrics (default: "/metrics")
	//+operator-sdk:csv:customresourcedefinitions:type=spec,displayName="Path",xDescriptors="urn:alm:descriptor:com.tectonic.ui:text"
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

// +operator-sdk:csv:customresourcedefinitions:displayName="OpenResty",resources={{ConfigMap,v1,openresty-cm}, {Pod,v1,openresty-app},{Deployment,v1,openresty-deployment}}
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
