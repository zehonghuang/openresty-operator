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
	"bytes"
	"context"
	"fmt"
	"github.com/go-logr/logr"
	"github.com/google/go-cmp/cmp"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/tools/record"
	"k8s.io/utils/ptr"
	webv1alpha1 "openresty-operator/api/v1alpha1"
	"openresty-operator/internal/metrics"
	"openresty-operator/internal/template"
	"openresty-operator/internal/utils"
	"reflect"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"strings"
	"time"

	_template "text/template"
)

// OpenRestyReconciler reconciles a OpenResty object
type OpenRestyReconciler struct {
	client.Client
	Scheme   *runtime.Scheme
	Recorder record.EventRecorder
}

// +kubebuilder:rbac:groups=openresty.huangzehong.me,resources=openresties,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=openresty.huangzehong.me,resources=openresties/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=openresty.huangzehong.me,resources=openresties/finalizers,verbs=update

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

	var missingServers, notReadyServers, missingServerCMs []string
	for _, name := range app.Spec.Http.ServerRefs {
		var srv webv1alpha1.ServerBlock
		if err := r.Get(ctx, types.NamespacedName{Name: name, Namespace: app.Namespace}, &srv); err != nil {
			if errors.IsNotFound(err) {
				missingServers = append(missingServers, name)
				continue
			}
			return ctrl.Result{}, err
		}
		metrics.SetCRDRefStatus(app.Namespace, app.Name, srv.Kind, srv.Name, srv.Status.Ready)
		if !srv.Status.Ready {
			notReadyServers = append(notReadyServers, name)
			continue
		}
		// 检查对应 ConfigMap
		cmName := "serverblock-" + name
		var cm corev1.ConfigMap
		if err := r.Get(ctx, types.NamespacedName{Name: cmName, Namespace: app.Namespace}, &cm); err != nil {
			if errors.IsNotFound(err) {
				missingServerCMs = append(missingServerCMs, cmName)
				continue
			}
			return ctrl.Result{}, err
		}
	}

	var missingUpstreams, notReadyUpstreams, missingUpstreamCMs []string
	upstreamsType := map[string]webv1alpha1.UpstreamType{}
	for _, name := range app.Spec.Http.UpstreamRefs {
		var ups webv1alpha1.Upstream
		if err := r.Get(ctx, types.NamespacedName{Name: name, Namespace: app.Namespace}, &ups); err != nil {
			if errors.IsNotFound(err) {
				missingUpstreams = append(missingUpstreams, name)
				continue
			}
			return ctrl.Result{}, err
		}
		if !ups.Status.Ready {
			notReadyUpstreams = append(notReadyUpstreams, name)
			continue
		}
		cmName := "upstream-" + name
		upstreamsType[name] = ups.Spec.Type
		var cm corev1.ConfigMap
		if err := r.Get(ctx, types.NamespacedName{Name: cmName, Namespace: app.Namespace}, &cm); err != nil {
			if errors.IsNotFound(err) {
				missingUpstreamCMs = append(missingUpstreamCMs, cmName)
				continue
			}
			return ctrl.Result{}, err
		}
	}

	if len(missingServers)+len(notReadyServers)+len(missingServerCMs)+
		len(missingUpstreams)+len(notReadyUpstreams)+len(missingUpstreamCMs) > 0 {
		msg := fmt.Sprintf("ServerRefs missing=%v notReady=%v noCM=%v; UpstreamRefs missing=%v notReady=%v noCM=%v",
			missingServers, notReadyServers, missingServerCMs,
			missingUpstreams, notReadyUpstreams, missingUpstreamCMs)

		log.Info("Dependency check failed", "details", msg)
		r.Recorder.Eventf(&app, corev1.EventTypeWarning, "InvalidRefs", msg)

		app.Status.Ready = false
		app.Status.Reason = msg
		if err := r.Status().Update(ctx, &app); err != nil {
			if errors.IsConflict(err) {
				log.Info("OpenResty status conflict, skipping update")
			} else {
				log.Error(err, "Failed to update OpenResty status")
			}
		}

		return ctrl.Result{RequeueAfter: 30 * time.Second}, nil
	}

	var includeLines []string
	for _, name := range app.Spec.Http.ServerRefs {
		includeLines = append(includeLines,
			fmt.Sprintf("include %s/%s/%s.conf;", utils.NginxServerConfigDir, name, name))
	}
	for _, name := range app.Spec.Http.UpstreamRefs {
		if upstreamsType[name] == webv1alpha1.UpstreamTypeAddress {
			includeLines = append(includeLines,
				fmt.Sprintf("include %s/%s/%s.conf;", utils.NginxUpstreamConfigDir, name, name))
		}
	}

	nginxConf := renderNginxConf(app.Spec.Http, app.Spec.MetricsServer, includeLines)

	cm := buildMainNginxConfConfigMap(&app, nginxConf)

	if err := createOrUpdateConfigMap(ctx, r.Client, r.Scheme, &app, cm, log); err != nil {
		return ctrl.Result{}, err
	}

	if err := r.deployOpenResty(ctx, &app, upstreamsType, log); err != nil {
		return ctrl.Result{}, err
	}

	for _, name := range app.Spec.Http.ServerRefs {
		var server webv1alpha1.ServerBlock
		if err := r.Get(ctx, types.NamespacedName{Name: name, Namespace: app.Namespace}, &server); err != nil {
			return ctrl.Result{}, err
		}

		svc := generateServiceForServer(&app, &server)
		if err := r.createOrUpdateService(ctx, &app, svc, log); err != nil {
			return ctrl.Result{}, err
		}
	}

	app.Status.Ready = true
	app.Status.Reason = ""
	app.Status.AvailableReplicas = *app.Spec.Replicas
	if err := r.Status().Update(ctx, &app); err != nil {
		if errors.IsConflict(err) {
			log.Info("OpenResty status conflict, skipping update")
		} else {
			log.Error(err, "Failed to update OpenResty status")
		}
	}

	return ctrl.Result{}, nil
}

