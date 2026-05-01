//go:build windows

package pythread

func expectedRunningOnValgrind(value string) int {
	if value == "0" {
		return 1
	}
	return 0
}
