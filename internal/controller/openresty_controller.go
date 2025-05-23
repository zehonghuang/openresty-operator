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
	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	webv1alpha1 "openresty-operator/api/v1alpha1"
	"openresty-operator/internal/constants"
	"openresty-operator/internal/handler"
	"openresty-operator/internal/runtime/metrics"
	"openresty-operator/internal/utils"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
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

	app, err := r.fetchOpenResty(ctx, req)
	if err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	serverStatus := handler.ValidateServerRefs(r.Get, app)
	upstreamStatus := handler.ValidateUpstreamRefs(r.Get, app)

	if !serverStatus.AllReady || !upstreamStatus.AllReady {
		reason := handler.ComposeDependencyFailureReason(serverStatus, upstreamStatus)
		r.handleDependencyFailure(ctx, app, reason, log)
		return ctrl.Result{RequeueAfter: DefaultRequeue}, nil
	}

	nginxConf := handler.RenderNginxConf(
		app.Spec.Http,
		app.Spec.MetricsServer,
		handler.BuildIncludeLines(app, upstreamStatus))

	if err := handler.CreateOrUpdateConfigMap(
		ctx, r.Client, r.Scheme, app,
		"openresty-"+app.Name+"-main",
		app.Namespace,
		constants.BuildCommonLabels(app, "configmap"),
		map[string]string{"nginx.conf": nginxConf},
		log, nil, nil); err != nil {
		return ctrl.Result{}, err
	}

	if err, _ = handler.DeployOpenRestyPod(ctx, r.Client, r.Scheme, app, upstreamStatus.UpstreamsType, log); err != nil {
		return ctrl.Result{}, err
	}

	if err := handler.DeployServerBlockServices(ctx, r.Client, r.Scheme, app, log); err != nil {
		return ctrl.Result{}, err
	}

	if app.Spec.ServiceMonitor.Enable {
		if err := handler.CreateOrUpdateMetricsService(ctx, r.Client, r.Scheme, app); err != nil {
			log.Error(err, "Failed to create or update MetricsService")
			return ctrl.Result{}, err
		}
		if err := handler.CreateOrUpdateServiceMonitor(ctx, r.Client, r.Scheme, app, app.Spec.ServiceMonitor.Labels, app.Spec.ServiceMonitor.Annotations, log); err != nil {
			log.Error(err, "Failed to create or update ServiceMonitor")
			return ctrl.Result{}, err
		}
	}

	app.Status.Ready = true
	app.Status.Reason = ""
	app.Status.AvailableReplicas = *app.Spec.Replicas
	if err := r.Status().Update(ctx, app); err != nil {
		if errors.IsConflict(err) {
			log.Info("OpenResty status conflict, skipping update")
		} else {
			log.Error(err, "Failed to update OpenResty status")
		}
	}

	return ctrl.Result{}, nil
}

func (r *OpenRestyReconciler) fetchOpenResty(ctx context.Context, req ctrl.Request) (*webv1alpha1.OpenResty, error) {
	var openresty webv1alpha1.OpenResty
	if err := r.Get(ctx, req.NamespacedName, &openresty); err != nil {
		return nil, err
	}
	return &openresty, nil
}

func (r *OpenRestyReconciler) handleDependencyFailure(ctx context.Context, app *webv1alpha1.OpenResty, reason string, log logr.Logger) {
	r.Recorder.Eventf(app, corev1.EventTypeWarning, "DependencyFailure", reason)

	app.Status.Ready = false
	app.Status.Reason = reason
	app.Status.AvailableReplicas = 0

	if err := r.Status().Update(ctx, app); err != nil {
		if errors.IsConflict(err) {
			log.Info("OpenResty status conflict during dependency failure update, skipping")
		} else {
			log.Error(err, "Failed to update OpenResty status after dependency failure")
		}
	}
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
		Owns(&monitoringv1.ServiceMonitor{}).
		WithEventFilter(predicate.Funcs{
			DeleteFunc: func(e event.DeleteEvent) bool {
				if obj, ok := e.Object.(*webv1alpha1.OpenResty); ok {
					for _, serverRef := range obj.Spec.Http.ServerRefs {
						metrics.OpenRestyCRDRefStatus.DeleteLabelValues(obj.Namespace, obj.Name, webv1alpha1.ServerBlock{}.Kind, serverRef)
					}
					for _, upstreamRef := range obj.Spec.Http.UpstreamRefs {
						metrics.OpenRestyCRDRefStatus.DeleteLabelValues(obj.Namespace, obj.Name, webv1alpha1.Upstream{}.Kind, upstreamRef)
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
						metrics.OpenRestyCRDRefStatus.DeleteLabelValues(oldObj.Namespace, oldObj.Name, webv1alpha1.ServerBlock{}.Kind, serverRef)
					}
				}

				oldSet = utils.SetFrom(oldObj.Spec.Http.UpstreamRefs)
				newSet = utils.SetFrom(newObj.Spec.Http.UpstreamRefs)

				for upstreamRef := range oldSet {
					if _, stillPresent := newSet[upstreamRef]; !stillPresent {
						metrics.OpenRestyCRDRefStatus.DeleteLabelValues(oldObj.Namespace, oldObj.Name, webv1alpha1.Upstream{}.Kind, upstreamRef)
					}
				}
				return true
			},
		}).
		Complete(r)
}
