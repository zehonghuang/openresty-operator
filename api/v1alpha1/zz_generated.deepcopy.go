//go:build !ignore_autogenerated

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

// Code generated by controller-gen. DO NOT EDIT.

package v1alpha1

import (
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	runtime "k8s.io/apimachinery/pkg/runtime"
)

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *CacheConf) DeepCopyInto(out *CacheConf) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new CacheConf.
func (in *CacheConf) DeepCopy() *CacheConf {
	if in == nil {
		return nil
	}
	out := new(CacheConf)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *GzipConf) DeepCopyInto(out *GzipConf) {
	*out = *in
	if in.Types != nil {
		in, out := &in.Types, &out.Types
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new GzipConf.
func (in *GzipConf) DeepCopy() *GzipConf {
	if in == nil {
		return nil
	}
	out := new(GzipConf)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *HttpBlock) DeepCopyInto(out *HttpBlock) {
	*out = *in
	if in.Include != nil {
		in, out := &in.Include, &out.Include
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
	if in.Extra != nil {
		in, out := &in.Extra, &out.Extra
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
	if in.ServerRefs != nil {
		in, out := &in.ServerRefs, &out.ServerRefs
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
	if in.UpstreamRefs != nil {
		in, out := &in.UpstreamRefs, &out.UpstreamRefs
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new HttpBlock.
func (in *HttpBlock) DeepCopy() *HttpBlock {
	if in == nil {
		return nil
	}
	out := new(HttpBlock)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Location) DeepCopyInto(out *Location) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	out.Status = in.Status
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Location.
func (in *Location) DeepCopy() *Location {
	if in == nil {
		return nil
	}
	out := new(Location)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *Location) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *LocationEntry) DeepCopyInto(out *LocationEntry) {
	*out = *in
	if in.Headers != nil {
		in, out := &in.Headers, &out.Headers
		*out = make([]NginxKV, len(*in))
		copy(*out, *in)
	}
	if in.HeadersFromSecret != nil {
		in, out := &in.HeadersFromSecret, &out.HeadersFromSecret
		*out = make([]ValueFromSecret, len(*in))
		copy(*out, *in)
	}
	if in.Timeout != nil {
		in, out := &in.Timeout, &out.Timeout
		*out = new(Timeouts)
		**out = **in
	}
	if in.AccessLog != nil {
		in, out := &in.AccessLog, &out.AccessLog
		*out = new(bool)
		**out = **in
	}
	if in.LimitReq != nil {
		in, out := &in.LimitReq, &out.LimitReq
		*out = new(string)
		**out = **in
	}
	if in.Gzip != nil {
		in, out := &in.Gzip, &out.Gzip
		*out = new(GzipConf)
		(*in).DeepCopyInto(*out)
	}
	if in.Cache != nil {
		in, out := &in.Cache, &out.Cache
		*out = new(CacheConf)
		**out = **in
	}
	if in.Lua != nil {
		in, out := &in.Lua, &out.Lua
		*out = new(LuaBlock)
		**out = **in
	}
	if in.Extra != nil {
		in, out := &in.Extra, &out.Extra
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new LocationEntry.
func (in *LocationEntry) DeepCopy() *LocationEntry {
	if in == nil {
		return nil
	}
	out := new(LocationEntry)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *LocationList) DeepCopyInto(out *LocationList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]Location, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new LocationList.
func (in *LocationList) DeepCopy() *LocationList {
	if in == nil {
		return nil
	}
	out := new(LocationList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *LocationList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *LocationSpec) DeepCopyInto(out *LocationSpec) {
	*out = *in
	if in.Entries != nil {
		in, out := &in.Entries, &out.Entries
		*out = make([]LocationEntry, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new LocationSpec.
func (in *LocationSpec) DeepCopy() *LocationSpec {
	if in == nil {
		return nil
	}
	out := new(LocationSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *LocationStatus) DeepCopyInto(out *LocationStatus) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new LocationStatus.
func (in *LocationStatus) DeepCopy() *LocationStatus {
	if in == nil {
		return nil
	}
	out := new(LocationStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *LogVolumeSpec) DeepCopyInto(out *LogVolumeSpec) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new LogVolumeSpec.
func (in *LogVolumeSpec) DeepCopy() *LogVolumeSpec {
	if in == nil {
		return nil
	}
	out := new(LogVolumeSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *LuaBlock) DeepCopyInto(out *LuaBlock) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new LuaBlock.
func (in *LuaBlock) DeepCopy() *LuaBlock {
	if in == nil {
		return nil
	}
	out := new(LuaBlock)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *MetricsServer) DeepCopyInto(out *MetricsServer) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new MetricsServer.
func (in *MetricsServer) DeepCopy() *MetricsServer {
	if in == nil {
		return nil
	}
	out := new(MetricsServer)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *NginxKV) DeepCopyInto(out *NginxKV) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new NginxKV.
func (in *NginxKV) DeepCopy() *NginxKV {
	if in == nil {
		return nil
	}
	out := new(NginxKV)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *NormalizeRule) DeepCopyInto(out *NormalizeRule) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	out.Status = in.Status
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new NormalizeRule.
func (in *NormalizeRule) DeepCopy() *NormalizeRule {
	if in == nil {
		return nil
	}
	out := new(NormalizeRule)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *NormalizeRule) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *NormalizeRuleList) DeepCopyInto(out *NormalizeRuleList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]NormalizeRule, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new NormalizeRuleList.
func (in *NormalizeRuleList) DeepCopy() *NormalizeRuleList {
	if in == nil {
		return nil
	}
	out := new(NormalizeRuleList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *NormalizeRuleList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *NormalizeRuleSpec) DeepCopyInto(out *NormalizeRuleSpec) {
	*out = *in
	if in.Request != nil {
		in, out := &in.Request, &out.Request
		*out = new(RequestSpec)
		(*in).DeepCopyInto(*out)
	}
	if in.Response != nil {
		in, out := &in.Response, &out.Response
		*out = make(map[string]v1.JSON, len(*in))
		for key, val := range *in {
			(*out)[key] = *val.DeepCopy()
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new NormalizeRuleSpec.
func (in *NormalizeRuleSpec) DeepCopy() *NormalizeRuleSpec {
	if in == nil {
		return nil
	}
	out := new(NormalizeRuleSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *NormalizeRuleStatus) DeepCopyInto(out *NormalizeRuleStatus) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new NormalizeRuleStatus.
func (in *NormalizeRuleStatus) DeepCopy() *NormalizeRuleStatus {
	if in == nil {
		return nil
	}
	out := new(NormalizeRuleStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *OpenResty) DeepCopyInto(out *OpenResty) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	out.Status = in.Status
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new OpenResty.
func (in *OpenResty) DeepCopy() *OpenResty {
	if in == nil {
		return nil
	}
	out := new(OpenResty)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *OpenResty) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *OpenRestyList) DeepCopyInto(out *OpenRestyList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]OpenResty, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new OpenRestyList.
func (in *OpenRestyList) DeepCopy() *OpenRestyList {
	if in == nil {
		return nil
	}
	out := new(OpenRestyList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *OpenRestyList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *OpenRestySpec) DeepCopyInto(out *OpenRestySpec) {
	*out = *in
	if in.Replicas != nil {
		in, out := &in.Replicas, &out.Replicas
		*out = new(int32)
		**out = **in
	}
	in.Resources.DeepCopyInto(&out.Resources)
	if in.Http != nil {
		in, out := &in.Http, &out.Http
		*out = new(HttpBlock)
		(*in).DeepCopyInto(*out)
	}
	if in.MetricsServer != nil {
		in, out := &in.MetricsServer, &out.MetricsServer
		*out = new(MetricsServer)
		**out = **in
	}
	if in.ServiceMonitor != nil {
		in, out := &in.ServiceMonitor, &out.ServiceMonitor
		*out = new(ServiceMonitor)
		(*in).DeepCopyInto(*out)
	}
	if in.ReloadAgentEnv != nil {
		in, out := &in.ReloadAgentEnv, &out.ReloadAgentEnv
		*out = make([]corev1.EnvVar, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	if in.NodeSelector != nil {
		in, out := &in.NodeSelector, &out.NodeSelector
		*out = make(map[string]string, len(*in))
		for key, val := range *in {
			(*out)[key] = val
		}
	}
	if in.Tolerations != nil {
		in, out := &in.Tolerations, &out.Tolerations
		*out = make([]corev1.Toleration, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	if in.Affinity != nil {
		in, out := &in.Affinity, &out.Affinity
		*out = new(corev1.Affinity)
		(*in).DeepCopyInto(*out)
	}
	if in.TerminationGracePeriodSeconds != nil {
		in, out := &in.TerminationGracePeriodSeconds, &out.TerminationGracePeriodSeconds
		*out = new(int64)
		**out = **in
	}
	out.LogVolume = in.LogVolume
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new OpenRestySpec.
func (in *OpenRestySpec) DeepCopy() *OpenRestySpec {
	if in == nil {
		return nil
	}
	out := new(OpenRestySpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *OpenRestyStatus) DeepCopyInto(out *OpenRestyStatus) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new OpenRestyStatus.
func (in *OpenRestyStatus) DeepCopy() *OpenRestyStatus {
	if in == nil {
		return nil
	}
	out := new(OpenRestyStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *RateLimitPolicy) DeepCopyInto(out *RateLimitPolicy) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	out.Spec = in.Spec
	out.Status = in.Status
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new RateLimitPolicy.
func (in *RateLimitPolicy) DeepCopy() *RateLimitPolicy {
	if in == nil {
		return nil
	}
	out := new(RateLimitPolicy)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *RateLimitPolicy) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *RateLimitPolicyList) DeepCopyInto(out *RateLimitPolicyList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]RateLimitPolicy, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new RateLimitPolicyList.
func (in *RateLimitPolicyList) DeepCopy() *RateLimitPolicyList {
	if in == nil {
		return nil
	}
	out := new(RateLimitPolicyList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *RateLimitPolicyList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *RateLimitPolicySpec) DeepCopyInto(out *RateLimitPolicySpec) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new RateLimitPolicySpec.
func (in *RateLimitPolicySpec) DeepCopy() *RateLimitPolicySpec {
	if in == nil {
		return nil
	}
	out := new(RateLimitPolicySpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *RateLimitPolicyStatus) DeepCopyInto(out *RateLimitPolicyStatus) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new RateLimitPolicyStatus.
func (in *RateLimitPolicyStatus) DeepCopy() *RateLimitPolicyStatus {
	if in == nil {
		return nil
	}
	out := new(RateLimitPolicyStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *RequestSpec) DeepCopyInto(out *RequestSpec) {
	*out = *in
	if in.Body != nil {
		in, out := &in.Body, &out.Body
		*out = make(map[string]v1.JSON, len(*in))
		for key, val := range *in {
			(*out)[key] = *val.DeepCopy()
		}
	}
	if in.Query != nil {
		in, out := &in.Query, &out.Query
		*out = make(map[string]v1.JSON, len(*in))
		for key, val := range *in {
			(*out)[key] = *val.DeepCopy()
		}
	}
	if in.QueryFromSecret != nil {
		in, out := &in.QueryFromSecret, &out.QueryFromSecret
		*out = make([]ValueFromSecret, len(*in))
		copy(*out, *in)
	}
	if in.Headers != nil {
		in, out := &in.Headers, &out.Headers
		*out = make([]NginxKV, len(*in))
		copy(*out, *in)
	}
	if in.HeadersFromSecret != nil {
		in, out := &in.HeadersFromSecret, &out.HeadersFromSecret
		*out = make([]ValueFromSecret, len(*in))
		copy(*out, *in)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new RequestSpec.
func (in *RequestSpec) DeepCopy() *RequestSpec {
	if in == nil {
		return nil
	}
	out := new(RequestSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ServerBlock) DeepCopyInto(out *ServerBlock) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	in.Status.DeepCopyInto(&out.Status)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ServerBlock.
func (in *ServerBlock) DeepCopy() *ServerBlock {
	if in == nil {
		return nil
	}
	out := new(ServerBlock)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *ServerBlock) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ServerBlockList) DeepCopyInto(out *ServerBlockList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]ServerBlock, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ServerBlockList.
func (in *ServerBlockList) DeepCopy() *ServerBlockList {
	if in == nil {
		return nil
	}
	out := new(ServerBlockList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *ServerBlockList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ServerBlockSpec) DeepCopyInto(out *ServerBlockSpec) {
	*out = *in
	if in.Headers != nil {
		in, out := &in.Headers, &out.Headers
		*out = make([]NginxKV, len(*in))
		copy(*out, *in)
	}
	if in.LocationRefs != nil {
		in, out := &in.LocationRefs, &out.LocationRefs
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
	if in.Extra != nil {
		in, out := &in.Extra, &out.Extra
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ServerBlockSpec.
func (in *ServerBlockSpec) DeepCopy() *ServerBlockSpec {
	if in == nil {
		return nil
	}
	out := new(ServerBlockSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ServerBlockStatus) DeepCopyInto(out *ServerBlockStatus) {
	*out = *in
	if in.LocationRef != nil {
		in, out := &in.LocationRef, &out.LocationRef
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ServerBlockStatus.
func (in *ServerBlockStatus) DeepCopy() *ServerBlockStatus {
	if in == nil {
		return nil
	}
	out := new(ServerBlockStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ServiceMonitor) DeepCopyInto(out *ServiceMonitor) {
	*out = *in
	if in.Labels != nil {
		in, out := &in.Labels, &out.Labels
		*out = make(map[string]string, len(*in))
		for key, val := range *in {
			(*out)[key] = val
		}
	}
	if in.Annotations != nil {
		in, out := &in.Annotations, &out.Annotations
		*out = make(map[string]string, len(*in))
		for key, val := range *in {
			(*out)[key] = val
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ServiceMonitor.
func (in *ServiceMonitor) DeepCopy() *ServiceMonitor {
	if in == nil {
		return nil
	}
	out := new(ServiceMonitor)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Timeouts) DeepCopyInto(out *Timeouts) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Timeouts.
func (in *Timeouts) DeepCopy() *Timeouts {
	if in == nil {
		return nil
	}
	out := new(Timeouts)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Upstream) DeepCopyInto(out *Upstream) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	in.Status.DeepCopyInto(&out.Status)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Upstream.
func (in *Upstream) DeepCopy() *Upstream {
	if in == nil {
		return nil
	}
	out := new(Upstream)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *Upstream) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *UpstreamList) DeepCopyInto(out *UpstreamList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]Upstream, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new UpstreamList.
func (in *UpstreamList) DeepCopy() *UpstreamList {
	if in == nil {
		return nil
	}
	out := new(UpstreamList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *UpstreamList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *UpstreamServer) DeepCopyInto(out *UpstreamServer) {
	*out = *in
	if in.NormalizeRequestRef != nil {
		in, out := &in.NormalizeRequestRef, &out.NormalizeRequestRef
		*out = new(corev1.LocalObjectReference)
		**out = **in
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new UpstreamServer.
func (in *UpstreamServer) DeepCopy() *UpstreamServer {
	if in == nil {
		return nil
	}
	out := new(UpstreamServer)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *UpstreamServerStatus) DeepCopyInto(out *UpstreamServerStatus) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new UpstreamServerStatus.
func (in *UpstreamServerStatus) DeepCopy() *UpstreamServerStatus {
	if in == nil {
		return nil
	}
	out := new(UpstreamServerStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *UpstreamSpec) DeepCopyInto(out *UpstreamSpec) {
	*out = *in
	if in.Servers != nil {
		in, out := &in.Servers, &out.Servers
		*out = make([]UpstreamServer, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new UpstreamSpec.
func (in *UpstreamSpec) DeepCopy() *UpstreamSpec {
	if in == nil {
		return nil
	}
	out := new(UpstreamSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *UpstreamStatus) DeepCopyInto(out *UpstreamStatus) {
	*out = *in
	if in.Servers != nil {
		in, out := &in.Servers, &out.Servers
		*out = make([]UpstreamServerStatus, len(*in))
		copy(*out, *in)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new UpstreamStatus.
func (in *UpstreamStatus) DeepCopy() *UpstreamStatus {
	if in == nil {
		return nil
	}
	out := new(UpstreamStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ValueFromSecret) DeepCopyInto(out *ValueFromSecret) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ValueFromSecret.
func (in *ValueFromSecret) DeepCopy() *ValueFromSecret {
	if in == nil {
		return nil
	}
	out := new(ValueFromSecret)
	in.DeepCopyInto(out)
	return out
}
