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
	"openresty-operator/internal/constants"
	"openresty-operator/internal/handler"
	"openresty-operator/internal/metrics"
	"openresty-operator/internal/utils"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"strings"
	"time"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	webv1alpha1 "openresty-operator/api/v1alpha1"
)

// ServerBlockReconciler reconciles a ServerBlock object
type ServerBlockReconciler struct {
	client.Client
	Scheme   *runtime.Scheme
	Recorder record.EventRecorder
}

// +kubebuilder:rbac:groups=openresty.huangzehong.me,resources=serverblocks,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=openresty.huangzehong.me,resources=serverblocks/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=openresty.huangzehong.me,resources=serverblocks/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the ServerBlock object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.19.0/pkg/reconcile
func (r *ServerBlockReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := log.FromContext(ctx).WithValues("serverblock", req.NamespacedName)

	server, err := r.fetchServerBlock(ctx, req)
	if err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	allLocations := make(map[string]*webv1alpha1.Location)

	for _, ref := range server.Spec.LocationRefs {
		var loc webv1alpha1.Location
		if err := r.Get(ctx, types.NamespacedName{Name: ref, Namespace: server.Namespace}, &loc); err != nil {
			if errors.IsNotFound(err) {
				allLocations[ref] = nil
			} else {
				return ctrl.Result{}, err
			}
		} else {
			allLocations[ref] = &loc
			metrics.SetCRDRefStatus(server.Namespace, server.Name, loc.Kind, ref, loc.Status.Ready)
		}
	}

	valid, problems := handler.ValidateLocationRefs(allLocations, server.Spec.LocationRefs)

	if !valid {
		msg := strings.Join(problems, " | ")
		r.Recorder.Eventf(server, corev1.EventTypeWarning, "InvalidRefs", msg)
		metrics.Recorder(server.Kind, server.Namespace, server.Name, corev1.EventTypeWarning, msg)
		_ = r.updateServerStatus(ctx, server, false, msg, log)
		return ctrl.Result{RequeueAfter: DefaultRequeue}, nil
	}

	conf := handler.GenerateServerBlockConfig(server)

	if err := r.createOrUpdateConfigMap(ctx, server, conf, log); err != nil {
		return ctrl.Result{}, err
	}

	_ = r.updateServerStatus(ctx, server, true, "", log)
	return ctrl.Result{RequeueAfter: DefaultRequeue}, nil
}

func (r *ServerBlockReconciler) updateServerStatus(ctx context.Context, srv *webv1alpha1.ServerBlock, ready bool, reason string, log logr.Logger) error {
	srv.Status.Ready = ready
	srv.Status.Version = fmt.Sprintf("%d", srv.Generation)
	srv.Status.Reason = reason
	isTriggerOpenResty := !utils.EqualSlices(srv.Spec.LocationRefs, srv.Status.LocationRef)
	srv.Status.LocationRef = srv.Spec.LocationRefs

	if err := r.Status().Update(ctx, srv); err != nil {
		if errors.IsConflict(err) {
			log.Info("ServerBlock status conflict, skipping update")
		} else {
			log.Error(err, "Failed to update ServerBlock status")
		}
	}

	if ready && isTriggerOpenResty {
		return r.updateOpenResty(ctx, srv)
	}

	return nil
}

func (r *ServerBlockReconciler) createOrUpdateConfigMap(ctx context.Context, sb *webv1alpha1.ServerBlock, content string, log logr.Logger) error {
	name := "serverblock-" + sb.Name
	dataName := sb.Name + ".conf"
	cm := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: sb.Namespace,
			Labels:    constants.BuildCommonLabels(sb, "configmap"),
		},
		Data: map[string]string{
			dataName: content,
		},
	}

	if err := ctrl.SetControllerReference(sb, cm, r.Scheme); err != nil {
		return err
	}

	var existing corev1.ConfigMap
	err := r.Get(ctx, types.NamespacedName{Name: name, Namespace: sb.Namespace}, &existing)
	if err != nil {
		if errors.IsNotFound(err) {
			log.Info("Creating ConfigMap", "name", name)
			return r.Create(ctx, cm)
		}
		return err
	}

	if existing.Data[dataName] != content {
		log.Info("Updating ConfigMap", "name", name)
		existing.Data[dataName] = content
		return r.Update(ctx, &existing)
	}

	return nil
}

func (r *ServerBlockReconciler) updateOpenResty(ctx context.Context, sb *webv1alpha1.ServerBlock) error {
	var appList webv1alpha1.OpenRestyList
	if err := r.List(ctx, &appList,
		client.MatchingFields{"spec.http.serverRefs": fmt.Sprintf("%s/%s", sb.Namespace, sb.Name)},
	); err != nil {
		return err
	}

	for _, app := range appList.Items {
		_ = r.triggerReconcile(ctx, &app)
	}

	return nil
}

func (r *ServerBlockReconciler) triggerReconcile(ctx context.Context, app *webv1alpha1.OpenResty) error {
	patched := app.DeepCopy()

	if patched.Annotations == nil {
		patched.Annotations = map[string]string{}
	}

	patched.Annotations[constants.AnnotationTriggerHash] = fmt.Sprintf("%d", time.Now().UnixNano())

	return r.Patch(ctx, patched, client.MergeFrom(app))
}

func (r *ServerBlockReconciler) fetchServerBlock(ctx context.Context, req ctrl.Request) (*webv1alpha1.ServerBlock, error) {
	var server webv1alpha1.ServerBlock
	if err := r.Get(ctx, req.NamespacedName, &server); err != nil {
		return nil, err
	}
	return &server, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *ServerBlockReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&webv1alpha1.ServerBlock{}).
		Owns(&corev1.ConfigMap{}).
		WithEventFilter(predicate.Funcs{
			UpdateFunc: func(e event.UpdateEvent) bool {
				return utils.IsSpecChanged(e.ObjectOld, e.ObjectNew)
			},
		}).
		Complete(r)
}
