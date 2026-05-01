package pycrossinterp

import "testing"

func TestPickleRoundTripXIData(t *testing.T) {
	x := NewXIData()
	if err := PickleToXIData(map[string]any{"a": 1.0}, "main.py", x); err != nil {
		t.Fatalf("PickleToXIData returned error: %v", err)
	}
	got, err := LoadPickleFromXIData(x)
	if err != nil {
		t.Fatalf("LoadPickleFromXIData returned error: %v", err)
	}
	value := got.(map[string]any)["a"]
	if value != 1.0 {
		t.Fatalf("roundtrip value = %#v", got)
	}
}

func TestMarshalRoundTripXIData(t *testing.T) {
	x := NewXIData()
	if err := MarshalToXIData([]any{"x", 2.0}, x); err != nil {
		t.Fatalf("MarshalToXIData returned error: %v", err)
	}
	got, err := ReadMarshalFromXIData(x)
	if err != nil {
		t.Fatalf("ReadMarshalFromXIData returned error: %v", err)
	}
	if len(got.([]any)) != 2 {
		t.Fatalf("roundtrip = %#v", got)
	}
}

func TestScriptToXIData(t *testing.T) {
	x := NewXIData()
	if err := ScriptToXIData("print('x')", false, x); err != nil {
		t.Fatalf("ScriptToXIData returned error: %v", err)
	}
	if got, err := x.NewObject(x); err != nil || got.(string) != "print('x')" {
		t.Fatalf("script new object = (%v, %v)", got, err)
	}
	if err := ScriptToXIData("return 1", false, NewXIData()); err == nil {
		t.Fatal("expected invalid script error")
	}
	if err := ScriptToXIData("global x\nprint(x)", true, NewXIData()); err == nil {
		t.Fatal("expected invalid pure script error")
	}
}

func TestScriptCodeAndFunctionValidation(t *testing.T) {
	x := NewXIData()
	err := ScriptToXIData(ScriptCode{
		Source:       "pass",
		ReturnsValue: false,
		Stateless:    true,
		Checked:      true,
	}, true, x)
	if err != nil {
		t.Fatalf("ScriptCode validation error: %v", err)
	}
	if got, err := x.NewObject(x); err != nil || got.(string) != "pass" {
		t.Fatalf("script code new object = (%v, %v)", got, err)
	}

	if err := ScriptToXIData(ScriptCode{
		Source:       "x = 1",
		ArgCount:     1,
		ReturnsValue: false,
	}, false, NewXIData()); err == nil {
		t.Fatal("expected arg validation error")
	}
	if err := ScriptToXIData(ScriptFunction{
		Code: ScriptCode{
			Source:       "value = 1",
			ReturnsValue: false,
		},
		Stateless: false,
	}, true, NewXIData()); err == nil {
		t.Fatal("expected stateless function validation error")
	}
}
