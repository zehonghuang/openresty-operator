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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// RateLimitPolicySpec defines the desired state of RateLimitPolicy
type RateLimitPolicySpec struct {
	// ZoneName is the name of the rate limiting zone defined via `limit_req_zone`
	//+operator-sdk:csv:customresourcedefinitions:type=spec,displayName="ZoneName",xDescriptors="urn:alm:descriptor:com.tectonic.ui:text"
	ZoneName string `json:"zoneName"`

	// Rate defines the rate limit, such as "10r/s" for 10 requests per second
	//+operator-sdk:csv:customresourcedefinitions:type=spec,displayName="Rate",xDescriptors="urn:alm:descriptor:com.tectonic.ui:text"
	Rate string `json:"rate"`

	// Key specifies the key to identify a client for rate limiting (default: "$binary_remote_addr")
	//+operator-sdk:csv:customresourcedefinitions:type=spec,displayName="Key",xDescriptors="urn:alm:descriptor:com.tectonic.ui:text"
	Key string `json:"key,omitempty"`

	// ZoneSize is the size of the shared memory zone (default: "10m")
	//+operator-sdk:csv:customresourcedefinitions:type=spec,displayName="ZoneSize",xDescriptors="urn:alm:descriptor:com.tectonic.ui:text"
	ZoneSize string `json:"zoneSize,omitempty"`

	// Burst specifies the maximum burst of requests allowed beyond the rate
	//+operator-sdk:csv:customresourcedefinitions:type=spec,displayName="Burst",xDescriptors="urn:alm:descriptor:com.tectonic.ui:text"
	Burst int `json:"burst,omitempty"`

	// NoDelay controls whether to allow burst requests to be served immediately without delay
	//+operator-sdk:csv:customresourcedefinitions:type=spec,displayName="ZoneName",xDescriptors="urn:alm:descriptor:com.tectonic.ui:booleanSwitch"
	NoDelay bool `json:"nodelay,omitempty"`
}

type RateLimitPolicyStatus struct {
	Ready   bool   `json:"ready,omitempty"`
	Version string `json:"version,omitempty"`
	Reason  string `json:"reason,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// RateLimitPolicy is the Schema for the ratelimitpolicies API
// +operator-sdk:csv:customresourcedefinitions:displayName="RateLimitPolicy",resources={{ConfigMap,v1,ratelimit-cm}}
type RateLimitPolicy struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   RateLimitPolicySpec   `json:"spec,omitempty"`
	Status RateLimitPolicyStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// RateLimitPolicyList contains a list of RateLimitPolicy
type RateLimitPolicyList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []RateLimitPolicy `json:"items"`
}

func init() {
	SchemeBuilder.Register(&RateLimitPolicy{}, &RateLimitPolicyList{})
}
