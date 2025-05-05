package health

import (
	"fmt"
	"github.com/go-logr/logr"
	"net"
	"openresty-operator/internal/utils"
	"sync"
	"time"

	"k8s.io/client-go/util/workqueue"
)

var (
	Checker *checker
)

type CheckResult struct {
	Address string
	IPs     []string
	Comment string
	Alive   bool
	Reason  string
}

func Init(workerCount int, timeout time.Duration, log logr.Logger) {
	Checker = newChecker(workerCount, timeout, log)
}

type checker struct {
	queue       workqueue.TypedRateLimitingInterface[string]
	statusMap   map[string]*CheckResult
	failures    map[string]int
	refCount    map[string]int
	maxFailures int
	lock        sync.RWMutex
	numWorker   int
	timeout     time.Duration
	log         logr.Logger
}

func newChecker(workerCount int, timeout time.Duration, log logr.Logger) *checker {
	c := &checker{
		queue: workqueue.NewTypedRateLimitingQueueWithConfig(workqueue.DefaultTypedControllerRateLimiter[string](),
			workqueue.TypedRateLimitingQueueConfig[string]{
				Name: "healthcheck",
			}),
		statusMap:   make(map[string]*CheckResult),
		failures:    make(map[string]int),
		maxFailures: 10,
		numWorker:   workerCount,
		timeout:     timeout,
		refCount:    make(map[string]int),
		log:         log.WithName("health-checker"),
	}
	for i := 0; i < c.numWorker; i++ {
		go c.worker()
	}
	return c
}

func (c *checker) Submit(addresses []string) map[string]*CheckResult {
	results := make(map[string]*CheckResult)
	c.lock.RLock()
	for _, addr := range addresses {
		val, ok := c.statusMap[addr]
		if !ok {
			c.lock.RUnlock()
			c.lock.Lock()
			c.refCount[addr]++
			c.lock.Unlock()
			c.queue.Add(addr)
			c.lock.RLock()
			results[addr] = nil
		} else {
			results[addr] = val
		}
	}
	c.lock.RUnlock()
	return results
}

func (c *checker) Release(addresses []string) {
	c.lock.Lock()
	defer c.lock.Unlock()
	for _, addr := range addresses {
		if count, ok := c.refCount[addr]; ok {
			if count <= 1 {
				delete(c.refCount, addr)
				delete(c.statusMap, addr)
				delete(c.failures, addr)
			} else {
				c.refCount[addr]--
			}
		}
	}
}

func (c *checker) worker() {
	for {
		item, shutdown := c.queue.Get()
		if shutdown {
			return
		}
		addr := item
		result := c.performHealthCheck(addr)
		c.log.Info("Health check completed", "address", addr, "alive", result.Alive, "reason", result.Reason, "time", time.Now().Format(time.RFC3339))

		c.lock.Lock()
		c.statusMap[addr] = result
		if result.Alive {
			c.failures[addr] = 0
		} else {
			c.failures[addr]++
		}
		shouldRequeue := c.failures[addr] < c.maxFailures
		if !shouldRequeue {
			delete(c.statusMap, addr)
			delete(c.failures, addr)
			delete(c.refCount, addr)
		}
		c.lock.Unlock()

		if shouldRequeue {
			c.queue.AddAfter(addr, 60*time.Second)
		}

		c.queue.Done(item)
	}
}

func (c *checker) performHealthCheck(addr string) *CheckResult {
	host, port, _ := utils.SplitHostPort(addr)
	ips := c.lookupHost(host)
	dnsIsReady := len(ips) > 0
	tcpIsReady := true
	if dnsIsReady {
		tcpIsReady = c.testTCP(net.JoinHostPort(host, port))
	}

	return &CheckResult{
		Address: addr,
		Alive:   tcpIsReady && dnsIsReady,
		IPs:     ips,
		Reason: func() string {
			if !dnsIsReady {
				return "DNS_ERROR"
			}
			if !tcpIsReady {
				return "TCP_FAIL"
			}
			return ""
		}(),
		Comment: func() string {
			if !dnsIsReady {
				return fmt.Sprintf("# server %s;  // DNS error", addr)
			}
			if !tcpIsReady {
				return fmt.Sprintf("# server %s;  // tcp unreachable", addr)
			}
			return ""
		}(),
	}
}

func (c *checker) testTCP(address string) bool {
	conn, err := net.DialTimeout("tcp", address, c.timeout)
	if err != nil {
		return false
	}
	_ = conn.Close()
	return true
}

func (c *checker) lookupHost(address string) []string {
	ips, err := net.LookupHost(address)
	if err != nil {
		return nil
	}
	return ips
}
