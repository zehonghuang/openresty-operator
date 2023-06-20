package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
)

var (
	UpstreamDNSResolvable = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: "openresty",
			Subsystem: "upstream",
			Name:      "dns_resolvable",
			Help:      "Whether the upstream DNS is resolvable (1=yes, 0=no).",
		},
		[]string{"namespace", "upstream", "server", "ip"},
	)
)

func SetUpstreamDNSResolvable(namespace, upstream, server, ip string, resolvable bool) {
	val := 0.0
	if resolvable {
		val = 1.0
	}
	UpstreamDNSResolvable.WithLabelValues(namespace, upstream, server, ip).Set(val)
}
