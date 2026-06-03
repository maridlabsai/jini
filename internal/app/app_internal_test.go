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
	exitCode := runLegacyPython([]string{"publish-readiness", "--format", "json"}, &stdout, &stderr)
	if exitCode != 0 {
		t.Fatalf("expected exit code 0, got %d\nstdout:\n%s\nstderr:\n%s", exitCode, stdout.String(), stderr.String())
	}
	if !strings.Contains(stdout.String(), `"argv": ["publish-readiness", "--format", "json"]`) {
		t.Fatalf("expected legacy argv passthrough, got:\n%s", stdout.String())
	}
}
