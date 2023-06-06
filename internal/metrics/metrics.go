package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	_metrics "sigs.k8s.io/controller-runtime/pkg/metrics"
)

var (
	Registry = _metrics.Registry
)

// RegisterAll 注册所有模块的指标
func RegisterAll() error {
	collectors := []prometheus.Collector{
		OpenRestyCRDRefStatus,
		UpstreamDNSResolvable,
	}

	for _, c := range collectors {
		if err := Registry.Register(c); err != nil {
			return err
		}
	}
	return nil
}
