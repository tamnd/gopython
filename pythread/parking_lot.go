package pythread

import (
	"sync"
	"time"

	"github.com/tamnd/gopython/pytime"
)

const (
	ParkOK = iota
	ParkAgain
	ParkTimeout
	ParkIntr
)

type Semaphore struct {
	ch chan struct{}
}

func (s *Semaphore) Init() {
	s.ch = make(chan struct{}, 1)
}

func (s *Semaphore) Wait(timeout pytime.Time, detach bool) int {
	if timeout < 0 {
		<-s.ch
		return ParkOK
	}
	timer := timeAfter(timeout)
	select {
	case <-s.ch:
		return ParkOK
	case <-timer:
		return ParkTimeout
	}
}

func (s *Semaphore) Wakeup() {
	select {
	case s.ch <- struct{}{}:
	default:
	}
}

func (s *Semaphore) Destroy() {}

type waitEntry struct {
	parkArg     any
	addr        uintptr
	sema        Semaphore
	isUnparking bool
}

type parkingBucket struct {
	mu      sync.Mutex
	waiters []*waitEntry
}

type ParkingLot struct {
	buckets [257]parkingBucket
}

var defaultParkingLot ParkingLot

func timeAfter(timeout pytime.Time) <-chan time.Time {
	return time.After(time.Duration(timeout))
}

func (p *ParkingLot) bucket(addr uintptr) *parkingBucket {
	return &p.buckets[addr%uintptr(len(p.buckets))]
}

func (p *ParkingLot) Park(addr uintptr, canPark func() bool, timeout pytime.Time, parkArg any, detach bool) int {
	wait := &waitEntry{
		parkArg: parkArg,
		addr:    addr,
	}

	bucket := p.bucket(addr)
	bucket.mu.Lock()
	if !canPark() {
		bucket.mu.Unlock()
		return ParkAgain
	}
	wait.sema.Init()
	bucket.waiters = append(bucket.waiters, wait)
	bucket.mu.Unlock()

	res := wait.sema.Wait(timeout, detach)
	if res == ParkOK {
		return res
	}

	bucket.mu.Lock()
	if wait.isUnparking {
		bucket.mu.Unlock()
		for {
			if wait.sema.Wait(-1, false) == ParkOK {
				return ParkOK
			}
		}
	}
	for i, queued := range bucket.waiters {
		if queued == wait {
			bucket.waiters = append(bucket.waiters[:i], bucket.waiters[i+1:]...)
			break
		}
	}
	bucket.mu.Unlock()
	return res
}

func (p *ParkingLot) Unpark(addr uintptr, fn func(arg any, parkArg any, hasMoreWaiters bool), arg any) {
	bucket := p.bucket(addr)

	bucket.mu.Lock()
	var waiter *waitEntry
	for i, queued := range bucket.waiters {
		if queued.addr == addr {
			waiter = queued
			bucket.waiters = append(bucket.waiters[:i], bucket.waiters[i+1:]...)
			waiter.isUnparking = true
			break
		}
	}
	hasMore := false
	for _, queued := range bucket.waiters {
		if queued.addr == addr {
			hasMore = true
			break
		}
	}
	if waiter != nil {
		fn(arg, waiter.parkArg, hasMore)
	} else {
		fn(arg, nil, false)
	}
	bucket.mu.Unlock()

	if waiter != nil {
		waiter.sema.Wakeup()
	}
}

func (p *ParkingLot) UnparkAll(addr uintptr) {
	bucket := p.bucket(addr)
	bucket.mu.Lock()
	waiters := make([]*waitEntry, 0, len(bucket.waiters))
	filtered := bucket.waiters[:0]
	for _, waiter := range bucket.waiters {
		if waiter.addr == addr {
			waiter.isUnparking = true
			waiters = append(waiters, waiter)
		} else {
			filtered = append(filtered, waiter)
		}
	}
	bucket.waiters = filtered
	bucket.mu.Unlock()

	for _, waiter := range waiters {
		waiter.sema.Wakeup()
	}
}

func (p *ParkingLot) AfterFork() {
	for i := range p.buckets {
		p.buckets[i] = parkingBucket{}
	}
}
