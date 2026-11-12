package app

import (
	"os"
	"path/filepath"
	"testing"
)

// r1aTestWorkspace returns a fresh, canonical (symlink-resolved) absolute
// workspace directory.
func r1aTestWorkspace(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	real, err := filepath.EvalSymlinks(dir)
	if err != nil {
		t.Fatalf("eval symlinks: %v", err)
	}
	return real
}

// r1aSetHome points the protected JINI_HOME at an isolated temp dir for the test.
func r1aSetHome(t *testing.T) {
	t.Helper()
	home := t.TempDir()
	old := executionModeHomeDir
	executionModeHomeDir = func() (string, error) { return home, nil }
	t.Cleanup(func() { executionModeHomeDir = old })
}

// r1aWriteFile writes a file at root/rel with an exact mode (umask-proof).
func r1aWriteFile(t *testing.T, root, rel string, data []byte, mode os.FileMode) string {
	t.Helper()
	p := filepath.Join(root, rel)
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(p, data, mode); err != nil {
		t.Fatalf("write: %v", err)
	}
	if err := os.Chmod(p, mode); err != nil {
		t.Fatalf("chmod: %v", err)
	}
	return p
}

// r1aReadDisk reads a workspace file directly from disk (test-only).
func r1aReadDisk(t *testing.T, root, rel string) []byte {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(root, rel))
	if err != nil {
		t.Fatalf("read disk %s: %v", rel, err)
	}
	return data
}
