package utils

const (
	NginxConfPath          = "/usr/local/openresty/nginx/conf/nginx.conf"
	NginxMimeTypesPath     = "/usr/local/openresty/nginx/conf/mime.types"
	NginxMainConfigMapName = "main-config"
	NginxConfDir           = "/etc/nginx/conf.d"
	NginxServerConfigDir   = NginxConfDir + "/servers"
	NginxLocationConfigDir = NginxConfDir + "/locations"
	NginxUpstreamConfigDir = NginxConfDir + "/upstreams"
	NginxLuaLibDir         = "/usr/local/openresty/lualib"
	NginxLuaLibUpstreamDir = NginxLuaLibDir + "/upstreams"
	NginxLuaLibSecretDir   = NginxLuaLibDir + "/secrets"
	NginxLogDir            = "/var/log/nginx"
	NginxTemplate          = `
worker_processes auto;
events { worker_connections 1024; }
http {
	
    resolver kube-dns.kube-system.svc.cluster.local valid=30s;
	
	lua_shared_dict secrets_store 10m;
    lua_shared_dict prometheus_metrics 10M;
    init_worker_by_lua_block {
		require("secrets.secrets_loader").reload()
{{ indent .InitLua 8 }}
    }
{{- if .EnableMetrics }}
    server {
        listen {{ .MetricsPort }};
        location {{ .MetricsPath }} {
            content_by_lua_block {
                prometheus:collect()
            }
        }
    }
{{- end }}
{{- range .Includes }}
    include {{ . }};
{{- end }}
{{- if .LogFormat }}log_format main '{{ .LogFormat }}';{{ end }}
{{- if .AccessLog }}access_log {{ .AccessLog }};{{ end }}
{{- if .ErrorLog }}error_log {{ .ErrorLog }};{{ end }}
{{- if .ClientMaxBodySize }}client_max_body_size {{ .ClientMaxBodySize }};{{ end }}
{{- if .Gzip }}gzip on;{{ end }}
{{- range .Extra }}
    {{ . }}
{{- end }}
{{- range .IncludeSnippets }}
    {{ . }}
{{- end }}
}
`
)
