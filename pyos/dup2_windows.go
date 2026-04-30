//go:build windows

package pyos

import (
	"syscall"
	"unsafe"
)

var (
	msvcrtDLL       = syscall.NewLazyDLL("msvcrt.dll")
	procCRTdup2     = msvcrtDLL.NewProc("_dup2")
	procCRTclose    = msvcrtDLL.NewProc("_close")
	procCRTgetErrno = msvcrtDLL.NewProc("_errno")
)

func dup2(fd1, fd2 int) (int, error) {
	if fd1 == fd2 {
		return fd2, nil
	}
	r1, _, callErr := procCRTdup2.Call(uintptr(fd1), uintptr(fd2))
	if int32(r1) == -1 {
		if errno, ok := crtErrno(); ok {
			return -1, errno
		}
		if callErr != syscall.Errno(0) {
			return -1, callErr
		}
		return -1, syscall.EINVAL
	}
	return fd2, nil
}

func crtErrno() (syscall.Errno, bool) {
	ptr, _, err := procCRTgetErrno.Call()
	if ptr == 0 {
		return 0, false
	}
	value := *(*int32)(unsafe.Pointer(ptr))
	if value == 0 {
		if err != syscall.Errno(0) {
			return 0, false
		}
		return 0, false
	}
	return syscall.Errno(value), true
}
