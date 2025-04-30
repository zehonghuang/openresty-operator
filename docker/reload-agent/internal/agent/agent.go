package agent

import (
	"bufio"
	"fmt"
	"os/exec"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"
)

type ReloadPolicy struct {
	Window    int `yaml:"window"`
	MaxEvents int `yaml:"maxEvents"`
}

type ReloadAgent struct {
	mu             sync.Mutex
	events         []time.Time
	window         time.Duration
	maxEvents      int
	lastReloadTime time.Time
	OnReload       func()
}

func NewReloadAgent(windowSeconds int, maxEvents int) *ReloadAgent {
	return &ReloadAgent{
		events:         make([]time.Time, 0),
		window:         time.Duration(windowSeconds) * time.Second,
		maxEvents:      maxEvents,
		lastReloadTime: time.Now(),
	}
}

func (r *ReloadAgent) RecordChange() {
	r.mu.Lock()
	defer r.mu.Unlock()

	now := time.Now()
	r.events = append(r.events, now)

	// 🔥 满足 maxEvents，立即触发
	if len(r.events) >= r.maxEvents {
		fmt.Printf("[reload-agent] 🔥 event threshold met (%d events), reloading early\n", len(r.events))
		r.reload(now)
	}
}

func (r *ReloadAgent) StartTicker() {
	ticker := time.NewTicker(1 * time.Second)
	go func() {
		for range ticker.C {
			r.mu.Lock()
			now := time.Now()

			// ⏳ 超过 window，强制判断（即使事件不够）
			if now.Sub(r.lastReloadTime) >= r.window && len(r.events) > 0 {
				fmt.Printf("[reload-agent] ⏳ window elapsed (%.0fs), checking\n", r.window.Seconds())
				r.reload(now)
			}
			r.mu.Unlock()
		}
	}()
}

func (r *ReloadAgent) reload(now time.Time) {
	fmt.Printf("[reload-agent] ✅ triggering nginx reload (events=%d in %.0fs)\n",
		len(r.events), now.Sub(r.lastReloadTime).Seconds())

	r.OnReload()

	if err := sendReloadSignalToNginx(); err != nil {
		fmt.Printf("[reload-agent] ❌ reload failed: %v\n", err)
		return
	}
	// 清空 & 重置窗口
	r.lastReloadTime = now
	r.events = nil
}

func sendReloadSignalToNginx() error {
	cmd := exec.Command("ps", "-eo", "pid,comm,args")
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("failed to get ps output: %w", err)
	}
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to run ps: %w", err)
	}

	scanner := bufio.NewScanner(stdout)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.Contains(line, "nginx: master") {
			fields := strings.Fields(line)
			if len(fields) >= 1 {
				pid, err := strconv.Atoi(fields[0])
				if err == nil {
					fmt.Printf("[agent] ✅ reloading nginx (pid=%d)\n", pid)
					return syscall.Kill(pid, syscall.SIGHUP)
				}
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("error reading ps output: %w", err)
	}

	fmt.Println("[agent] ❌ nginx master PID not found")
	return fmt.Errorf("nginx master PID not found")
}
