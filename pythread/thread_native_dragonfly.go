//go:build dragonfly

package pythread

import "syscall"

const dragonflyLWPGettid = 496

var currentNativeThreadID = func() uint64 {
	tid, _, _ := syscall.RawSyscall(dragonflyLWPGettid, 0, 0, 0)
	return uint64(tid)
}
