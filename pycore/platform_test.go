package pycore

import (
	"runtime"
	"strings"
	"testing"
)

func TestConfigDictGet(t *testing.T) {
	dict := map[string]any{"answer": 42}
	item, err := ConfigDictGet(dict, "answer")
	if err != nil {
		t.Fatal(err)
	}
	if item != 42 {
		t.Fatalf("item = %v", item)
	}
	if _, err := ConfigDictGet(dict, "missing"); err == nil ||
		err.Error() != "missing config key: missing" {
		t.Fatalf("missing err = %v", err)
	}
	if err := ConfigDictInvalidType("x"); err == nil ||
		err.Error() != "invalid config type: x" {
		t.Fatalf("type err = %v", err)
	}
}

func TestPythonDirectoryReadme(t *testing.T) {
	want := "Miscellaneous source files for the main Python shared library.\n"
	if PythonDirectoryReadme != want {
		t.Fatalf("README = %q", PythonDirectoryReadme)
	}
}

func TestVersionConstants(t *testing.T) {
	if PYVersion != "3.14.4+" {
		t.Fatalf("PYVersion = %q", PYVersion)
	}
	if PYVersionHex != 0x030e04f0 {
		t.Fatalf("PYVersionHex = %#x", PYVersionHex)
	}
}

func TestPlatformMetadata(t *testing.T) {
	if !strings.Contains(PyGetCopyright(), "Python Software Foundation") {
		t.Fatalf("copyright = %q", PyGetCopyright())
	}
	if PyGetCompiler() != "[Go "+runtime.Version()+"]" {
		t.Fatalf("compiler = %q", PyGetCompiler())
	}
	if PyGetPlatform() != runtime.GOOS {
		t.Fatalf("platform = %q", PyGetPlatform())
	}
	if PyGetBuildInfo() != "main, Jan 01 1970, 00:00:00" {
		t.Fatalf("build info = %q", PyGetBuildInfo())
	}
	if !strings.HasPrefix(PyGetVersion(), "3.14.4+ (main, Jan 01 1970, 00:00:00)") {
		t.Fatalf("version = %q", PyGetVersion())
	}
}

func TestPyFPEdummy(t *testing.T) {
	if PyFPEdummy(nil) != 1.0 {
		t.Fatal("PyFPEdummy mismatch")
	}
}
