package pythread

import (
	"errors"
	"math"
	"runtime"
	"sync/atomic"
	"time"
)

const (
	UnsetTimeout = int64(-1)
)

var (
	threadInitialized atomic.Bool
	threadStackSize   atomic.Uint64
	threadInitHook    = func() {}
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
}

type Info struct {
	Name    string
	Lock    string
	Version string
}

func InitThread() {
	if !threadInitialized.CompareAndSwap(false, true) {
		return
	}
	threadInitHook()
}

func GetStackSize() uint64 {
	return threadStackSize.Load()
}

func SetStackSize(uint64) int {
	return -2
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
	}
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
