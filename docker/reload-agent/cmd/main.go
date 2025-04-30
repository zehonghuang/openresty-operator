package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"reload-agent/internal/agent"
	"reload-agent/internal/watcher"
)

func main() {
	fmt.Println("[reload-agent] starting")

	var (
		reloadTotal = prometheus.NewCounter(prometheus.CounterOpts{
			Name: "reload_agent_reload_total",
			Help: "Total number of nginx reloads triggered by the agent",
		})

		lastReloadTimestamp = prometheus.NewGauge(prometheus.GaugeOpts{
			Name: "reload_agent_reload_last_timestamp_seconds",
			Help: "Unix timestamp of the last reload triggered",
		})
	)

	prometheus.MustRegister(reloadTotal, lastReloadTimestamp)

	go func() {
		http.Handle("/metrics", promhttp.Handler())
		log.Println("[reload-agent] Prometheus metrics available at :19091/metrics")
		log.Fatal(http.ListenAndServe(":19091", nil))
	}()

	jsonData := os.Getenv("RELOAD_POLICY")
	var policy agent.ReloadPolicy

	if err := json.Unmarshal([]byte(jsonData), &policy); err != nil {
		log.Fatalf("invalid RELOAD_POLICY: %v", err)
	}

	r := agent.NewReloadAgent(policy.Window, policy.MaxEvents)
	r.OnReload = func() {
		reloadTotal.Inc()
		lastReloadTimestamp.SetToCurrentTime()
	}
	r.StartTicker()

	dirs, err := watcher.DiscoverWatchDirs("/etc/nginx/conf.d")
	if err != nil {
		log.Fatalf("failed to discover watch dirs: %v", err)
	}
	for _, dir := range dirs {
		dir := dir
		go func() {
			err := watcher.WatchDirectory(dir, func() {
				r.RecordChange()
			})
			if err != nil {
				fmt.Fprintf(os.Stderr, "[reload-agent] failed to watch %s: %v\n", dir, err)
			}
		}()
	}

	select {}
}
