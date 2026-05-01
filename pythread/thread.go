package pythread

import (
	"bytes"
	"errors"
	"fmt"
	"math"
	"runtime"
	"sync"
	"sync/atomic"
	"time"
)

const (
	UnsetTimeout = int64(-1)
)

var (
	threadInitialized atomic.Bool
	threadStackSize   atomic.Uint64
	nextThreadID      atomic.Uint64
	threadInitHook    = func() {}
	threadIDs         sync.Map
)

type LockStatus int

const (
	LockFailure LockStatus = iota
	LockAcquired
	LockIntr
)

type TimedLocker interface {
	AcquireTimed(timeoutMicroseconds int64, detach bool) LockStatus
}

type TSSKey struct {
	initialized bool
	values      sync.Map
}

type Info struct {
	Name    string
	Lock    string
	Version string
}

type ThreadIdent uint64

const InvalidThreadID ThreadIdent = 0

type ThreadHandle struct {
	done     chan struct{}
	mu       sync.Mutex
	detached bool
	joined   bool
}

type tlsKey struct {
	values sync.Map
}

var (
	tlsKeysMu  sync.Mutex
	tlsKeys    = map[int]*tlsKey{}
	nextTLSKey = 1
)

func InitThread() {
	if !threadInitialized.CompareAndSwap(false, true) {
		return
	}
	if nextThreadID.Load() == 0 {
		nextThreadID.Store(1)
	}
	threadIDs.Store(currentGoroutineID(), ThreadIdent(1))
	threadInitHook()
}

func GetStackSize() uint64 {
	return threadStackSize.Load()
}

func SetStackSize(size uint64) int {
	return setStackSizePlatform(size)
}

func ParseTimeoutSeconds(seconds *float64, blocking bool) (int64, error) {
	if seconds == nil {
		if blocking {
			return UnsetTimeout, nil
		}
		return 0, nil
	}
	if !blocking {
		return 0, errors.New("can't specify a timeout for a non-blocking call")
	}
	if *seconds < 0 {
		return 0, errors.New("timeout value must be a non-negative number")
	}
	if math.IsNaN(*seconds) || math.IsInf(*seconds, 0) {
		return 0, errors.New("timeout value is too large")
	}
	if *seconds > float64(timeoutMaxMicroseconds())/1e6 {
		return 0, errors.New("timeout value is too large")
	}

	timeout := int64(math.Ceil(*seconds * float64(time.Second)))
	return timeout, nil
}

func AcquireLockTimedWithRetries(lock TimedLocker, timeout int64, pending func() error) LockStatus {
	endtime := int64(0)
	if timeout > 0 {
		endtime = time.Now().UnixNano() + timeout
	}

	var status LockStatus
	for {
		microseconds := timeoutToMicroseconds(timeout)
		status = lock.AcquireTimed(0, false)
		if status == LockFailure && microseconds != 0 {
			status = lock.AcquireTimed(microseconds, true)
		}
		if status != LockIntr {
			return status
		}
		if pending != nil {
			if err := pending(); err != nil {
				return LockIntr
			}
		}
		if timeout > 0 {
			timeout = endtime - time.Now().UnixNano()
			if timeout < 0 {
				return LockFailure
			}
		}
	}
}

func AllocTSS() *TSSKey {
	return &TSSKey{}
}

func FreeTSS(key *TSSKey) {
	if key != nil {
		DeleteTSS(key)
	}
}

func IsTSSCreated(key *TSSKey) bool {
	if key == nil {
		return false
	}
	return key.initialized
}

func CreateTSS(key *TSSKey) int {
	if key == nil {
		return -1
	}
	key.initialized = true
	return 0
}

func DeleteTSS(key *TSSKey) {
	if key != nil {
		key.initialized = false
		key.values = sync.Map{}
	}
}

func SetTSS(key *TSSKey, value any) int {
	if key == nil || !key.initialized {
		return -1
	}
	key.values.Store(currentThreadIdent(), value)
	return 0
}

func GetTSS(key *TSSKey) any {
	if key == nil || !key.initialized {
		return nil
	}
	value, _ := key.values.Load(currentThreadIdent())
	return value
}

func GetInfo() Info {
	name := "pthread"
	lock := "mutex+cond"
	if runtime.GOOS == "windows" {
		name = "nt"
		lock = ""
	}
	if runtime.GOOS == "js" {
		name = "pthread-stubs"
		lock = ""
	}
	return Info{
		Name:    name,
		Lock:    lock,
		Version: "",
	}
}

func CreateTLSKey() int {
	tlsKeysMu.Lock()
	defer tlsKeysMu.Unlock()

	key := nextTLSKey
	nextTLSKey++
	tlsKeys[key] = &tlsKey{}
	return key
}

