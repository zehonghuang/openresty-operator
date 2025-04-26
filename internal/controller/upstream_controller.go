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
	"net/url"
	webv1alpha1 "openresty-operator/api/v1alpha1"
	"openresty-operator/internal/metrics"
	"openresty-operator/internal/utils"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sort"
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

type serverResult struct {
	Address string
	Config  string
	Alive   bool
	Index   int
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

	var upstream webv1alpha1.Upstream
	if err := r.Get(ctx, req.NamespacedName, &upstream); err != nil {
		if errors.IsNotFound(err) {
			log.Info("Upstream resource not found")
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}

	const maxConcurrentChecks = 10

	var (
		wg          sync.WaitGroup
		sem         = make(chan struct{}, maxConcurrentChecks)
		resultsChan = make(chan serverResult, len(upstream.Spec.Servers))
	)

	for i, addr := range upstream.Spec.Servers {
		wg.Add(1)
		sem <- struct{}{}

		go func(addr string, i int) {
			defer wg.Done()
			defer func() { <-sem }()

			host, port, err := splitHostPort(addr)
			if err != nil {
				log.Error(err, "Invalid format", "server", addr)
				resultsChan <- serverResult{Index: i, Address: addr, Alive: false, Config: fmt.Sprintf("# server %s;  // invalid format", addr)}
				return
			}

			ips, err := net.LookupHost(host)
			if err != nil {
				r.Recorder.Eventf(&upstream, corev1.EventTypeWarning, "DNSError", "Failed to resolve host %s: %v", host, err)
				resultsChan <- serverResult{Index: i, Address: addr, Alive: false, Config: fmt.Sprintf("# server %s;  // DNS error", addr)}
				metrics.Recorder(upstream.Kind, upstream.Namespace, upstream.Name, corev1.EventTypeWarning, fmt.Sprintf("# server %s;  // DNS error", addr))
				return
			} else {
				DnsCache.Lock()
				DnsCache.Data[host] = ips
				DnsCache.Unlock()
			}

			alive, err := testTCP(host, port)
			if alive {
				resultsChan <- serverResult{Index: i, Address: addr, Alive: true, Config: fmt.Sprintf("server %s;", addr)}
			} else {
				reason := "dead"
				if err != nil {
					reason = err.Error()
				}
				r.Recorder.Eventf(
					&upstream,
					corev1.EventTypeWarning,
					"ConnectionError",
					"TCP test failed for host %s: %v",
					host, err,
				)
				resultsChan <- serverResult{Index: i, Address: addr, Alive: false, Config: fmt.Sprintf("# server %s;  // %s", addr, reason)}
			}
		}(addr, i)
	}

	wg.Wait()
	close(resultsChan)

	results := utils.DrainChan(resultsChan)
	sort.Slice(results, func(i, j int) bool {
		return results[i].Index < results[j].Index
	})

	var (
		configLines []string
		statusList  []webv1alpha1.UpstreamServerStatus
	)

	for _, r := range results {
		configLines = append(configLines, r.Config)
		statusList = append(statusList, webv1alpha1.UpstreamServerStatus{
			Address: r.Address,
			Alive:   r.Alive,
		})
		metrics.SetUpstreamDNSResolvable(upstream.Namespace, upstream.Name, r.Address, "ALL", r.Alive)
		host, _, _ := splitHostPort(r.Address)
		for _, ip := range DnsCache.Data[host] {
			metrics.SetUpstreamDNSResolvable(upstream.Namespace, upstream.Name, r.Address, ip, r.Alive)
		}
	}

	nginxConfig := ""
	if upstream.Spec.Type == webv1alpha1.UpstreamTypeAddress {
		nginxConfig = renderNginxUpstreamBlock(utils.SanitizeName(upstream.Name), configLines)
	}
	if upstream.Spec.Type == webv1alpha1.UpstreamTypeFullURL {
		nginxConfig = renderNginxUpstreamLua(utils.SanitizeName(upstream.Name), statusList)
	}
	// 写入 ConfigMap
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
	dataName := upstream.Name + UpstreamRenderTypeMap[upstream.Spec.Type]
	cm := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: upstream.Namespace,
			Annotations: map[string]string{
				"openresty.huangzehong.me/generated-from-generation": fmt.Sprintf("%d", upstream.GetGeneration()),
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
			"openresty.huangzehong.me/generated-from-generation": fmt.Sprintf("%d", upstream.GetGeneration()),
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

func renderNginxUpstreamLua(upstreamName string, rs []webv1alpha1.UpstreamServerStatus) string {
	var b strings.Builder

	b.WriteString(fmt.Sprintf("-- upstream-%s.lua\n", upstreamName))
	b.WriteString("local random = require(\"upstreams.random_weighted\")\n\n")

	b.WriteString("local servers = {\n")
	alives := 0
	for _, s := range rs {
		if s.Alive {
			b.WriteString(fmt.Sprintf("    { address = \"%s\", weight = %d },\n", s.Address, 1)) // 可扩展 weight 字段
			alives++
		} else {
			b.WriteString(fmt.Sprintf("--    { address = \"%s\", weight = %d },\n", s.Address, 1))
		}
	}
	if alives == 0 {
		return ""
	}
	b.WriteString("}\n\n")
	b.WriteString("random.init(servers)\n\n")

	b.WriteString("return function()\n")
	b.WriteString("    local picked = random.pick()\n")
	b.WriteString("    local uri = ngx.var.uri or \"/\"\n")
	b.WriteString("    local prefix = ngx.var.location_prefix or \"/\"\n\n")

	b.WriteString("    if prefix:sub(1,1) == \"^\" then\n")
	b.WriteString("        prefix = prefix:sub(2)\n")
	b.WriteString("    end\n\n")

	b.WriteString("    local from, to = ngx.re.find(uri, \"^\" .. prefix, \"jo\")\n")
	b.WriteString("    if from == 1 and to then\n")
	b.WriteString("        uri = \"/\" .. uri:sub(to + 1)\n")
	b.WriteString("    end\n\n")

	b.WriteString("    if picked:sub(-1) == \"/\" and uri:sub(1,1) == \"/\" then\n")
	b.WriteString("        picked = picked:sub(1, -2)\n")
	b.WriteString("    end\n\n")

	b.WriteString("    ngx.var.target = picked .. uri\n")
	b.WriteString("end\n")

	return b.String()
}

func splitHostPort(input string) (string, string, error) {
	// 处理带 http/https schema 的 URL
	if strings.HasPrefix(input, "http://") || strings.HasPrefix(input, "https://") {
		u, err := url.Parse(input)
		if err != nil {
			return "", "", fmt.Errorf("invalid URL: %v", err)
		}

		host := u.Hostname()
		port := u.Port()
		if port == "" {
			if u.Scheme == "http" {
				port = "80"
			} else if u.Scheme == "https" {
				port = "443"
			}
		}
		return host, port, nil
	}

	// 处理 host:port 的格式
	if strings.Contains(input, ":") {
		host, port, err := net.SplitHostPort(input)
		if err == nil {
			return host, port, nil
		}
		// 可能是域名中带冒号但格式不合法，比如 IPv6 缺 []
		return "", "", fmt.Errorf("invalid host:port format: %v", err)
	}

	// fallback，只有 host 没有端口
	return input, "80", nil
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
		WithEventFilter(predicate.Funcs{
			DeleteFunc: func(e event.DeleteEvent) bool {
				if obj, ok := e.Object.(*webv1alpha1.Upstream); ok {
					for _, server := range obj.Spec.Servers {
						host, _, _ := splitHostPort(server)
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
