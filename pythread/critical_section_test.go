package pythread

import "testing"

func TestCriticalSectionRecursiveLockSkipsRelock(t *testing.T) {
	interp := NewInterpreterState()
	ts := NewThreadState(interp)
	var mu Mutex
	var outer, inner CriticalSection

	ts.BeginCriticalSection(&outer, &mu)
	if mu.LockFast() {
		t.Fatal("mutex should be held by outer critical section")
	}

	ts.BeginCriticalSectionSlow(&inner, &mu)
	if inner.mutex != nil {
		t.Fatal("inner critical section should skip relocking")
	}

	ts.EndCriticalSection(&inner)
	if mu.LockFast() {
		t.Fatal("outer lock should still be held")
	}

	ts.EndCriticalSection(&outer)
	if !mu.LockFast() {
		t.Fatal("mutex should be released after outer end")
	}
	mu.Unlock()
}

func TestCriticalSectionSuspendAndResume(t *testing.T) {
	interp := NewInterpreterState()
	ts := NewThreadState(interp)
	var mu Mutex
	var cs CriticalSection

	ts.BeginCriticalSection(&cs, &mu)
	ts.SuspendCriticalSections()

	if !mu.LockFast() {
		t.Fatal("mutex should be released while suspended")
	}
	mu.Unlock()

	ts.ResumeCriticalSection()
	if mu.LockFast() {
		t.Fatal("mutex should be re-locked after resume")
	}

	ts.EndCriticalSection(&cs)
	if !mu.LockFast() {
		t.Fatal("mutex should be released after end")
	}
	mu.Unlock()
}

func TestCriticalSection2LocksInStableOrder(t *testing.T) {
	interp := NewInterpreterState()
	ts := NewThreadState(interp)
	var mu1, mu2 Mutex
	var cs CriticalSection2

	ts.BeginCriticalSection2(&cs, &mu2, &mu1)
	if cs.base.mutex == nil || cs.mutex2 == nil {
		t.Fatal("two-mutex critical section should hold both locks")
	}
	if mu1.LockFast() || mu2.LockFast() {
		t.Fatal("both mutexes should be locked")
	}

	ts.EndCriticalSection2(&cs)
	if !mu1.LockFast() {
		t.Fatal("first mutex should be released")
	}
	mu1.Unlock()
	if !mu2.LockFast() {
		t.Fatal("second mutex should be released")
	}
	mu2.Unlock()
}

func TestCriticalSectionSkipsLockWhenWorldStopped(t *testing.T) {
	interp := NewInterpreterState()
	interp.StopTheWorld.WorldStopped = true
	ts := NewThreadState(interp)
	var mu Mutex
	var cs CriticalSection

	ts.BeginCriticalSectionSlow(&cs, &mu)
	if cs.mutex != nil {
		t.Fatal("world-stopped path should not lock the mutex")
	}
	if !mu.LockFast() {
		t.Fatal("mutex should remain unlocked")
	}
	mu.Unlock()
}
