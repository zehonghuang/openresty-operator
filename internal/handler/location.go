package handler

import (
	"fmt"
	"net/url"
	"openresty-operator/api/v1alpha1"
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

func GenerateLocationConfig(entries []v1alpha1.LocationEntry) string {
	var b strings.Builder
	for _, e := range entries {
		b.WriteString(fmt.Sprintf("location %s {\n", e.Path))

		if e.ProxyPassIsFullURL {
			if e.Lua != nil && e.Lua.Content != "" {
				b.WriteString("    content_by_lua_block {\n")
				b.WriteString(e.Lua.Content)
				b.WriteString("    }\n")
			} else {
				b.WriteString("    set $target \"\";\n")
				b.WriteString(fmt.Sprintf("    set $location_prefix \"%s\";\n", e.Path))

				b.WriteString("    rewrite_by_lua_block {\n")
				b.WriteString(fmt.Sprintf("        require(\"upstreams.%s.%s\")()\n", safeName(e.ProxyPass), safeName(e.ProxyPass)))
				b.WriteString("    }\n")
			}
			b.WriteString("    proxy_pass $target;\n")
		} else if e.ProxyPass != "" {
			b.WriteString(fmt.Sprintf("    proxy_pass %s;\n", e.ProxyPass))
		}

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
			b.WriteString("        local addr = (ngx.var.upstream_addr or \"unknown\"):match(\"^[^,]+\")\n")
			b.WriteString("        local status = ngx.var.status\n")
			b.WriteString("        local latency = tonumber(ngx.var.upstream_response_time) or 0\n")
			b.WriteString("        metric_upstream_latency:observe(latency, {addr})\n")
			b.WriteString("        metric_upstream_total:inc(1, {addr, status})\n")
			b.WriteString("    }\n")
		}

		b.WriteString("}\n\n")
	}

	return b.String()
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
