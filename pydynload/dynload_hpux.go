//go:build hpux

package pydynload

func dynLoadFiletab() []string {
	return []string{".sl"}
}

func findSharedFuncptr(prefix, shortname, pathname string) (uintptr, error) {
	return 0, ErrUnsupported
}
