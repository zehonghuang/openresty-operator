package handler

import (
	"encoding/json"
	"fmt"
	corev1 "k8s.io/api/core/v1"
	"openresty-operator/api/v1alpha1"
	"strings"
)

func RenderNormalizeRuleLua(rule *v1alpha1.NormalizeRule, getSecretFunc func(ns, name string) (*corev1.Secret, error)) string {
	var builder strings.Builder

	builder.WriteString("local cjson = require(\"cjson\")\n")
	builder.WriteString("cjson.encode_escape_forward_slash(false)\n")
	builder.WriteString("local utils = require(\"normalize.utils\")\n")
	builder.WriteString("return {\n")

	if rule.Spec.Request != nil {
		builder.WriteString("  request = function()\n")
		builder.WriteString("    ngx.req.read_body()\n")
		builder.WriteString("    local requestObj = cjson.decode(ngx.req.get_body_data()) or {}\n")
		builder.WriteString("    local output = {}\n\n")

		if len(rule.Spec.Request.Body) > 0 {
			for key, val := range rule.Spec.Request.Body {
				// Try parse as string
				var str string
				if err := json.Unmarshal(val.Raw, &str); err == nil {
					builder.WriteString(fmt.Sprintf("    output[%q] = utils.get(requestObj, %q)\n", key, str))
					continue
				}

				// Try parse as {"lua": "..."} or {"value": "..."}
				var obj map[string]interface{}
				if err := json.Unmarshal(val.Raw, &obj); err == nil {
					if luaVal, ok := obj["lua"]; ok {
						if luaStr, ok := luaVal.(string); ok {
							builder.WriteString(fmt.Sprintf("    output[%q] = (function()\n      %s  end)()\n", key, indentLua(luaStr, "        ")))
							continue
						}
					}
					if staticVal, ok := obj["value"]; ok {
						builder.WriteString(fmt.Sprintf("    output[%q] = %q\n", key, fmt.Sprintf("%v", staticVal)))
						continue
					}
				}

				// Fallback comment
				builder.WriteString(fmt.Sprintf("    -- could not render normalize field %q\n", key))
			}
		}

		if len(rule.Spec.Request.Query) > 0 || len(rule.Spec.Request.QueryFromSecret) > 0 {
			builder.WriteString("    local query = {}\n")
			for key, val := range rule.Spec.Request.Query {
				// Try parse as string
				var str string
				if err := json.Unmarshal(val.Raw, &str); err == nil {
					builder.WriteString(fmt.Sprintf("    query[%q] = utils.get(requestObj, %q)\n", key, str))
					continue
				}

				// Try parse as {"lua": "..."} or {"value": "..."}
				var obj map[string]interface{}
				if err := json.Unmarshal(val.Raw, &obj); err == nil {
					if luaVal, ok := obj["lua"]; ok {
						if luaStr, ok := luaVal.(string); ok {
							builder.WriteString(fmt.Sprintf("    query[%q] = (function()\n      %s  end)()\n", key, indentLua(luaStr, "        ")))
							continue
						}
					}
					if staticVal, ok := obj["value"]; ok {
						builder.WriteString(fmt.Sprintf("    query[%q] = %q\n", key, fmt.Sprintf("%v", staticVal)))
						continue
					}
				}

				// Fallback comment
				builder.WriteString(fmt.Sprintf("    -- could not render normalize query param %q\n", key))
			}
			for _, val := range rule.Spec.Request.QueryFromSecret {
				secret, err := getSecretFunc(rule.Namespace, val.SecretName)
				if err != nil || secret == nil {
					builder.WriteString(fmt.Sprintf("    -- failed to get secret %q\n", val.SecretName))
					continue
				}
				if b64, ok := secret.Data[val.SecretKey]; ok {
					builder.WriteString(fmt.Sprintf("    query[%q] = %q\n", val.Name, string(b64)))
				} else {
					builder.WriteString(fmt.Sprintf("    -- key %q not found in secret %q\n", val.SecretKey, val.SecretName))
				}
			}
			builder.WriteString("    local queryStr = ngx.encode_args(query)\n")
			builder.WriteString("    if not ngx.var.target:find(\"?\", 1, true) then\n")
			builder.WriteString("      ngx.var.target = ngx.var.target .. \"?\" .. queryStr\n")
			builder.WriteString("    else\n")
			builder.WriteString("      ngx.var.target = ngx.var.target .. \"&\" .. queryStr\n")
			builder.WriteString("    end\n")
		}

		for headerKey, val := range rule.Spec.Request.Headers {
			builder.WriteString(fmt.Sprintf("    ngx.req.set_header(%q, %q)\n", headerKey, val.Value))
		}

		for _, val := range rule.Spec.Request.HeadersFromSecret {
			secret, err := getSecretFunc(rule.Namespace, val.SecretName)
			if err != nil || secret == nil {
				builder.WriteString(fmt.Sprintf("    -- failed to get secret %q\n", val.SecretName))
				continue
			}
			if b64, ok := secret.Data[val.SecretKey]; ok {
				builder.WriteString(fmt.Sprintf("    ngx.req.set_header(%q, %q)\n", val.Name, string(b64)))
			} else {
				builder.WriteString(fmt.Sprintf("    -- key %q not found in secret %q\n", val.SecretKey, val.SecretName))
			}
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

		builder.WriteString("\n    local final_body = cjson.encode(output)\n")
		builder.WriteString("\n    ngx.arg[1] = final_body\n")
		builder.WriteString("\n    ngx.arg[2] = true\n")
		builder.WriteString("  end\n")
	}

	builder.WriteString("}\n")

	return builder.String()
}
