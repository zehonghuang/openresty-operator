package health

import (
	"net"
	"sync"
	"time"

	"k8s.io/client-go/util/workqueue"
)

type Checker struct {
	queue       workqueue.TypedRateLimitingInterface[string]
	statusMap   map[string]bool
	failures    map[string]int
	maxFailures int
	lock        sync.RWMutex
	numWorker   int
	timeout     time.Duration
}

func NewChecker(workerCount int, timeout time.Duration) *Checker {
	c := &Checker{
		queue: workqueue.NewTypedRateLimitingQueueWithConfig(workqueue.DefaultTypedControllerRateLimiter[string](),
			workqueue.TypedRateLimitingQueueConfig[string]{
				Name: "healthcheck",
			}),
		statusMap:   make(map[string]bool),
		failures:    make(map[string]int),
		maxFailures: 3,
		numWorker:   workerCount,
		timeout:     timeout,
	}
	for i := 0; i < c.numWorker; i++ {
		go c.worker()
	}
	return c
}

func (c *Checker) Submit(addresses []string) map[string]bool {
	results := make(map[string]bool)
	c.lock.RLock()
	for _, addr := range addresses {
		val, ok := c.statusMap[addr]
		if !ok {
			c.lock.RUnlock()
			c.queue.Add(addr)
			c.lock.RLock()
			results[addr] = false // unknown yet
		} else {
			results[addr] = val
		}
	}
	c.lock.RUnlock()
	return results
}

func (c *Checker) worker() {
	for {
		item, shutdown := c.queue.Get()
		if shutdown {
			return
		}
		addr := item
		ready := c.healthCheck(addr)

		c.lock.Lock()
		c.statusMap[addr] = ready
		if ready {
			c.failures[addr] = 0
		} else {
			c.failures[addr]++
		}
		shouldRequeue := c.failures[addr] < c.maxFailures
		c.lock.Unlock()

		if shouldRequeue {
			c.queue.AddAfter(addr, 60*time.Second)
		}

		c.queue.Done(item)
	}
}

func (c *Checker) healthCheck(address string) bool {
	conn, err := net.DialTimeout("tcp", address, c.timeout)
	if err != nil {
		return false
	}
	_ = conn.Close()
	return true
}
