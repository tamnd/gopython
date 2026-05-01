package pythread

import (
	"testing"
	"time"

	"github.com/tamnd/gopython/pytime"
)

func TestPyMutexLockUnlockAndTimeout(t *testing.T) {
	var m PyMutex
	if status := m.LockTimed(0, 0); status != LockAcquired {
		t.Fatalf("initial lock = %v, want LockAcquired", status)
	}
	if status := m.LockTimed(0, 0); status != LockFailure {
		t.Fatalf("second lock = %v, want LockFailure", status)
	}
	m.Unlock()
	if m.IsLocked() {
		t.Fatal("mutex should be unlocked")
	}
}

func TestPyMutexWakesWaiter(t *testing.T) {
	var m PyMutex
	m.Lock()
	done := make(chan LockStatus, 1)
	go func() {
		done <- m.LockTimed(pytime.Time(100_000_000), LockDetach)
	}()
	time.Sleep(10 * time.Millisecond)
	m.Unlock()
	if got := <-done; got != LockAcquired {
		t.Fatalf("waiter status = %v, want LockAcquired", got)
	}
	m.Unlock()
}

func TestRawMutexWakeup(t *testing.T) {
	var m RawMutex
	m.Lock()
	done := make(chan struct{}, 1)
	go func() {
		m.Lock()
		done <- struct{}{}
		m.Unlock()
	}()
	time.Sleep(10 * time.Millisecond)
	m.Unlock()
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("raw mutex waiter did not wake up")
	}
}

func TestEventNotify(t *testing.T) {
	var e Event
	done := make(chan bool, 1)
	go func() {
		done <- e.WaitTimed(pytime.Time(100_000_000), true)
	}()
	time.Sleep(10 * time.Millisecond)
	e.Notify()
	if !<-done {
		t.Fatal("event waiter should observe notify")
	}
}

func TestOnceFlagCallOnceSlow(t *testing.T) {
	var o OnceFlag
	calls := 0
	fn := func(arg any) int {
		calls++
		return 0
	}
	if got := o.CallOnceSlow(fn, nil); got != 0 {
		t.Fatalf("first CallOnceSlow = %d, want 0", got)
	}
	if got := o.CallOnceSlow(fn, nil); got != 0 {
		t.Fatalf("second CallOnceSlow = %d, want 0", got)
	}
	if calls != 1 {
		t.Fatalf("calls = %d, want 1", calls)
	}
}

func TestRecursiveMutex(t *testing.T) {
	prev := currentThreadIdent
	currentThreadIdent = func() ThreadIdent { return 11 }
	t.Cleanup(func() { currentThreadIdent = prev })

	var m RecursiveMutex
	m.Lock()
	if !m.IsLockedByCurrentThread() {
		t.Fatal("recursive mutex should be owned")
	}
	m.Lock()
	if m.level != 1 {
		t.Fatalf("level = %d, want 1", m.level)
	}
	m.Unlock()
	m.Unlock()
}

func TestRWMutex(t *testing.T) {
	var m RWMutex
	m.RLock()
	readDone := make(chan struct{}, 1)
	go func() {
		m.RLock()
		readDone <- struct{}{}
		m.RUnlock()
	}()
	select {
	case <-readDone:
	case <-time.After(time.Second):
		t.Fatal("second reader should acquire")
	}
	m.RUnlock()

	m.Lock()
	writerDone := make(chan struct{}, 1)
	go func() {
		m.Lock()
		writerDone <- struct{}{}
		m.Unlock()
	}()
	time.Sleep(10 * time.Millisecond)
	m.Unlock()
	select {
	case <-writerDone:
	case <-time.After(time.Second):
		t.Fatal("writer should wake after unlock")
	}
}

func TestSeqLock(t *testing.T) {
	var s SeqLock
	s.LockWrite()
	s.UnlockWrite()
	prev := s.BeginRead()
	if prev%2 == 1 {
		t.Fatal("BeginRead should not return an updating sequence")
	}
	if !s.EndRead(prev) {
		t.Fatal("EndRead should succeed without intervening write")
	}
}
