package pyruntime

import (
	"errors"
	"testing"
)

func TestAnyFileObject(t *testing.T) {
	called := ""
	if err := AnyFileObject("", true, func(name string) error {
		called = "interactive:" + name
		return nil
	}, nil); err != nil {
		t.Fatalf("AnyFileObject interactive returned error: %v", err)
	}
	if called != "interactive:???" {
		t.Fatalf("called = %q", called)
	}
	if err := AnyFileObject("file.py", false, nil, func(name string) error {
		called = "simple:" + name
		return nil
	}); err != nil {
		t.Fatalf("AnyFileObject simple returned error: %v", err)
	}
	if called != "simple:file.py" {
		t.Fatalf("called = %q", called)
	}
}

func TestInteractiveLoop(t *testing.T) {
	count := 0
	prompts, err := InteractiveLoop(">>> ", "... ", nil, func() (int, error) {
		count++
		if count == 1 {
			return 0, errors.New("MemoryError")
		}
		return EOF, nil
	})
	if err != nil {
		t.Fatalf("InteractiveLoop returned error: %v", err)
	}
	if prompts["ps1"] != ">>> " || prompts["ps2"] != "... " {
		t.Fatalf("prompts = %#v", prompts)
	}
}
