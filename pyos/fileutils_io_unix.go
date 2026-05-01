//go:build !windows

package pyos

import "syscall"

func readFDPlatform(fd int, buf []byte) (int, error) {
	return syscall.Read(fd, buf)
}

func writeFDPlatform(fd int, buf []byte) (int, error) {
	return syscall.Write(fd, buf)
}

func isInterrupted(err error) bool {
	return err == syscall.EINTR
}
