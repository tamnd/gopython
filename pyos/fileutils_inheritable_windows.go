//go:build windows

package pyos

import "golang.org/x/sys/windows"

func getInheritable(fd int) (bool, error) {
	var flags uint32
	err := windows.GetHandleInformation(windows.Handle(fd), &flags)
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
