package app

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// Build identity — release channels (stable / beta / nightly). buildVersion and
// buildChannel are set at release time via -ldflags
//   -X github.com/maridlabsai/jini/internal/app.buildVersion=0.1.3
//   -X github.com/maridlabsai/jini/internal/app.buildChannel=beta
// A plain source build leaves them empty; jiniVersion() then falls back to the
// VERSION file (or "dev") and the "source" channel, so `jini version` always
// says something true about how this binary was produced.
var (
	buildVersion string
	buildChannel string
)

const versionFallback = "dev"

func jiniVersion() string {
	if v := strings.TrimSpace(buildVersion); v != "" {
		return v
	}
	if v := versionFromFile(); v != "" {
		return v
	}
	return versionFallback
}

func jiniChannel() string {
	if c := strings.TrimSpace(buildChannel); c != "" {
		return c
	}
	return "source"
}

// versionFromFile reads the repo VERSION file for source builds (best-effort).
func versionFromFile() string {
	if exe, err := os.Executable(); err == nil {
		if data, err := os.ReadFile(filepath.Join(filepath.Dir(exe), "VERSION")); err == nil {
			if v := strings.TrimSpace(string(data)); v != "" {
				return v
			}
		}
	}
	if data, err := os.ReadFile("VERSION"); err == nil {
		return strings.TrimSpace(string(data))
	}
	return ""
}

// runVersion prints "jini <version> (<channel>)" — the one line a user or a bug
// report needs. Kept intentionally minimal.
func runVersion(stdout io.Writer) int {
	fmt.Fprintf(stdout, "jini %s (%s)\n", jiniVersion(), jiniChannel())
	return 0
}
