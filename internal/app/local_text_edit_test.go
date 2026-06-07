package app

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParseLocalTextEditIntentTitleCasedUnquotedIntent(t *testing.T) {
	workDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(workDir, "pear vc script.txt"), []byte("intro\n"), 0o644); err != nil {
		t.Fatalf("write target: %v", err)
	}
	t.Chdir(workDir)

	intent, ok := parseLocalTextEditIntent("Add A Line Saying Jini Was Here In The Pear Vc Script Txt File In This Folder")
	if !ok {
		t.Fatalf("expected title-cased unquoted edit intent; split line=%q", firstUnquotedSayingText("Add A Line Saying Jini Was Here In The Pear Vc Script Txt File In This Folder"))
	}
	if intent.Line != "jini was here" {
		t.Fatalf("expected normalized jini line, got %q", intent.Line)
	}
}

func TestParseLocalTextEditIntentPreservesUnquotedLineContainingIn(t *testing.T) {
	workDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(workDir, "notes.txt"), []byte("intro\n"), 0o644); err != nil {
		t.Fatalf("write target: %v", err)
	}
	t.Chdir(workDir)

	intent, ok := parseLocalTextEditIntent("add a line saying check in at 9 in notes.txt")
	if !ok {
		t.Fatal("expected unquoted edit intent")
	}
	if intent.Line != "check in at 9" {
		t.Fatalf("expected line containing in to be preserved, got %q", intent.Line)
	}
}
