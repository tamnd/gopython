//go:build !windows && !darwin && !linux

package pydynload

func dynLoadFiletab() []string {
	return nil
}

func findSharedFuncptr(prefix, shortname, pathname string) (uintptr, error) {
	return 0, ErrUnsupported
}
