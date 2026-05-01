//go:build darwin

package pythread

import (
	"syscall"
)

const darwinThreadSelfID = 372

var currentNativeThreadID = func() uint64 {
	tid, _, _ := syscall.RawSyscall(darwinThreadSelfID, 0, 0, 0)
	return uint64(tid)
}
