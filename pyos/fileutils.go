package pyos

import (
	"bytes"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"unicode/utf8"
)

type StatInfo struct {
	Mode fs.FileMode
	Size int64
}

type ErrorHandler int

const (
	ErrorStrict ErrorHandler = iota
	ErrorSurrogateEscape
)

type DecodeLocaleError struct {
	Reason string
	Pos    int
}

func (e *DecodeLocaleError) Error() string {
	return e.Reason
}

type EncodeLocaleError struct {
	Reason string
	Pos    int
}

func (e *EncodeLocaleError) Error() string {
	return e.Reason
}

var forceASCIICache = -1

func DeviceEncoding(fd int) string {
	if fd < 0 {
		return ""
	}
	return deviceEncoding(fd)
}

func GetCwd() (string, error) {
	return os.Getwd()
}

func IsAbs(path string) bool {
	if runtime.GOOS == "windows" && path != "" {
		switch path[0] {
		case '\\', '/':
			return true
		}
	}
	return filepath.IsAbs(path)
}

func AbsPath(path string) (string, error) {
	if path == "" || path == "." {
		return GetCwd()
	}
	if IsAbs(path) {
		return path, nil
	}
	cwd, err := GetCwd()
	if err != nil {
		return "", err
	}
	return filepath.Join(cwd, path), nil
}

func ReadLink(path string) (string, error) {
	return os.Readlink(path)
}

func WReadLink(path []rune) ([]rune, error) {
	nativePath, err := runesToNativePath(path)
	if err != nil {
		return nil, err
	}
	link, err := ReadLink(nativePath)
	if err != nil {
		return nil, err
	}
	return DecodeLocale([]byte(link))
}

func RealPath(path string) (string, error) {
	return filepath.EvalSymlinks(path)
}

func WRealPath(path []rune) ([]rune, error) {
	nativePath, err := runesToNativePath(path)
	if err != nil {
		return nil, err
	}
	resolved, err := RealPath(nativePath)
	if err != nil {
		return nil, err
	}
	return DecodeLocale([]byte(resolved))
}

func StatPath(path string) (fs.FileInfo, error) {
	return os.Stat(path)
}

func LstatPath(path string) (fs.FileInfo, error) {
	return os.Lstat(path)
}

func WStat(path string) (fs.FileInfo, error) {
	return os.Stat(path)
}

func WStatRunes(path []rune) (fs.FileInfo, error) {
	nativePath, err := runesToNativePath(path)
	if err != nil {
		return nil, err
	}
	return os.Stat(nativePath)
}

func GetInheritable(fd int) (bool, error) {
	if fd < 0 {
		return false, errors.New("invalid file descriptor")
	}
	return getInheritable(fd)
}

func SetInheritable(fd int, inheritable bool) error {
	if fd < 0 {
		return errors.New("invalid file descriptor")
	}
	return setInheritable(fd, inheritable)
}

func IsValidFD(fd int) bool {
	if fd < 0 {
		return false
	}
	return isValidFD(fd)
}

func OpenFile(path string, flag int, perm fs.FileMode) (*os.File, error) {
	fd, err := OpenFD(path, flag, perm)
	if err != nil {
		return nil, err
	}
	file := os.NewFile(uintptr(fd), path)
	if file == nil {
		closeFD(fd)
		return nil, errors.New("failed to wrap file descriptor")
	}
	return file, nil
}

func OpenFileNoRaise(path string, flag int, perm fs.FileMode) (*os.File, error) {
	fd, err := OpenFDNoRaise(path, flag, perm)
	if err != nil {
		return nil, err
	}
	file := os.NewFile(uintptr(fd), path)
	if file == nil {
		closeFD(fd)
		return nil, errors.New("failed to wrap file descriptor")
	}
	return file, nil
}

func OpenFD(path string, flag int, perm fs.FileMode) (int, error) {
	return openFD(path, flag, perm, true)
}

func OpenFDNoRaise(path string, flag int, perm fs.FileMode) (int, error) {
	return openFD(path, flag, perm, false)
}