func buildMainNginxConfConfigMap(app *webv1alpha1.OpenResty, nginxConf string) *corev1.ConfigMap {
	return &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "openresty-" + app.Name + "-main",
			Namespace: app.Namespace,
			Labels: map[string]string{
				"app": app.Name,
			},
		},
		Data: map[string]string{
			"nginx.conf": nginxConf,
		},
	}
}

func createOrUpdateConfigMap(ctx context.Context, c client.Client, scheme *runtime.Scheme, owner client.Object, cm *corev1.ConfigMap, log logr.Logger) error {
	if err := ctrl.SetControllerReference(owner, cm, scheme); err != nil {
		return err
	}

	var existing corev1.ConfigMap
	err := c.Get(ctx, types.NamespacedName{Name: cm.Name, Namespace: cm.Namespace}, &existing)
	if err != nil {
		if errors.IsNotFound(err) {
			log.Info("Creating ConfigMap", "name", cm.Name)
			return c.Create(ctx, cm)
		}
		return err
	}

	if !reflect.DeepEqual(existing.Data, cm.Data) {
		log.Info("Updating ConfigMap", "name", cm.Name)
		existing.Data = cm.Data
		return c.Update(ctx, &existing)
	}

	return nil
}

type nginxConfData struct {
	InitLua           string
	EnableMetrics     bool
	MetricsPort       string
	MetricsPath       string
	Includes          []string
	LogFormat         string
	AccessLog         string
	ErrorLog          string
	ClientMaxBodySize string
	Gzip              bool
	Extra             []string
	IncludeSnippets   []string
}

func defaultOr(s, fallback string) string {
	if s == "" {
		return fallback
	}
	return s
}

