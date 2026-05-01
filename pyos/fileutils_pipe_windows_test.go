//go:build windows

package pyos

import "syscall"

func createPipe() (int, int, error) {
	var fds [2]syscall.Handle
	if err := syscall.Pipe(fds[:]); err != nil {
		return 0, 0, err
	}
	return int(fds[0]), int(fds[1]), nil
}

func closePipeFD(fd int) error {
	return syscall.Close(syscall.Handle(fd))
}
