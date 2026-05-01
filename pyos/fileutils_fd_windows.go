//go:build windows

package pyos

import "os"

func FstatNoRaise(fd int) (StatInfo, error) {
	file := os.NewFile(uintptr(fd), "")
	if file == nil {
		return StatInfo{}, os.ErrInvalid
	}
	info, err := file.Stat()
	if err != nil {
		return StatInfo{}, err
	}
	return StatInfo{
		Mode: info.Mode(),
		Size: info.Size(),
	}, nil
}

func Fstat(fd int) (StatInfo, error) {
	return FstatNoRaise(fd)
}

func closeFD(fd int) error {
	file := os.NewFile(uintptr(fd), "")
	if file == nil {
		return os.ErrInvalid
	}
	return file.Close()
}