func WFopen(path []rune, mode string) (*os.File, error) {
	nativePath, err := runesToNativePath(path)
	if err != nil {
		return nil, err
	}
	flags, perm, err := fopenMode(mode)
	if err != nil {
		return nil, err
	}
	return OpenFileNoRaise(nativePath, flags, perm)
}

func Fopen(path string, mode string) (*os.File, error) {
	flags, perm, err := fopenMode(mode)
	if err != nil {
		return nil, err
	}
	return OpenFile(path, flags, perm)
}

func Fclose(file *os.File) error {
	if file == nil {
		return nil
	}
	return file.Close()
}

func WAbsPath(path []rune) ([]rune, error) {
	nativePath, err := runesToNativePath(path)
	if err != nil {
		return nil, err
	}
	absPath, err := AbsPath(nativePath)
	if err != nil {
		return nil, err
	}
	return DecodeLocale([]byte(absPath))
}

func GetForceASCII() int {
	if forceASCIICache == -1 {
		if checkForceASCII() {
			forceASCIICache = 1
		} else {
			forceASCIICache = 0
		}
	}
	return forceASCIICache
}

func ResetForceASCII() {
	forceASCIICache = -1
}

func DecodeLocale(arg []byte) ([]rune, error) {
	return DecodeLocaleEx(arg, false, ErrorSurrogateEscape)
}

func DecodeLocaleEx(arg []byte, currentLocale bool, errors ErrorHandler) ([]rune, error) {
	useUTF8 := true
	if !currentLocale && GetForceASCII() == 1 {
		useUTF8 = false
	}
	if currentLocale && GetForceASCII() == 1 {
		useUTF8 = false
	}
	if useUTF8 {
		return decodeUTF8WithSurrogates(arg, errors)
	}
	return decodeASCII(arg, errors)
}

func EncodeLocale(text []rune) ([]byte, error) {
	return EncodeLocaleEx(text, false, ErrorSurrogateEscape)
}

func EncodeLocaleRaw(text []rune) ([]byte, error) {
	return EncodeLocaleEx(text, false, ErrorSurrogateEscape)
}

func EncodeLocaleEx(text []rune, currentLocale bool, errors ErrorHandler) ([]byte, error) {
	useUTF8 := true
	if !currentLocale && GetForceASCII() == 1 {
		useUTF8 = false
	}
	if currentLocale && GetForceASCII() == 1 {
		useUTF8 = false
	}
	if useUTF8 {
		return encodeUTF8WithSurrogates(text, errors)
	}
	return encodeASCII(text, errors)
}

func LocaleEncoding() string {
	if runtime.GOOS == "windows" {
		return "cp1252"
	}
	lang := localeEnv()
	if lang == "" {
		return "utf-8"
	}
	if strings.Contains(strings.ToUpper(lang), "UTF-8") || strings.Contains(strings.ToUpper(lang), "UTF8") {
		return "utf-8"
	}
	if GetForceASCII() == 1 {
		return "ascii"
	}
	return "utf-8"
}

func localeEnv() string {
	for _, key := range []string{"LC_ALL", "LC_CTYPE", "LANG"} {
		if value := os.Getenv(key); value != "" {
			return value
		}
	}
	return ""
}

func checkForceASCII() bool {
	if runtime.GOOS == "windows" {
		return false
	}
	loc := localeEnv()
	if loc == "" {
		return false
	}
	base := loc
	if i := strings.IndexByte(base, '.'); i >= 0 {
		base = base[:i]
	}
	if base != "C" && base != "POSIX" {
		return false
	}
	codeset := ""
	if i := strings.IndexByte(loc, '.'); i >= 0 && i+1 < len(loc) {
		codeset = strings.ToLower(loc[i+1:])
	}
	if codeset == "" {
		return true
	}
	asciiAliases := map[string]bool{
		"ascii": true, "646": true, "ansi_x3.4_1968": true, "ansi_x3.4_1986": true,
		"ansi_x3_4_1968": true, "cp367": true, "csascii": true, "ibm367": true,
		"iso646_us": true, "iso_646.irv_1991": true, "iso_ir_6": true, "us": true, "us_ascii": true,
	}
	return asciiAliases[codeset]
}

