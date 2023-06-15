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

// LocationSpec defines the desired state of Location
type LocationSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Foo is an example field of Location. Edit location_types.go to remove/update
	Entries []LocationEntry `json:"entries"`
}

type LocationEntry struct {
	Path string `json:"path"`

	ProxyPass string `json:"proxyPass,omitempty"`

	Headers               []NginxKV  `json:"headers,omitempty"` // proxy_set_header / add_header
	Timeout               *Timeouts  `json:"timeout,omitempty"`
	AccessLog             *bool      `json:"accessLog,omitempty"` // true/false
	LimitReq              *string    `json:"limitReq,omitempty"`  // zone=api burst=10 nodelay
	Gzip                  *GzipConf  `json:"gzip,omitempty"`
	Cache                 *CacheConf `json:"cache,omitempty"`
	Lua                   *LuaBlock  `json:"lua,omitempty"`
	EnableUpstreamMetrics bool       `json:"enableUpstreamMetrics,omitempty"`

	Extra []string `json:"extra,omitempty"` // 自定义指令
}

type NginxKV struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type Timeouts struct {
	Connect string `json:"connect,omitempty"` // 5s
	Send    string `json:"send,omitempty"`    // 10s
	Read    string `json:"read,omitempty"`    // 10s
}

type GzipConf struct {
	Enable bool     `json:"enable"`
	Types  []string `json:"types,omitempty"`
}

type CacheConf struct {
	Zone  string `json:"zone,omitempty"`
	Valid string `json:"valid,omitempty"` // 如 "200 1m"
}

type LuaBlock struct {
	Access  string `json:"access,omitempty"`
	Content string `json:"content,omitempty"`
}

// LocationStatus defines the observed state of Location
type LocationStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	Ready   bool   `json:"ready"`
	Version string `json:"version,omitempty"` // 对应 generation
	Reason  string `json:"reason,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:webhookserver:path=/validate-location,mutating=false,failurePolicy=fail,groups=openresty.huangzehong.me,resources=locations,verbs=create;update,versions=v1alpha1,name=validation.location.webhookserver.chillyroom.com,sideEffects=None,admissionReviewVersions=v1

// Location is the Schema for the locations API
type Location struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   LocationSpec   `json:"spec,omitempty"`
	Status LocationStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// LocationList contains a list of Location
type LocationList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Location `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Location{}, &LocationList{})
}
