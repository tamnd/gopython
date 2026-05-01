//go:build windows

package pyos

import (
	"syscall"
	"unsafe"

	"golang.org/x/sys/windows"
)

var getHandleInformationProc = windows.NewLazySystemDLL("kernel32.dll").NewProc("GetHandleInformation")

func getHandleInformation(handle windows.Handle) (uint32, error) {
	var flags uint32
	r1, _, err := getHandleInformationProc.Call(uintptr(handle), uintptr(unsafe.Pointer(&flags)))
	if r1 == 0 {
		if err != nil && err != syscall.Errno(0) {
			return 0, err
		}
		return 0, syscall.EINVAL
	}
	return flags, nil
}

func getInheritable(fd int) (bool, error) {
	flags, err := getHandleInformation(windows.Handle(fd))
	if err != nil {
		return false, err
	}
	return flags&windows.HANDLE_FLAG_INHERIT != 0, nil
}

func setInheritable(fd int, inheritable bool) error {
	flags := uint32(0)
	if inheritable {
		flags = windows.HANDLE_FLAG_INHERIT
	}
	return windows.SetHandleInformation(
		windows.Handle(fd),
		windows.HANDLE_FLAG_INHERIT,
		flags,
	)
}

func isValidFD(fd int) bool {
	_, err := getHandleInformation(windows.Handle(fd))
	return err == nil
}
