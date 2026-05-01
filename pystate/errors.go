package pystate

import "errors"

var (
	ErrNoInterpreter    = errors.New("no interpreter")
	ErrNoWatcherSlot    = errors.New("no more context watcher IDs available")
	ErrInvalidWatcherID = errors.New("invalid context watcher ID")
	ErrMissingWatcher   = errors.New("no context watcher set for ID")
)