func renderNginxConf(http *webv1alpha1.HttpBlock, metrics *webv1alpha1.MetricsServer, includes []string) string {

	data := nginxConfData{
		InitLua:           template.DefaultInitLua,
		EnableMetrics:     metrics != nil && metrics.Enable,
		MetricsPort:       defaultOr(metrics.Listen, "9091"),
		MetricsPath:       defaultOr(metrics.Path, "/metrics"),
		Includes:          http.Include,
		LogFormat:         utils.SanitizeLogFormat(http.LogFormat),
		AccessLog:         http.AccessLog,
		ErrorLog:          http.ErrorLog,
		ClientMaxBodySize: http.ClientMaxBodySize,
		Gzip:              http.Gzip,
		Extra:             http.Extra,
		IncludeSnippets:   includes,
	}

	tmpl := _template.Must(_template.New("nginx").Funcs(_template.FuncMap{
		"indent": func(s string, spaces int) string {
			pad := strings.Repeat(" ", spaces)
			lines := strings.Split(s, "\n")
			for i := range lines {
				lines[i] = pad + lines[i]
			}
			return strings.Join(lines, "\n")
		},
	}).Parse(utils.NginxTemplate))

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return fmt.Sprintf("# failed to render: %v", err)
	}

	return buf.String()
}

