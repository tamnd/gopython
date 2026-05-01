package pycrossinterp

import (
	"fmt"
	"strings"
)

type ExcInfoType struct {
	Name     string
	QualName string
	Module   string
}

type ExcInfo struct {
	Type       ExcInfoType
	Message    string
	ErrDisplay string
}

type ExcSnapshotObject struct {
	Type       map[string]string
	Msg        string
	Formatted  string
	ErrDisplay string
}

func NewExcInfo(exc any) (*ExcInfo, error) {
	switch value := exc.(type) {
	case nil:
		return nil, fmt.Errorf("missing exc")
	case ExcSnapshotObject:
		return InitExcInfoFromObject(value)
	case *ExcSnapshotObject:
		if value == nil {
			return nil, fmt.Errorf("missing exc")
		}
		return InitExcInfoFromObject(*value)
	case error:
		return InitExcInfoFromException("Exception", "Exception", "builtins", value.Error(), value.Error()), nil
	default:
		return InitExcInfoFromException("Exception", "Exception", "builtins", fmt.Sprint(value), fmt.Sprint(value)), nil
	}
}

func FreeExcInfo(info *ExcInfo) {
	if info != nil {
		info.Clear()
	}
}

func FormatExcInfo(info *ExcInfo) string {
	if info == nil {
		return ""
	}
	return info.Format()
}

func ExcInfoAsObject(info *ExcInfo) ExcSnapshotObject {
	if info == nil {
		return ExcSnapshotObject{}
	}
	return info.AsObject()
}

func (info *ExcInfo) Clear() {
	*info = ExcInfo{}
}

func (info ExcInfo) IsSet() bool {
	return info.Type.Name != "" || info.Message != ""
}

func (info ExcInfo) Format() string {
	module := info.Type.Module
	qualname := info.Type.QualName
	if qualname == "" {
		qualname = info.Type.Name
	}
	if module == "builtins" || module == "__main__" {
		module = ""
	}
	if qualname != "" {
		if module != "" {
			if info.Message != "" {
				return fmt.Sprintf("%s.%s: %s", module, qualname, info.Message)
			}
			return fmt.Sprintf("%s.%s", module, qualname)
		}
		if info.Message != "" {
			return fmt.Sprintf("%s: %s", qualname, info.Message)
		}
		return qualname
	}
	return info.Message
}

func InitExcInfoFromException(typeName, qualName, module, message, errDisplay string) *ExcInfo {
	return &ExcInfo{
		Type: ExcInfoType{
			Name:     typeName,
			QualName: qualName,
			Module:   module,
		},
		Message:    message,
		ErrDisplay: strings.TrimSuffix(errDisplay, "\n"),
	}
}

func InitExcInfoFromObject(obj ExcSnapshotObject) (*ExcInfo, error) {
	typeName, ok := obj.Type["__name__"]
	if !ok {
		return nil, fmt.Errorf("exception snapshot missing 'type' attribute")
	}
	return &ExcInfo{
		Type: ExcInfoType{
			Name:     typeName,
			QualName: obj.Type["__qualname__"],
			Module:   obj.Type["__module__"],
		},
		Message:    obj.Msg,
		ErrDisplay: obj.ErrDisplay,
	}, nil
}

func (info ExcInfo) Apply() string {
	if info.ErrDisplay != "" {
		return info.ErrDisplay
	}
	return info.Format()
}

func (info ExcInfo) TypeAsObject() map[string]string {
	result := map[string]string{}
	if info.Type.Name != "" {
		result["__name__"] = info.Type.Name
	}
	if info.Type.QualName != "" {
		result["__qualname__"] = info.Type.QualName
	}
	if info.Type.Module != "" {
		result["__module__"] = info.Type.Module
	}
	if len(result) == 0 {
		return nil
	}
	return result
}

func (info ExcInfo) AsObject() ExcSnapshotObject {
	return ExcSnapshotObject{
		Type:       info.TypeAsObject(),
		Msg:        info.Message,
		Formatted:  info.Format(),
		ErrDisplay: info.ErrDisplay,
	}
}