func decodeASCII(arg []byte, errors ErrorHandler) ([]rune, error) {
	out := make([]rune, 0, len(arg))
	for i, ch := range arg {
		if ch < 128 {
			out = append(out, rune(ch))
			continue
		}
		if errors != ErrorSurrogateEscape {
			return nil, &DecodeLocaleError{Reason: "decoding error", Pos: i}
		}
		out = append(out, rune(0xdc00)+rune(ch))
	}
	return out, nil
}

func decodeUTF8WithSurrogates(arg []byte, errors ErrorHandler) ([]rune, error) {
	out := make([]rune, 0, len(arg))
	for i := 0; i < len(arg); {
		r, size := utf8.DecodeRune(arg[i:])
		if r == utf8.RuneError && size == 1 {
			if errors != ErrorSurrogateEscape {
				return nil, &DecodeLocaleError{Reason: "decoding error", Pos: i}
			}
			out = append(out, rune(0xdc00)+rune(arg[i]))
			i++
			continue
		}
		if 0xD800 <= r && r <= 0xDFFF {
			if errors != ErrorSurrogateEscape {
				return nil, &DecodeLocaleError{Reason: "decoding error", Pos: i}
			}
			for _, b := range arg[i : i+size] {
				out = append(out, rune(0xdc00)+rune(b))
			}
			i += size
			continue
		}
		out = append(out, r)
		i += size
	}
	return out, nil
}

func encodeASCII(text []rune, errors ErrorHandler) ([]byte, error) {
	out := bytes.Buffer{}
	for i, r := range text {
		switch {
		case r <= 0x7f:
			out.WriteByte(byte(r))
		case errors == ErrorSurrogateEscape && 0xdc80 <= r && r <= 0xdcff:
			out.WriteByte(byte(r - 0xdc00))
		default:
			return nil, &EncodeLocaleError{Reason: "encoding error", Pos: i}
		}
	}
	return out.Bytes(), nil
}

func encodeUTF8WithSurrogates(text []rune, errors ErrorHandler) ([]byte, error) {
	out := bytes.Buffer{}
	var tmp [utf8.UTFMax]byte
	for i, r := range text {
		if errors == ErrorSurrogateEscape && 0xdc80 <= r && r <= 0xdcff {
			out.WriteByte(byte(r - 0xdc00))
			continue
		}
		if 0xD800 <= r && r <= 0xDFFF {
			return nil, &EncodeLocaleError{Reason: "encoding error", Pos: i}
		}
		n := utf8.EncodeRune(tmp[:], r)
		out.Write(tmp[:n])
	}
	return out.Bytes(), nil
}

func runesToNativePath(path []rune) (string, error) {
	encoded, err := EncodeLocaleRaw(path)
	if err != nil {
		return "", err
	}
	return string(encoded), nil
}

func fopenMode(mode string) (int, fs.FileMode, error) {
	if mode == "" {
		return 0, 0, errors.New("empty mode")
	}
	readWrite := false
	exclusive := false
	for _, ch := range mode[1:] {
		switch ch {
		case '+':
			readWrite = true
		case 'x':
			exclusive = true
		case 'b', 't':
		default:
			return 0, 0, errors.New("invalid fopen mode")
		}
	}

	flags := 0
	switch mode[0] {
	case 'r':
		if readWrite {
			flags = os.O_RDWR
		} else {
			flags = os.O_RDONLY
		}
		if exclusive {
			return 0, 0, errors.New("exclusive flag requires write mode")
		}
	case 'w':
		if readWrite {
			flags = os.O_RDWR
		} else {
			flags = os.O_WRONLY
		}
		flags |= os.O_CREATE | os.O_TRUNC
		if exclusive {
			flags |= os.O_EXCL
		}
	case 'a':
		if readWrite {
			flags = os.O_RDWR
		} else {
			flags = os.O_WRONLY
		}
		flags |= os.O_CREATE | os.O_APPEND
		if exclusive {
			flags |= os.O_EXCL
		}
	default:
		return 0, 0, errors.New("invalid fopen mode")
	}
	return flags, 0o666, nil
}
