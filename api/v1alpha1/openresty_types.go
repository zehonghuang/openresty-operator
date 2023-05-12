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
	// 副本数量
	Replicas *int32 `json:"replicas,omitempty"`
	// 镜像地址
	Image string `json:"image,omitempty"`
	// Service 的端口
	Http *HttpBlock `json:"http"` // http 配置 + serverRefs
}

// http 层配置
type HttpBlock struct {
	Include           []string `json:"include,omitempty"`           // include 文件，如 mime.types
	LogFormat         string   `json:"logFormat,omitempty"`         // log_format 指令
	AccessLog         string   `json:"accessLog,omitempty"`         // access_log 路径
	ErrorLog          string   `json:"errorLog,omitempty"`          // error_log 路径
	ClientMaxBodySize string   `json:"clientMaxBodySize,omitempty"` // client_max_body_size
	Gzip              bool     `json:"gzip,omitempty"`              // gzip on/off
	Extra             []string `json:"extra,omitempty"`             // 其他 http 层自定义指令
	ServerRefs        []string `json:"serverRefs"`                  // 引用 ServerBlock 名称
}

// OpenRestyStatus defines the observed state of OpenResty
type OpenRestyStatus struct {
	AvailableReplicas int32 `json:"availableReplicas,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// OpenResty is the Schema for the openresties API
type OpenResty struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   OpenRestySpec   `json:"spec,omitempty"`
	Status OpenRestyStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// OpenRestyList contains a list of OpenResty
type OpenRestyList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []OpenResty `json:"items"`
}

func init() {
	SchemeBuilder.Register(&OpenResty{}, &OpenRestyList{})
}
