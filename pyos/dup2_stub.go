//go:build !unix && !windows

package pyos

import "errors"

var ErrDup2Unsupported = errors.New("dup2 is unsupported on this platform")

func dup2(fd1, fd2 int) (int, error) {
	return -1, ErrDup2Unsupported
}
