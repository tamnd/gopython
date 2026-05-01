//go:build windows

package pyos

import (
	"io/fs"

	"golang.org/x/sys/windows"
)

func openFD(path string, flag int, perm fs.FileMode, retryOnEINTR bool) (int, error) {
	_ = retryOnEINTR
	handle, err := windows.Open(path, flag, uint32(perm.Perm()))
	if err != nil {
		return -1, err
	}
	if err := setInheritable(int(handle), false); err != nil {
		windows.CloseHandle(handle)
		return -1, err
	}
	return int(handle), nil
}
