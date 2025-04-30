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
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	"openresty-operator/internal/handler"
	"openresty-operator/internal/metrics"
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

const DefaultRequeue = 30 * time.Second

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

	location, err := r.fetchLocation(ctx, req)
	if err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	valid, problems := handler.ValidateLocationEntries(location.Spec.Entries)
	if !valid {
		msg := strings.Join(problems, " | ")
		log.Error(nil, "Path validation failed", "details", msg)
		r.Recorder.Eventf(location, corev1.EventTypeWarning, "InvalidPath", msg)
		metrics.Recorder(location.Kind, location.Namespace, location.Name, corev1.EventTypeWarning, msg)

		r.updateLocationStatus(ctx, location, false, msg, log)
		return ctrl.Result{RequeueAfter: DefaultRequeue}, nil

	}

	conf := handler.GenerateLocationConfig(location.Spec.Entries)

	if err := r.createOrUpdateConfigMap(ctx, location, conf, log); err != nil {
		return ctrl.Result{}, err
	}

	r.updateLocationStatus(ctx, location, true, "", log)

	return ctrl.Result{RequeueAfter: DefaultRequeue}, nil

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

func (r *LocationReconciler) updateLocationStatus(
	ctx context.Context,
	current *webv1alpha1.Location,
	ready bool,
	reason string,
	log logr.Logger,
) {
	current.Status.Ready = ready
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

func (r *LocationReconciler) fetchLocation(ctx context.Context, req ctrl.Request) (*webv1alpha1.Location, error) {
	var location webv1alpha1.Location
	if err := r.Get(ctx, req.NamespacedName, &location); err != nil {
		return nil, err
	}
	return &location, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *LocationReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&webv1alpha1.Location{}).
		Owns(&corev1.ConfigMap{}).
		Owns(&monitoringv1.ServiceMonitor{}).
		Complete(r)
}
