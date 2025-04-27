package handler

import (
	"github.com/stretchr/testify/assert"
	webv1alpha1 "openresty-operator/api/v1alpha1"
	"strings"
	"testing"
)

func TestValidateLocationEntries(t *testing.T) {
	tests := []struct {
		name         string
		entries      []webv1alpha1.LocationEntry
		wantValid    bool
		wantProblems []string
	}{
		{
			name: "Valid unique paths",
			entries: []webv1alpha1.LocationEntry{
				{Path: "/foo"},
				{Path: "/bar"},
			},
			wantValid: true,
		},
		{
			name: "Invalid path format",
			entries: []webv1alpha1.LocationEntry{
				{Path: "foo"},
			},
			wantValid:    false,
			wantProblems: []string{"Invalid path: foo"},
		},
		{
			name: "Duplicate paths",
			entries: []webv1alpha1.LocationEntry{
				{Path: "/foo"},
				{Path: "/foo"},
			},
			wantValid:    false,
			wantProblems: []string{"Duplicate path: /foo"},
		},
		{
			name: "Invalid and Duplicate paths",
			entries: []webv1alpha1.LocationEntry{
				{Path: "foo"},
				{Path: "foo"},
			},
			wantValid:    false,
			wantProblems: []string{"Invalid path: foo", "Duplicate path: foo"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			valid, problems := ValidateLocationEntries(tt.entries)

			assert.Equal(t, tt.wantValid, valid)

			for _, expectedProblem := range tt.wantProblems {
				found := false
				for _, p := range problems {
					if p == expectedProblem || strings.Contains(p, expectedProblem) {
						found = true
						break
					}
				}
				assert.True(t, found, "expected problem containing %q, but got %v", expectedProblem, problems)
			}
		})
	}
}

func TestGenerateLocationConfig(t *testing.T) {
	tests := []struct {
		name         string
		entries      []webv1alpha1.LocationEntry
		wantContains []string
	}{
		{
			name: "Simple proxy_pass",
			entries: []webv1alpha1.LocationEntry{
				{
					Path:      "/foo",
					ProxyPass: "http://backend",
				},
			},
			wantContains: []string{
				"location /foo {",
				"proxy_pass http://backend;",
			},
		},
		{
			name: "FullURL with content_by_lua",
			entries: []webv1alpha1.LocationEntry{
				{
					Path:               "/bar",
					ProxyPassIsFullURL: true,
					Lua: &webv1alpha1.LuaBlock{
						Content: "ngx.say('Hello World')",
					},
				},
			},
			wantContains: []string{
				"location /bar {",
				"content_by_lua_block {",
				"ngx.say('Hello World')",
			},
		},
		{
			name:         "Empty entries",
			entries:      []webv1alpha1.LocationEntry{},
			wantContains: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GenerateLocationConfig(tt.entries)

			for _, expect := range tt.wantContains {
				assert.Contains(t, got, expect, "expected rendered config to contain %q", expect)
			}
		})
	}
}
