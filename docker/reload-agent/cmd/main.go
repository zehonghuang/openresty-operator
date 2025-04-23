package main

import (
	"encoding/json"
	"fmt"
	"gopkg.in/yaml.v3"
	"log"
	"os"

	"reload-agent/internal/agent"
	"reload-agent/internal/watcher"
)

func main() {
	fmt.Println("[reload-agent] starting")

	jsonData := os.Getenv("RELOAD_POLICIES")
	var policies []agent.ReloadPolicy
	if err := json.Unmarshal([]byte(jsonData), &policies); err != nil {
		log.Fatalf("invalid RELOAD_POLICIES: %v", err)
	}

	r := agent.NewReloadAgent(policies)

	err := watcher.WatchDirectory("/etc/nginx/conf.d", func() {
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
