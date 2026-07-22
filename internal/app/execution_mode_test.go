package app

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func withExecutionModeHome(t *testing.T) string {
	t.Helper()
	home := t.TempDir()
	oldHome := executionModeHomeDir
	executionModeHomeDir = func() (string, error) { return home, nil }
	t.Cleanup(func() { executionModeHomeDir = oldHome })
	t.Setenv("JINI_MODE", "")
	return home
}

func TestEffectiveExecutionModeDefaultsToAuto(t *testing.T) {
	withExecutionModeHome(t)
	if mode := effectiveExecutionMode(); mode != "auto" {
		t.Fatalf("expected auto, got %q", mode)
	}
}

func TestEffectiveExecutionModeRoundTrip(t *testing.T) {
	withExecutionModeHome(t)
	if err := saveExecutionMode("ask"); err != nil {
		t.Fatal(err)
	}
	if mode := effectiveExecutionMode(); mode != "ask" {
		t.Fatalf("expected ask, got %q", mode)
	}
	if err := saveExecutionMode("auto"); err != nil {
		t.Fatal(err)
	}
	if mode := effectiveExecutionMode(); mode != "auto" {
		t.Fatalf("expected auto, got %q", mode)
	}
}

func TestEffectiveExecutionModeEnvOverridesFile(t *testing.T) {
	withExecutionModeHome(t)
	if err := saveExecutionMode("auto"); err != nil {
		t.Fatal(err)
	}
	t.Setenv("JINI_MODE", "ask")
	if mode := effectiveExecutionMode(); mode != "ask" {
		t.Fatalf("expected ask from env, got %q", mode)
	}
}

func TestEffectiveExecutionModeCorruptFileFailsClosed(t *testing.T) {
	home := withExecutionModeHome(t)
	dir := filepath.Join(home, ".jini")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "mode.json"), []byte("{corrupt"), 0o600); err != nil {
		t.Fatal(err)
	}
	var warnings bytes.Buffer
	oldWarn := executionModeWarnings
	executionModeWarnings = &warnings
	defer func() { executionModeWarnings = oldWarn }()
	if mode := effectiveExecutionMode(); mode != "ask" {
		t.Fatalf("corrupt file must fail closed to ask, got %q", mode)
	}
	if !strings.Contains(warnings.String(), "mode setting unreadable") {
		t.Fatalf("expected warning, got %q", warnings.String())
	}
}

func TestEffectiveExecutionModeGarbageEnvFailsClosed(t *testing.T) {
	withExecutionModeHome(t)
	t.Setenv("JINI_MODE", "garbage")
	var warnings bytes.Buffer
	oldWarn := executionModeWarnings
	executionModeWarnings = &warnings
	defer func() { executionModeWarnings = oldWarn }()
	if mode := effectiveExecutionMode(); mode != "ask" {
		t.Fatalf("garbage env must fail closed to ask, got %q", mode)
	}
}

func TestEffectiveExecutionModeHomeDirFailureFailsClosed(t *testing.T) {
	oldHome := executionModeHomeDir
	executionModeHomeDir = func() (string, error) { return "", os.ErrPermission }
	defer func() { executionModeHomeDir = oldHome }()
	t.Setenv("JINI_MODE", "")
	var warnings bytes.Buffer
	oldWarn := executionModeWarnings
	executionModeWarnings = &warnings
	defer func() { executionModeWarnings = oldWarn }()
	if mode := effectiveExecutionMode(); mode != "ask" {
		t.Fatalf("homedir failure must fail closed to ask, got %q", mode)
	}
}

func TestSaveExecutionModeAtomicNoTempDroppings(t *testing.T) {
	home := withExecutionModeHome(t)
	if err := saveExecutionMode("ask"); err != nil {
		t.Fatal(err)
	}
	entries, err := os.ReadDir(filepath.Join(home, ".jini"))
	if err != nil {
		t.Fatal(err)
	}
	for _, entry := range entries {
		if entry.Name() != "mode.json" {
			t.Fatalf("unexpected file left behind: %s", entry.Name())
		}
	}
	info, err := os.Stat(filepath.Join(home, ".jini", "mode.json"))
	if err != nil {
		t.Fatal(err)
	}
	if perm := info.Mode().Perm(); perm != 0o600 {
		t.Fatalf("expected 0600, got %o", perm)
	}
}

func TestSaveExecutionModeRejectsInvalid(t *testing.T) {
	withExecutionModeHome(t)
	if err := saveExecutionMode("banana"); err == nil {
		t.Fatal("expected error for invalid mode")
	}
}
