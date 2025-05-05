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
	webv1alpha1 "openresty-operator/api/v1alpha1"
	"openresty-operator/internal/constants"
	"openresty-operator/internal/handler"
	"openresty-operator/internal/metrics"
	"openresty-operator/internal/utils"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sync"
	"time"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

// UpstreamReconciler reconciles a Upstream object
type UpstreamReconciler struct {
	client.Client
	Scheme   *runtime.Scheme
	Recorder record.EventRecorder
}

var DnsCache = struct {
	sync.RWMutex
	Data map[string][]string
}{
	Data: make(map[string][]string),
}

const (
	// 文件扩展名
	UpstreamRenderTypeConf = ".conf"
	UpstreamRenderTypeLua  = ".lua"
)

var UpstreamRenderTypeMap = map[webv1alpha1.UpstreamType]string{
	webv1alpha1.UpstreamTypeAddress: UpstreamRenderTypeConf,
	webv1alpha1.UpstreamTypeFullURL: UpstreamRenderTypeLua,
}

// +kubebuilder:rbac:groups=openresty.huangzehong.me,resources=upstreams,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=openresty.huangzehong.me,resources=upstreams/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=openresty.huangzehong.me,resources=upstreams/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Upstream object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.19.0/pkg/reconcile
func (r *UpstreamReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := log.FromContext(ctx).WithValues("upstream", req.NamespacedName)

	upstream, err := r.fetchUpstream(ctx, req)
	if err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	var statusList []webv1alpha1.UpstreamServerStatus
	results := handler.ProbeUpstreamServers(ctx, upstream)
	for addr, check := range results {
		if check == nil {
			r.Recorder.Eventf(
				upstream,
				corev1.EventTypeNormal,
				"HealthCheckPending",
				fmt.Sprintf("Address %s is pending health check", addr),
			)
		} else if !check.Alive {
			r.Recorder.Eventf(
				upstream,
				corev1.EventTypeWarning,
				check.Reason,
				check.Comment,
			)
		} else {
			for _, ip := range check.IPs {
				metrics.SetUpstreamDNSResolvable(upstream.Namespace, upstream.Name, check.Address, ip, check.Alive)
			}
			metrics.SetUpstreamDNSResolvable(upstream.Namespace, upstream.Name, check.Address, "ALL", check.Alive)
			statusList = append(statusList, webv1alpha1.UpstreamServerStatus{
				Address: check.Address,
				Alive:   check.Alive,
			})
		}
	}

	nginxConfig := handler.GenerateUpstreamConfig(upstream, utils.MapValuesNonNil(results))

	// 写入 ConfigMap
	allDown := false
	if len(nginxConfig) > 0 {
		if err := r.createOrUpdateConfigMap(ctx, upstream, nginxConfig, log); err != nil {
			log.Error(err, "Failed to update ConfigMap")
			return ctrl.Result{}, err
		}
	} else {
		allDown = true
	}

	// 更新 Status
	if allDown {
		r.updateStatus(ctx, upstream, false, nginxConfig, statusList, "All servers unavailable or DNS failed", log)

	} else {
		r.updateStatus(ctx, upstream, true, nginxConfig, statusList, "", log)
	}

	return reconcile.Result{RequeueAfter: 15 * time.Second}, nil
}

func (r *UpstreamReconciler) createOrUpdateConfigMap(ctx context.Context, upstream *webv1alpha1.Upstream, config string, log logr.Logger) error {
	name := "upstream-" + upstream.Name
	dataName := upstream.Name + UpstreamRenderTypeMap[upstream.Spec.Type]
	cm := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: upstream.Namespace,
			Labels:    constants.BuildCommonLabels(upstream, "configmap"),
			Annotations: map[string]string{
				constants.AnnotationGeneratedFromGeneration: fmt.Sprintf("%d", upstream.GetGeneration()),
			},
		},
		Data: map[string]string{
			dataName: config,
		},
	}

	if err := ctrl.SetControllerReference(upstream, cm, r.Scheme); err != nil {
		return err
	}

	var existing corev1.ConfigMap
	err := r.Get(ctx, types.NamespacedName{Name: name, Namespace: upstream.Namespace}, &existing)
	if err != nil {
		if errors.IsNotFound(err) {
			log.Info("Creating ConfigMap", "name", name)
			return r.Create(ctx, cm)
		}
		return err
	}

	needsUpdate := false
	if existing.Data[dataName] != config {
		needsUpdate = true
	}
	if _, ok := existing.Data[dataName]; !ok {
		needsUpdate = true
	}

	if needsUpdate {
		log.Info("Updating ConfigMap", "name", name)
		existing.Data = map[string]string{
			dataName: config,
		}
		existing.Annotations = map[string]string{
			constants.AnnotationGeneratedFromGeneration: fmt.Sprintf("%d", upstream.GetGeneration()),
		}
		return r.Update(ctx, &existing)
	}
	return nil
}

func (r *UpstreamReconciler) updateStatus(
	ctx context.Context,
	current *webv1alpha1.Upstream,
	ready bool,
	nginxConfig string,
	statusList []webv1alpha1.UpstreamServerStatus,
	reason string,
	log logr.Logger,
) {
	current.Status.Ready = ready
	current.Status.NginxConfig = nginxConfig
	current.Status.Servers = statusList
	current.Status.Version = fmt.Sprintf("%d", current.Generation)
	current.Status.Reason = reason

	if err := r.Status().Update(ctx, current); err != nil {
		if errors.IsConflict(err) {
			log.Info("Location status conflict, skipping update")
		} else {
			log.Error(err, "Failed to update Location status")
		}
	}
}

func (r *UpstreamReconciler) fetchUpstream(ctx context.Context, req ctrl.Request) (*webv1alpha1.Upstream, error) {
	var upstream webv1alpha1.Upstream
	if err := r.Get(ctx, req.NamespacedName, &upstream); err != nil {
		return nil, err
	}
	return &upstream, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *UpstreamReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		// Uncomment the following line adding a pointer to an instance of the controlled resource as an argument
		For(&webv1alpha1.Upstream{}).
		Owns(&corev1.ConfigMap{}).
		WithEventFilter(predicate.Funcs{
			DeleteFunc: func(e event.DeleteEvent) bool {
				if obj, ok := e.Object.(*webv1alpha1.Upstream); ok {
					for _, server := range obj.Spec.Servers {
						host, _, _ := utils.SplitHostPort(server)
						metrics.UpstreamDNSResolvable.DeleteLabelValues(obj.Namespace, obj.Name, host)
					}
				}
				return false
			},
			UpdateFunc: func(e event.UpdateEvent) bool {
				oldObj, ok1 := e.ObjectOld.(*webv1alpha1.Upstream)
				newObj, ok2 := e.ObjectNew.(*webv1alpha1.Upstream)
				if !ok1 || !ok2 {
					return true
				}

				oldSet := utils.SetFrom(oldObj.Spec.Servers)
				newSet := utils.SetFrom(newObj.Spec.Servers)

				for server := range oldSet {
					if _, stillPresent := newSet[server]; !stillPresent {
						metrics.UpstreamDNSResolvable.DeleteLabelValues(oldObj.Namespace, oldObj.Name, server)
					}
				}
				return true
			},
		}).
		Complete(r)
}
