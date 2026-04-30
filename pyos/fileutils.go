package pyos

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
	"syscall"
)

type StatInfo struct {
	Mode fs.FileMode
	Size int64
}

func DeviceEncoding(fd int) string {
	if fd < 0 {
		return ""
	}
	if runtime.GOOS == "windows" {
		return ""
	}
	return "utf-8"
}

func GetCwd() (string, error) {
	return os.Getwd()
}

func IsAbs(path string) bool {
	return filepath.IsAbs(path)
}

func AbsPath(path string) (string, error) {
	if path == "" || path == "." {
		return GetCwd()
	}
	if IsAbs(path) {
		return path, nil
	}
	cwd, err := GetCwd()
	if err != nil {
		return "", err
	}
	return filepath.Join(cwd, path), nil
}

func ReadLink(path string) (string, error) {
	return os.Readlink(path)
}

func RealPath(path string) (string, error) {
	return filepath.EvalSymlinks(path)
}

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

func StatPath(path string) (fs.FileInfo, error) {
	return os.Stat(path)
}

func LstatPath(path string) (fs.FileInfo, error) {
	return os.Lstat(path)
}

func WStat(path string) (fs.FileInfo, error) {
	return os.Stat(path)
}

func SetInheritable(fd int, inheritable bool) error {
	if fd < 0 {
		return errors.New("invalid file descriptor")
	}
	return nil
}
