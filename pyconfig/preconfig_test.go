package pyconfig

import (
	"reflect"
	"testing"
)

func TestFileSystemEncodingState(t *testing.T) {
	ClearFileSystemEncoding()
	SetFileSystemEncoding("utf-8", "surrogateescape")
	if got := FileSystemEncoding(); got != "utf-8" {
		t.Fatalf("FileSystemEncoding = %q, want utf-8", got)
	}
	if got := FileSystemErrors(); got != "surrogateescape" {
		t.Fatalf("FileSystemErrors = %q, want surrogateescape", got)
	}
	ClearFileSystemEncoding()
	if FileSystemEncoding() != "" || FileSystemErrors() != "" {
		t.Fatal("ClearFileSystemEncoding should clear both values")
	}
}

func TestArgvAsWideStringListBytes(t *testing.T) {
	args := &Argv{
		UseBytesArgv: true,
		BytesArgv: [][]byte{
			[]byte("python"),
			{'a', 0x80, 'b'},
		},
	}
	var list WideStringList
	if err := ArgvAsWideStringList(args, &list); err != nil {
		t.Fatalf("ArgvAsWideStringList returned error: %v", err)
	}
	if got := string(list.Items[0]); got != "python" {
		t.Fatalf("argv[0] = %q", got)
	}
	want := []rune{'a', 0xdc80, 'b'}
	if !reflect.DeepEqual(list.Items[1], want) {
		t.Fatalf("argv[1] = %#v, want %#v", list.Items[1], want)
	}
}

func TestArgvAsWideStringListWideCopy(t *testing.T) {
	args := &Argv{
		WideArgv: [][]rune{[]rune("python"), []rune("script.py")},
	}
	var list WideStringList
	if err := ArgvAsWideStringList(args, &list); err != nil {
		t.Fatalf("ArgvAsWideStringList returned error: %v", err)
	}
	args.WideArgv[0][0] = 'X'
	if got := string(list.Items[0]); got != "python" {
		t.Fatalf("copied argv mutated to %q", got)
	}
}

func TestPreConfigInitializers(t *testing.T) {
	var compat PreConfig
	InitCompatPreConfig(&compat)
	if compat.ConfigInit != ConfigInitCompat || compat.UseEnvironment != -1 || compat.UTF8Mode != 0 {
		t.Fatalf("compat = %#v", compat)
	}

	var python PreConfig
	InitPythonPreConfig(&python)
	if python.ConfigInit != ConfigInitPython || python.ParseArgv != 1 || python.UseEnvironment != 1 || python.UTF8Mode != -1 {
		t.Fatalf("python = %#v", python)
	}

	var isolated PreConfig
	InitIsolatedPreConfig(&isolated)
	if isolated.ConfigInit != ConfigInitIsolated || isolated.Isolated != 1 || isolated.UseEnvironment != 0 || isolated.ConfigureLocale != 0 {
		t.Fatalf("isolated = %#v", isolated)
	}
}

func TestPreConfigFromConfigHelpers(t *testing.T) {
	var config Config
	InitPythonConfig(&config)
	config.ParseArgv = 1
	config.DevMode = 1

	var pre PreConfig
	InitPreConfigFromConfig(&pre, &config)
	if pre.ConfigInit != ConfigInitPython || pre.ParseArgv != 1 || pre.DevMode != 1 {
		t.Fatalf("pre = %#v", pre)
	}

	var copied PreConfig
	GetPreConfigFromConfig(&copied, &config)
	if copied != pre {
		t.Fatalf("copied = %#v pre = %#v", copied, pre)
	}
}

func TestGetEnvHelpers(t *testing.T) {
	t.Setenv("PYCONFIG_TEST_FLAG", "2")
	flag := 0
	GetEnvFlag(1, &flag, "PYCONFIG_TEST_FLAG")
	if flag != 2 {
		t.Fatalf("flag = %d, want 2", flag)
	}
	if GetEnv(0, "PYCONFIG_TEST_FLAG") != "" {
		t.Fatal("GetEnv should ignore env when disabled")
	}
}

func TestGetEnvFlagInvalidFallsBackToOne(t *testing.T) {
	t.Setenv("PYCONFIG_TEST_BAD", "bad")
	flag := 0
	GetEnvFlag(1, &flag, "PYCONFIG_TEST_BAD")
	if flag != 1 {
		t.Fatalf("flag = %d, want 1", flag)
	}
}

func TestGetXOption(t *testing.T) {
	options := WideStringList{Items: [][]rune{[]rune("dev"), []rune("utf8=1")}}
	if got := GetXOption(options, "dev"); string(got) != "dev" {
		t.Fatalf("GetXOption(dev) = %q", string(got))
	}
	if got := GetXOption(options, "utf8"); got == nil {
		t.Fatal("GetXOption should match option names before =")
	}
}

func TestValidateUTF8XOption(t *testing.T) {
	if got, err := ValidateUTF8XOption([]rune("utf8")); err != nil || got != 1 {
		t.Fatalf("ValidateUTF8XOption utf8 = (%d, %v)", got, err)
	}
	if got, err := ValidateUTF8XOption([]rune("utf8=0")); err != nil || got != 0 {
		t.Fatalf("ValidateUTF8XOption utf8=0 = (%d, %v)", got, err)
	}
	if _, err := ValidateUTF8XOption([]rune("utf8=bad")); err == nil {
		t.Fatal("ValidateUTF8XOption should reject invalid values")
	}
}
