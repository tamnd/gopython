//go:build linux || darwin

package pydynload

func dynLoadFiletab() []string {
	return []string{".cpython-314.so", ".abi3.so", ".so"}
}

func findSharedFuncptr(prefix, shortname, pathname string) (uintptr, error) {
	return 0, ErrUnsupported
}
