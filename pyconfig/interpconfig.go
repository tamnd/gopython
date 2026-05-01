package pyconfig

import "fmt"

type GILMode int

const (
	DefaultGIL GILMode = iota
	SharedGIL
	OwnGIL
)

const (
	RuntimeFlagUseMainObmalloc uint32 = 1 << iota
	RuntimeFlagFork
	RuntimeFlagExec
	RuntimeFlagThreads
	RuntimeFlagDaemonThreads
	RuntimeFlagMultiInterpExtensions
)

type InterpreterConfig struct {
	UseMainObmalloc            bool
	AllowFork                  bool
	AllowExec                  bool
	AllowThreads               bool
	AllowDaemonThreads         bool
	CheckMultiInterpExtensions bool
	GIL                        GILMode
}

type InterpreterStateSnapshot struct {
	FeatureFlags uint32
	OwnGIL       bool
}

func (mode GILMode) String() (string, error) {
	switch mode {
	case DefaultGIL:
		return "default", nil
	case SharedGIL:
		return "shared", nil
	case OwnGIL:
		return "own", nil
	default:
		return "", fmt.Errorf("invalid interpreter config 'gil' value")
	}
}

func (config InterpreterConfig) AsMap() (map[string]any, error) {
	gil, err := config.GIL.String()
	if err != nil {
		return nil, err
	}
	return map[string]any{
		"use_main_obmalloc":             config.UseMainObmalloc,
		"allow_fork":                    config.AllowFork,
		"allow_exec":                    config.AllowExec,
		"allow_threads":                 config.AllowThreads,
		"allow_daemon_threads":          config.AllowDaemonThreads,
		"check_multi_interp_extensions": config.CheckMultiInterpExtensions,
		"gil":                           gil,
	}, nil
}

func InitInterpreterConfigFromMap(config *InterpreterConfig, dict map[string]any) error {
	if config == nil {
		return fmt.Errorf("nil interpreter config")
	}
	return interpreterConfigFromMap(config, dict, false)
}

func UpdateInterpreterConfigFromMap(config *InterpreterConfig, dict map[string]any) error {
	if config == nil {
		return fmt.Errorf("nil interpreter config")
	}
	return interpreterConfigFromMap(config, dict, true)
}

func (config *InterpreterConfig) InitFromState(interp InterpreterStateSnapshot) {
	config.UseMainObmalloc = interp.FeatureFlags&RuntimeFlagUseMainObmalloc != 0
	config.AllowFork = interp.FeatureFlags&RuntimeFlagFork != 0
	config.AllowExec = interp.FeatureFlags&RuntimeFlagExec != 0
	config.AllowThreads = interp.FeatureFlags&RuntimeFlagThreads != 0
	config.AllowDaemonThreads = interp.FeatureFlags&RuntimeFlagDaemonThreads != 0
	config.CheckMultiInterpExtensions = interp.FeatureFlags&RuntimeFlagMultiInterpExtensions != 0
	if interp.OwnGIL {
		config.GIL = OwnGIL
	} else {
		config.GIL = SharedGIL
	}
}

func interpreterConfigFromMap(config *InterpreterConfig, dict map[string]any, missingAllowed bool) error {
	working := make(map[string]any, len(dict))
	for key, value := range dict {
		working[key] = value
	}

	if err := copyBoolField(working, "use_main_obmalloc", &config.UseMainObmalloc, missingAllowed); err != nil {
		return err
	}
	if err := copyBoolField(working, "allow_fork", &config.AllowFork, missingAllowed); err != nil {
		return err
	}
	if err := copyBoolField(working, "allow_exec", &config.AllowExec, missingAllowed); err != nil {
		return err
	}
	if err := copyBoolField(working, "allow_threads", &config.AllowThreads, missingAllowed); err != nil {
		return err
	}
	if err := copyBoolField(working, "allow_daemon_threads", &config.AllowDaemonThreads, missingAllowed); err != nil {
		return err
	}
	if err := copyBoolField(working, "check_multi_interp_extensions", &config.CheckMultiInterpExtensions, missingAllowed); err != nil {
		return err
	}

	if value, ok := working["gil"]; ok {
		text, ok := value.(string)
		if !ok {
			return invalidConfigType("gil")
		}
		mode, err := parseGILMode(text)
		if err != nil {
			return err
		}
		config.GIL = mode
		delete(working, "gil")
	} else if !missingAllowed {
		return missingConfigKey("gil")
	}

	if extra := len(working); extra == 1 {
		return fmt.Errorf("config dict has 1 extra item")
	} else if extra > 0 {
		return fmt.Errorf("config dict has %d extra items", extra)
	}
	return nil
}

func copyBoolField(dict map[string]any, name string, dst *bool, missingAllowed bool) error {
	value, ok := dict[name]
	if !ok {
		if missingAllowed {
			return nil
		}
		return missingConfigKey(name)
	}
	flag, ok := value.(bool)
	if !ok {
		return invalidConfigType(name)
	}
	*dst = flag
	delete(dict, name)
	return nil
}

func parseGILMode(text string) (GILMode, error) {
	switch text {
	case "", "default":
		return DefaultGIL, nil
	case "shared":
		return SharedGIL, nil
	case "own":
		return OwnGIL, nil
	default:
		return DefaultGIL, fmt.Errorf("unsupported interpreter config .gil value %q", text)
	}
}

func missingConfigKey(name string) error {
	return fmt.Errorf("missing config key: %s", name)
}

func invalidConfigType(name string) error {
	return fmt.Errorf("invalid config type: %s", name)
}
