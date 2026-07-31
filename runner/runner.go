// Package runner is the PUBLIC entrypoint wrapper for the Jini CLI.
//
// A separate module (e.g. ../jini-commercial) cannot import internal/app, but
// it can import this package to build its own binary that links a paid
// Autopilot strategy (registered via the autopilot package) into the same
// process. The public cmd/jini binary and a commercial binary therefore run
// the same core, differing only by what extensions are registered at init.
package runner

import (
	"io"

	"github.com/maridlabsai/jini/internal/app"
)

// Run executes the Jini CLI with the given args and streams, returning the
// process exit code. Mirrors what cmd/jini/main.go does.
func Run(args []string, stdin io.Reader, stdout, stderr io.Writer) int {
	return app.RunInteractive(args, stdin, stdout, stderr)
}
