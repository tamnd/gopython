package pythread

import (
	"errors"
	"runtime"
	"sync/atomic"
	"testing"
)

type fakeTimedLocker struct {
	results  []LockStatus
	timeouts []int64
	detachs  []bool
}

func (l *fakeTimedLocker) AcquireTimed(timeoutMicroseconds int64, detach bool) LockStatus {
	l.timeouts = append(l.timeouts, timeoutMicroseconds)
	l.detachs = append(l.detachs, detach)
	status := l.results[0]
	l.results = l.results[1:]
	return status
}

func TestInitThreadRunsOnlyOnce(t *testing.T) {
	threadInitialized = atomic.Bool{}
	calls := 0
	threadInitHook = func() { calls++ }
	t.Cleanup(func() { threadInitHook = func() {} })

	InitThread()
	InitThread()

	if calls != 1 {
		t.Fatalf("init hook calls = %d, want 1", calls)
	}
}

func TestParseTimeoutSeconds(t *testing.T) {
	if got, err := ParseTimeoutSeconds(nil, true); err != nil || got != UnsetTimeout {
		t.Fatalf("blocking none = (%d, %v), want (%d, nil)", got, err, UnsetTimeout)
	}
	if got, err := ParseTimeoutSeconds(nil, false); err != nil || got != 0 {
		t.Fatalf("non-blocking none = (%d, %v), want (0, nil)", got, err)
	}

	value := 0.0015
	got, err := ParseTimeoutSeconds(&value, true)
	if err != nil {
		t.Fatalf("ParseTimeoutSeconds returned error: %v", err)
	}
	if got != 1500000 {
		t.Fatalf("timeout = %d, want 1500000", got)
	}
}

func TestParseTimeoutSecondsErrors(t *testing.T) {
	value := 1.0
	if _, err := ParseTimeoutSeconds(&value, false); err == nil {
		t.Fatal("expected non-blocking timeout error")
	}

	negative := -1.0
	if _, err := ParseTimeoutSeconds(&negative, true); err == nil {
		t.Fatal("expected negative timeout error")
	}

	overflow := float64(timeoutMaxMicroseconds())/1e6 + 1
	if _, err := ParseTimeoutSeconds(&overflow, true); err == nil {
		t.Fatal("expected overflow error")
	}
}

func TestAcquireLockTimedWithRetries(t *testing.T) {
	lock := &fakeTimedLocker{
		results: []LockStatus{LockFailure, LockIntr, LockFailure, LockAcquired},
	}
	pendingCalls := 0

	status := AcquireLockTimedWithRetries(lock, 2_000_000, func() error {
		pendingCalls++
		return nil
	})

	if status != LockAcquired {
		t.Fatalf("status = %v, want LockAcquired", status)
	}
	if pendingCalls != 1 {
		t.Fatalf("pending calls = %d, want 1", pendingCalls)
	}
	if len(lock.timeouts) < 4 {
		t.Fatalf("unexpected AcquireTimed call count: %d", len(lock.timeouts))
	}
	if lock.timeouts[0] != 0 || lock.detachs[0] {
		t.Fatal("first call should be non-blocking without detach")
	}
	if lock.timeouts[1] <= 0 || !lock.detachs[1] {
		t.Fatal("second call should be timed and detached")
	}
}

func TestAcquireLockTimedWithRetriesPropagatesPendingError(t *testing.T) {
	lock := &fakeTimedLocker{results: []LockStatus{LockIntr}}
	status := AcquireLockTimedWithRetries(lock, 0, func() error {
		return errors.New("stop")
	})
	if status != LockIntr {
		t.Fatalf("status = %v, want LockIntr", status)
	}
}

func TestTSSLifecycle(t *testing.T) {
	key := AllocTSS()
	if key == nil {
		t.Fatal("AllocTSS returned nil")
	}
	if IsTSSCreated(key) {
		t.Fatal("new key should not be initialized")
	}
	if CreateTSS(key) != 0 {
		t.Fatal("CreateTSS should succeed")
	}
	if !IsTSSCreated(key) {
		t.Fatal("key should be initialized after CreateTSS")
	}
	FreeTSS(key)
	if IsTSSCreated(key) {
		t.Fatal("key should be cleared after FreeTSS")
	}
}

func TestGetInfo(t *testing.T) {
	info := GetInfo()
	if info.Name == "" {
		t.Fatal("thread name should not be empty")
	}
	if runtime.GOOS == "windows" && info.Name != "nt" {
		t.Fatalf("windows thread name = %q, want nt", info.Name)
	}
	if runtime.GOOS != "windows" && runtime.GOOS != "js" && info.Lock != "mutex+cond" {
		t.Fatalf("lock name = %q, want mutex+cond", info.Lock)
	}
}
