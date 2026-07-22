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
	executionModeHomeDir            = os.UserHomeDir
)

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
