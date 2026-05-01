package pyos

import (
	"os"
	"path/filepath"
	"strings"
	"sync"
)

type Config struct {
	ProgramFullPath            string
	Prefix                     string
	ExecPrefix                 string
	StdlibDir                  string
	ModuleSearchPath           string
	CalculatedModuleSearchPath string
	ProgramName                string
	Home                       string
	Executable                 string
	ModuleSearchPaths          []string
	IsPythonBuild              bool
}

var (
	pathConfigMu        sync.RWMutex
	pathConfig          Config
	moduleSearchPathSet bool
)

func GetGlobalModuleSearchPath() string {
	pathConfigMu.RLock()
	defer pathConfigMu.RUnlock()
	return pathConfig.ModuleSearchPath
}

func ClearGlobalPathConfig() {
	pathConfigMu.Lock()
	defer pathConfigMu.Unlock()
	pathConfig = Config{}
	moduleSearchPathSet = false
}

func ReadGlobalPathConfig(config *Config) {
	pathConfigMu.RLock()
	defer pathConfigMu.RUnlock()

	if pathConfig.Prefix != "" && config.Prefix == "" {
		config.Prefix = pathConfig.Prefix
	}
	if pathConfig.ExecPrefix != "" && config.ExecPrefix == "" {
		config.ExecPrefix = pathConfig.ExecPrefix
	}
	if pathConfig.StdlibDir != "" && config.StdlibDir == "" {
		config.StdlibDir = pathConfig.StdlibDir
	}
	if pathConfig.ProgramName != "" && config.ProgramName == "" {
		config.ProgramName = pathConfig.ProgramName
	}
	if pathConfig.Home != "" && config.Home == "" {
		config.Home = pathConfig.Home
	}
	if pathConfig.ProgramFullPath != "" && config.Executable == "" {
		config.Executable = pathConfig.ProgramFullPath
	}
	if pathConfig.IsPythonBuild && !config.IsPythonBuild {
		config.IsPythonBuild = true
	}
}

func UpdateGlobalPathConfig(config Config) {
	pathConfigMu.Lock()
	defer pathConfigMu.Unlock()

	if config.Prefix != "" {
		pathConfig.Prefix = config.Prefix
	}
	if config.ExecPrefix != "" {
		pathConfig.ExecPrefix = config.ExecPrefix
	}
	if config.StdlibDir != "" {
		pathConfig.StdlibDir = config.StdlibDir
	}
	if config.ProgramName != "" {
		pathConfig.ProgramName = config.ProgramName
	}
	if config.Home != "" {
		pathConfig.Home = config.Home
	}
	if config.Executable != "" {
		pathConfig.ProgramFullPath = config.Executable
	}
	if config.IsPythonBuild {
		pathConfig.IsPythonBuild = true
	}

	pathConfig.ModuleSearchPath = ""
	pathConfig.CalculatedModuleSearchPath = strings.Join(config.ModuleSearchPaths, string(os.PathListSeparator))
	moduleSearchPathSet = false
}

func SetPath(path string) {
	pathConfigMu.Lock()
	defer pathConfigMu.Unlock()

	pathConfig.Prefix = ""
	pathConfig.ExecPrefix = ""
	if pathConfig.Home != "" {
		pathConfig.StdlibDir = pathConfig.Home
	} else {
		pathConfig.StdlibDir = ""
	}
	pathConfig.ModuleSearchPath = path
	pathConfig.CalculatedModuleSearchPath = ""
	moduleSearchPathSet = true
}

func SetPythonHome(home string) {
	pathConfigMu.Lock()
	defer pathConfigMu.Unlock()
	pathConfig.Home = home
}

func SetProgramName(programName string) {
	pathConfigMu.Lock()
	defer pathConfigMu.Unlock()
	pathConfig.ProgramName = programName
}

func GetPath() string {
	pathConfigMu.RLock()
	defer pathConfigMu.RUnlock()
	if moduleSearchPathSet {
		return pathConfig.ModuleSearchPath
	}
	return pathConfig.CalculatedModuleSearchPath
}

func GetStdlibDir() string {
	pathConfigMu.RLock()
	defer pathConfigMu.RUnlock()
	if pathConfig.StdlibDir == "" {
		return ""
	}
	return pathConfig.StdlibDir
}

func GetPrefix() string {
	pathConfigMu.RLock()
	defer pathConfigMu.RUnlock()
	return pathConfig.Prefix
}

func GetExecPrefix() string {
	pathConfigMu.RLock()
	defer pathConfigMu.RUnlock()
	return pathConfig.ExecPrefix
}

func GetProgramFullPath() string {
	pathConfigMu.RLock()
	defer pathConfigMu.RUnlock()
	return pathConfig.ProgramFullPath
}

func GetPythonHome() string {
	pathConfigMu.RLock()
	defer pathConfigMu.RUnlock()
	return pathConfig.Home
}

func GetProgramName() string {
	pathConfigMu.RLock()
	defer pathConfigMu.RUnlock()
	return pathConfig.ProgramName
}

func ComputeSysPath0(argv []string) (string, bool, error) {
	if len(argv) == 0 {
		return "", false, nil
	}

	argv0 := argv[0]
	haveModuleArg := argv0 == "-m"
	haveScriptArg := !haveModuleArg && argv0 != "-c"
	path0 := argv0

	if haveModuleArg {
		cwd, err := os.Getwd()
		if err != nil {
			return "", false, nil
		}
		return cwd, true, nil
	}

	if !haveScriptArg {
		return "", true, nil
	}

	if link, err := os.Readlink(path0); err == nil {
		switch {
		case filepath.IsAbs(link):
			path0 = link
		case !strings.ContainsRune(link, os.PathSeparator):
			// Keep path0 unchanged.
		default:
			dir := filepath.Dir(path0)
			path0 = filepath.Join(dir, link)
		}
	}

	if abs, err := filepath.Abs(path0); err == nil {
		path0 = abs
	}
	if resolved, err := filepath.EvalSymlinks(path0); err == nil {
		path0 = resolved
	}

	dir := filepath.Dir(path0)
	if dir == "." {
		dir = ""
	}
	if dir == string(os.PathSeparator) {
		return dir, true, nil
	}
	return strings.TrimRight(dir, string(os.PathSeparator)), true, nil
}
