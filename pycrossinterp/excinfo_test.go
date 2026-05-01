package pycrossinterp

import "testing"

func TestExcInfoFormat(t *testing.T) {
	info := ExcInfo{
		Type:    ExcInfoType{Name: "ValueError", QualName: "ValueError", Module: "builtins"},
		Message: "bad",
	}
	if got := info.Format(); got != "ValueError: bad" {
		t.Fatalf("Format = %q", got)
	}
	info.Type.Module = "pkg"
	if got := info.Format(); got != "pkg.ValueError: bad" {
		t.Fatalf("Format = %q", got)
	}
}

func TestExcInfoFromObjectAndApply(t *testing.T) {
	info, err := InitExcInfoFromObject(ExcSnapshotObject{
		Type: map[string]string{
			"__name__":     "TypeError",
			"__qualname__": "TypeError",
			"__module__":   "builtins",
		},
		Msg:        "boom",
		ErrDisplay: "trace",
	})
	if err != nil {
		t.Fatalf("InitExcInfoFromObject returned error: %v", err)
	}
	if got := info.Apply(); got != "trace" {
		t.Fatalf("Apply = %q", got)
	}
	obj := info.AsObject()
	if obj.Formatted == "" || obj.Type["__name__"] != "TypeError" {
		t.Fatalf("AsObject = %#v", obj)
	}
}

func TestExcInfoClearAndIsSet(t *testing.T) {
	info := InitExcInfoFromException("RuntimeError", "", "__main__", "bad", "bad\n")
	if !info.IsSet() {
		t.Fatal("info should be set")
	}
	if info.ErrDisplay != "bad" {
		t.Fatalf("ErrDisplay = %q", info.ErrDisplay)
	}
	info.Clear()
	if info.IsSet() {
		t.Fatal("info should be cleared")
	}
}
