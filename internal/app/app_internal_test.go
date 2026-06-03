package app

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestSafelyRunInteractiveRecoversPanics(t *testing.T) {
	var stderr bytes.Buffer
	exitCode := safelyRunInteractive(&stderr, func() int {
		panic("boom")
	})
	if exitCode != 1 {
		t.Fatalf("expected exit code 1, got %d", exitCode)
	}
	if !strings.Contains(stderr.String(), "Jini hit an unexpected internal error and stopped safely.") {
		t.Fatalf("expected safe panic message, got:\n%s", stderr.String())
	}
}

func TestRunLegacyPythonDelegatesUnsupportedCommand(t *testing.T) {
	root := t.TempDir()
	toolsDir := filepath.Join(root, "tools")
	if err := os.MkdirAll(toolsDir, 0o755); err != nil {
		t.Fatalf("mkdir tools: %v", err)
	}
	scriptPath := filepath.Join(toolsDir, "jini_validate.py")
	script := "import json, sys\nprint(json.dumps({'argv': sys.argv[1:]}))\n"
	if err := os.WriteFile(scriptPath, []byte(script), 0o755); err != nil {
		t.Fatalf("write legacy script: %v", err)
	}

	t.Setenv("JINI_SOURCE_DIR", root)

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	exitCode := runLegacyPython([]string{"publish-readiness", "--format", "json"}, nil, &stdout, &stderr)
	if exitCode != 0 {
		t.Fatalf("expected exit code 0, got %d\nstdout:\n%s\nstderr:\n%s", exitCode, stdout.String(), stderr.String())
	}
	if !strings.Contains(stdout.String(), `"argv": ["publish-readiness", "--format", "json"]`) {
		t.Fatalf("expected legacy argv passthrough, got:\n%s", stdout.String())
	}
}

func TestRunLegacyPythonPreservesCallerWorkingDirectory(t *testing.T) {
	root := t.TempDir()
	toolsDir := filepath.Join(root, "tools")
	if err := os.MkdirAll(toolsDir, 0o755); err != nil {
		t.Fatalf("mkdir tools: %v", err)
	}
	scriptPath := filepath.Join(toolsDir, "jini_validate.py")
	script := "import os\nprint(os.getcwd())\n"
	if err := os.WriteFile(scriptPath, []byte(script), 0o755); err != nil {
		t.Fatalf("write legacy script: %v", err)
	}

	callerDir := t.TempDir()
	previousDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	if err := os.Chdir(callerDir); err != nil {
		t.Fatalf("chdir: %v", err)
	}
	defer func() {
		_ = os.Chdir(previousDir)
	}()

	t.Setenv("JINI_SOURCE_DIR", root)

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	exitCode := runLegacyPython(nil, nil, &stdout, &stderr)
	if exitCode != 0 {
		t.Fatalf("expected exit code 0, got %d\nstdout:\n%s\nstderr:\n%s", exitCode, stdout.String(), stderr.String())
	}
	if !strings.Contains(stdout.String(), callerDir) {
		t.Fatalf("expected legacy run to preserve caller cwd %q, got:\n%s", callerDir, stdout.String())
	}
}

func TestRecognizedGoCommandFallsBackToLegacyForUnsupportedFlags(t *testing.T) {
	root := t.TempDir()
	toolsDir := filepath.Join(root, "tools")
	if err := os.MkdirAll(toolsDir, 0o755); err != nil {
		t.Fatalf("mkdir tools: %v", err)
	}
	scriptPath := filepath.Join(toolsDir, "jini_validate.py")
	script := "import json, sys\nprint(json.dumps({'argv': sys.argv[1:]}))\n"
	if err := os.WriteFile(scriptPath, []byte(script), 0o755); err != nil {
		t.Fatalf("write legacy script: %v", err)
	}

	t.Setenv("JINI_SOURCE_DIR", root)

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	exitCode := Run([]string{"doctor", "--format", "json"}, &stdout, &stderr)
	if exitCode != 0 {
		t.Fatalf("expected exit code 0, got %d\nstdout:\n%s\nstderr:\n%s", exitCode, stdout.String(), stderr.String())
	}
	if !strings.Contains(stdout.String(), `"argv": ["doctor", "--format", "json"]`) {
		t.Fatalf("expected legacy doctor argv passthrough, got:\n%s", stdout.String())
	}
}

func TestLegacyPythonEnvPrependsConfiguredPythonPath(t *testing.T) {
	t.Setenv("PYTHONPATH", "/existing/site-packages")
	t.Setenv("JINI_LEGACY_PYTHONPATH", "/stable/vendor")

	env := legacyPythonEnv("/tmp/jini-source")
	joined := strings.Join(env, "\n")

	if !strings.Contains(joined, "JINI_SOURCE_DIR=/tmp/jini-source") {
		t.Fatalf("expected source dir in env, got:\n%s", joined)
	}
	if !strings.Contains(joined, "PYTHONPATH=/stable/vendor"+string(os.PathListSeparator)+"/existing/site-packages") {
		t.Fatalf("expected prepended PYTHONPATH, got:\n%s", joined)
	}
}

func TestShouldUseLegacyFrontDoorHonorsEnv(t *testing.T) {
	t.Setenv("JINI_USE_LEGACY_FRONT_DOOR", "1")
	if !shouldUseLegacyFrontDoor() {
		t.Fatalf("expected legacy front door to be enabled")
	}

	t.Setenv("JINI_USE_LEGACY_FRONT_DOOR", "0")
	if shouldUseLegacyFrontDoor() {
		t.Fatalf("expected legacy front door to be disabled")
	}
}
