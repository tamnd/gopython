package pyos

import (
	"errors"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"sync/atomic"
	"syscall"
	"testing"

	"golang.org/x/term"
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

	wlink, err := WReadLink([]rune(link))
	if err != nil {
		t.Fatalf("WReadLink returned error: %v", err)
	}
	if string(wlink) != target {
		t.Fatalf("WReadLink = %q, want %q", string(wlink), target)
	}

	wreal, err := WRealPath([]rune(link))
	if err != nil {
		t.Fatalf("WRealPath returned error: %v", err)
	}
	if string(wreal) != real {
		t.Fatalf("WRealPath = %q, want %q", string(wreal), real)
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

func TestGetSetInheritable(t *testing.T) {
	file, err := os.CreateTemp(t.TempDir(), "inheritable")
	if err != nil {
		t.Fatalf("CreateTemp returned error: %v", err)
	}
	defer file.Close()

	fd := int(file.Fd())
	original, err := GetInheritable(fd)
	if err != nil {
		t.Fatalf("GetInheritable returned error: %v", err)
	}

	for _, inheritable := range []bool{false, true, original} {
		if err := SetInheritable(fd, inheritable); err != nil {
			t.Fatalf("SetInheritable(%t) returned error: %v", inheritable, err)
		}
		got, err := GetInheritable(fd)
		if err != nil {
			t.Fatalf("GetInheritable after SetInheritable(%t) returned error: %v", inheritable, err)
		}
		if got != inheritable {
			t.Fatalf("GetInheritable after SetInheritable(%t) = %t", inheritable, got)
		}
	}
}

func TestOpenFileAndValidFD(t *testing.T) {
	path := filepath.Join(t.TempDir(), "opened.txt")
	file, err := OpenFile(path, os.O_CREATE|os.O_RDWR|os.O_TRUNC, 0o600)
	if err != nil {
		t.Fatalf("OpenFile returned error: %v", err)
	}
	fd := int(file.Fd())
	if !IsValidFD(fd) {
		t.Fatal("IsValidFD should report the opened file descriptor as valid")
	}
	inheritable, err := GetInheritable(fd)
	if err != nil {
		t.Fatalf("GetInheritable on opened file returned error: %v", err)
	}
	if inheritable {
		t.Fatal("OpenFile should clear inheritable on the returned descriptor")
	}
	if _, err := io.WriteString(file, "hello"); err != nil {
		t.Fatalf("WriteString returned error: %v", err)
	}
	if err := file.Close(); err != nil {
		t.Fatalf("Close returned error: %v", err)
	}
	if IsValidFD(fd) {
		t.Fatal("IsValidFD should report a closed descriptor as invalid")
	}
}

func TestOpenFDNoRaiseAndWFopen(t *testing.T) {
	path := filepath.Join(t.TempDir(), "fd-opened.txt")
	fd, err := OpenFDNoRaise(path, os.O_CREATE|os.O_RDWR|os.O_TRUNC, 0o600)
	if err != nil {
		t.Fatalf("OpenFDNoRaise returned error: %v", err)
	}
	if !IsValidFD(fd) {
		t.Fatal("OpenFDNoRaise should return a valid descriptor")
	}
	file := os.NewFile(uintptr(fd), path)
	if file == nil {
		t.Fatal("os.NewFile returned nil for OpenFDNoRaise descriptor")
	}
	if _, err := io.WriteString(file, "hello"); err != nil {
		t.Fatalf("WriteString returned error: %v", err)
	}
	if err := file.Close(); err != nil {
		t.Fatalf("Close returned error: %v", err)
	}

	path2 := filepath.Join(t.TempDir(), "wfopen.txt")
	wfile, err := WFopen([]rune(path2), "w+")
	if err != nil {
		t.Fatalf("WFopen returned error: %v", err)
	}
	if _, err := io.WriteString(wfile, "content"); err != nil {
		t.Fatalf("WriteString to WFopen file returned error: %v", err)
	}
	if _, err := wfile.Seek(0, 0); err != nil {
		t.Fatalf("Seek returned error: %v", err)
	}
	data, err := io.ReadAll(wfile)
	if err != nil {
		t.Fatalf("ReadAll returned error: %v", err)
	}
	if string(data) != "content" {
		t.Fatalf("WFopen content = %q, want %q", string(data), "content")
	}
	if err := Fclose(wfile); err != nil {
		t.Fatalf("Fclose returned error: %v", err)
	}
	if _, err := WFopen([]rune(path2), "rx"); err == nil {
		t.Fatal("WFopen should reject exclusive read mode")
	}
}

func TestFopen(t *testing.T) {
	path := filepath.Join(t.TempDir(), "fopen.txt")
	file, err := Fopen(path, "w+")
	if err != nil {
		t.Fatalf("Fopen returned error: %v", err)
	}
	if _, err := io.WriteString(file, "hello"); err != nil {
		t.Fatalf("WriteString returned error: %v", err)
	}
	if _, err := file.Seek(0, 0); err != nil {
		t.Fatalf("Seek returned error: %v", err)
	}
	data, err := io.ReadAll(file)
	if err != nil {
		t.Fatalf("ReadAll returned error: %v", err)
	}
	if string(data) != "hello" {
		t.Fatalf("Fopen content = %q, want %q", string(data), "hello")
	}
	if err := Fclose(file); err != nil {
		t.Fatalf("Fclose returned error: %v", err)
	}
}

func TestReadFDAndWriteFD(t *testing.T) {
	readFD, writeFD, err := createPipe()
	if err != nil {
		t.Fatalf("createPipe returned error: %v", err)
	}
	defer closePipeFD(readFD)
	defer closePipeFD(writeFD)

	if n, err := WriteFD(writeFD, []byte("hello")); err != nil || n != 5 {
		t.Fatalf("WriteFD = (%d, %v), want (5, nil)", n, err)
	}

	buf := make([]byte, 8)
	n, err := ReadFD(readFD, buf)
	if err != nil {
		t.Fatalf("ReadFD returned error: %v", err)
	}
	if got := string(buf[:n]); got != "hello" {
		t.Fatalf("ReadFD content = %q, want %q", got, "hello")
	}
}

func TestReadFDAndWriteFDRetry(t *testing.T) {
	prevRead := readFDHook
	prevWrite := writeFDHook
	var readCalls atomic.Int32
	var writeCalls atomic.Int32
	readFDHook = func(fd int, p []byte) (int, error) {
		if readCalls.Add(1) == 1 {
			return -1, syscall.EINTR
		}
		copy(p, []byte("ok"))
		return 2, nil
	}
	writeFDHook = func(fd int, p []byte) (int, error) {
		if writeCalls.Add(1) == 1 {
			return -1, syscall.EINTR
		}
		return len(p), nil
	}
	t.Cleanup(func() {
		readFDHook = prevRead
		writeFDHook = prevWrite
	})

	buf := make([]byte, 4)
	n, err := ReadFD(0, buf)
	if err != nil || n != 2 || string(buf[:2]) != "ok" {
		t.Fatalf("ReadFD retry = (%d, %v, %q), want (2, nil, %q)", n, err, string(buf[:2]), "ok")
	}
	if _, err := WriteFD(1, []byte("abc")); err != nil {
		t.Fatalf("WriteFD retry returned error: %v", err)
	}
	if got := readCalls.Load(); got != 2 {
		t.Fatalf("read retry calls = %d, want 2", got)
	}
	if got := writeCalls.Load(); got != 2 {
		t.Fatalf("write retry calls = %d, want 2", got)
	}
}

func TestWriteFDNoRaiseDoesNotRetry(t *testing.T) {
	prevWrite := writeFDHook
	var writeCalls atomic.Int32
	writeFDHook = func(fd int, p []byte) (int, error) {
		writeCalls.Add(1)
		return -1, syscall.EINTR
	}
	t.Cleanup(func() {
		writeFDHook = prevWrite
	})

	if _, err := WriteFDNoRaise(1, []byte("abc")); !errors.Is(err, syscall.EINTR) {
		t.Fatalf("WriteFDNoRaise error = %v, want EINTR", err)
	}
	if got := writeCalls.Load(); got != 1 {
		t.Fatalf("write no-raise calls = %d, want 1", got)
	}
}

func TestWStatAndWAbsPath(t *testing.T) {
	file, err := os.CreateTemp(t.TempDir(), "wpath")
	if err != nil {
		t.Fatalf("CreateTemp returned error: %v", err)
	}
	defer file.Close()

	info, err := WStatRunes([]rune(file.Name()))
	if err != nil {
		t.Fatalf("WStatRunes returned error: %v", err)
	}
	if info.Name() == "" {
		t.Fatal("WStatRunes returned empty file name")
	}

	abs, err := WAbsPath([]rune("."))
	if err != nil {
		t.Fatalf("WAbsPath returned error: %v", err)
	}
	if string(abs) == "" {
		t.Fatal("WAbsPath returned empty path")
	}
}

func TestDecodeLocaleSurrogateEscape(t *testing.T) {
	got, err := DecodeLocale([]byte{'a', 0xff, 'b'})
	if err != nil {
		t.Fatalf("DecodeLocale returned error: %v", err)
	}
	want := []rune{'a', 0xdcff, 'b'}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("DecodeLocale = %#v, want %#v", got, want)
	}

	if _, err := DecodeLocaleEx([]byte{0xff}, false, ErrorStrict); err == nil {
		t.Fatal("DecodeLocaleEx strict should fail on invalid byte")
	}
}

func TestEncodeLocaleSurrogateEscape(t *testing.T) {
	got, err := EncodeLocale([]rune{'a', 0xdc80, 'b'})
	if err != nil {
		t.Fatalf("EncodeLocale returned error: %v", err)
	}
	want := []byte{'a', 0x80, 'b'}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("EncodeLocale = %#v, want %#v", got, want)
	}

	if _, err := EncodeLocaleEx([]rune{0xd800}, false, ErrorStrict); err == nil {
		t.Fatal("EncodeLocaleEx strict should fail on surrogate rune")
	}
}

