//go:build unix

package pyos

import "golang.org/x/sys/unix"

func dup2(fd1, fd2 int) (int, error) {
	if fd1 != fd2 {
		if _, err := unix.FcntlInt(uintptr(fd1), unix.F_GETFL, 0); err != nil {
			return -1, err
		}
		if _, err := unix.FcntlInt(uintptr(fd2), unix.F_GETFL, 0); err == nil {
			_ = unix.Close(fd2)
		}
		newfd, err := unix.FcntlInt(uintptr(fd1), unix.F_DUPFD, fd2)
		if err != nil {
			return -1, err
		}
		return newfd, nil
	}
	return fd2, nil
}
