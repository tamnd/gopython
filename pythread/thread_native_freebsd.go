//go:build freebsd

package pythread

import (
	"syscall"
	"unsafe"
)

const freebsdThrSelf = 432

var currentNativeThreadID = func() uint64 {
	var tid uint64
	_, _, _ = syscall.RawSyscall(freebsdThrSelf, uintptr(unsafe.Pointer(&tid)), 0, 0)
	return tid
}
