package pyconfig

import "github.com/tamnd/gopython/pyos"

const DefaultMaxStrDigits = 4300

type Config struct {
	ConfigInit            ConfigInit
	Isolated              int
	UseEnvironment        int
	DevMode               int
	InstallSignalHandlers int
	UseHashSeed           int
	HashSeed              uint64
	Faulthandler          int
	Tracemalloc           int
	PerfProfiling         int
	RemoteDebug           int
	ImportTime            int
	CodeDebugRanges       int
	ShowRefCount          int
	DumpRefs              int
	DumpRefsFile          []rune
	FilesystemEncoding    []rune
	FilesystemErrors      []rune
	PycachePrefix         []rune
	ParseArgv             int
	OrigArgv              WideStringList
	Argv                  WideStringList
	XOptions              WideStringList
	WarnOptions           WideStringList
	SiteImport            int
	BytesWarning          int
	WarnDefaultEncoding   int
	Inspect               int
	Interactive           int
	OptimizationLevel     int
	ParserDebug           int
	WriteBytecode         int
	Verbose               int
	Quiet                 int
	UserSiteDirectory     int
	ConfigureCStdio       int
	BufferedStdio         int
	StdioEncoding         []rune
	StdioErrors           []rune
	LegacyWindowsStdio    int
	CheckHashPycsMode     []rune
	UseFrozenModules      int
	SafePath              int
	IntMaxStrDigits       int
	ThreadInheritContext  int
	ContextAwareWarnings  int
	CPUCount              int
	PathconfigWarnings    int
	ProgramName           []rune
	PythonpathEnv         []rune
	Home                  []rune
	Platlibdir            []rune
	ModuleSearchPathsSet  int
	ModuleSearchPaths     WideStringList
	StdlibDir             []rune
	Executable            []rune
	BaseExecutable        []rune
	Prefix                []rune
	BasePrefix            []rune
	ExecPrefix            []rune
	BaseExecPrefix        []rune
	SkipSourceFirstLine   int
	RunCommand            []rune
	RunModule             []rune
	RunFilename           []rune
	SysPath0              []rune
	InstallImportlib      int
	InitMain              int
	IsPythonBuild         int
}

func InitCompatConfig(config *Config) {
	*config = Config{}
	config.ConfigInit = ConfigInitCompat
	config.ImportTime = -1
	config.Isolated = -1
	config.UseEnvironment = -1
	config.DevMode = -1
	config.InstallSignalHandlers = 1
	config.UseHashSeed = -1
	config.Faulthandler = -1
	config.Tracemalloc = -1
	config.PerfProfiling = -1
	config.RemoteDebug = -1
	config.ModuleSearchPathsSet = 0
	config.ParseArgv = 0
	config.SiteImport = -1
	config.BytesWarning = -1
	config.WarnDefaultEncoding = 0
	config.Inspect = -1
	config.Interactive = -1
	config.OptimizationLevel = -1
	config.ParserDebug = -1
	config.WriteBytecode = -1
	config.Verbose = -1
	config.Quiet = -1
	config.UserSiteDirectory = -1
	config.ConfigureCStdio = 0
	config.BufferedStdio = -1
	config.InstallImportlib = 1
	config.PathconfigWarnings = -1
	config.InitMain = 1
	config.LegacyWindowsStdio = -1
	config.UseFrozenModules = 1
	config.SafePath = 0
	config.IntMaxStrDigits = -1
	config.IsPythonBuild = 0
	config.CodeDebugRanges = 1
	config.CPUCount = -1
	config.ThreadInheritContext = 0
	config.ContextAwareWarnings = 0
}

func InitPythonConfig(config *Config) {
	initConfigDefaults(config)
	config.ConfigInit = ConfigInitPython
	config.ConfigureCStdio = 1
	config.ParseArgv = 1
}

func InitIsolatedConfig(config *Config) {
	initConfigDefaults(config)
	config.ConfigInit = ConfigInitIsolated
	config.Isolated = 1
	config.UseEnvironment = 0
	config.UserSiteDirectory = 0
	config.DevMode = 0
	config.InstallSignalHandlers = 0
	config.UseHashSeed = 0
	config.Tracemalloc = 0
	config.PerfProfiling = 0
	config.IntMaxStrDigits = DefaultMaxStrDigits
	config.SafePath = 1
	config.PathconfigWarnings = 0
	config.ThreadInheritContext = 0
	config.LegacyWindowsStdio = 0
}

func initConfigDefaults(config *Config) {
	InitCompatConfig(config)
	config.Isolated = 0
	config.UseEnvironment = 1
	config.SiteImport = 1
	config.BytesWarning = 0
	config.Inspect = 0
	config.Interactive = 0
	config.OptimizationLevel = 0
	config.ParserDebug = 0
	config.WriteBytecode = 1
	config.Verbose = 0
	config.Quiet = 0
	config.UserSiteDirectory = 1
	config.BufferedStdio = 1
	config.PathconfigWarnings = 1
	config.LegacyWindowsStdio = 0
}