func (r *OpenRestyReconciler) deployOpenResty(ctx context.Context, app *webv1alpha1.OpenResty,
	upstreamsType map[string]webv1alpha1.UpstreamType,
	log logr.Logger) error {

	var volumes []corev1.Volume
	var mounts []corev1.VolumeMount

	name := "openresty-" + app.Name
	replicas := int32(1)
	if app.Spec.Replicas != nil {
		replicas = *app.Spec.Replicas
	}

	image := "gintonic1glass/openresty:alpine-1.1.0"
	if app.Spec.Image != "" {
		image = app.Spec.Image
	}

	// + mount nginx.conf
	volumes = append(volumes, corev1.Volume{
		Name: "main-config",
		VolumeSource: corev1.VolumeSource{
			ConfigMap: &corev1.ConfigMapVolumeSource{
				LocalObjectReference: corev1.LocalObjectReference{
					Name: "openresty-" + app.Name + "-main",
				},
			},
		},
	})
	mounts = append(mounts, corev1.VolumeMount{
		Name:      "main-config",
		MountPath: utils.NginxConfPath,
		SubPath:   "nginx.conf",
	})

	locationSeen := map[string]bool{}

	for _, serverName := range app.Spec.Http.ServerRefs {
		volumes = append(volumes, corev1.Volume{
			Name: "serverblock-" + serverName,
			VolumeSource: corev1.VolumeSource{
				ConfigMap: &corev1.ConfigMapVolumeSource{
					LocalObjectReference: corev1.LocalObjectReference{Name: "serverblock-" + serverName},
				},
			},
		})
		mounts = append(mounts, corev1.VolumeMount{
			Name:      "serverblock-" + serverName,
			MountPath: utils.NginxServerConfigDir + "/" + serverName,
		})

		var server webv1alpha1.ServerBlock
		if err := r.Get(ctx, types.NamespacedName{Name: serverName, Namespace: app.Namespace}, &server); err != nil {
			return err
		}

		for _, locName := range server.Spec.LocationRefs {
			if locationSeen[locName] {
				continue
			}
			locationSeen[locName] = true

			volumes = append(volumes, corev1.Volume{
				Name: "location-" + locName,
				VolumeSource: corev1.VolumeSource{
					ConfigMap: &corev1.ConfigMapVolumeSource{
						LocalObjectReference: corev1.LocalObjectReference{Name: "location-" + locName},
					},
				},
			})
			mounts = append(mounts, corev1.VolumeMount{
				Name:      "location-" + locName,
				MountPath: utils.NginxLocationConfigDir + "/" + locName,
			})
		}
	}

	for _, upstreamName := range app.Spec.Http.UpstreamRefs {
		volumes = append(volumes, corev1.Volume{
			Name: "upstream-" + upstreamName,
			VolumeSource: corev1.VolumeSource{
				ConfigMap: &corev1.ConfigMapVolumeSource{
					LocalObjectReference: corev1.LocalObjectReference{Name: "upstream-" + upstreamName},
				},
			},
		})

		Path := ""
		if upstreamsType[upstreamName] == webv1alpha1.UpstreamTypeAddress {
			Path = utils.NginxUpstreamConfigDir + "/" + upstreamName
		} else {
			Path = utils.NginxLuaLibUpstreamDir + "/" + upstreamName
		}
		mounts = append(mounts, corev1.VolumeMount{
			Name:      "upstream-" + upstreamName,
			MountPath: Path,
		})

	}

	var metricsPort corev1.ContainerPort
	if app.Spec.MetricsServer != nil && app.Spec.MetricsServer.Enable {
		port := "8080"
		if app.Spec.MetricsServer.Listen != "" {
			port = app.Spec.MetricsServer.Listen
		}
		metricsPort.ContainerPort = utils.ParseListenPort(port)
		metricsPort.Name = "metrics"
		metricsPort.Protocol = corev1.ProtocolTCP
	}

	// 1. 获取 Deployment 的默认值模板
	defaulted := &appsv1.Deployment{}
	r.Scheme.Default(defaulted)

	// 2. 构造你的业务 Deployment（注意不要提前写 defaulted 字段）
	dep := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: app.Namespace,
			Labels:    map[string]string{"app": name},
		},
		Spec: defaulted.Spec, // 复制默认值
	}

	// 3. 设置你的业务逻辑字段
	dep.Spec.Replicas = &replicas
	dep.Spec.Selector = &metav1.LabelSelector{
		MatchLabels: map[string]string{"app": name},
	}
	dep.Spec.Template.ObjectMeta.Labels = map[string]string{"app": name}
	dep.Spec.Template.Annotations = buildPrometheusAnnotations(app.Spec.MetricsServer)
	if v, ok := app.Annotations["openresty.huangzehong.me/trigger-hash"]; ok {
		dep.Spec.Template.Annotations["openresty.huangzehong.me/trigger-hash"] = v
	}
	dep.Spec.Template.Spec.ShareProcessNamespace = ptr.To(true)
	dep.Spec.Template.Spec.Volumes = volumes
	dep.Spec.Template.Spec.Containers = []corev1.Container{
		{
			Name:  "openresty",
			Image: image,
			Ports: []corev1.ContainerPort{
				{
					Name:          "http",
					ContainerPort: 80,
					Protocol:      corev1.ProtocolTCP,
				},
				metricsPort,
			},
			VolumeMounts: mounts,
		},
		{
			Name:  "reload-agent",
			Image: "gintonic1glass/reload-agent:v0.1.5",
			Ports: []corev1.ContainerPort{
				{
					Name:          "metrics",
					ContainerPort: 19091,
					Protocol:      corev1.ProtocolTCP,
				},
			},
			Env:          app.Spec.ReloadAgentEnv,
			VolumeMounts: mounts[1:], // 不挂主配置
		},
	}

	existing := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      dep.Name,
			Namespace: dep.Namespace,
		},
	}

	_, err := controllerutil.CreateOrUpdate(ctx, r.Client, existing, func() error {
		if !reflect.DeepEqual(existing.Spec, dep.Spec) {
			diff := cmp.Diff(existing.Spec, dep.Spec)
			log.V(4).Info("Deployment spec has changed", "diff", diff)
			existing.Spec = dep.Spec
		}
		return ctrl.SetControllerReference(app, existing, r.Scheme)
	})

	return err
}

