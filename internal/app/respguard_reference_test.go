package app

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestBrokenPathReferences_FlagsMissingClaimedFiles(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "notes.txt"), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}

	// A file Jini says it updated, that exists → clean.
	if issues := brokenPathReferences("Updated notes.txt", dir); len(issues) != 0 {
		t.Fatalf("existing claimed file must be clean, got %+v", issues)
	}
	// A file Jini says it edited, that is missing → flagged.
	if issues := brokenPathReferences("· edited src/gone.go", dir); !hasKind(issues, issueBrokenReference) {
		t.Fatalf("missing claimed file must be flagged, got %v", kinds(issues))
	}
	if issues := brokenPathReferences("wrote 12 bytes to build/out.bin", dir); !hasKind(issues, issueBrokenReference) {
		t.Fatalf("missing 'wrote N bytes to' target must be flagged, got %v", kinds(issues))
	}
}

func TestBrokenPathReferences_ConservativeNoFalsePositives(t *testing.T) {
	dir := t.TempDir()
	// Prose that merely mentions files, placeholders, flags, URLs, or non-claim
	// verbs must NOT be flagged — only explicit file-claims are checked.
	clean := []string{
		"To fix this, edit your config file and retry.", // no concrete token after a claim verb
		"Updated <your-file>",                           // placeholder
		"Updated --permission-mode",                     // flag, not a path
		"See https://example.com/report.md for details", // URL
		"This explains what a mutex is.",                // ordinary prose
		"Run `go test ./...` to verify.",                // command, no claim verb
	}
	for _, c := range clean {
		if issues := brokenPathReferences(c, dir); hasKind(issues, issueBrokenReference) {
			t.Fatalf("false positive on %q: %+v", c, issues)
		}
	}
}

func TestLocalTextEdit_ConfirmationCitesRealFile(t *testing.T) {
	t.Setenv("JINI_PROVIDER", "local-preview")
	t.Setenv("JINI_STATE_DIR", t.TempDir())
	dir := t.TempDir()
	t.Chdir(dir)
	if err := os.WriteFile(filepath.Join(dir, "notes.txt"), []byte("first line\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	var out, errBuf bytes.Buffer
	RunInteractive([]string{"add", "a", "line", "saying", "hello", "to", "notes.txt"}, strings.NewReader(""), &out, &errBuf)
	combined := out.String() + "\n" + errBuf.String()

	// Whatever file Jini claims to have touched must actually exist (integrity).
	if issues := brokenPathReferences(combined, dir); hasKind(issues, issueBrokenReference) {
		t.Fatalf("edit confirmation cited a non-existent file:\n%s\nissues=%+v", combined, issues)
	}
}
