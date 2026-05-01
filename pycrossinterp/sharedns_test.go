package pycrossinterp

import "testing"

func TestSharedNamespaceLifecycle(t *testing.T) {
	ns, err := CreateSharedNamespace([]string{"a", "b"})
	if err != nil {
		t.Fatalf("CreateSharedNamespace error: %v", err)
	}
	err = ns.Fill(map[string]any{
		"a": map[string]any{"x": 1},
	}, XIDataFullFallback)
	if err != nil {
		t.Fatalf("Fill error: %v", err)
	}
	if ns.NumNames != 2 || ns.NumValues != 2 {
		t.Fatalf("ns counts = %#v", ns)
	}

	target := map[string]any{}
	if err := ns.Apply(target, "default"); err != nil {
		t.Fatalf("Apply error: %v", err)
	}
	if target["b"] != "default" {
		t.Fatalf("target = %#v", target)
	}
	got := target["a"].(map[string]any)
	if got["x"] != float64(1) {
		t.Fatalf("target[a] = %#v", target["a"])
	}

	ns.Free()
	if ns.NumNames != 0 || ns.NumValues != 0 {
		t.Fatalf("ns after free = %#v", ns)
	}
}

func TestSharedNamespaceDestroyUsesPendingRelease(t *testing.T) {
	ns, err := CreateSharedNamespace([]string{"a"})
	if err != nil {
		t.Fatalf("CreateSharedNamespace error: %v", err)
	}
	if err := ns.Fill(map[string]any{"a": "value"}, XIDataFullFallback); err != nil {
		t.Fatalf("Fill error: %v", err)
	}
	scheduled := int64(0)
	ns.Destroy(2, func(id int64) bool { return id == 1 }, func(id int64, fn func(any) int, arg any) {
		scheduled = id
		fn(arg)
	})
	if scheduled != 1 {
		t.Fatalf("scheduled = %d", scheduled)
	}
	if ns.NumNames != 0 || ns.NumValues != 0 {
		t.Fatalf("ns after destroy = %#v", ns)
	}
}

func TestSharedNamespaceErrors(t *testing.T) {
	if _, err := AllocSharedNamespace(0); err == nil {
		t.Fatal("expected empty namespace error")
	}
	if _, err := AllocSharedNamespace(-1); err == nil {
		t.Fatal("expected negative namespace error")
	}
}