func TestForceASCIIAndLocaleEncoding(t *testing.T) {
	ResetForceASCII()
	t.Setenv("LC_ALL", "C.ascii")
	if runtime.GOOS == "windows" {
		if got := GetForceASCII(); got != 0 {
			t.Fatalf("GetForceASCII = %d, want 0 on windows", got)
		}
		if got := LocaleEncoding(); got != "cp1252" {
			t.Fatalf("LocaleEncoding = %q, want cp1252 on windows", got)
		}
	} else {
		if got := GetForceASCII(); got != 1 {
			t.Fatalf("GetForceASCII = %d, want 1", got)
		}
		if got := LocaleEncoding(); got != "ascii" {
			t.Fatalf("LocaleEncoding = %q, want ascii", got)
		}
	}

	ResetForceASCII()
	t.Setenv("LC_ALL", "en_US.UTF-8")
	if got := GetForceASCII(); got != 0 {
		t.Fatalf("GetForceASCII UTF-8 = %d, want 0", got)
	}
	file, err := os.CreateTemp(t.TempDir(), "device-encoding")
	if err != nil {
		t.Fatalf("CreateTemp returned error: %v", err)
	}
	defer file.Close()
	if got := DeviceEncoding(int(file.Fd())); got != "" {
		t.Fatalf("DeviceEncoding(regular file) = %q, want empty string", got)
	}
	if term.IsTerminal(0) {
		if got := DeviceEncoding(0); got == "" {
			t.Fatal("DeviceEncoding should return a non-empty encoding for a terminal")
		}
	}
}
