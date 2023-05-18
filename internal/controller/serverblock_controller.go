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
	"k8s.io/client-go/util/retry"
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

// +kubebuilder:rbac:groups=web.chillyroom.com,resources=serverblocks,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=web.chillyroom.com,resources=serverblocks/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=web.chillyroom.com,resources=serverblocks/finalizers,verbs=update

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

	var server webv1alpha1.ServerBlock
	if err := r.Get(ctx, req.NamespacedName, &server); err != nil {
		if errors.IsNotFound(err) {
			log.Info("ServerBlock not found")
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}

	var missing []string
	var notReady []string
	for _, ref := range server.Spec.LocationRefs {
		var loc webv1alpha1.Location
		err := r.Get(ctx, types.NamespacedName{Name: ref, Namespace: server.Namespace}, &loc)
		if err != nil {
			if errors.IsNotFound(err) {
				missing = append(missing, ref)
			} else {
				return ctrl.Result{}, err
			}
		} else if !loc.Status.Ready {
			notReady = append(notReady, ref)
		}
	}

	if len(missing)+len(notReady) > 0 {
		msg := fmt.Sprintf("Missing/NotReady Locations: %v %v", missing, notReady)
		r.Recorder.Eventf(&server, corev1.EventTypeWarning, "InvalidRefs", msg)

		_ = r.updateServerStatus(ctx, req.NamespacedName, false, msg)
		return ctrl.Result{RequeueAfter: 30 * time.Second}, nil
	}

	conf := renderServerBlock(&server)
	if err := r.createOrUpdateConfigMap(ctx, &server, conf, log); err != nil {
		return ctrl.Result{}, err
	}

	log.Info("ServerBlock config updated successfully", "name", server.Name)
	r.Recorder.Eventf(&server, corev1.EventTypeNormal, "ConfigUpdated", "ConfigMap rendered and updated")

	_ = r.updateServerStatus(ctx, req.NamespacedName, true, "")
	return ctrl.Result{RequeueAfter: 30 * time.Second}, nil
}

func (r *ServerBlockReconciler) updateServerStatus(ctx context.Context, name types.NamespacedName, ready bool, reason string) error {
	return retry.RetryOnConflict(retry.DefaultBackoff, func() error {
		var latest webv1alpha1.ServerBlock
		if err := r.Get(ctx, name, &latest); err != nil {
			return err
		}
		latest = *latest.DeepCopy()
		latest.Status.Ready = ready
		latest.Status.Reason = reason
		return r.Status().Update(ctx, &latest)
	})
}

func renderServerBlock(s *webv1alpha1.ServerBlock) string {
	var b strings.Builder

	b.WriteString("server {\n")
	b.WriteString(fmt.Sprintf("    listen %s;\n", s.Spec.Listen))

	serverName := fmt.Sprintf("%s.%s.svc.cluster.local", s.Name, s.Namespace)
	b.WriteString(fmt.Sprintf("    server_name %s;\n", serverName))

	for _, ref := range s.Spec.LocationRefs {
		includePath := fmt.Sprintf("/etc/nginx/locations/%s.conf", ref)
		b.WriteString(fmt.Sprintf("    include %s;\n", includePath))
	}

	for _, h := range s.Spec.Headers {
		b.WriteString(fmt.Sprintf("    add_header %s %s;\n", h.Key, h.Value))
	}

	for _, line := range s.Spec.Extra {
		b.WriteString("    " + line + "\n")
	}

	b.WriteString("}\n")
	return b.String()
}

func (r *ServerBlockReconciler) createOrUpdateConfigMap(ctx context.Context, sb *webv1alpha1.ServerBlock, content string, log logr.Logger) error {
	name := "serverblock-" + sb.Name
	dataName := sb.Name + ".conf"
	cm := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: sb.Namespace,
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

// SetupWithManager sets up the controller with the Manager.
func (r *ServerBlockReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&webv1alpha1.ServerBlock{}).
		Owns(&corev1.ConfigMap{}).
		Complete(r)
}
