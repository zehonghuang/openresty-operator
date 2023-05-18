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
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	webv1alpha1 "openresty-operator/api/v1alpha1"
	"openresty-operator/internal/template"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"strings"
)

// OpenRestyReconciler reconciles a OpenResty object
type OpenRestyReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

const openRestyFinalizer = "openresty.finalizers.chillyroom.com"

// +kubebuilder:rbac:groups=web.chillyroom.com,resources=openresties,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=web.chillyroom.com,resources=openresties/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=web.chillyroom.com,resources=openresties/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the OpenResty object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.19.0/pkg/reconcile
func (r *OpenRestyReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := log.FromContext(ctx).WithValues("openresty", req.NamespacedName)

	var app webv1alpha1.OpenResty
	if err := r.Get(ctx, req.NamespacedName, &app); err != nil {
		if errors.IsNotFound(err) {
			log.Info("OpenResty not found")
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}
	return ctrl.Result{}, nil
}

const defaultInitLua = `
    prometheus = require("prometheus").init("prometheus_metrics")

    metric_upstream_latency = prometheus:histogram(
        "upstream_latency_seconds",
        "Upstream response time in seconds",
        {"upstream"}
    )

    metric_upstream_total = prometheus:counter(
        "upstream_requests_total",
        "Total upstream requests",
        {"upstream", "status"}
    )
`

func renderNginxConf(http *webv1alpha1.HttpBlock, includes []string) string {
	var b strings.Builder
	b.WriteString("worker_processes auto;\n")
	b.WriteString("events { worker_connections 1024; }\n")
	b.WriteString("http {\n")

	// Prometheus 指标共享内存
	b.WriteString("    lua_shared_dict prometheus_metrics 10M;\n\n")

	// Prometheus init_by_lua_block
	b.WriteString("    init_by_lua_block {\n")
	b.WriteString(indentLua(template.DefaultInitLua, "        "))
	b.WriteString("    }\n\n")

	for _, inc := range http.Include {
		b.WriteString(fmt.Sprintf("    include %s;\n", inc))
	}
	if http.LogFormat != "" {
		b.WriteString(fmt.Sprintf("    log_format main '%s';\n", strings.ReplaceAll(http.LogFormat, "\n", "'\n    '")))
	}
	if http.AccessLog != "" {
		b.WriteString(fmt.Sprintf("    access_log %s;\n", http.AccessLog))
	}
	if http.ErrorLog != "" {
		b.WriteString(fmt.Sprintf("    error_log %s;\n", http.ErrorLog))
	}
	if http.ClientMaxBodySize != "" {
		b.WriteString(fmt.Sprintf("    client_max_body_size %s;\n", http.ClientMaxBodySize))
	}
	if http.Gzip {
		b.WriteString("    gzip on;\n")
	}
	for _, line := range http.Extra {
		b.WriteString("    " + line + "\n")
	}

	for _, inc := range includes {
		b.WriteString(inc + "\n")
	}

	b.WriteString("}\n")
	return b.String()
}

func (r *OpenRestyReconciler) deployOpenResty(ctx context.Context, app *webv1alpha1.OpenResty,
	volumes []corev1.Volume, mounts []corev1.VolumeMount, log logr.Logger) error {

	name := "openresty-" + app.Name
	replicas := int32(1)
	if app.Spec.Replicas != nil {
		replicas = *app.Spec.Replicas
	}

	image := "gintonic1glass/openresty:with-prometheus"
	if app.Spec.Image != "" {
		image = app.Spec.Image
	}

	// + mount nginx.conf
	volumes = append(volumes, corev1.Volume{
		Name: "main-config",
		VolumeSource: corev1.VolumeSource{
			ConfigMap: &corev1.ConfigMapVolumeSource{
				LocalObjectReference: corev1.LocalObjectReference{Name: "openresty-" + app.Name},
			},
		},
	})
	mounts = append(mounts, corev1.VolumeMount{
		Name:      "main-config",
		MountPath: "/etc/nginx/nginx.conf",
		SubPath:   "nginx.conf",
	})

	dep := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: app.Namespace,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{"app": name},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{"app": name},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:         "openresty",
							Image:        image,
							VolumeMounts: mounts,
						},
					},
					Volumes: volumes,
				},
			},
		},
	}

	// 可扩展为 patch/update
	return r.Create(ctx, dep)
}

// SetupWithManager sets up the controller with the Manager.
func (r *OpenRestyReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&webv1alpha1.OpenResty{}).
		Owns(&appsv1.Deployment{}).
		Owns(&corev1.Service{}).
		Owns(&corev1.ConfigMap{}).
		Complete(r)
}
