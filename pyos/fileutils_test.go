package pyos

import (
	"os"
	"path/filepath"
	"testing"
)

func TestGetCwdAndAbsPath(t *testing.T) {
	cwd, err := GetCwd()
	if err != nil {
		t.Fatalf("GetCwd returned error: %v", err)
	}
	if cwd == "" {
		t.Fatal("GetCwd returned empty path")
	}

	got, err := AbsPath(".")
	if err != nil {
		t.Fatalf("AbsPath returned error: %v", err)
	}
	if got != cwd {
		t.Fatalf("AbsPath(.) = %q, want %q", got, cwd)
	}

	got, err = AbsPath("child")
	if err != nil {
		t.Fatalf("AbsPath child returned error: %v", err)
	}
	if !filepath.IsAbs(got) {
		t.Fatalf("AbsPath child = %q, want absolute path", got)
	}
}

func TestIsAbs(t *testing.T) {
	if !IsAbs(string(os.PathSeparator) + "tmp") {
		t.Fatal("unix-style absolute path should be absolute on this host")
	}
	if IsAbs("relative") {
		t.Fatal("relative path should not be absolute")
	}
}

func TestReadLinkAndRealPath(t *testing.T) {
	dir := t.TempDir()
	target := filepath.Join(dir, "target.txt")
	link := filepath.Join(dir, "link.txt")
	if err := os.WriteFile(target, []byte("hello"), 0o644); err != nil {
		t.Fatalf("WriteFile returned error: %v", err)
	}
	if err := os.Symlink(target, link); err != nil {
		t.Skipf("symlink unavailable: %v", err)
	}

	got, err := ReadLink(link)
	if err != nil {
		t.Fatalf("ReadLink returned error: %v", err)
	}
	if got != target {
		t.Fatalf("ReadLink = %q, want %q", got, target)
	}

	real, err := RealPath(link)
	if err != nil {
		t.Fatalf("RealPath returned error: %v", err)
	}
	realInfo, err := os.Stat(real)
	if err != nil {
		t.Fatalf("Stat(realpath) returned error: %v", err)
	}
	targetInfo, err := os.Stat(target)
	if err != nil {
		t.Fatalf("Stat(target) returned error: %v", err)
	}
	if !os.SameFile(realInfo, targetInfo) {
		t.Fatalf("RealPath = %q, want same file as %q", real, target)
	}
}

func TestStatWrappersAndFstat(t *testing.T) {
	file, err := os.CreateTemp(t.TempDir(), "fileutils")
	if err != nil {
		t.Fatalf("CreateTemp returned error: %v", err)
	}
	defer file.Close()
	if _, err := file.WriteString("hello"); err != nil {
		t.Fatalf("WriteString returned error: %v", err)
	}

	info, err := StatPath(file.Name())
	if err != nil {
		t.Fatalf("StatPath returned error: %v", err)
	}
	if info.Size() != 5 {
		t.Fatalf("StatPath size = %d, want 5", info.Size())
	}

	linfo, err := LstatPath(file.Name())
	if err != nil {
		t.Fatalf("LstatPath returned error: %v", err)
	}
	if linfo.Mode().Type() != 0 {
		t.Fatalf("LstatPath mode = %v, want regular file", linfo.Mode())
	}

	stat, err := Fstat(int(file.Fd()))
	if err != nil {
		t.Fatalf("Fstat returned error: %v", err)
	}
	if stat.Size != 5 {
		t.Fatalf("Fstat size = %d, want 5", stat.Size)
	}

	if _, err := FstatNoRaise(-1); err == nil {
		t.Fatal("FstatNoRaise should fail for invalid fd")
	}
}

func TestSetInheritableRejectsInvalidFD(t *testing.T) {
	if err := SetInheritable(-1, false); err == nil {
		t.Fatal("SetInheritable should reject invalid fd")
	}
}
