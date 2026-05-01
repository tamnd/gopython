//go:build openbsd

package pythread

import "syscall"

const openbsdGetthrid = 299

var currentNativeThreadID = func() uint64 {
	tid, _, _ := syscall.RawSyscall(openbsdGetthrid, 0, 0, 0)
	return uint64(tid)
}
