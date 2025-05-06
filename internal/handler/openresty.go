package handler

import (
	"bytes"
	"context"
	"fmt"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	webv1alpha1 "openresty-operator/api/v1alpha1"
	"openresty-operator/internal/runtime/metrics"
	"openresty-operator/internal/template"
	"openresty-operator/internal/utils"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"strings"
	text_template "text/template"
)

type GetFunc func(ctx context.Context, key types.NamespacedName, obj client.Object, opts ...client.GetOption) error

type ServerRefsStatus struct {
	AllReady         bool
	MissingServers   []string
	NotReadyServers  []string
	MissingServerCMs []string
}

type UpstreamRefsStatus struct {
	AllReady           bool
	MissingUpstreams   []string
	NotReadyUpstreams  []string
	MissingUpstreamCMs []string
	UpstreamsType      map[string]webv1alpha1.UpstreamType
}

func ValidateServerRefs(get GetFunc, app *webv1alpha1.OpenResty) ServerRefsStatus {
	ctx := context.Background()
	status := ServerRefsStatus{AllReady: true}

	for _, name := range app.Spec.Http.ServerRefs {
		var srv webv1alpha1.ServerBlock
		if err := get(ctx, types.NamespacedName{Name: name, Namespace: app.Namespace}, &srv); err != nil {
			if errors.IsNotFound(err) {
				status.MissingServers = append(status.MissingServers, name)
			} else {
				status.MissingServers = append(status.MissingServers, fmt.Sprintf("%s (error: %v)", name, err))
			}
			status.AllReady = false
			continue
		}

		metrics.SetCRDRefStatus(app.Namespace, app.Name, srv.Kind, srv.Name, srv.Status.Ready)

		if !srv.Status.Ready {
			status.NotReadyServers = append(status.NotReadyServers, name)
			status.AllReady = false
			continue
		}

		var cm corev1.ConfigMap
		cmName := "serverblock-" + name
		if err := get(ctx, types.NamespacedName{Name: cmName, Namespace: app.Namespace}, &cm); err != nil {
			if errors.IsNotFound(err) {
				status.MissingServerCMs = append(status.MissingServerCMs, cmName)
			} else {
				status.MissingServerCMs = append(status.MissingServerCMs, fmt.Sprintf("%s (error: %v)", cmName, err))
			}
			status.AllReady = false
		}
	}

	return status
}

func ValidateUpstreamRefs(get GetFunc, app *webv1alpha1.OpenResty) UpstreamRefsStatus {
	ctx := context.Background()
	status := UpstreamRefsStatus{
		AllReady:      true,
		UpstreamsType: make(map[string]webv1alpha1.UpstreamType),
	}

	for _, name := range app.Spec.Http.UpstreamRefs {
		var ups webv1alpha1.Upstream
		if err := get(ctx, types.NamespacedName{Name: name, Namespace: app.Namespace}, &ups); err != nil {
			if errors.IsNotFound(err) {
				status.MissingUpstreams = append(status.MissingUpstreams, name)
			} else {
				status.MissingUpstreams = append(status.MissingUpstreams, fmt.Sprintf("%s (error: %v)", name, err))
			}
			status.AllReady = false
			continue
		}

		status.UpstreamsType[name] = ups.Spec.Type
		metrics.SetCRDRefStatus(app.Namespace, app.Name, ups.Kind, ups.Name, ups.Status.Ready)

		if !ups.Status.Ready {
			status.NotReadyUpstreams = append(status.NotReadyUpstreams, name)
			status.AllReady = false
			continue
		}

		var cm corev1.ConfigMap
		cmName := "upstream-" + name
		if err := get(ctx, types.NamespacedName{Name: cmName, Namespace: app.Namespace}, &cm); err != nil {
			if errors.IsNotFound(err) {
				status.MissingUpstreamCMs = append(status.MissingUpstreamCMs, cmName)
			} else {
				status.MissingUpstreamCMs = append(status.MissingUpstreamCMs, fmt.Sprintf("%s (error: %v)", cmName, err))
			}
			status.AllReady = false
		}
	}

	return status
}

