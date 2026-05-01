package pyos

import (
	"os"
	"path/filepath"
	"testing"
)

func TestPathConfigReadUpdateAndClear(t *testing.T) {
	t.Cleanup(ClearGlobalPathConfig)
	ClearGlobalPathConfig()

	UpdateGlobalPathConfig(Config{
		Prefix:            "/prefix",
		ExecPrefix:        "/exec-prefix",
		StdlibDir:         "/stdlib",
		ProgramName:       "gopython",
		Home:              "/home",
		Executable:        "/bin/gopython",
		ModuleSearchPaths: []string{"a", "b", "c"},
		IsPythonBuild:     true,
	})

	if got := GetPath(); got != "a"+string(os.PathListSeparator)+"b"+string(os.PathListSeparator)+"c" {
		t.Fatalf("GetPath() = %q", got)
	}

	var cfg Config
	ReadGlobalPathConfig(&cfg)
	if cfg.Prefix != "/prefix" || cfg.ExecPrefix != "/exec-prefix" || cfg.StdlibDir != "/stdlib" {
		t.Fatalf("ReadGlobalPathConfig basic fields = %#v", cfg)
	}
	if cfg.ProgramName != "gopython" || cfg.Home != "/home" || cfg.Executable != "/bin/gopython" {
		t.Fatalf("ReadGlobalPathConfig name/home/executable = %#v", cfg)
	}
	if !cfg.IsPythonBuild {
		t.Fatalf("ReadGlobalPathConfig IsPythonBuild = false")
	}

	ClearGlobalPathConfig()
	if got := GetPath(); got != "" {
		t.Fatalf("GetPath after clear = %q", got)
	}
	if got := GetPrefix(); got != "" {
		t.Fatalf("GetPrefix after clear = %q", got)
	}
}

func TestSetPathUsesHomeForStdlib(t *testing.T) {
	t.Cleanup(ClearGlobalPathConfig)
	ClearGlobalPathConfig()

	SetPythonHome("/python-home")
	SetPath("/x:/y")

	if got := GetPrefix(); got != "" {
		t.Fatalf("GetPrefix() = %q, want empty", got)
	}
	if got := GetExecPrefix(); got != "" {
		t.Fatalf("GetExecPrefix() = %q, want empty", got)
	}
	if got := GetStdlibDir(); got != "/python-home" {
		t.Fatalf("GetStdlibDir() = %q", got)
	}
	if got := GetPath(); got != "/x:/y" {
		t.Fatalf("GetPath() = %q", got)
	}
}

func TestSetPathKeepsExplicitEmptyPath(t *testing.T) {
	t.Cleanup(ClearGlobalPathConfig)
	ClearGlobalPathConfig()

	UpdateGlobalPathConfig(Config{ModuleSearchPaths: []string{"a", "b"}})
	SetPath("")

	if got := GetPath(); got != "" {
		t.Fatalf("GetPath() = %q, want explicit empty path", got)
	}
}

func TestSetProgramNameAndHome(t *testing.T) {
	t.Cleanup(ClearGlobalPathConfig)
	ClearGlobalPathConfig()

	SetProgramName("prog")
	SetPythonHome("home")

	if got := GetProgramName(); got != "prog" {
		t.Fatalf("GetProgramName() = %q", got)
	}
	if got := GetPythonHome(); got != "home" {
		t.Fatalf("GetPythonHome() = %q", got)
	}
}

func TestComputeSysPath0EmptyArgv(t *testing.T) {
	got, ok, err := ComputeSysPath0(nil)
	if err != nil {
		t.Fatalf("ComputeSysPath0(nil) error = %v", err)
	}
	if ok {
		t.Fatalf("ComputeSysPath0(nil) ok = true")
	}
	if got != "" {
		t.Fatalf("ComputeSysPath0(nil) path = %q", got)
	}
}

func TestComputeSysPath0CommandString(t *testing.T) {
	got, ok, err := ComputeSysPath0([]string{"-c"})
	if err != nil {
		t.Fatalf("ComputeSysPath0(-c) error = %v", err)
	}
	if !ok {
		t.Fatalf("ComputeSysPath0(-c) ok = false")
	}
	if got != "" {
		t.Fatalf("ComputeSysPath0(-c) path = %q", got)
	}
}

func TestComputeSysPath0ModuleUsesCwd(t *testing.T) {
	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("os.Getwd() failed: %v", err)
	}
	got, ok, err := ComputeSysPath0([]string{"-m"})
	if err != nil {
		t.Fatalf("ComputeSysPath0(-m) error = %v", err)
	}
	if !ok {
		t.Fatalf("ComputeSysPath0(-m) ok = false")
	}
	if got != wd {
		t.Fatalf("ComputeSysPath0(-m) = %q, want %q", got, wd)
	}
}

func TestComputeSysPath0ScriptUsesDirectory(t *testing.T) {
	dir := t.TempDir()
	script := filepath.Join(dir, "main.py")
	if err := os.WriteFile(script, []byte("pass\n"), 0o644); err != nil {
		t.Fatalf("os.WriteFile() failed: %v", err)
	}

	got, ok, err := ComputeSysPath0([]string{script})
	if err != nil {
		t.Fatalf("ComputeSysPath0(script) error = %v", err)
	}
	if !ok {
		t.Fatalf("ComputeSysPath0(script) ok = false")
	}
	want, err := filepath.EvalSymlinks(dir)
	if err != nil {
		want = dir
	}
	if got != want {
		t.Fatalf("ComputeSysPath0(script) = %q, want %q", got, want)
	}
}

func TestComputeSysPath0ScriptSymlinkUsesTargetDirectory(t *testing.T) {
	dir := t.TempDir()
	targetDir := filepath.Join(dir, "target")
	linkDir := filepath.Join(dir, "link")
	if err := os.MkdirAll(targetDir, 0o755); err != nil {
		t.Fatalf("os.MkdirAll(targetDir) failed: %v", err)
	}
	if err := os.MkdirAll(linkDir, 0o755); err != nil {
		t.Fatalf("os.MkdirAll(linkDir) failed: %v", err)
	}

	target := filepath.Join(targetDir, "main.py")
	if err := os.WriteFile(target, []byte("pass\n"), 0o644); err != nil {
		t.Fatalf("os.WriteFile(target) failed: %v", err)
	}
	link := filepath.Join(linkDir, "main.py")
	if err := os.Symlink(target, link); err != nil {
		t.Skipf("os.Symlink() unavailable: %v", err)
	}

	got, ok, err := ComputeSysPath0([]string{link})
	if err != nil {
		t.Fatalf("ComputeSysPath0(link) error = %v", err)
	}
	if !ok {
		t.Fatalf("ComputeSysPath0(link) ok = false")
	}
	want, err := filepath.EvalSymlinks(targetDir)
	if err != nil {
		want = targetDir
	}
	if got != want {
		t.Fatalf("ComputeSysPath0(link) = %q, want %q", got, want)
	}
}
