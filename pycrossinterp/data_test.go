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

func TestXIDataAllocHelpers(t *testing.T) {
	x := NewXIData()
	x.InitWithSize(3, 8, "obj", func(v any) (any, error) { return v, nil })
	if data, ok := x.Data.([]byte); !ok || len(data) != 8 {
		t.Fatalf("x.Data = %#v", x.Data)
	}
	if err := FreeXIData(3, x); err != nil {
		t.Fatalf("FreeXIData returned error: %v", err)
	}
	if x.InterpID != -1 {
		t.Fatalf("InterpID = %d, want -1", x.InterpID)
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

func TestCallInInterpreterAndRawFree(t *testing.T) {
	called := 0
	got := CallInInterpreterAndRawFree(1, 1, func(any) int {
		called++
		return 4
	}, "arg", nil)
	if got != 4 || called != 1 {
		t.Fatalf("direct call = (%d, %d)", got, called)
	}
	scheduled := 0
	got = CallInInterpreterAndRawFree(1, 5, func(any) int { return 9 }, "arg", func(id int64, fn func(any) int, arg any) {
		scheduled = int(id)
	})
	if got != 0 || scheduled != 5 {
		t.Fatalf("scheduled call = (%d, %d)", got, scheduled)
	}
}

func TestObjectXIDataLookupAndFallback(t *testing.T) {
	var lookup LookupState
	lookup.Init()
	lookup.Global.Add("", func(v any) (any, error) { return "shared:" + v.(string), nil })

	x := NewXIData()
	if err := ObjectCheckXIData(&lookup, "abc"); err != nil {
		t.Fatalf("ObjectCheckXIData returned error: %v", err)
	}
	if err := ObjectGetXIData(&lookup, 7, "abc", XIDataOnly, x); err != nil {
		t.Fatalf("ObjectGetXIData returned error: %v", err)
	}
	if x.InterpID != 7 || x.Data != "shared:abc" {
		t.Fatalf("x = %#v", x)
	}

	y := NewXIData()
	if err := ObjectGetXIData(nil, 8, map[string]any{"a": 1}, XIDataFullFallback, y); err != nil {
		t.Fatalf("ObjectGetXIData fallback returned error: %v", err)
	}
	if y.InterpID != 8 {
		t.Fatalf("y = %#v", y)
	}

	z := NewXIData()
	if err := ObjectGetXIData(nil, 9, make(chan int), XIDataOnly, z); err == nil {
		t.Fatal("expected xidata failure")
	}
}
