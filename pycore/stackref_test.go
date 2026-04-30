package pycore

import (
	"strings"
	"testing"
)

func TestStackRefCreateGetAndClose(t *testing.T) {
	reg := NewStackRefRegistry(true)
	obj := &struct{ name string }{name: "value"}

	ref := reg.Create(obj, "testObject", "stackref_test.go", 10)
	if got := reg.GetObject(ref); got != obj {
		t.Fatal("GetObject returned the wrong object")
	}
	if !reg.Is(ref, ref) {
		t.Fatal("Is should compare underlying object identity")
	}

	if got := reg.Close(ref, "stackref_test.go", 20); got != obj {
		t.Fatal("Close returned the wrong object")
	}
}

func TestStackRefBuiltinAssociationUsesReservedSlot(t *testing.T) {
	reg := NewStackRefRegistry(false)
	obj := &struct{ name string }{name: "none"}

	reg.AssociateBuiltin(obj, "NoneType", StackRefNone)
	if got := reg.GetObject(StackRefNone); got != obj {
		t.Fatal("reserved stackref should resolve to associated builtin")
	}
	if got := reg.Close(StackRefNone, "stackref_test.go", 30); got != obj {
		t.Fatal("closing builtin stackref should keep the associated object")
	}
	if got := reg.GetObject(StackRefNone); got != obj {
		t.Fatal("builtin stackref should remain associated after close")
	}
}

func TestStackRefBorrowAndLeakReport(t *testing.T) {
	reg := NewStackRefRegistry(false)
	obj := &struct{ name string }{name: "borrowed"}

	ref := reg.Create(obj, "testObject", "stackref_test.go", 40)
	reg.RecordBorrow(ref, "stackref_test.go", 41)

	err := reg.ReportLeaks(func(any) bool { return false })
	if err == nil {
		t.Fatal("ReportLeaks should detect the unclosed stackref")
	}
	if !strings.Contains(err.Error(), "Last borrow at stackref_test.go:41") {
		t.Fatalf("unexpected leak message: %v", err)
	}
}

func TestStackRefDoubleCloseReportsClosedSite(t *testing.T) {
	reg := NewStackRefRegistry(true)
	ref := reg.Create(&struct{}{}, "testObject", "stackref_test.go", 50)
	reg.Close(ref, "stackref_test.go", 51)

	defer func() {
		recovered := recover()
		if recovered == nil {
			t.Fatal("expected double-close panic")
		}
		msg := recovered.(string)
		if !strings.Contains(msg, "Double close of ref ID") {
			t.Fatalf("unexpected panic: %s", msg)
		}
		if !strings.Contains(msg, "Closed at stackref_test.go:51") {
			t.Fatalf("panic should report the original close site: %s", msg)
		}
	}()

	reg.Close(ref, "stackref_test.go", 52)
}

func TestTaggedIntRoundTrip(t *testing.T) {
	for _, value := range []int64{-7, 0, 19} {
		ref := TagInt(value)
		if !IsTaggedInt(ref) {
			t.Fatalf("value %d should produce a tagged int", value)
		}
		if got := UntagInt(ref); got != value {
			t.Fatalf("untagged value = %d, want %d", got, value)
		}
	}
	if !IsNullOrInt(StackRefNull) {
		t.Fatal("null should satisfy IsNullOrInt")
	}
}
