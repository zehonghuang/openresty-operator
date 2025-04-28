package handler

import (
	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	webv1alpha1 "openresty-operator/api/v1alpha1"
	"testing"
)

func TestValidateLocationRefs(t *testing.T) {
	tests := []struct {
		name         string
		locations    map[string]*webv1alpha1.Location
		locationRefs []string
		wantValid    bool
		wantProblems []string
	}{
		{
			name: "All Locations Ready and Unique",
			locations: map[string]*webv1alpha1.Location{
				"loc1": {Status: webv1alpha1.LocationStatus{Ready: true}, Spec: webv1alpha1.LocationSpec{Entries: []webv1alpha1.LocationEntry{{Path: "/foo"}}}},
				"loc2": {Status: webv1alpha1.LocationStatus{Ready: true}, Spec: webv1alpha1.LocationSpec{Entries: []webv1alpha1.LocationEntry{{Path: "/bar"}}}},
			},
			locationRefs: []string{"loc1", "loc2"},
			wantValid:    true,
		},
		{
			name: "Missing Location",
			locations: map[string]*webv1alpha1.Location{
				"loc1": {Status: webv1alpha1.LocationStatus{Ready: true}},
			},
			locationRefs: []string{"loc1", "loc2"},
			wantValid:    false,
			wantProblems: []string{"Missing Location: loc2"},
		},
		{
			name: "Location Not Ready",
			locations: map[string]*webv1alpha1.Location{
				"loc1": {Status: webv1alpha1.LocationStatus{Ready: false}},
			},
			locationRefs: []string{"loc1"},
			wantValid:    false,
			wantProblems: []string{"Location not ready: loc1"},
		},
		{
			name: "Duplicated Paths Across Locations",
			locations: map[string]*webv1alpha1.Location{
				"loc1": {Status: webv1alpha1.LocationStatus{Ready: true}, Spec: webv1alpha1.LocationSpec{Entries: []webv1alpha1.LocationEntry{{Path: "/foo"}}}},
				"loc2": {Status: webv1alpha1.LocationStatus{Ready: true}, Spec: webv1alpha1.LocationSpec{Entries: []webv1alpha1.LocationEntry{{Path: "/foo"}}}},
			},
			locationRefs: []string{"loc1", "loc2"},
			wantValid:    false,
			wantProblems: []string{"Duplicated path '/foo' in loc1 and loc2"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			valid, problems := ValidateLocationRefs(tt.locations, tt.locationRefs)

			assert.Equal(t, tt.wantValid, valid)
			for _, expectedProblem := range tt.wantProblems {
				found := false
				for _, p := range problems {
					if p == expectedProblem || (len(expectedProblem) > 0 && p == expectedProblem) {
						found = true
						break
					}
				}
				assert.True(t, found, "expected problem containing %q, got %v", expectedProblem, problems)
			}
		})
	}
}

func TestGenerateServerBlockConfig(t *testing.T) {
	s := &webv1alpha1.ServerBlock{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-server",
			Namespace: "default",
		},
		Spec: webv1alpha1.ServerBlockSpec{
			Listen: "80",
			LocationRefs: []string{
				"loc1",
				"loc2",
			},
			Headers: []webv1alpha1.NginxKV{
				{Key: "X-Frame-Options", Value: "DENY"},
			},
			Extra: []string{
				"client_max_body_size 20m;",
			},
		},
	}

	conf := GenerateServerBlockConfig(s)

	assert.Contains(t, conf, "listen 80;")
	assert.Contains(t, conf, "server_name test-server.default.svc.cluster.local;")
	assert.Contains(t, conf, "include /etc/nginx/conf.d/locations/loc1/loc1.conf;")
	assert.Contains(t, conf, "include /etc/nginx/conf.d/locations/loc2/loc2.conf;")
	assert.Contains(t, conf, "add_header X-Frame-Options DENY;")
	assert.Contains(t, conf, "client_max_body_size 20m;")
}
