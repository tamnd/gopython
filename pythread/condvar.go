package pythread

import (
	"time"
)

type Cond struct {
	waitersMu Mutex
	waiters   []chan struct{}
}

func (c *Cond) Wait(m *Mutex) int {
	waiter := c.addWaiter()
	m.Unlock()
	<-waiter
	m.Lock()
	return 0
}

func (c *Cond) TimedWait(m *Mutex, timeout time.Duration) int {
	waiter := c.addWaiter()
	m.Unlock()

	if timeout < 0 {
		<-waiter
		m.Lock()
		return 0
	}

	timer := time.NewTimer(timeout)
	defer timer.Stop()

	result := 0
	select {
	case <-waiter:
		result = 0
	case <-timer.C:
		if c.removeWaiter(waiter) {
			result = 1
		} else {
			<-waiter
			result = 0
		}
	}

	m.Lock()
	return result
}

func (c *Cond) Signal() int {
	c.waitersMu.Lock()
	defer c.waitersMu.Unlock()

	if len(c.waiters) == 0 {
		return 0
	}
	waiter := c.waiters[0]
	c.waiters = c.waiters[1:]
	close(waiter)
	return 0
}

func (c *Cond) Broadcast() int {
	c.waitersMu.Lock()
	defer c.waitersMu.Unlock()

	for _, waiter := range c.waiters {
		close(waiter)
	}
	c.waiters = nil
	return 0
}

func (c *Cond) addWaiter() chan struct{} {
	waiter := make(chan struct{})
	c.waitersMu.Lock()
	c.waiters = append(c.waiters, waiter)
	c.waitersMu.Unlock()
	return waiter
}

func (c *Cond) removeWaiter(target chan struct{}) bool {
	c.waitersMu.Lock()
	defer c.waitersMu.Unlock()

	for i, waiter := range c.waiters {
		if waiter == target {
			c.waiters = append(c.waiters[:i], c.waiters[i+1:]...)
			return true
		}
	}
	return false
}
