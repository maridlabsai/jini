package app

// Auto/Ask execution mode — specs/auto-ask-execution-mode-design.md.
// GLOBAL setting (~/.jini/mode.json), deliberately diverging from the
// per-project sessionStateRoot() pattern: supervision is a property of the
// user, not the workspace. Parsing fails CLOSED: a corrupted safety toggle
// must degrade to supervision ("ask"), not autonomy — this deliberately
// diverges from router_settings.go's fail-open load. An absent file is not
// an error; that is the normal "auto" default.

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

const (
	executionModeAuto = "auto"
	executionModeAsk  = "ask"
)

var (
	executionModeWarnings io.Writer = os.Stderr
	// executionModeHomeDir resolves the home that holds GLOBAL per-user Jini
	// state (~/.jini/mode.json, trusted-dirs.json, savings-ledger.json). It
	// honors JINI_HOME so a dogfood or test run can redirect all of that to a
	// sandbox instead of polluting the real ledger/trust; without the override
	// it is the OS home. (Repo-scoped state uses sessionStateRoot/JINI_STATE_DIR
	// separately.) Overridable as a var so TestMain can pin it.
	executionModeHomeDir = jiniHomeDir
)

// jiniHomeDir returns JINI_HOME when set (sandbox override), else the OS home.
func jiniHomeDir() (string, error) {
	if override := strings.TrimSpace(os.Getenv("JINI_HOME")); override != "" {
		return override, nil
	}
	return os.UserHomeDir()
}

type savedExecutionMode struct {
	SchemaVersion string `json:"schema_version"`
	ContextType   string `json:"context_type"`
	Mode          string `json:"mode"`
}

func executionModePath() (string, error) {
	home, err := executionModeHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".jini", "mode.json"), nil
}

// effectiveExecutionMode resolves JINI_MODE env, then ~/.jini/mode.json,
// then "auto". Every unreadable or unknown source fails closed to "ask".
func effectiveExecutionMode() string {
	if raw := strings.TrimSpace(os.Getenv("JINI_MODE")); raw != "" {
		return parseExecutionMode(raw, "JINI_MODE")
	}
	path, err := executionModePath()
	if err != nil {
		return warnExecutionModeUnreadable("home directory")
	}
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return executionModeAuto
	}
	if err != nil {
		return warnExecutionModeUnreadable(path)
	}
	var payload savedExecutionMode
	if err := json.Unmarshal(data, &payload); err != nil {
		return warnExecutionModeUnreadable(path)
	}
	return parseExecutionMode(payload.Mode, path)
}

func parseExecutionMode(raw, source string) string {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case executionModeAuto:
		return executionModeAuto
	case executionModeAsk:
		return executionModeAsk
	default:
		return warnExecutionModeUnreadable(source)
	}
}

func warnExecutionModeUnreadable(source string) string {
	if executionModeWarnings != nil {
		fmt.Fprintf(executionModeWarnings, "mode setting unreadable (%s); treating as Ask (supervised)\n", source)
	}
	return executionModeAsk
}

func saveExecutionMode(mode string) error {
	if mode != executionModeAuto && mode != executionModeAsk {
		return fmt.Errorf("unknown mode %q", mode)
	}
	path, err := executionModePath()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	payload := savedExecutionMode{SchemaVersion: "0.1.0", ContextType: "JiniExecutionMode", Mode: mode}
	data, err := json.MarshalIndent(payload, "", "  ")
	if err != nil {
		return err
	}
	tmp, err := os.CreateTemp(filepath.Dir(path), "mode-*.json.tmp")
	if err != nil {
		return err
	}
	tmpName := tmp.Name()
	if _, err := tmp.Write(append(data, '\n')); err != nil {
		tmp.Close()
		os.Remove(tmpName)
		return err
	}
	if err := tmp.Chmod(0o600); err != nil {
		tmp.Close()
		os.Remove(tmpName)
		return err
	}
	if err := tmp.Close(); err != nil {
		os.Remove(tmpName)
		return err
	}
	return os.Rename(tmpName, path)
}

func executionModeDisplayLine(mode string) string {
	if mode == executionModeAsk {
		return "Ask — Jini checks with you before side effects and before resuming throttled work."
	}
	return "Auto — Jini picks for you and keeps going."
}

func runMode(args []string, stdout, stderr io.Writer) int {
	if len(args) == 0 {
		mode := effectiveExecutionMode()
		line := executionModeDisplayLine(mode)
		if env := strings.TrimSpace(os.Getenv("JINI_MODE")); env != "" {
			saved := executionModeAuto
			if path, err := executionModePath(); err == nil {
				if data, readErr := os.ReadFile(path); readErr == nil {
					var payload savedExecutionMode
					if json.Unmarshal(data, &payload) == nil && strings.EqualFold(payload.Mode, executionModeAsk) {
						saved = executionModeAsk
					}
				}
			}
			line = strings.TrimSuffix(line, ".") + fmt.Sprintf(" (from JINI_MODE; saved setting is %s).", titleCase(saved))
		}
		fmt.Fprintln(stdout, line)
		if mode == executionModeAsk {
			fmt.Fprintln(stdout, "Switch with `jini mode auto`.")
		} else {
			fmt.Fprintln(stdout, "Switch with `jini mode ask`.")
		}
		return 0
	}
	requested := strings.ToLower(strings.TrimSpace(args[0]))
	if requested == "autopilot" || requested == "managed-throttle-recovery" {
		// Paid Autopilot ships from the Jini commercial repo; this public build
		// only fails closed and names the free equivalent.
		renderFeatureAccessDenied(stderr, featureAccessForID("commercial-autopilot"))
		return 1
	}
	if requested != executionModeAuto && requested != executionModeAsk {
		fmt.Fprintf(stderr, "Unknown mode %q. Use `jini mode auto` or `jini mode ask`.\n", args[0])
		return 1
	}
	if err := saveExecutionMode(requested); err != nil {
		fmt.Fprintln(stderr, err.Error())
		return 1
	}
	fmt.Fprintln(stdout, executionModeDisplayLine(requested))
	if strings.TrimSpace(os.Getenv("JINI_MODE")) != "" {
		fmt.Fprintln(stdout, "Note: JINI_MODE is set and still wins in this shell.")
	}
	return 0
}
