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
	"encoding/json"
	"fmt"
	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	"k8s.io/utils/pointer"
	"openresty-operator/api/v1alpha1"
	"openresty-operator/internal/constants"
	"openresty-operator/internal/handler"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

// NormalizeRuleReconciler reconciles a NormalizeRule object
type NormalizeRuleReconciler struct {
	client.Client
	Scheme   *runtime.Scheme
	Recorder record.EventRecorder
}

// +kubebuilder:rbac:groups=openresty.huangzehong.me,resources=normalizerules,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=openresty.huangzehong.me,resources=normalizerules/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=openresty.huangzehong.me,resources=normalizerules/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the NormalizeRule object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.19.0/pkg/reconcile
func (r *NormalizeRuleReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	var normalizeRule v1alpha1.NormalizeRule
	if err := r.Get(ctx, req.NamespacedName, &normalizeRule); err != nil {
		// unable to fetch NormalizeRule - ignore not-found errors, otherwise log and return
		logger.Error(err, "unable to fetch NormalizeRule")
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	validateField := func(field apiextensionsv1.JSON, fieldName string) bool {
		var str string
		if err := json.Unmarshal(field.Raw, &str); err == nil {
			if len(str) == 0 {
				r.Recorder.Eventf(&normalizeRule, "Warning", "ValidationFailed", "Validation warning: field %s string value is empty", fieldName)
				return false
			}
			return true
		}

		var obj map[string]interface{}
		if err := json.Unmarshal(field.Raw, &obj); err == nil {
			luaVal, ok := obj["lua"]
			if !ok {
				r.Recorder.Eventf(&normalizeRule, "Warning", "ValidationFailed", "Validation warning: field %s object missing 'lua' key", fieldName)
				return false
			}
			if _, ok := luaVal.(string); !ok {
				r.Recorder.Eventf(&normalizeRule, "Warning", "ValidationFailed", "Validation warning: field %s 'lua' value is not a string", fieldName)
				return false
			}
			return true
		}

		r.Recorder.Eventf(&normalizeRule, "Warning", "ValidationFailed", "Validation warning: field %s has unsupported JSON type", fieldName)
		return false
	}

	valid := true
	if normalizeRule.Spec.Request != nil {
		for i, item := range normalizeRule.Spec.Request.Body {
			fieldName := fmt.Sprintf("spec.request[%s]", i)
			valid = valid && validateField(item, fieldName)
		}
	}

	for i, item := range normalizeRule.Spec.Response {
		fieldName := fmt.Sprintf("spec.response[%s]", i)
		valid = valid && validateField(item, fieldName)
	}

	if !controllerutil.ContainsFinalizer(&normalizeRule, constants.NormalizeRuleFinalizer) {
		controllerutil.AddFinalizer(&normalizeRule, constants.NormalizeRuleFinalizer)
		_ = r.Update(ctx, &normalizeRule)
	}

	if !normalizeRule.ObjectMeta.DeletionTimestamp.IsZero() {
		if controllerutil.ContainsFinalizer(&normalizeRule, constants.NormalizeRuleFinalizer) {
			handler.CreateOrUpdateConfigMap(ctx, r.Client, r.Scheme, &normalizeRule, normalizeRule.Namespace+"-normalize",
				normalizeRule.Namespace, constants.BuildCommonLabels(&normalizeRule, "configmap"), nil, logger, func(reference *metav1.OwnerReference) {
					reference.Controller = pointer.Bool(false)
					reference.BlockOwnerDeletion = pointer.Bool(true)
				}, []string{normalizeRule.Name + UpstreamRenderTypeLua})
			controllerutil.RemoveFinalizer(&normalizeRule, constants.NormalizeRuleFinalizer)
			if err := r.Update(ctx, &normalizeRule); err != nil {
				return ctrl.Result{}, err
			}
		}
		return ctrl.Result{}, nil
	}

	if valid {
		lua := handler.RenderNormalizeRuleLua(&normalizeRule, func(ns, name string) (*corev1.Secret, error) {
			s := corev1.Secret{}
			if err := r.Get(ctx, types.NamespacedName{Namespace: ns, Name: name}, &s); err != nil {
				return nil, err
			}
			return &s, nil
		})
		handler.CreateOrUpdateConfigMap(ctx, r.Client, r.Scheme, &normalizeRule, normalizeRule.Namespace+"-normalize",
			normalizeRule.Namespace, constants.BuildCommonLabels(&normalizeRule, "configmap"), map[string]string{
				normalizeRule.Name + UpstreamRenderTypeLua: lua,
			}, logger, func(reference *metav1.OwnerReference) {
				reference.Controller = pointer.Bool(false)
				reference.BlockOwnerDeletion = pointer.Bool(true)
			}, nil)
	}

	r.updateNormalizeRuleStatus(ctx, &normalizeRule, valid, "", logger)

	return ctrl.Result{}, nil
}

func (r *NormalizeRuleReconciler) fetchNormalizeRule(ctx context.Context, req ctrl.Request) (*v1alpha1.NormalizeRule, error) {
	var rule v1alpha1.NormalizeRule
	if err := r.Get(ctx, req.NamespacedName, &rule); err != nil {
		return nil, err
	}
	return &rule, nil
}

func (r *NormalizeRuleReconciler) updateNormalizeRuleStatus(
	ctx context.Context,
	current *v1alpha1.NormalizeRule,
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

// SetupWithManager sets up the controller with the Manager.
func (r *NormalizeRuleReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&v1alpha1.NormalizeRule{}).
		Complete(r)
}
