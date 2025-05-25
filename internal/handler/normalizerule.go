package handler

import (
	"encoding/json"
	"fmt"
	"openresty-operator/api/v1alpha1"
	"strings"
)

func RenderNormalizeRuleLua(rule *v1alpha1.NormalizeRule) string {
	var builder strings.Builder

	builder.WriteString("return {\n")

	if len(rule.Spec.Request) > 0 {
		builder.WriteString("  request = function()\n")
		builder.WriteString("    ngx.req.read_body()\n")
		builder.WriteString("    local cjson = require(\"cjson.safe\")\n")
		builder.WriteString("    local utils = require(\"normalize.utils\")\n")
		builder.WriteString("    local requestObj = cjson.decode(ngx.req.get_body_data()) or {}\n")
		builder.WriteString("    local output = {}\n\n")

		for key, val := range rule.Spec.Request {
			// Try parse as string
			var str string
			if err := json.Unmarshal(val.Raw, &str); err == nil {
				builder.WriteString(fmt.Sprintf("    output[%q] = utils.get(requestObj, %q)\n", key, str))
				continue
			}

			// Try parse as {"lua": "..."}
			var obj map[string]interface{}
			if err := json.Unmarshal(val.Raw, &obj); err == nil {
				if luaVal, ok := obj["lua"]; ok {
					if luaStr, ok := luaVal.(string); ok {
						builder.WriteString(fmt.Sprintf("    output[%q] = (function()\n      %s  end)()\n", key, indentLua(luaStr, "        ")))
						continue
					}
				}
			}

			// Fallback comment
			builder.WriteString(fmt.Sprintf("    -- could not render normalize field %q\n", key))
		}

		builder.WriteString("\n    ngx.req.set_body_data(cjson.encode(output))\n")
		builder.WriteString("  end")
		if len(rule.Spec.Response) > 0 {
			builder.WriteString(",\n")
		} else {
			builder.WriteString("\n")
		}
	}

	if len(rule.Spec.Response) > 0 {
		builder.WriteString("  response = function()\n")
		builder.WriteString("    if not ngx.ctx.body_buffer then\n")
		builder.WriteString("      ngx.ctx.body_buffer = {}\n")
		builder.WriteString("    end\n")
		builder.WriteString("    table.insert(ngx.ctx.body_buffer, ngx.arg[1])\n\n")
		builder.WriteString("    if not ngx.arg[2] then\n")
		builder.WriteString("      ngx.arg[1] = nil\n")
		builder.WriteString("      return\n")
		builder.WriteString("    end\n\n")
		builder.WriteString("    local full_body = table.concat(ngx.ctx.body_buffer)\n")
		builder.WriteString("    local cjson = require(\"cjson.safe\")\n")
		builder.WriteString("    local utils = require(\"normalize.utils\")\n")
		builder.WriteString("    local responseObj = cjson.decode(full_body) or {}\n")
		builder.WriteString("    local output = {}\n\n")

		for key, val := range rule.Spec.Response {
			// Try parse as string
			var str string
			if err := json.Unmarshal(val.Raw, &str); err == nil {
				builder.WriteString(fmt.Sprintf("    output[%q] = utils.get(responseObj, %q)\n", key, str))
				continue
			}

			// Try parse as {"lua": "..."}
			var obj map[string]interface{}
			if err := json.Unmarshal(val.Raw, &obj); err == nil {
				if luaVal, ok := obj["lua"]; ok {
					if luaStr, ok := luaVal.(string); ok {
						builder.WriteString(fmt.Sprintf("    output[%q] = (function()\n      %s  end)()\n", key, indentLua(luaStr, "        ")))
						continue
					}
				}
			}

			// Fallback comment
			builder.WriteString(fmt.Sprintf("    -- could not render normalize field %q\n", key))
		}

		builder.WriteString("\n    ngx.arg[1] = cjson.encode(output)\n")
		builder.WriteString("  end\n")
	}

	builder.WriteString("}\n")

	return builder.String()
}
