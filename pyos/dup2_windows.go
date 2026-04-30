//go:build windows

package pyos

import (
	"syscall"
)

var (
	msvcrtDLL    = syscall.NewLazyDLL("msvcrt.dll")
	procCRTdup2  = msvcrtDLL.NewProc("_dup2")
	procCRTclose = msvcrtDLL.NewProc("_close")
)

func dup2(fd1, fd2 int) (int, error) {
	if fd1 == fd2 {
		return fd2, nil
	}
	r1, _, callErr := procCRTdup2.Call(uintptr(fd1), uintptr(fd2))
	if int32(r1) == -1 {
		if callErr != syscall.Errno(0) {
			return -1, callErr
		}
		return -1, syscall.EINVAL
	}
	return fd2, nil
}
