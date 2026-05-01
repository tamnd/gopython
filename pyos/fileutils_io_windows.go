//go:build windows

package pyos

import "syscall"

func readFDPlatform(fd int, buf []byte) (int, error) {
	return syscall.Read(syscall.Handle(fd), buf)
}

func writeFDPlatform(fd int, buf []byte) (int, error) {
	return syscall.Write(syscall.Handle(fd), buf)
}

func isInterrupted(err error) bool {
	return err == syscall.EINTR
}
