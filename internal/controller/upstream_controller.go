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
	"net"
	webv1alpha1 "openresty-operator/api/v1alpha1"
	"openresty-operator/internal/metrics"
	"openresty-operator/internal/utils"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"strings"
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

// +kubebuilder:rbac:groups=web.chillyroom.com,resources=upstreams,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=web.chillyroom.com,resources=upstreams/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=web.chillyroom.com,resources=upstreams/finalizers,verbs=update

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

	var upstream webv1alpha1.Upstream
	if err := r.Get(ctx, req.NamespacedName, &upstream); err != nil {
		if errors.IsNotFound(err) {
			log.Info("Upstream resource not found")
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}

	const maxConcurrentChecks = 10

	type result struct {
		Address string
		Config  string
		Alive   bool
	}

	var (
		wg      sync.WaitGroup
		sem     = make(chan struct{}, maxConcurrentChecks)
		results = make(chan result, len(upstream.Spec.Servers))
	)

	for _, addr := range upstream.Spec.Servers {
		wg.Add(1)
		sem <- struct{}{}

		go func(addr string) {
			defer wg.Done()
			defer func() { <-sem }()

			host, port, err := splitHostPort(addr)
			if err != nil {
				log.Error(err, "Invalid format", "server", addr)
				results <- result{Address: addr, Alive: false, Config: fmt.Sprintf("# server %s;  // invalid format", addr)}
				return
			}

			if _, err := net.LookupHost(host); err != nil {
				r.Recorder.Eventf(&upstream, corev1.EventTypeWarning, "DNSError", "Failed to resolve host %s: %v", host, err)
				results <- result{Address: addr, Alive: false, Config: fmt.Sprintf("# server %s;  // DNS error", addr)}
				return
			}

			alive, err := testTCP(host, port)
			if alive {
				results <- result{Address: addr, Alive: true, Config: fmt.Sprintf("server %s;", addr)}
			} else {
				reason := "dead"
				if err != nil {
					reason = err.Error()
				}
				results <- result{Address: addr, Alive: false, Config: fmt.Sprintf("# server %s;  // %s", addr, reason)}
			}
		}(addr)
	}

	wg.Wait()
	close(results)

	var (
		configLines []string
		statusList  []webv1alpha1.UpstreamServerStatus
	)

	for r := range results {
		configLines = append(configLines, r.Config)
		statusList = append(statusList, webv1alpha1.UpstreamServerStatus{
			Address: r.Address,
			Alive:   r.Alive,
		})
		metrics.SetUpstreamDNSResolvable(upstream.Namespace, upstream.Name, r.Address, r.Alive)
	}

	// 写入 ConfigMap
	nginxConfig := renderNginxUpstreamBlock(utils.SanitizeName(upstream.Name), configLines)
	allDown := false
	if len(nginxConfig) > 0 {
		if err := r.createOrUpdateConfigMap(ctx, &upstream, nginxConfig, log); err != nil {
			log.Error(err, "Failed to update ConfigMap")
			return ctrl.Result{}, err
		}
	} else {
		allDown = true
	}

	// 更新 Status
	if allDown {
		r.updateLocationStatus(ctx, upstream, false, nginxConfig, statusList, "All servers unavailable or DNS failed", log)

	} else {
		r.updateLocationStatus(ctx, upstream, true, nginxConfig, statusList, "", log)
	}

	return reconcile.Result{RequeueAfter: 30 * time.Second}, nil
}

func (r *UpstreamReconciler) createOrUpdateConfigMap(ctx context.Context, upstream *webv1alpha1.Upstream, config string, log logr.Logger) error {
	name := "upstream-" + upstream.Name
	dataName := upstream.Name + ".conf"
	cm := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: upstream.Namespace,
			Annotations: map[string]string{
				"web.chillyroom.com/generated-from-generation": fmt.Sprintf("%d", upstream.GetGeneration()),
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

	if existing.Data[dataName] != config {
		log.Info("Updating ConfigMap", "name", name)
		existing.Data[dataName] = config
		existing.Annotations = map[string]string{
			"web.chillyroom.com/generated-from-generation": fmt.Sprintf("%d", upstream.GetGeneration()),
		}
		return r.Update(ctx, &existing)
	}

	return nil
}

func renderNginxUpstreamBlock(name string, lines []string) string {
	if len(lines) == 0 {
		return ""
	}
	var b strings.Builder
	b.WriteString(fmt.Sprintf("upstream %s {\n", name))
	for _, line := range lines {
		b.WriteString("    " + line + "\n")
	}
	b.WriteString("}\n")
	return b.String()
}

func splitHostPort(input string) (string, string, error) {
	parts := strings.Split(input, ":")
	if len(parts) != 2 {
		return "", "", fmt.Errorf("invalid server format: %s", input)
	}
	return parts[0], parts[1], nil
}

func testTCP(ip, port string) (bool, error) {
	conn, err := net.DialTimeout("tcp", net.JoinHostPort(ip, port), 1*time.Second)
	if err != nil {
		return false, err
	}
	conn.Close()
	return true, nil
}

func (r *UpstreamReconciler) updateLocationStatus(
	ctx context.Context,
	current webv1alpha1.Upstream,
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

	if err := r.Status().Update(ctx, &current); err != nil {
		if errors.IsConflict(err) {
			log.Info("Location status conflict, skipping update")
		} else {
			log.Error(err, "Failed to update Location status")
		}
	}
}

// SetupWithManager sets up the controller with the Manager.
func (r *UpstreamReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		// Uncomment the following line adding a pointer to an instance of the controlled resource as an argument
		For(&webv1alpha1.Upstream{}).
		Owns(&corev1.ConfigMap{}).
		Complete(r)
}
