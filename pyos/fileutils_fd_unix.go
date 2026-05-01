//go:build !windows

package pyos

import (
	"io/fs"
	"syscall"
)

func FstatNoRaise(fd int) (StatInfo, error) {
	var stat syscall.Stat_t
	if err := syscall.Fstat(fd, &stat); err != nil {
		return StatInfo{}, err
	}
	return StatInfo{
		Mode: fs.FileMode(stat.Mode),
		Size: stat.Size,
	}, nil
}

func Fstat(fd int) (StatInfo, error) {
	return FstatNoRaise(fd)
}

func closeFD(fd int) error {
	return syscall.Close(fd)
}
