//go:build windows

package pydynload

func dynLoadFiletab() []string {
	return []string{".cp314-win_amd64.pyd", ".pyd"}
}

func findSharedFuncptr(prefix, shortname, pathname string) (uintptr, error) {
	return 0, ErrUnsupported
}
