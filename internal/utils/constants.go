package utils

const (
	NginxConfPath          = "/usr/local/openresty/nginx/conf/nginx.conf"
	NginxMimeTypesPath     = "/usr/local/openresty/nginx/conf/mime.types"
	NginxMainConfigMapName = "main-config"
	NginxConfDir           = "/etc/nginx/conf.d"
	NginxServerConfigDir   = NginxConfDir + "/servers"
	NginxLocationConfigDir = NginxConfDir + "/locations"
	NginxUpstreamConfigDir = NginxConfDir + "/upstreams"
)
