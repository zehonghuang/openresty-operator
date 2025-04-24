package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"

	"reload-agent/internal/agent"
	"reload-agent/internal/watcher"
)

func main() {
	fmt.Println("[reload-agent] starting")

	jsonData := os.Getenv("RELOAD_POLICY")
	var policy agent.ReloadPolicy

	if err := json.Unmarshal([]byte(jsonData), &policy); err != nil {
		log.Fatalf("invalid RELOAD_POLICY: %v", err)
	}

	r := agent.NewReloadAgent(policy.Window, policy.MaxEvents)
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
