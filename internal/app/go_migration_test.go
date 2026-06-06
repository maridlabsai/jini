package app_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func gitTrackedFiles(t *testing.T) []string {
	t.Helper()
	output, err := exec.Command("git", "ls-files").Output()
	if err != nil {
		t.Fatalf("git ls-files: %v", err)
	}
	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	if len(lines) == 1 && lines[0] == "" {
		return nil
	}
	return lines
}

func TestRepositoryHasNoTrackedPythonFiles(t *testing.T) {
	for _, path := range gitTrackedFiles(t) {
		if strings.HasSuffix(path, ".py") {
			t.Fatalf("tracked Python file remains after Go migration: %s", path)
		}
	}
}

func TestRequiredGatesDoNotInvokePython(t *testing.T) {
	root := repoRootForMigrationTest(t)
	scanned := []string{
		"Makefile",
		"install.sh",
		"tools/run_required_gates.sh",
	}
	for _, rel := range scanned {
		data, err := os.ReadFile(filepath.Join(root, rel))
		if err != nil {
			t.Fatalf("read %s: %v", rel, err)
		}
		text := string(data)
		for _, forbidden := range []string{"python3", "jini_validate", "tools/jini.py"} {
			if strings.Contains(text, forbidden) {
				t.Fatalf("Python invocation remains in %s: %s", rel, forbidden)
			}
		}
	}
}

func repoRootForMigrationTest(t *testing.T) string {
	t.Helper()
	output, err := exec.Command("git", "rev-parse", "--show-toplevel").Output()
	if err != nil {
		t.Fatalf("git rev-parse: %v", err)
	}
	return strings.TrimSpace(string(output))
}
