package pyconfig

import "testing"

func TestConfigInitializers(t *testing.T) {
	var compat Config
	InitCompatConfig(&compat)
	if compat.ConfigInit != ConfigInitCompat || compat.UseEnvironment != -1 || compat.ConfigureCStdio != 0 {
		t.Fatalf("compat = %#v", compat)
	}

	var python Config
	InitPythonConfig(&python)
	if python.ConfigInit != ConfigInitPython || python.ParseArgv != 1 || python.UseEnvironment != 1 || python.ConfigureCStdio != 1 {
		t.Fatalf("python = %#v", python)
	}

	var isolated Config
	InitIsolatedConfig(&isolated)
	if isolated.ConfigInit != ConfigInitIsolated || isolated.Isolated != 1 || isolated.SafePath != 1 || isolated.IntMaxStrDigits != DefaultMaxStrDigits {
		t.Fatalf("isolated = %#v", isolated)
	}
}

func TestConfigClear(t *testing.T) {
	config := Config{
		PycachePrefix:        []rune("cache"),
		FilesystemEncoding:   []rune("utf-8"),
		Argv:                 WideStringList{Items: [][]rune{[]rune("python")}},
		ModuleSearchPaths:    WideStringList{Items: [][]rune{[]rune("/tmp")}},
		ModuleSearchPathsSet: 1,
	}
	config.Clear()
	if config.PycachePrefix != nil || config.FilesystemEncoding != nil {
		t.Fatalf("clear left string fields populated: %#v", config)
	}
	if len(config.Argv.Items) != 0 || len(config.ModuleSearchPaths.Items) != 0 || config.ModuleSearchPathsSet != 0 {
		t.Fatalf("clear left list fields populated: %#v", config)
	}
}

func TestSetStringAndBytesString(t *testing.T) {
	var config Config
	var field []rune
	SetString(&config, &field, []rune("value"))
	if string(field) != "value" {
		t.Fatalf("SetString = %q", string(field))
	}
	if err := SetBytesString(&config, &field, []byte{'a', 0x80, 'b'}); err != nil {
		t.Fatalf("SetBytesString returned error: %v", err)
	}
	if len(field) != 3 || field[1] != 0xdc80 {
		t.Fatalf("SetBytesString = %#v", field)
	}
}

func TestConfigCopyFromClones(t *testing.T) {
	src := &Config{
		FilesystemEncoding: []rune("utf-8"),
		Argv:               WideStringList{Items: [][]rune{[]rune("python")}},
	}
	var dst Config
	if err := dst.CopyFrom(src); err != nil {
		t.Fatalf("CopyFrom returned error: %v", err)
	}
	src.FilesystemEncoding[0] = 'X'
	src.Argv.Items[0][0] = 'X'
	if string(dst.FilesystemEncoding) != "utf-8" {
		t.Fatalf("copied encoding mutated to %q", string(dst.FilesystemEncoding))
	}
	if string(dst.Argv.Items[0]) != "python" {
		t.Fatalf("copied argv mutated to %q", string(dst.Argv.Items[0]))
	}
}

func TestWideStringListHelpers(t *testing.T) {
	var list WideStringList
	if err := list.Insert(0, []rune("b")); err != nil {
		t.Fatalf("Insert returned error: %v", err)
	}
	if err := list.Insert(0, []rune("a")); err != nil {
		t.Fatalf("Insert returned error: %v", err)
	}
	if err := list.AppendWide([]rune("c")); err != nil {
		t.Fatalf("AppendWide returned error: %v", err)
	}
	if !list.Find([]rune("b")) || list.Find([]rune("missing")) {
		t.Fatalf("Find returned wrong result for %#v", list.Items)
	}

	var other WideStringList
	other.CopyFrom(list)
	list.Items[0][0] = 'X'
	if string(other.Items[0]) != "a" {
		t.Fatalf("CopyFrom mutated to %q", string(other.Items[0]))
	}
}