func ComposeDependencyFailureReason(serverStatus ServerRefsStatus, upstreamStatus UpstreamRefsStatus) string {
	var parts []string

	if len(serverStatus.MissingServers) > 0 {
		parts = append(parts, fmt.Sprintf("Missing Servers: %s", strings.Join(serverStatus.MissingServers, ", ")))
	}
	if len(serverStatus.NotReadyServers) > 0 {
		parts = append(parts, fmt.Sprintf("NotReady Servers: %s", strings.Join(serverStatus.NotReadyServers, ", ")))
	}
	if len(serverStatus.MissingServerCMs) > 0 {
		parts = append(parts, fmt.Sprintf("Missing Server ConfigMaps: %s", strings.Join(serverStatus.MissingServerCMs, ", ")))
	}

	if len(upstreamStatus.MissingUpstreams) > 0 {
		parts = append(parts, fmt.Sprintf("Missing Upstreams: %s", strings.Join(upstreamStatus.MissingUpstreams, ", ")))
	}
	if len(upstreamStatus.NotReadyUpstreams) > 0 {
		parts = append(parts, fmt.Sprintf("NotReady Upstreams: %s", strings.Join(upstreamStatus.NotReadyUpstreams, ", ")))
	}
	if len(upstreamStatus.MissingUpstreamCMs) > 0 {
		parts = append(parts, fmt.Sprintf("Missing Upstream ConfigMaps: %s", strings.Join(upstreamStatus.MissingUpstreamCMs, ", ")))
	}

	if len(parts) == 0 {
		return "Unknown dependency error"
	}
	return strings.Join(parts, " | ")
}

func BuildIncludeLines(app *webv1alpha1.OpenResty, upstreamStatus UpstreamRefsStatus) []string {
	var lines []string

	for _, name := range app.Spec.Http.ServerRefs {
		line := fmt.Sprintf("include %s/%s/%s.conf;", utils.NginxServerConfigDir, name, name)
		lines = append(lines, line)
	}

	for _, name := range app.Spec.Http.UpstreamRefs {
		if upstreamStatus.UpstreamsType[name] == webv1alpha1.UpstreamTypeAddress {
			line := fmt.Sprintf("include %s/%s/%s.conf;", utils.NginxUpstreamConfigDir, name, name)
			lines = append(lines, line)
		}
	}

	return lines
}

type nginxConfData struct {
	InitLua           string
	EnableMetrics     bool
	MetricsPort       string
	MetricsPath       string
	Includes          []string
	LogFormat         string
	AccessLog         string
	ErrorLog          string
	ClientMaxBodySize string
	Gzip              bool
	Extra             []string
	IncludeSnippets   []string
}

func RenderNginxConf(http *webv1alpha1.HttpBlock, metrics *webv1alpha1.MetricsServer, includeLines []string) string {
	data := nginxConfData{
		InitLua:           template.DefaultInitLua,
		EnableMetrics:     metrics != nil && metrics.Enable,
		MetricsPort:       defaultOr(metrics.Listen, "9091"),
		MetricsPath:       defaultOr(metrics.Path, "/metrics"),
		Includes:          http.Include,
		LogFormat:         utils.SanitizeLogFormat(http.LogFormat),
		AccessLog:         http.AccessLog,
		ErrorLog:          http.ErrorLog,
		ClientMaxBodySize: http.ClientMaxBodySize,
		Gzip:              http.Gzip,
		Extra:             http.Extra,
		IncludeSnippets:   includeLines,
	}

	tmpl := text_template.Must(text_template.New("nginx").Funcs(text_template.FuncMap{
		"indent": func(s string, spaces int) string {
			pad := strings.Repeat(" ", spaces)
			lines := strings.Split(s, "\n")
			for i := range lines {
				lines[i] = pad + lines[i]
			}
			return strings.Join(lines, "\n")
		},
	}).Parse(utils.NginxTemplate))

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return fmt.Sprintf("# failed to render: %v", err)
	}

	return buf.String()
}

func defaultOr(s, fallback string) string {
	if s == "" {
		return fallback
	}
	return s
}
