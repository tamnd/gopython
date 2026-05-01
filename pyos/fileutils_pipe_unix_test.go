//go:build !windows

package pyos

import "syscall"

func createPipe() (int, int, error) {
	var fds [2]int
	if err := syscall.Pipe(fds[:]); err != nil {
		return 0, 0, err
	}
	return fds[0], fds[1], nil
}

func closePipeFD(fd int) error {
	return syscall.Close(fd)
}