func buildPrometheusAnnotations(metrics *webv1alpha1.MetricsServer) map[string]string {
	if metrics == nil || !metrics.Enable {
		return map[string]string{}
	}
	port := defaultOr(metrics.Listen, "9091")
	path := defaultOr(metrics.Path, "/metrics")

	return map[string]string{
		"prometheus.io/scrape": "true",
		"prometheus.io/port":   port,
		"prometheus.io/path":   path,
	}
}

func generateServiceForServer(app *webv1alpha1.OpenResty, server *webv1alpha1.ServerBlock) *corev1.Service {
	port := utils.ParseListenPort(server.Spec.Listen)

	return &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      server.Name,
			Namespace: app.Namespace,
			Labels: map[string]string{
				"app":  "openresty-" + app.Name,
				"type": "server-service",
			},
		},
		Spec: corev1.ServiceSpec{
			Selector: map[string]string{
				"app": "openresty-" + app.Name,
			},
			Ports: []corev1.ServicePort{
				{
					Name:       "http",
					Port:       port,
					TargetPort: intstr.FromInt32(int32(port)),
					Protocol:   corev1.ProtocolTCP,
				},
			},
		},
	}
}

func (r *OpenRestyReconciler) createOrUpdateService(ctx context.Context, app *webv1alpha1.OpenResty, svc *corev1.Service, log logr.Logger) error {
	if err := ctrl.SetControllerReference(app, svc, r.Scheme); err != nil {
		return err
	}
	var existing corev1.Service
	err := r.Get(ctx, types.NamespacedName{Name: svc.Name, Namespace: svc.Namespace}, &existing)
	if err != nil {
		if errors.IsNotFound(err) {
			log.Info("Creating Service", "name", svc.Name)
			return r.Create(ctx, svc)
		}
		return err
	}
	svc.ResourceVersion = existing.ResourceVersion
	return r.Update(ctx, svc)
}

// SetupWithManager sets up the controller with the Manager.
func (r *OpenRestyReconciler) SetupWithManager(mgr ctrl.Manager) error {
	if err := mgr.GetFieldIndexer().IndexField(
		context.Background(),
		&webv1alpha1.OpenResty{},
		"spec.http.serverRefs",
		func(obj client.Object) []string {
			app := obj.(*webv1alpha1.OpenResty)
			var keys []string
			if app.Spec.Http != nil {
				for _, serverRef := range app.Spec.Http.ServerRefs {
					keys = append(keys, fmt.Sprintf("%s/%s", app.Namespace, serverRef))
				}
			}
			return keys
		},
	); err != nil {
		return err
	}

	return ctrl.NewControllerManagedBy(mgr).
		For(&webv1alpha1.OpenResty{}).
		Owns(&appsv1.Deployment{}).
		Owns(&corev1.Service{}).
		Owns(&corev1.ConfigMap{}).
		WithEventFilter(predicate.Funcs{
			DeleteFunc: func(e event.DeleteEvent) bool {
				if obj, ok := e.Object.(*webv1alpha1.OpenResty); ok {
					for _, serverRef := range obj.Spec.Http.ServerRefs {
						metrics.OpenRestyCRDRefStatus.DeleteLabelValues(obj.Namespace, obj.Name, serverRef)
					}
				}
				return false
			},
			UpdateFunc: func(e event.UpdateEvent) bool {
				oldObj, ok1 := e.ObjectOld.(*webv1alpha1.OpenResty)
				newObj, ok2 := e.ObjectNew.(*webv1alpha1.OpenResty)
				if !ok1 || !ok2 {
					return true
				}

				oldSet := utils.SetFrom(oldObj.Spec.Http.ServerRefs)
				newSet := utils.SetFrom(newObj.Spec.Http.ServerRefs)

				for serverRef := range oldSet {
					if _, stillPresent := newSet[serverRef]; !stillPresent {
						metrics.OpenRestyCRDRefStatus.DeleteLabelValues(oldObj.Namespace, oldObj.Name, serverRef)
					}
				}
				return true
			},
		}).
		Complete(r)
}
