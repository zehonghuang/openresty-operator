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

package controller

import (
	"context"
	"fmt"
	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	"net/url"
	"openresty-operator/internal/metrics"
	"openresty-operator/internal/utils"
	"regexp"
	"strings"
	"time"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	webv1alpha1 "openresty-operator/api/v1alpha1"
)

// LocationReconciler reconciles a Location object
type LocationReconciler struct {
	client.Client
	Scheme   *runtime.Scheme
	Recorder record.EventRecorder
}

// +kubebuilder:rbac:groups=openresty.huangzehong.me,resources=locations,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=openresty.huangzehong.me,resources=locations/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=openresty.huangzehong.me,resources=locations/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Location object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.19.0/pkg/reconcile
func (r *LocationReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := log.FromContext(ctx).WithValues("location", req.NamespacedName)

	var location webv1alpha1.Location
	if err := r.Get(ctx, req.NamespacedName, &location); err != nil {
		if errors.IsNotFound(err) {
			log.Info("Location resource not found, likely deleted")
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}

	// 校验 location 合法性以及是否重复
	pathSeen := make(map[string]struct{})
	var invalidPaths []string
	var duplicatePaths []string

	for _, entry := range location.Spec.Entries {
		path := entry.Path
		valid, reason := utils.ValidateLocationPath(path)
		if !valid {
			invalidPaths = append(invalidPaths, fmt.Sprintf("%s (%s)", path, reason))
		}
		if _, exists := pathSeen[path]; exists {
			duplicatePaths = append(duplicatePaths, path)
		} else {
			pathSeen[path] = struct{}{}
		}
	}

	if len(invalidPaths) > 0 || len(duplicatePaths) > 0 {
		var allProblems []string
		if len(invalidPaths) > 0 {
			allProblems = append(allProblems, fmt.Sprintf("Invalid paths: %s", strings.Join(invalidPaths, ", ")))
		}
		if len(duplicatePaths) > 0 {
			allProblems = append(allProblems, fmt.Sprintf("Duplicate paths: %s", strings.Join(duplicatePaths, ", ")))
		}

		msg := strings.Join(allProblems, " | ")
		log.Error(nil, "Path validation failed", "details", msg)
		r.Recorder.Eventf(&location, corev1.EventTypeWarning, "InvalidPath", msg)
		metrics.Recorder(location.Kind, location.Namespace, location.Name, corev1.EventTypeWarning, msg)

		r.updateLocationStatus(ctx, location, false, msg, log)

		return ctrl.Result{RequeueAfter: 30 * time.Second}, nil
	}

	conf := renderLocationEntries(location.Spec.Entries)

	if err := r.createOrUpdateConfigMap(ctx, &location, conf, log); err != nil {
		return ctrl.Result{}, err
	}

	r.updateLocationStatus(ctx, location, true, "", log)

	return ctrl.Result{RequeueAfter: 30 * time.Second}, nil

}

func (r *LocationReconciler) createOrUpdateConfigMap(ctx context.Context, loc *webv1alpha1.Location, conf string, log logr.Logger) error {
	name := "location-" + loc.Name
	dataName := loc.Name + ".conf"
	cm := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: loc.Namespace,
			Annotations: map[string]string{
				"openresty.huangzehong.me/generated-from-generation": fmt.Sprintf("%d", loc.GetGeneration()),
			},
		},
		Data: map[string]string{
			dataName: conf,
		},
	}

	if err := ctrl.SetControllerReference(loc, cm, r.Scheme); err != nil {
		return err
	}

	var existing corev1.ConfigMap
	err := r.Get(ctx, types.NamespacedName{Name: name, Namespace: loc.Namespace}, &existing)
	if err != nil {
		if errors.IsNotFound(err) {
			log.Info("Creating ConfigMap", "name", name)
			return r.Create(ctx, cm)
		}
		return err
	}

	if existing.Data[dataName] != conf {
		log.Info("Updating ConfigMap", "name", name)
		existing.Data[dataName] = conf
		existing.Annotations = map[string]string{
			"openresty.huangzehong.me/generated-from-generation": fmt.Sprintf("%d", loc.GetGeneration()),
		}
		return r.Update(ctx, &existing)
	}

	return nil
}

