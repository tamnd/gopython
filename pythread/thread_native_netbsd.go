//go:build netbsd

package pythread

import "syscall"

const netbsdLWPSelf = 311

var currentNativeThreadID = func() uint64 {
	tid, _, _ := syscall.RawSyscall(netbsdLWPSelf, 0, 0, 0)
	return uint64(tid)
}
