package pydynload

import (
	"errors"
	"testing"
)

func TestEncodedNameASCII(t *testing.T) {
	encoded, prefix, err := EncodedName("pkg.mod-name")
	if err != nil {
		t.Fatalf("EncodedName() error = %v", err)
	}
	if encoded != "mod_name" {
		t.Fatalf("EncodedName() encoded = %q", encoded)
	}
	if prefix != asciiOnlyPrefix {
		t.Fatalf("EncodedName() prefix = %q", prefix)
	}
}

func TestEncodedNameNonASCII(t *testing.T) {
	encoded, prefix, err := EncodedName("pkg.m\xc3\xb3d")
	if err != nil {
		t.Fatalf("EncodedName() error = %v", err)
	}
	if encoded == "" {
		t.Fatalf("EncodedName() encoded empty")
	}
	if prefix != nonASCIIprefix {
		t.Fatalf("EncodedName() prefix = %q", prefix)
	}
}

func TestInitLoaderInfo(t *testing.T) {
	info, err := InitLoaderInfo("pkg.mod", "/tmp/mod.so", OriginDynamic)
	if err != nil {
		t.Fatalf("InitLoaderInfo() error = %v", err)
	}
	if info.Name != "pkg.mod" || info.Path != "/tmp/mod.so" || info.Origin != OriginDynamic {
		t.Fatalf("InitLoaderInfo() = %#v", info)
	}
	if info.HookPrefix != asciiOnlyPrefix || info.NameEncoded != "mod" {
		t.Fatalf("InitLoaderInfo encoded fields = %#v", info)
	}
}

func TestInitLoaderInfoBuiltin(t *testing.T) {
	info, err := InitLoaderInfoForBuiltin("builtins")
	if err != nil {
		t.Fatalf("InitLoaderInfoForBuiltin() error = %v", err)
	}
	if info.Origin != OriginBuiltin || info.Path != "builtins" {
		t.Fatalf("InitLoaderInfoForBuiltin() = %#v", info)
	}
}

func TestInitLoaderInfoCore(t *testing.T) {
	info, err := InitLoaderInfoForCore("sys")
	if err != nil {
		t.Fatalf("InitLoaderInfoForCore() error = %v", err)
	}
	if info.Origin != OriginCore {
		t.Fatalf("InitLoaderInfoForCore() = %#v", info)
	}
}

func TestFindSharedFuncptrUnsupported(t *testing.T) {
	_, err := FindSharedFuncptr("PyInit", "mod", "/tmp/mod.so")
	if !errors.Is(err, ErrUnsupported) {
		t.Fatalf("FindSharedFuncptr() error = %v, want ErrUnsupported", err)
	}
}

func TestDynLoadFiletab(t *testing.T) {
	tab := DynLoadFiletab()
	if tab == nil {
		t.Fatalf("DynLoadFiletab() returned nil slice")
	}
}
