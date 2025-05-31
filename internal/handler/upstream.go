package handler

import (
	"context"
	"fmt"
	webv1alpha1 "openresty-operator/api/v1alpha1"
	"openresty-operator/internal/runtime/health"
	"openresty-operator/internal/utils"
	"strings"
)

func ProbeUpstreamServers(ctx context.Context, upstream *webv1alpha1.Upstream) map[string]*health.CheckResult {
	return health.Checker.Submit(utils.MapList(upstream.Spec.Servers, func(server webv1alpha1.UpstreamServer) string {
		return server.Address
	}))
}

func GenerateUpstreamConfig(upstream *webv1alpha1.Upstream, results []*health.CheckResult) string {
	name := utils.SanitizeName(upstream.Name)

	switch upstream.Spec.Type {
	case webv1alpha1.UpstreamTypeAddress:
		return renderNginxUpstreamBlock(name, buildConfigLines(results))
	case webv1alpha1.UpstreamTypeFullURL:
		return renderNginxUpstreamLua(name, results, upstream.Spec.Servers)
	default:
		return ""
	}
}

func buildConfigLines(results []*health.CheckResult) []string {
	var lines []string
	for _, r := range results {
		h, p, _ := utils.SplitHostPort(r.Address)
		if r.Alive {
			ipList := make([]string, 0, len(r.IPs))
			for _, ip := range r.IPs {
				ipList = append(ipList, fmt.Sprintf("\"%s\"", ip))
			}
			lines = append(lines, fmt.Sprintf(
				"{ host = \"%s\", port = %s, weight = 1, ips = { %s } },",
				h, p, strings.Join(ipList, ", "),
			))
		}
	}
	return lines
}

func renderNginxUpstreamBlock(name string, lines []string) string {
	if len(lines) == 0 {
		return ""
	}
	var b strings.Builder
	b.WriteString(fmt.Sprintf("upstream %s {\n", name))
	b.WriteString("    server 0.0.0.1;\n")
	b.WriteString("    balancer_by_lua_block {\n")
	b.WriteString("        local servers = {\n")
	for _, line := range lines {
		b.WriteString("            " + line + "\n")
	}
	b.WriteString("        }\n\n")
	b.WriteString("        require(\"upstreams.balancer\").randomWeightedBalance(servers)\n")
	b.WriteString("    }\n")
	b.WriteString("}\n")
	return b.String()
}

/*
*

	type UpstreamServer struct {
		Address string `json:"address"`

		// NormalizeRequestRef refers to a reusable NormalizeRequest CRD
		NormalizeRequestRef *corev1.LocalObjectReference `json:"normalizeRequestRef,omitempty"`
	}
*/
func renderNginxUpstreamLua(name string, results []*health.CheckResult, servers []webv1alpha1.UpstreamServer) string {
	var b strings.Builder

	b.WriteString(fmt.Sprintf("-- upstream-%s.lua\n", name))
	b.WriteString("local random = require(\"upstreams.random_weighted\")\n\n")

	alives := 0
	b.WriteString("local servers = {\n")
	for _, s := range results {
		if s.Alive {
			b.WriteString(fmt.Sprintf("    { address = \"%s\", weight = 1 },\n", s.Address))
			alives++
		} else {
			b.WriteString(fmt.Sprintf("--    { address = \"%s\", weight = 1 },\n", s.Address))
		}
	}
	b.WriteString("}\n\n")

	b.WriteString("local normalizeFunc = {\n")
	for _, i2 := range servers {
		if i2.NormalizeRequestRef != nil {
			b.WriteString(fmt.Sprintf("    [\"%s\"] = \"normalizerules.%s\",\n", i2.Address, i2.NormalizeRequestRef.Name))
		}
	}
	b.WriteString("}\n\n")

	if alives == 0 {
		return ""
	}

	b.WriteString("random.init(servers)\n\n")
	b.WriteString("return {\n")
	b.WriteString("  default = function()\n")
	b.WriteString("    local picked = random.pick()\n")
	b.WriteString("    ngx.ctx.server_host = picked\n")
	b.WriteString("    local uri = ngx.var.uri or \"/\"\n")
	b.WriteString("    local prefix = ngx.var.location_prefix or \"/\"\n\n")
	b.WriteString("    if prefix:sub(1,1) == \"^\" then\n")
	b.WriteString("        prefix = prefix:sub(2)\n")
	b.WriteString("    end\n\n")
	b.WriteString("    local from, to = ngx.re.find(uri, \"^\" .. prefix, \"jo\")\n")
	b.WriteString("    if from == 1 and to then\n")
	b.WriteString("        uri = \"/\" .. uri:sub(to + 1)\n")
	b.WriteString("    end\n\n")
	b.WriteString("    local addr = picked\n")
	b.WriteString("    ngx.ctx.req_address = addr\n")
	b.WriteString("    if addr:match(\"https?://[^/]+/.+\") then\n")
	b.WriteString("        ngx.var.target = addr\n")
	b.WriteString("    else\n")
	b.WriteString("    	if addr:sub(-1) == \"/\" and uri:sub(1,1) == \"/\" then\n")
	b.WriteString("        	addr = addr:sub(1, -2)\n")
	b.WriteString("      end\n")
	b.WriteString("    	ngx.var.target = addr .. uri\n")
	b.WriteString("    end\n\n")

	b.WriteString("    local normalize_module = normalizeFunc[picked]\n")
	b.WriteString("    if normalize_module then\n")
	b.WriteString("      local normalize = require(normalize_module)\n")
	b.WriteString("      if normalize and normalize.request then\n")
	b.WriteString("        local ok, err = pcall(normalize.request)\n")
	b.WriteString("        if not ok then\n")
	b.WriteString("          ngx.log(ngx.ERR, \"normalizeRequest failed: \", err)\n")
	b.WriteString("        end\n")
	b.WriteString("      end\n")
	b.WriteString("    end\n")
	b.WriteString("  end,\n")

	b.WriteString("  normalizeResponse = function()\n")
	b.WriteString("    local addr = ngx.ctx.req_address\n")
	b.WriteString("    if not addr then\n")
	b.WriteString("      ngx.log(ngx.ERR, \"normalizeResponse: missing ngx.ctx.req_address\")\n")
	b.WriteString("      return\n")
	b.WriteString("    end\n")
	b.WriteString("\n")
	b.WriteString("    local normalize_module = normalizeFunc[addr]\n")
	b.WriteString("    if normalize_module then\n")
	b.WriteString("      local normalize = require(normalize_module)\n")
	b.WriteString("      if normalize and normalize.response then\n")
	b.WriteString("        local ok, err = pcall(normalize.response)\n")
	b.WriteString("        if not ok then\n")
	b.WriteString("          ngx.log(ngx.ERR, \"normalizeResponse failed: \", err)\n")
	b.WriteString("        end\n")
	b.WriteString("      end\n")
	b.WriteString("    end\n")
	b.WriteString("\n")

	b.WriteString("  end\n")
	b.WriteString("}\n")

	return b.String()
}
