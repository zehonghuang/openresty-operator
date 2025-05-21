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
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/client-go/tools/record"
	"openresty-operator/api/v1alpha1"

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

	validateField := func(field apiextensionsv1.JSON, fieldName string) {
		// 尝试解析成 string（即 JSONPath）
		var str string
		if err := json.Unmarshal(field.Raw, &str); err == nil {
			if len(str) == 0 || str[0] != '$' {
				r.Recorder.Eventf(&normalizeRule, "Warning", "ValidationFailed", "Validation warning: field %s string value does not start with '$'", fieldName)
			}
			return
		}

		// 尝试解析为 map[string]interface{} 并检测是否包含 lua
		var obj map[string]interface{}
		if err := json.Unmarshal(field.Raw, &obj); err == nil {
			luaVal, ok := obj["lua"]
			if !ok {
				r.Recorder.Eventf(&normalizeRule, "Warning", "ValidationFailed", "Validation warning: field %s object missing 'lua' key", fieldName)
				return
			}
			if _, ok := luaVal.(string); !ok {
				r.Recorder.Eventf(&normalizeRule, "Warning", "ValidationFailed", "Validation warning: field %s 'lua' value is not a string", fieldName)
			}
			return
		}

		r.Recorder.Eventf(&normalizeRule, "Warning", "ValidationFailed", "Validation warning: field %s has unsupported JSON type", fieldName)
	}

	for i, item := range normalizeRule.Spec.Request {
		fieldName := fmt.Sprintf("spec.request[%s]", i)
		validateField(item, fieldName)
	}

	for i, item := range normalizeRule.Spec.Response {
		fieldName := fmt.Sprintf("spec.response[%s]", i)
		validateField(item, fieldName)
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *NormalizeRuleReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&v1alpha1.NormalizeRule{}).
		Complete(r)
}
