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

	// Entries is a list of individual location configuration entries
	Entries []LocationEntry `json:"entries"`
}

// LocationEntry defines a single Nginx `location` block and its behavior
type LocationEntry struct {
	// Path is the location match path (e.g., "/", "/api", etc.)
	Path string `json:"path"`

	// ProxyPass sets the backend address to proxy traffic to
	ProxyPass string `json:"proxyPass,omitempty"`

	// Headers defines a list of headers to set via proxy_set_header or add_header
	Headers []NginxKV `json:"headers,omitempty"`

	// Timeout configures upstream timeout values (connect/send/read)
	Timeout *Timeouts `json:"timeout,omitempty"`

	// AccessLog enables or disables access logging for this location
	AccessLog *bool `json:"accessLog,omitempty"`

	// LimitReq applies request rate limiting (e.g., "zone=api burst=10 nodelay")
	LimitReq *string `json:"limitReq,omitempty"`

	// Gzip enables gzip compression for specific content types
	Gzip *GzipConf `json:"gzip,omitempty"`

	// Cache defines caching configuration for the location
	Cache *CacheConf `json:"cache,omitempty"`

	// Lua allows embedding custom Lua logic via access/content phases
	Lua *LuaBlock `json:"lua,omitempty"`

	// EnableUpstreamMetrics enables automatic Prometheus metrics collection for upstream requests
	EnableUpstreamMetrics bool `json:"enableUpstreamMetrics,omitempty"`

	// Extra allows defining custom raw Nginx directives
	Extra []string `json:"extra,omitempty"`
}

type NginxKV struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

// Timeouts defines upstream timeout configuration
type Timeouts struct {
	// Connect is the maximum time to establish a connection
	Connect string `json:"connect,omitempty"`

	// Send is the timeout for sending a request to the upstream
	Send string `json:"send,omitempty"`

	// Read is the timeout for reading a response from the upstream
	Read string `json:"read,omitempty"`
}

// GzipConf configures gzip compression
type GzipConf struct {
	// Enable toggles gzip compression
	Enable bool `json:"enable"`

	// Types lists MIME types to compress
	Types []string `json:"types,omitempty"`
}

// CacheConf configures caching for responses
type CacheConf struct {
	// Zone specifies the cache zone name
	Zone string `json:"zone,omitempty"`

	// Valid defines cache duration per status code (e.g., "200 1m")
	Valid string `json:"valid,omitempty"`
}

// LuaBlock defines embedded Lua logic for access/content phases
type LuaBlock struct {
	// Access contains Lua code to execute during access phase
	Access string `json:"access,omitempty"`

	// Content contains Lua code to execute during content phase
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
