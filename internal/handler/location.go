package handler

import (
	"context"
	"encoding/json"
	"fmt"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"net/url"
	"openresty-operator/api/v1alpha1"
	"openresty-operator/internal/constants"
	"openresty-operator/internal/utils"
	"regexp"
	"strings"
)

func ValidateLocationEntries(entries []v1alpha1.LocationEntry) (bool, []string) {
	pathSeen := make(map[string]struct{})
	var problems []string

	for _, entry := range entries {
		path := entry.Path

		// 校验路径是否合法
		valid, reason := utils.ValidateLocationPath(path)
		if !valid {
			problems = append(problems, fmt.Sprintf("Invalid path: %s (%s)", path, reason))
		}

		// 检查重复路径
		if _, exists := pathSeen[path]; exists {
			problems = append(problems, fmt.Sprintf("Duplicate path: %s", path))
		} else {
			pathSeen[path] = struct{}{}
		}
	}

	return len(problems) == 0, problems
}

func GenerateLocationConfig(name, namespace string, entries []v1alpha1.LocationEntry) string {
	var b strings.Builder
	for _, e := range entries {
		b.WriteString(fmt.Sprintf("location %s {\n", e.Path))

		needRewrite := e.ProxyPassIsFullURL || len(e.HeadersFromSecret) > 0
		b.WriteString(fmt.Sprintf("    set $location_path \"%s\";\n", e.Path))
		if needRewrite {
			if e.ProxyPassIsFullURL {
				b.WriteString("    set $target \"\";\n")
			}
			b.WriteString(fmt.Sprintf("    set $location_prefix \"%s\";\n", e.Path))
			b.WriteString("    rewrite_by_lua_block {\n")

			if len(e.HeadersFromSecret) > 0 {
				b.WriteString(fmt.Sprintf("        local namespace = ngx.var.namespace or \"%s\"\n", namespace))
				b.WriteString(fmt.Sprintf("        local locationName = \"%s\"\n", name))
				b.WriteString(fmt.Sprintf("        local path = \"%s\"\n", e.Path))
				b.WriteString("        local headers = {\n")
				for _, h := range e.HeadersFromSecret {
					b.WriteString(fmt.Sprintf("            \"%s\",\n", h.Name))
				}
				b.WriteString("        }\n")
				b.WriteString("        for _, headerName in ipairs(headers) do\n")
				b.WriteString("            local key = namespace .. \"/\" .. locationName .. \"/\" .. path .. \"/\" .. headerName\n")
				b.WriteString("            local value = ngx.shared.secrets_store:get(key)\n")
				b.WriteString("            if value then\n")
				b.WriteString("                ngx.req.set_header(headerName, value)\n")
				b.WriteString("            end\n")
				b.WriteString("        end\n")
			}

			// FullURL upstream动态分流
			if e.ProxyPassIsFullURL {
				b.WriteString(fmt.Sprintf("        require(\"upstreams.%s.%s\").default()\n", safeName(e.ProxyPass), safeName(e.ProxyPass)))
			}

			b.WriteString("    }\n")
		}
		if e.ProxyPassIsFullURL {
			b.WriteString("    header_filter_by_lua_block\n {\n")
			b.WriteString("      ngx.header[\"Content-Length\"] = nil")
			b.WriteString("    }\n")

			b.WriteString("    body_filter_by_lua_block {\n")
			b.WriteString(fmt.Sprintf("        require(\"upstreams.%s.%s\").normalizeResponse()\n", safeName(e.ProxyPass), safeName(e.ProxyPass)))
			b.WriteString("    }\n")
		}

		if e.ProxyPassIsFullURL {
			b.WriteString("    proxy_pass $target;\n")
		} else if e.ProxyPass != "" {
			b.WriteString(fmt.Sprintf("    proxy_pass %s;\n", e.ProxyPass))
		}

		// 明文 Headers
		for _, h := range e.Headers {
			b.WriteString(fmt.Sprintf("    proxy_set_header %s %s;\n", h.Key, h.Value))
		}

		if e.Timeout != nil {
			if e.Timeout.Connect != "" {
				b.WriteString(fmt.Sprintf("    proxy_connect_timeout %s;\n", e.Timeout.Connect))
			}
			if e.Timeout.Send != "" {
				b.WriteString(fmt.Sprintf("    proxy_send_timeout %s;\n", e.Timeout.Send))
			}
			if e.Timeout.Read != "" {
				b.WriteString(fmt.Sprintf("    proxy_read_timeout %s;\n", e.Timeout.Read))
			}
		}

		if e.AccessLog != nil && !*e.AccessLog {
			b.WriteString("    access_log off;\n")
		}

		if e.LimitReq != nil {
			b.WriteString(fmt.Sprintf("    limit_req %s;\n", *e.LimitReq))
		}

		if e.Gzip != nil && e.Gzip.Enable {
			b.WriteString("    gzip on;\n")
			if len(e.Gzip.Types) > 0 {
				b.WriteString(fmt.Sprintf("    gzip_types %s;\n", strings.Join(e.Gzip.Types, " ")))
			}
		}

		if e.Cache != nil {
			if e.Cache.Zone != "" {
				b.WriteString(fmt.Sprintf("    proxy_cache %s;\n", e.Cache.Zone))
			}
			if e.Cache.Valid != "" {
				b.WriteString(fmt.Sprintf("    proxy_cache_valid %s;\n", e.Cache.Valid))
			}
		}

		if e.Lua != nil && e.Lua.Access != "" {
			b.WriteString("    access_by_lua_block {\n")
			b.WriteString(indentLua(e.Lua.Access, "        "))
			b.WriteString("    }\n")
		}

		for _, extra := range e.Extra {
			b.WriteString(fmt.Sprintf("    %s\n", extra))
		}

		if e.EnableUpstreamMetrics {
			b.WriteString("    log_by_lua_block {\n")
			//b.WriteString("        local addr = (ngx.var.upstream_addr or \"unknown\"):match(\"^[^,]+\")\n")
			//b.WriteString("        local status = ngx.var.status\n")
			//b.WriteString("        local latency = tonumber(ngx.var.upstream_response_time) or 0\n")
			//b.WriteString("        metric_upstream_latency:observe(latency, {addr})\n")
			//b.WriteString("        metric_upstream_total:inc(1, {addr, status})\n")
			b.WriteString("        require(\"metrics\").record()\n")
			b.WriteString("    }\n")
		}

		b.WriteString("}\n\n")
	}
	return b.String()
}

