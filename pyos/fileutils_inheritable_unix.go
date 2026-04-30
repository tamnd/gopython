//go:build unix

package pyos

import "golang.org/x/sys/unix"

func getInheritable(fd int) (bool, error) {
	flags, err := unix.FcntlInt(uintptr(fd), unix.F_GETFD, 0)
	if err != nil {
		return false, err
	}
	return flags&unix.FD_CLOEXEC == 0, nil
}

func setInheritable(fd int, inheritable bool) error {
	flags, err := unix.FcntlInt(uintptr(fd), unix.F_GETFD, 0)
	if err != nil {
		return err
	}
	newFlags := flags
	if inheritable {
		newFlags &^= unix.FD_CLOEXEC
	} else {
		newFlags |= unix.FD_CLOEXEC
	}
	if newFlags == flags {
		return nil
	}
	_, err = unix.FcntlInt(uintptr(fd), unix.F_SETFD, newFlags)
	return err
}

func isValidFD(fd int) bool {
	_, err := unix.FcntlInt(uintptr(fd), unix.F_GETFD, 0)
	return err == nil
}
