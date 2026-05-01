package pycrossinterp

import "testing"

func TestXIDataLifecycle(t *testing.T) {
	x := NewXIData()
	if x.InterpID != -1 {
		t.Fatalf("InterpID = %d, want -1", x.InterpID)
	}
	x.Init(4, "data", "obj", func(v any) (any, error) { return v, nil })
	if err := CheckXIData(x); err != nil {
		t.Fatalf("CheckXIData returned error: %v", err)
	}
	if err := x.Clear(4); err != nil {
		t.Fatalf("Clear returned error: %v", err)
	}
	if x.InterpID != -1 || x.Data != nil || x.Obj != nil {
		t.Fatalf("cleared xidata = %#v", x)
	}
}

func TestXIDataErrors(t *testing.T) {
	x := NewXIData()
	x.InterpID = -1
	if err := CheckXIData(x); err == nil {
		t.Fatal("expected missing interp error")
	}
	x.InterpID = 1
	if err := CheckXIData(x); err == nil {
		t.Fatal("expected missing new_object func error")
	}
}

func TestCallInInterpreter(t *testing.T) {
	called := 0
	got := CallInInterpreter(1, 1, func(any) int {
		called++
		return 7
	}, nil, nil)
	if got != 7 || called != 1 {
		t.Fatalf("direct call = (%d, %d)", got, called)
	}
	scheduled := 0
	got = CallInInterpreter(1, 2, func(any) int {
		scheduled++
		return 9
	}, nil, func(id int64, fn func(any) int, arg any) {
		scheduled = int(id)
	})
	if got != 0 || scheduled != 2 {
		t.Fatalf("scheduled call = (%d, %d)", got, scheduled)
	}
}
