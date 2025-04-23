package main

import (
	"fmt"
	"gopkg.in/yaml.v3"
	"os"

	"reload-agent/internal/agent"
	"reload-agent/internal/watcher"
)

func main() {
	fmt.Println("[reload-agent] starting")

	cfg, err := loadConfig("config/default.yaml")
	if err != nil {
		fmt.Fprintf(os.Stderr, "[reload-agent] failed to load config: %v\n", err)
		os.Exit(1)
	}

	r := agent.NewReloadAgent(cfg.ReloadPolicies)

	err = watcher.WatchDirectory("/etc/nginx/conf.d", func() {
		r.RecordChange()
	})

	if err != nil {
		fmt.Fprintln(os.Stderr, "[reload-agent] error:", err)
		os.Exit(1)
	}
}

func loadConfig(path string) (*agent.Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read config error: %w", err)
	}

	var cfg agent.Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parse yaml error: %w", err)
	}

	return &cfg, nil
}