func DeleteTLSKey(key int) {
	tlsKeysMu.Lock()
	delete(tlsKeys, key)
	tlsKeysMu.Unlock()
}

func DeleteTLSKeyValue(key int) {
	tlsKeysMu.Lock()
	entry := tlsKeys[key]
	tlsKeysMu.Unlock()
	if entry == nil {
		return
	}
	entry.values.Delete(currentThreadIdent())
}

func SetTLSKeyValue(key int, value any) int {
	tlsKeysMu.Lock()
	entry := tlsKeys[key]
	tlsKeysMu.Unlock()
	if entry == nil {
		return -1
	}
	entry.values.Store(currentThreadIdent(), value)
	return 0
}

func GetTLSKeyValue(key int) any {
	tlsKeysMu.Lock()
	entry := tlsKeys[key]
	tlsKeysMu.Unlock()
	if entry == nil {
		return nil
	}
	value, _ := entry.values.Load(currentThreadIdent())
	return value
}

func ReInitTLS() {}

func defaultCurrentThreadIdent() ThreadIdent {
	if !threadInitialized.Load() {
		InitThread()
	}
	gid := currentGoroutineID()
	if ident, ok := threadIDs.Load(gid); ok {
		return ident.(ThreadIdent)
	}
	ident := ThreadIdent(nextThreadID.Add(1))
	actual, _ := threadIDs.LoadOrStore(gid, ident)
	return actual.(ThreadIdent)
}

func GetThreadIdentEx() ThreadIdent {
	if !threadInitialized.Load() {
		InitThread()
	}
	return currentThreadIdent()
}

func GetThreadIdent() uint64 {
	return uint64(GetThreadIdentEx())
}

func GetThreadNativeID() uint64 {
	return currentNativeThreadID()
}

func ExitThread() {
	if !threadInitialized.Load() {
		panic("exit 0")
	}
	runtime.Goexit()
}

func HangThread() {
	for {
		time.Sleep(24 * time.Hour)
	}
}

func StartJoinableThread(fn func(any), arg any) (ThreadIdent, *ThreadHandle, error) {
	if fn == nil {
		return InvalidThreadID, nil, errors.New("thread function must not be nil")
	}
	if !threadInitialized.Load() {
		InitThread()
	}

	handle := &ThreadHandle{done: make(chan struct{})}
	id := ThreadIdent(nextThreadID.Add(1))
	go func() {
		runtime.LockOSThread()
		defer runtime.UnlockOSThread()
		gid := currentGoroutineID()
		threadIDs.Store(gid, id)
		defer threadIDs.Delete(gid)
		defer close(handle.done)
		fn(arg)
	}()
	return id, handle, nil
}

func StartNewThread(fn func(any), arg any) ThreadIdent {
	id, handle, err := StartJoinableThread(fn, arg)
	if err != nil {
		return InvalidThreadID
	}
	_ = handle.Detach()
	return id
}

func (h *ThreadHandle) Join() error {
	if h == nil {
		return errors.New("thread handle is nil")
	}

	h.mu.Lock()
	if h.detached {
		h.mu.Unlock()
		return errors.New("thread handle is detached")
	}
	if h.joined {
		h.mu.Unlock()
		return errors.New("thread handle already joined")
	}
	h.joined = true
	done := h.done
	h.mu.Unlock()

	<-done
	return nil
}

func (h *ThreadHandle) Detach() error {
	if h == nil {
		return errors.New("thread handle is nil")
	}

	h.mu.Lock()
	defer h.mu.Unlock()

	if h.joined {
		return errors.New("thread handle already joined")
	}
	if h.detached {
		return nil
	}
	h.detached = true
	return nil
}

func timeoutMaxMicroseconds() int64 {
	if runtime.GOOS == "windows" {
		const maxWaitMilliseconds = int64(0xFFFFFFFE)
		if maxWaitMilliseconds < math.MaxInt64/1000 {
			return maxWaitMilliseconds * 1000
		}
	}
	return math.MaxInt64
}

func timeoutToMicroseconds(timeout int64) int64 {
	if timeout < 0 {
		return timeout
	}
	return int64(math.Ceil(float64(timeout) / 1000))
}

func currentGoroutineID() uint64 {
	var buf [64]byte
	n := runtime.Stack(buf[:], false)
	prefix := []byte("goroutine ")
	stack := buf[:n]
	if !bytes.HasPrefix(stack, prefix) {
		panic("unexpected runtime stack header")
	}
	stack = stack[len(prefix):]
	end := bytes.IndexByte(stack, ' ')
	if end < 0 {
		panic("unexpected runtime stack header")
	}
	var gid uint64
	if _, err := fmt.Sscanf(string(stack[:end]), "%d", &gid); err != nil {
		panic("unexpected runtime stack header")
	}
	return gid
}
