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
		return renderNginxUpstreamLua(name, results)
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

func renderNginxUpstreamLua(name string, results []*health.CheckResult) string {
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

	if alives == 0 {
		return ""
	}

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
