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

func TestSimpleStringAndFile(t *testing.T) {
	var got string
	if err := SimpleString("print(1)", "mod", func(command string, name string) error {
		got = command + "@" + name
		return nil
	}); err != nil {
		t.Fatalf("SimpleString returned error: %v", err)
	}
	if got != "print(1)@mod" {
		t.Fatalf("got = %q", got)
	}

	if err := SimpleFile("x.pyc", true, true, func(name string) error {
		got = "pyc:" + name
		return nil
	}, nil); err != nil {
		t.Fatalf("SimpleFile pyc returned error: %v", err)
	}
	if got != "pyc:x.pyc" {
		t.Fatalf("got = %q", got)
	}

	if err := SimpleFile("x.py", true, false, nil, func(name string) error {
		got = "src:" + name
		return nil
	}); err != nil {
		t.Fatalf("SimpleFile src returned error: %v", err)
	}
	if got != "src:x.py" {
		t.Fatalf("got = %q", got)
	}
}

func TestHandleSystemExit(t *testing.T) {
	if handled, _, _ := HandleSystemExit("KeyboardInterrupt", false, nil); handled {
		t.Fatal("keyboard interrupt should not exit here")
	}
	if handled, _, _ := HandleSystemExit("SystemExit", true, nil); handled {
		t.Fatal("inspect should suppress exit")
	}
	handled, code, text := HandleSystemExit("SystemExit", false, 7)
	if !handled || code != 7 || text != "" {
		t.Fatalf("int exit = (%t, %d, %q)", handled, code, text)
	}
	handled, code, text = HandleSystemExit("SystemExit", false, "bye")
	if !handled || code != 1 || text != "bye" {
		t.Fatalf("string exit = (%t, %d, %q)", handled, code, text)
	}
}

func TestExceptionFormattingHelpers(t *testing.T) {
	if got := ExceptionMessage("m", "Error", "bad"); got != "m.Error: bad" {
		t.Fatalf("ExceptionMessage = %q", got)
	}
	if got := ExceptionMessage("builtins", "Error", ""); got != "Error" {
		t.Fatalf("ExceptionMessage builtins = %q", got)
	}
	if got := RenderChainedExceptions([]string{"first", "second"}); got != "first\n\nsecond" {
		t.Fatalf("RenderChainedExceptions = %q", got)
	}
}