func GenerateSecretFromLocations(ctx context.Context, location *v1alpha1.Location, getSecretFunc func(ns, name string) (*corev1.Secret, error)) (*corev1.Secret, error) {
	data := make(map[string]string)

	for _, entry := range location.Spec.Entries {
		if len(entry.HeadersFromSecret) == 0 {
			continue
		}

		for _, h := range entry.HeadersFromSecret {
			secret, err := getSecretFunc(location.Namespace, h.SecretName)
			if err != nil {
				return nil, fmt.Errorf("failed to get secret %s/%s: %w", location.Namespace, h.SecretName, err)
			}

			val, ok := secret.Data[h.SecretKey]
			if !ok {
				return nil, fmt.Errorf("key %s not found in secret %s/%s", h.SecretKey, location.Namespace, h.SecretName)
			}

			key := fmt.Sprintf("%s/%s/%s/%s", location.Namespace, location.Name, entry.Path, h.Name)
			data[key] = string(val)
		}
	}

	if len(data) == 0 {
		return nil, nil
	}

	jsonBytes, err := json.Marshal(data)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal headers JSON: %w", err)
	}

	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("secret-headers-%s", location.Name),
			Namespace: location.Namespace,
			Labels:    constants.BuildCommonLabels(location, "secret"),
			Annotations: map[string]string{
				constants.AnnotationSecretHeaders: fmt.Sprintf("%s", location.Name),
			},
		},
		Type: corev1.SecretTypeOpaque,
		Data: map[string][]byte{
			"keys.json": jsonBytes,
		},
	}

	return secret, nil
}

func safeName(proxyPass string) string {
	u, err := url.Parse(proxyPass)
	if err != nil || u.Host == "" {
		return "invalid-proxypass"
	}

	host := u.Host

	// 将 host 中的非法字符替换成 "-"
	host = strings.ToLower(host)
	host = strings.ReplaceAll(host, ".", "-")
	host = strings.ReplaceAll(host, ":", "-")

	// 确保只包含合法字符
	reg := regexp.MustCompile(`[^a-z0-9\-]`)
	host = reg.ReplaceAllString(host, "")

	// 最长限制 63 字符（K8s 对象名规范）
	if len(host) > 63 {
		host = host[:63]
	}

	return host
}

func indentLua(code, prefix string) string {
	lines := strings.Split(code, "\n")
	for i, line := range lines {
		lines[i] = prefix + line
	}
	return strings.Join(lines, "\n") + "\n"
}
