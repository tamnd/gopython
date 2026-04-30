package pycore

import "testing"

func TestStdlibModuleNames(t *testing.T) {
	if len(StdlibModuleNames) != 297 {
		t.Fatalf("len = %d", len(StdlibModuleNames))
	}
	for i := 1; i < len(StdlibModuleNames); i++ {
		if StdlibModuleNames[i-1] >= StdlibModuleNames[i] {
			t.Fatalf("names not strictly sorted at %d", i)
		}
	}
	for _, name := range []string{
		"__future__",
		"_ast",
		"_zstd",
		"abc",
		"sys",
		"zoneinfo",
	} {
		if !IsStdlibModule(name) {
			t.Fatalf("%q should be a stdlib module", name)
		}
	}
	for _, name := range []string{"", "requests", "numpy", "not_in_stdlib"} {
		if IsStdlibModule(name) {
			t.Fatalf("%q should not be a stdlib module", name)
		}
	}
}
