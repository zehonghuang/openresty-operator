package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
)

var (
	// OpenRestyCRDRefStatus 表示 OpenResty 所引用的 CRD 资源的就绪状态
	// value = 1 表示引用的资源 Ready，0 表示 NotReady 或 Missing
	OpenRestyCRDRefStatus = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "openresty_crd_ref_status",
			Help: "Tracks the readiness of CRDs referenced by OpenRestyApp",
		},
		[]string{"namespace", "from", "to", "kind"},
	)
)

func SetCRDRefStatus(namespace, from, kind, refName string, ready bool) {
	val := float64(0)
	if ready {
		val = 1
	}
	OpenRestyCRDRefStatus.WithLabelValues(namespace, from, refName, kind).Set(val)
}