func renderLocationEntries(entries []webv1alpha1.LocationEntry) string {
	var b strings.Builder
	for _, e := range entries {
		b.WriteString(fmt.Sprintf("location %s {\n", e.Path))

		if e.ProxyPassIsFullURL {
			if e.Lua != nil && e.Lua.Content != "" {
				b.WriteString("    content_by_lua_block {\n")
				b.WriteString(e.Lua.Content)
				b.WriteString("    }\n")
			} else {
				b.WriteString("    content_by_lua_block {\n")
				b.WriteString(fmt.Sprintf("        ngx.var.target = require(\"upstreams/%s\"):pick() .. ngx.var.request_uri\n", safeName(e.ProxyPass)))
				b.WriteString("    }\n")
			}
			b.WriteString("    proxy_pass $target;\n")
		} else if e.ProxyPass != "" {
			b.WriteString(fmt.Sprintf("    proxy_pass %s;\n", e.ProxyPass))
		}

		for _, h := range e.Headers {
			b.WriteString(fmt.Sprintf("    proxy_set_header %s %s;\n", h.Key, h.Value))
		}

		if e.Timeout != nil {
			if e.Timeout.Connect != "" {
				b.WriteString(fmt.Sprintf("    proxy_connect_timeout %s;\n", e.Timeout.Connect))
			}
			if e.Timeout.Send != "" {
				b.WriteString(fmt.Sprintf("    proxy_send_timeout %s;\n", e.Timeout.Send))
			}
			if e.Timeout.Read != "" {
				b.WriteString(fmt.Sprintf("    proxy_read_timeout %s;\n", e.Timeout.Read))
			}
		}

		if e.AccessLog != nil && !*e.AccessLog {
			b.WriteString("    access_log off;\n")
		}

		if e.LimitReq != nil {
			b.WriteString(fmt.Sprintf("    limit_req %s;\n", *e.LimitReq))
		}

		if e.Gzip != nil && e.Gzip.Enable {
			b.WriteString("    gzip on;\n")
			if len(e.Gzip.Types) > 0 {
				b.WriteString(fmt.Sprintf("    gzip_types %s;\n", strings.Join(e.Gzip.Types, " ")))
			}
		}

		if e.Cache != nil {
			if e.Cache.Zone != "" {
				b.WriteString(fmt.Sprintf("    proxy_cache %s;\n", e.Cache.Zone))
			}
			if e.Cache.Valid != "" {
				b.WriteString(fmt.Sprintf("    proxy_cache_valid %s;\n", e.Cache.Valid))
			}
		}

		if e.Lua != nil && e.Lua.Access != "" {
			b.WriteString("    access_by_lua_block {\n")
			b.WriteString(indentLua(e.Lua.Access, "        "))
			b.WriteString("    }\n")
		}

		for _, extra := range e.Extra {
			b.WriteString(fmt.Sprintf("    %s\n", extra))
		}

		if e.EnableUpstreamMetrics {
			b.WriteString("    log_by_lua_block {\n")
			b.WriteString("        local addr = (ngx.var.upstream_addr or \"unknown\"):match(\"^[^,]+\")\n")
			b.WriteString("        local status = ngx.var.status\n")
			b.WriteString("        local latency = tonumber(ngx.var.upstream_response_time) or 0\n")
			b.WriteString("        metric_upstream_latency:observe(latency, {addr})\n")
			b.WriteString("        metric_upstream_total:inc(1, {addr, status})\n")
			b.WriteString("    }\n")
		}

		b.WriteString("}\n\n")
	}

	return b.String()
}

func safeName(proxyPass string) string {
	u, err := url.Parse(proxyPass)
	if err != nil || u.Host == "" {
		return "invalid-proxypass"
	}

	host := u.Host

	// 将 host 中的非法字符替换成 "-"
	host = strings.ToLower(host)
	host = strings.ReplaceAll(host, ".", "-")
	host = strings.ReplaceAll(host, ":", "-")

	// 确保只包含合法字符
	reg := regexp.MustCompile(`[^a-z0-9\-]`)
	host = reg.ReplaceAllString(host, "")

	// 最长限制 63 字符（K8s 对象名规范）
	if len(host) > 63 {
		host = host[:63]
	}

	return host
}

func indentLua(code, prefix string) string {
	lines := strings.Split(code, "\n")
	for i, line := range lines {
		lines[i] = prefix + line
	}
	return strings.Join(lines, "\n") + "\n"
}

func (r *LocationReconciler) updateLocationStatus(
	ctx context.Context,
	current webv1alpha1.Location,
	ready bool,
	reason string,
	log logr.Logger,
) {
	current.Status.Ready = ready
	current.Status.Version = fmt.Sprintf("%d", current.Generation)
	current.Status.Reason = reason

	if err := r.Status().Update(ctx, &current); err != nil {
		if errors.IsConflict(err) {
			log.Info("Location status conflict, skipping update")
		} else {
			log.Error(err, "Failed to update Location status")
		}
	}
}

// SetupWithManager sets up the controller with the Manager.
func (r *LocationReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&webv1alpha1.Location{}).
		Owns(&corev1.ConfigMap{}).
		Complete(r)
}
