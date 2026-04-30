//go:build !windows

package pythread

func runningOnValgrindEnvIsEnabled(value string) bool {
	return value != "0"
}
