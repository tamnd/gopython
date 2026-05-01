package pyconfig

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"sync"

	"github.com/tamnd/gopython/pyos"
)

type ConfigInit int

const (
	ConfigInitCompat   ConfigInit = 1
	ConfigInitPython   ConfigInit = 2
	ConfigInitIsolated ConfigInit = 3
)

const MemoryAllocatorNotSet = -1

type PreConfig struct {
	ConfigInit              ConfigInit
	ParseArgv               int
	Isolated                int
	UseEnvironment          int
	ConfigureLocale         int
	CoerceCLocale           int
	CoerceCLocaleWarn       int
	LegacyWindowsFSEncoding int
	UTF8Mode                int
	DevMode                 int
	Allocator               int
}

type Argv struct {
	UseBytesArgv bool
	BytesArgv    [][]byte
	WideArgv     [][]rune
}

type WideStringList struct {
	Items [][]rune
}

type fileSystemEncodingState struct {
	mu       sync.Mutex
	encoding string
	errors   string
}

var fsEncodingState fileSystemEncodingState

func ClearFileSystemEncoding() {
	fsEncodingState.mu.Lock()
	defer fsEncodingState.mu.Unlock()
	fsEncodingState.encoding = ""
	fsEncodingState.errors = ""
}

func SetFileSystemEncoding(encoding, errors string) {
	fsEncodingState.mu.Lock()
	defer fsEncodingState.mu.Unlock()
	fsEncodingState.encoding = encoding
	fsEncodingState.errors = errors
}

func FileSystemEncoding() string {
	fsEncodingState.mu.Lock()
	defer fsEncodingState.mu.Unlock()
	return fsEncodingState.encoding
}

func FileSystemErrors() string {
	fsEncodingState.mu.Lock()
	defer fsEncodingState.mu.Unlock()
	return fsEncodingState.errors
}

func ArgvAsWideStringList(args *Argv, list *WideStringList) error {
	var result WideStringList
	if args.UseBytesArgv {
		result.Items = make([][]rune, 0, len(args.BytesArgv))
		for _, arg := range args.BytesArgv {
			text, err := pyos.DecodeLocale(arg)
			if err != nil {
				return err
			}
			result.Items = append(result.Items, append([]rune(nil), text...))
		}
	} else {
		result.Items = cloneRuneList(args.WideArgv)
	}
	list.Clear()
	list.Items = result.Items
	return nil
}

func (list *WideStringList) Clear() {
	list.Items = nil
}

func (list *WideStringList) Append(item []rune) {
	list.Items = append(list.Items, append([]rune(nil), item...))
}

func InitCompatPreConfig(config *PreConfig) {
	*config = PreConfig{
		ConfigInit:              ConfigInitCompat,
		ParseArgv:               0,
		Isolated:                -1,
		UseEnvironment:          -1,
		ConfigureLocale:         1,
		CoerceCLocale:           0,
		CoerceCLocaleWarn:       0,
		LegacyWindowsFSEncoding: -1,
		UTF8Mode:                0,
		DevMode:                 -1,
		Allocator:               MemoryAllocatorNotSet,
	}
}

func InitPythonPreConfig(config *PreConfig) {
	InitCompatPreConfig(config)
	config.ConfigInit = ConfigInitPython
	config.Isolated = 0
	config.ParseArgv = 1
	config.UseEnvironment = 1
	config.CoerceCLocale = -1
	config.CoerceCLocaleWarn = -1
	config.UTF8Mode = -1
	config.LegacyWindowsFSEncoding = 0
}

func InitIsolatedPreConfig(config *PreConfig) {
	InitCompatPreConfig(config)
	config.ConfigInit = ConfigInitIsolated
	config.ConfigureLocale = 0
	config.Isolated = 1
	config.UseEnvironment = 0
	config.UTF8Mode = 0
	config.DevMode = 0
	config.LegacyWindowsFSEncoding = 0
}

func (config PreConfig) AsMap() map[string]int {
	return map[string]int{
		"_config_init":               int(config.ConfigInit),
		"parse_argv":                 config.ParseArgv,
		"isolated":                   config.Isolated,
		"use_environment":            config.UseEnvironment,
		"configure_locale":           config.ConfigureLocale,
		"coerce_c_locale":            config.CoerceCLocale,
		"coerce_c_locale_warn":       config.CoerceCLocaleWarn,
		"legacy_windows_fs_encoding": config.LegacyWindowsFSEncoding,
		"utf8_mode":                  config.UTF8Mode,
		"dev_mode":                   config.DevMode,
		"allocator":                  config.Allocator,
	}
}

func GetEnv(useEnvironment int, name string) string {
	if useEnvironment <= 0 {
		return ""
	}
	value := os.Getenv(name)
	if value == "" {
		return ""
	}
	return value
}

func StrToInt(text string) (int, error) {
	value, err := strconv.ParseInt(text, 10, 0)
	if err != nil {
		return 0, err
	}
	return int(value), nil
}

func GetEnvFlag(useEnvironment int, flag *int, name string) {
	value := GetEnv(useEnvironment, name)
	if value == "" {
		return
	}
	parsed, err := StrToInt(value)
	if err != nil || parsed < 0 {
		parsed = 1
	}
	if *flag < parsed {
		*flag = parsed
	}
}

func GetXOption(xoptions WideStringList, name string) []rune {
	nameRunes := []rune(name)
	for _, option := range xoptions.Items {
		text := string(option)
		if index := strings.IndexRune(text, '='); index >= 0 {
			option = option[:index]
		}
		if runesEqual(option, nameRunes) {
			return option
		}
	}
	return nil
}

func cloneRuneList(src [][]rune) [][]rune {
	if len(src) == 0 {
		return nil
	}
	dst := make([][]rune, len(src))
	for i, item := range src {
		dst[i] = append([]rune(nil), item...)
	}
	return dst
}

func runesEqual(left, right []rune) bool {
	if len(left) != len(right) {
		return false
	}
	for i := range left {
		if left[i] != right[i] {
			return false
		}
	}
	return true
}

func (config *PreConfig) CopyFrom(other PreConfig) {
	*config = other
}

func (list WideStringList) StringItems() []string {
	out := make([]string, len(list.Items))
	for i, item := range list.Items {
		out[i] = string(item)
	}
	return out
}

func ValidateUTF8XOption(option []rune) (int, error) {
	text := string(option)
	if text == "utf8" {
		return 1, nil
	}
	if strings.HasPrefix(text, "utf8=") {
		switch strings.TrimPrefix(text, "utf8=") {
		case "0":
			return 0, nil
		case "1":
			return 1, nil
		default:
			return 0, fmt.Errorf("invalid -X utf8 option value")
		}
	}
	return -1, nil
}
