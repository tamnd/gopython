//go:build unix

package pyos

import "golang.org/x/term"

func deviceEncoding(fd int) string {
	if !term.IsTerminal(fd) {
		return ""
	}
	return LocaleEncoding()
}
