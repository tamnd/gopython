package pyruntime

import "errors"

var (
	errSinglePhaseExtension = errors.New("per-interpreter obmalloc does not support single-phase init extension modules")
	errInvalidGIL           = errors.New("invalid interpreter config 'gil' value")
)
