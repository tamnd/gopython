//go:build unix

package pyos

import (
	"io/fs"

	"golang.org/x/sys/unix"
)

func openFD(path string, flag int, perm fs.FileMode, retryOnEINTR bool) (int, error) {
	openFlags := flag | unix.O_CLOEXEC
	for {
		fd, err := unix.Open(path, openFlags, uint32(perm.Perm()))
		if err == nil {
			if err := setInheritable(fd, false); err != nil {
				unix.Close(fd)
				return -1, err
			}
			return fd, nil
		}
		if retryOnEINTR && err == unix.EINTR {
			continue
		}
		return -1, err
	}
}
