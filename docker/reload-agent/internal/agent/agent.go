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

type Config struct {
	ReloadPolicies []ReloadPolicy `yaml:"reloadPolicies"`
}

type ReloadPolicy struct {
	Window    int `yaml:"window"`
	MaxEvents int `yaml:"maxEvents"`
}

type ReloadAgent struct {
	mu       sync.Mutex
	events   []time.Time
	policies []ReloadPolicy
}

func NewReloadAgent(policies []ReloadPolicy) *ReloadAgent {
	return &ReloadAgent{
		policies: policies,
		events:   make([]time.Time, 0),
	}
}

func (r *ReloadAgent) RecordChange() {
	r.mu.Lock()
	defer r.mu.Unlock()

	now := time.Now()
	r.events = append(r.events, now)

	// 清理过期事件（保留最长窗口期内的事件）
	cutoff := now.Add(-r.maxWindow())
	valid := make([]time.Time, 0, len(r.events))
	for _, t := range r.events {
		if t.After(cutoff) {
			valid = append(valid, t)
		}
	}
	r.events = valid

	// 判断是否满足任一策略
	for _, policy := range r.policies {
		threshold := now.Add(-time.Duration(policy.Window) * time.Second)
		count := 0
		for _, t := range r.events {
			if t.After(threshold) {
				count++
			}
		}
		if count >= policy.MaxEvents {
			fmt.Println("[reload-agent] triggering nginx reload")
			err := sendReloadSignalToNginx()
			if err != nil {
				// TODO: 错误处理、metrics、日志
			} else {

			}
			r.events = nil
			return
		}
	}
}

func (r *ReloadAgent) maxWindow() time.Duration {
	_max := time.Second
	for _, p := range r.policies {
		w := time.Duration(p.Window) * time.Second
		if w > _max {
			_max = w
		}
	}
	return _max
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
