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
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	webv1alpha1 "openresty-operator/api/v1alpha1"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"time"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

// RateLimitPolicyReconciler reconciles a RateLimitPolicy object
type RateLimitPolicyReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=web.chillyroom.com,resources=ratelimitpolicies,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=web.chillyroom.com,resources=ratelimitpolicies/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=web.chillyroom.com,resources=ratelimitpolicies/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the RateLimitPolicy object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.19.0/pkg/reconcile
func (r *RateLimitPolicyReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx).WithValues("ratelimitpolicy", req.NamespacedName)

	var policy webv1alpha1.RateLimitPolicy
	if err := r.Get(ctx, req.NamespacedName, &policy); err != nil {
		if errors.IsNotFound(err) {
			logger.Info("RateLimitPolicy not found")
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}

	conf := renderLimitReqZone(&policy)
	configMap := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "ratelimit-" + policy.Name,
			Namespace: policy.Namespace,
		},
		Data: map[string]string{
			policy.Name + ".conf": conf,
		},
	}

	if err := controllerutil.SetControllerReference(&policy, configMap, r.Scheme); err != nil {
		return ctrl.Result{}, err
	}

	var existing corev1.ConfigMap
	err := r.Get(ctx, types.NamespacedName{Name: configMap.Name, Namespace: configMap.Namespace}, &existing)
	if err != nil {
		if errors.IsNotFound(err) {
			logger.Info("Creating ConfigMap", "name", configMap.Name)
			if err := r.Create(ctx, configMap); err != nil {
				return ctrl.Result{}, err
			}
		} else {
			return ctrl.Result{}, err
		}
	} else {
		if existing.Data[policy.Name+".conf"] != conf {
			existing.Data = configMap.Data
			logger.Info("Updating ConfigMap", "name", configMap.Name)
			if err := r.Update(ctx, &existing); err != nil {
				return ctrl.Result{}, err
			}
		}
	}

	policy.Status.Ready = true
	policy.Status.Version = fmt.Sprintf("%d", policy.Generation)
	policy.Status.Reason = ""

	if err := r.Status().Update(ctx, &policy); err != nil {
		if errors.IsConflict(err) {
			logger.Info("Location status conflict, skipping update")
		} else {
			logger.Error(err, "Failed to update Location status")
		}
	}

	logger.Info("RateLimitPolicy reconciled successfully")
	return ctrl.Result{RequeueAfter: 60 * time.Second}, nil
}

func renderLimitReqZone(p *webv1alpha1.RateLimitPolicy) string {
	key := p.Spec.Key
	if key == "" {
		key = "$binary_remote_addr"
	}
	zoneSize := p.Spec.ZoneSize
	if zoneSize == "" {
		zoneSize = "10m"
	}
	return fmt.Sprintf("limit_req_zone %s zone=%s:%s rate=%s;", key, p.Spec.ZoneName, zoneSize, p.Spec.Rate)
}

// SetupWithManager sets up the controller with the Manager.
func (r *RateLimitPolicyReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&webv1alpha1.RateLimitPolicy{}).
		Owns(&corev1.ConfigMap{}).
		Complete(r)
}
