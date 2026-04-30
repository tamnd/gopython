//go:build windows

package pyos

import (
	"fmt"

	"golang.org/x/sys/windows"
)

func deviceEncoding(fd int) string {
	handle := windows.Handle(fd)
	var mode uint32
	if err := windows.GetConsoleMode(handle, &mode); err != nil {
		return ""
	}
	var cp uint32
	var err error
	if fd == 0 {
		cp, err = windows.GetConsoleCP()
	} else {
		cp, err = windows.GetConsoleOutputCP()
	}
	if err != nil || cp == 0 {
		return ""
	}
	return fmt.Sprintf("cp%d", cp)
}
