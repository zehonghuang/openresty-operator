package handler

import (
	"openresty-operator/internal/runtime/health"
	"testing"

	webv1alpha1 "openresty-operator/api/v1alpha1"

	"github.com/stretchr/testify/assert"
)

func TestGenerateUpstreamConfig(t *testing.T) {
	tests := []struct {
		name     string
		upstream *webv1alpha1.Upstream
		results  []*health.CheckResult
		wantPart string // 只验证一部分核心内容
	}{
		{
			name: "Address mode with alive servers",
			upstream: &webv1alpha1.Upstream{
				Spec: webv1alpha1.UpstreamSpec{
					Type: webv1alpha1.UpstreamTypeAddress,
				},
			},
			results: []*health.CheckResult{
				{Address: "127.0.0.1:80", Alive: true, Comment: "server 127.0.0.1:80;"},
				{Address: "127.0.0.2:80", Alive: false, Comment: "# server 127.0.0.2:80;  // tcp unreachable"},
			},
			wantPart: "upstream",
		},
		{
			name: "FullURL mode with alive servers",
			upstream: &webv1alpha1.Upstream{
				Spec: webv1alpha1.UpstreamSpec{
					Type: webv1alpha1.UpstreamTypeFullURL,
				},
			},
			results: []*health.CheckResult{
				{Address: "https://foo.com", Alive: true},
				{Address: "https://bar.com", Alive: false},
			},
			wantPart: "local random = require",
		},
		{
			name: "All servers dead",
			upstream: &webv1alpha1.Upstream{
				Spec: webv1alpha1.UpstreamSpec{
					Type: webv1alpha1.UpstreamTypeAddress,
				},
			},
			results: []*health.CheckResult{
				{Address: "127.0.0.1:80", Alive: false},
				{Address: "127.0.0.2:80", Alive: false},
			},
			wantPart: "", // 应该返回空
		},
		{
			name: "Empty server list",
			upstream: &webv1alpha1.Upstream{
				Spec: webv1alpha1.UpstreamSpec{
					Type: webv1alpha1.UpstreamTypeAddress,
				},
			},
			results:  []*health.CheckResult{},
			wantPart: "",
		},
		{
			name: "Unknown UpstreamType",
			upstream: &webv1alpha1.Upstream{
				Spec: webv1alpha1.UpstreamSpec{
					Type: "Unknown",
				},
			},
			results:  []*health.CheckResult{},
			wantPart: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GenerateUpstreamConfig(tt.upstream, tt.results)
			if tt.wantPart == "" {
				assert.Equal(t, "", got, "expected empty config")
			} else {
				assert.Contains(t, got, tt.wantPart, "expected config to contain expected content")
			}
		})
	}
}