func (config *Config) Clear() {
	config.PycachePrefix = nil
	config.PythonpathEnv = nil
	config.Home = nil
	config.ProgramName = nil
	config.Argv.Clear()
	config.WarnOptions.Clear()
	config.XOptions.Clear()
	config.ModuleSearchPaths.Clear()
	config.ModuleSearchPathsSet = 0
	config.StdlibDir = nil
	config.Executable = nil
	config.BaseExecutable = nil
	config.Prefix = nil
	config.BasePrefix = nil
	config.ExecPrefix = nil
	config.BaseExecPrefix = nil
	config.Platlibdir = nil
	config.SysPath0 = nil
	config.FilesystemEncoding = nil
	config.FilesystemErrors = nil
	config.StdioEncoding = nil
	config.StdioErrors = nil
	config.RunCommand = nil
	config.RunModule = nil
	config.RunFilename = nil
	config.CheckHashPycsMode = nil
	config.OrigArgv.Clear()
}

func SetString(config *Config, field *[]rune, text []rune) {
	_ = config
	*field = cloneRunes(text)
}

func SetBytesString(config *Config, field *[]rune, text []byte) error {
	_ = config
	if text == nil {
		*field = nil
		return nil
	}
	decoded, err := pyos.DecodeLocale(text)
	if err != nil {
		return err
	}
	*field = cloneRunes(decoded)
	return nil
}

func (config *Config) CopyFrom(other *Config) error {
	config.Clear()
	*config = *other
	config.DumpRefsFile = cloneRunes(other.DumpRefsFile)
	config.FilesystemEncoding = cloneRunes(other.FilesystemEncoding)
	config.FilesystemErrors = cloneRunes(other.FilesystemErrors)
	config.PycachePrefix = cloneRunes(other.PycachePrefix)
	config.StdioEncoding = cloneRunes(other.StdioEncoding)
	config.StdioErrors = cloneRunes(other.StdioErrors)
	config.CheckHashPycsMode = cloneRunes(other.CheckHashPycsMode)
	config.ProgramName = cloneRunes(other.ProgramName)
	config.PythonpathEnv = cloneRunes(other.PythonpathEnv)
	config.Home = cloneRunes(other.Home)
	config.Platlibdir = cloneRunes(other.Platlibdir)
	config.StdlibDir = cloneRunes(other.StdlibDir)
	config.Executable = cloneRunes(other.Executable)
	config.BaseExecutable = cloneRunes(other.BaseExecutable)
	config.Prefix = cloneRunes(other.Prefix)
	config.BasePrefix = cloneRunes(other.BasePrefix)
	config.ExecPrefix = cloneRunes(other.ExecPrefix)
	config.BaseExecPrefix = cloneRunes(other.BaseExecPrefix)
	config.RunCommand = cloneRunes(other.RunCommand)
	config.RunModule = cloneRunes(other.RunModule)
	config.RunFilename = cloneRunes(other.RunFilename)
	config.SysPath0 = cloneRunes(other.SysPath0)
	config.OrigArgv.Items = cloneRuneList(other.OrigArgv.Items)
	config.Argv.Items = cloneRuneList(other.Argv.Items)
	config.XOptions.Items = cloneRuneList(other.XOptions.Items)
	config.WarnOptions.Items = cloneRuneList(other.WarnOptions.Items)
	config.ModuleSearchPaths.Items = cloneRuneList(other.ModuleSearchPaths.Items)
	return nil
}

func (list *WideStringList) Insert(index int, item []rune) error {
	if index < 0 {
		return invalidWideStringListIndex()
	}
	if index > len(list.Items) {
		index = len(list.Items)
	}
	itemCopy := cloneRunes(item)
	list.Items = append(list.Items, nil)
	copy(list.Items[index+1:], list.Items[index:])
	list.Items[index] = itemCopy
	return nil
}

func (list *WideStringList) AppendWide(item []rune) error {
	return list.Insert(len(list.Items), item)
}

func (list *WideStringList) Extend(other WideStringList) error {
	for _, item := range other.Items {
		if err := list.AppendWide(item); err != nil {
			return err
		}
	}
	return nil
}

func (list *WideStringList) CopyFrom(other WideStringList) {
	list.Items = cloneRuneList(other.Items)
}

func (list WideStringList) Find(item []rune) bool {
	for _, current := range list.Items {
		if runesEqual(current, item) {
			return true
		}
	}
	return false
}

func cloneRunes(src []rune) []rune {
	if src == nil {
		return nil
	}
	return append([]rune(nil), src...)
}

func invalidWideStringListIndex() error {
	return missingConfigKey("PyWideStringList_Insert index must be >= 0")
}
