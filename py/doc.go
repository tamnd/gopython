// Package py is the public entry point for embedding gopython.
//
// Phase 0 contains repository and CI scaffolding only. Python execution is not
// implemented yet.
package py

const Version = "v0.0.0"

// Status returns the current project status for callers and the CLI.
func Status() string {
	return "repository scaffold only"
}
