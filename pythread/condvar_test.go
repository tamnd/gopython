package pythread

import (
	"testing"
	"time"
)

func TestCondWaitAndSignal(t *testing.T) {
	var mu Mutex
	var cond Cond
	done := make(chan struct{})

	mu.Lock()
	go func() {
		time.Sleep(10 * time.Millisecond)
		cond.Signal()
		close(done)
	}()

	if got := cond.Wait(&mu); got != 0 {
		t.Fatalf("Wait = %d, want 0", got)
	}
	mu.Unlock()

	<-done
}

func TestCondTimedWaitTimeout(t *testing.T) {
	var mu Mutex
	var cond Cond

	mu.Lock()
	start := time.Now()
	got := cond.TimedWait(&mu, 20*time.Millisecond)
	elapsed := time.Since(start)
	mu.Unlock()

	if got != 1 {
		t.Fatalf("TimedWait = %d, want 1", got)
	}
	if elapsed < 15*time.Millisecond {
		t.Fatalf("TimedWait returned too early: %s", elapsed)
	}
}

func TestCondTimedWaitSignalBeatsTimeout(t *testing.T) {
	var mu Mutex
	var cond Cond

	mu.Lock()
	go func() {
		time.Sleep(10 * time.Millisecond)
		cond.Signal()
	}()

	if got := cond.TimedWait(&mu, 200*time.Millisecond); got != 0 {
		t.Fatalf("TimedWait = %d, want 0", got)
	}
	mu.Unlock()
}

func TestCondBroadcastWakesAll(t *testing.T) {
	var mu Mutex
	var cond Cond
	done := make(chan struct{}, 2)

	waiter := func() {
		mu.Lock()
		cond.Wait(&mu)
		mu.Unlock()
		done <- struct{}{}
	}
	go waiter()
	go waiter()

	time.Sleep(10 * time.Millisecond)
	cond.Broadcast()

	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("first waiter was not released")
	}
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("second waiter was not released")
	}
}
