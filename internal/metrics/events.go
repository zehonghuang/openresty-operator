package metrics

import "github.com/prometheus/client_golang/prometheus"

var (
	// OpenrestyOperatorEventInfo 表示 OpenResty 所引用的 CRD 资源的就绪状态
	// value = 1 表示引用的资源 Ready，0 表示 NotReady 或 Missing
	OpenrestyOperatorEventInfo = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "openresty_operator_event_info",
			Help: "Openresty-Operator events converted to labeled metrics",
		},
		[]string{
			"kind",
			"namespace",
			"name",
			"type",
			"reason",
		},
	)
)

func Recorder(kind, namespace, name, eventType, reason string) {
	OpenrestyOperatorEventInfo.WithLabelValues(kind, namespace, name, eventType, reason).Inc()
}
