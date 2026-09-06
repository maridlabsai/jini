package app

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// initSideEffectGitRepo creates a hermetic git repo with one committed file and
// makes it the process working directory for the test.
func initSideEffectGitRepo(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	runGitCommandForInternalTest(t, dir, "init", "-q")
	writeTestFile(t, filepath.Join(dir, "tracked.txt"), "original\n")
	runGitCommandForInternalTest(t, dir, "add", "tracked.txt")
	runGitCommandForInternalTest(t, dir,
		"-c", "user.name=Jini Test",
		"-c", "user.email=jini@example.test",
		"-c", "commit.gpgsign=false",
		"commit", "-q", "-m", "seed",
	)
	t.Chdir(dir)
	return dir
}

func TestCLIHandoffReceiptRecordsSideEffectsAndRollbackPath(t *testing.T) {
	t.Setenv("JINI_STATE_DIR", t.TempDir())
	dir := initSideEffectGitRepo(t)
	writeProviderFakeExecutable(t, t.TempDir(), "claude", strings.Join([]string{
		`printf 'edited\n' >> tracked.txt`,
		`printf 'brand new\n' > created.txt`,
		`echo done`,
	}, "\n"))

	_, receipt, err := runCLIHandoff(context.Background(), "claude-code", "change two files")
	if err != nil {
		t.Fatalf("runCLIHandoff: %v", err)
	}
	if receipt == nil {
		t.Fatal("expected a receipt")
	}
	if got, want := receipt.SideEffectCount, 2; got != want {
		t.Fatalf("side effect count = %d, want %d (receipt %#v)", got, want, receipt)
	}
	for _, want := range []string{"created.txt", "tracked.txt"} {
		if !containsString(receipt.SideEffects, want) {
			t.Fatalf("side effects %v missing %q", receipt.SideEffects, want)
		}
	}
	if !strings.Contains(receipt.RollbackHint, "git restore") {
		t.Fatalf("rollback hint %q does not mention git restore", receipt.RollbackHint)
	}
	if !strings.Contains(receipt.RollbackHint, "tracked.txt") {
		t.Fatalf("rollback hint %q does not name the tracked file", receipt.RollbackHint)
	}
	if !strings.Contains(receipt.RollbackHint, "created.txt") ||
		!strings.Contains(receipt.RollbackHint, "removed manually") {
		t.Fatalf("rollback hint %q does not flag the new file as manual cleanup", receipt.RollbackHint)
	}
	// The rollback path stays advisory: nothing was undone for the user.
	if data, readErr := os.ReadFile(filepath.Join(dir, "created.txt")); readErr != nil || !strings.Contains(string(data), "brand new") {
		t.Fatalf("expected the new file to remain on disk, got %q err %v", string(data), readErr)
	}

	summary := formatCLIHandoffReceiptSummary(receipt)
	if !containsLineWithPrefix(summary, "Side effects: 2 (") {
		t.Fatalf("summary missing side effects line: %#v", summary)
	}
	if !containsLineWithPrefix(summary, "Rollback: ") {
		t.Fatalf("summary missing rollback line: %#v", summary)
	}

	var rendered bytes.Buffer
	renderCLIHandoffReceiptSummary(&rendered, summary)
	block := rendered.String()
	if !strings.Contains(block, "Last CLI handoff") ||
		!strings.Contains(block, "Side effects: 2 (created.txt, tracked.txt)") ||
		!strings.Contains(block, "Rollback: Review: git diff; revert tracked files with: git restore tracked.txt") {
		t.Fatalf("rendered handoff block missing side effects or rollback:\n%s", block)
	}
	t.Logf("rendered receipt block:\n%s", block)
}

func TestCLIHandoffReceiptReadOnlyRunHasNoSideEffects(t *testing.T) {
	t.Setenv("JINI_STATE_DIR", t.TempDir())
	initSideEffectGitRepo(t)
	writeProviderFakeExecutable(t, t.TempDir(), "claude", `cat tracked.txt >/dev/null; echo "read only"`)

	_, receipt, err := runCLIHandoff(context.Background(), "claude-code", "just read the file")
	if err != nil {
		t.Fatalf("runCLIHandoff: %v", err)
	}
	if len(receipt.SideEffects) != 0 || receipt.SideEffectCount != 0 || receipt.RollbackHint != "" {
		t.Fatalf("read-only run reported side effects: %#v", receipt)
	}
	for _, line := range formatCLIHandoffReceiptSummary(receipt) {
		if strings.HasPrefix(line, "Side effects:") || strings.HasPrefix(line, "Rollback:") {
			t.Fatalf("read-only summary should stay silent, got %q", line)
		}
	}
}

func TestCLIHandoffReceiptOutsideGitRepoReportsNothing(t *testing.T) {
	t.Setenv("JINI_STATE_DIR", t.TempDir())
	dir := t.TempDir()
	t.Chdir(dir)
	if _, ok := cliHandoffWorkTreeStatus(); ok {
		t.Skip("temp dir is inside a git work tree; cannot test the untracked case")
	}
	writeProviderFakeExecutable(t, t.TempDir(), "claude", `printf 'x\n' > created.txt; echo done`)

	_, receipt, err := runCLIHandoff(context.Background(), "claude-code", "write a file")
	if err != nil {
		t.Fatalf("runCLIHandoff: %v", err)
	}
	if _, statErr := os.Stat(filepath.Join(dir, "created.txt")); statErr != nil {
		t.Fatalf("fake CLI did not write the file: %v", statErr)
	}
	// Honest degrade: outside a work tree Jini cannot attribute changes, so it
	// claims none rather than guessing.
	if len(receipt.SideEffects) != 0 || receipt.SideEffectCount != 0 || receipt.RollbackHint != "" {
		t.Fatalf("expected empty side effects outside a git repo, got %#v", receipt)
	}
}

func TestCLIHandoffSideEffectsAreCappedWithHonestCount(t *testing.T) {
	before := map[string]string{}
	after := map[string]string{}
	for i := 0; i < 20; i++ {
		after[string(rune('a'+i))+".txt"] = "??"
	}
	receipt := &cliHandoffReceipt{}
	annotateCLIHandoffSideEffectsFor(receipt, before, after)
	if receipt.SideEffectCount != 20 {
		t.Fatalf("expected honest count 20, got %d", receipt.SideEffectCount)
	}
	if len(receipt.SideEffects) != cliHandoffSideEffectLimit {
		t.Fatalf("expected %d listed paths, got %d", cliHandoffSideEffectLimit, len(receipt.SideEffects))
	}
	if strings.Contains(receipt.RollbackHint, "git restore") {
		t.Fatalf("untracked-only rollback hint should not suggest git restore: %q", receipt.RollbackHint)
	}
	if !strings.Contains(receipt.RollbackHint, "removed manually") {
		t.Fatalf("rollback hint %q should flag manual cleanup", receipt.RollbackHint)
	}
	line := formatCLIHandoffSideEffectLine(receipt)
	if !strings.HasPrefix(line, "Side effects: 20 (") || !strings.Contains(line, "+14 more") {
		t.Fatalf("summary line lost the honest count: %q", line)
	}
}

func TestChangedWorkTreePathsIgnoresPreexistingDirtyFiles(t *testing.T) {
	before := map[string]string{"already.go": " M", "flipped.go": " M"}
	after := map[string]string{"already.go": " M", "flipped.go": "M ", "new.go": "??"}
	got := changedWorkTreePaths(before, after)
	want := []string{"flipped.go", "new.go"}
	if strings.Join(got, ",") != strings.Join(want, ",") {
		t.Fatalf("changed paths = %v, want %v", got, want)
	}
}

func TestCLIHandoffWorkTreeStatusHandlesSpacedAndRenamedPaths(t *testing.T) {
	dir := initSideEffectGitRepo(t)
	writeTestFile(t, filepath.Join(dir, "spaced name.txt"), "hello\n")
	runGitCommandForInternalTest(t, dir, "mv", "tracked.txt", "renamed.txt")

	entries, ok := cliHandoffWorkTreeStatus()
	if !ok {
		t.Fatal("expected a git work tree")
	}
	if state, found := entries["spaced name.txt"]; !found || state != "??" {
		t.Fatalf("spaced path missing or wrong state: %#v", entries)
	}
	if state, found := entries["renamed.txt"]; !found || state[0] != 'R' {
		t.Fatalf("rename destination missing or wrong state: %#v", entries)
	}
	if _, found := entries["tracked.txt"]; found {
		t.Fatalf("rename source should not be parsed as its own entry: %#v", entries)
	}
}

func containsLineWithPrefix(lines []string, prefix string) bool {
	for _, line := range lines {
		if strings.HasPrefix(line, prefix) {
			return true
		}
	}
	return false
}

// The direct-answer path (the common `jini "<task>"` flow) must SURFACE the
// side effects and rollback hint, not just record them on a receipt that only
// the saved-work thread summary renders. Guards against the "inert receipt"
// regression.
func TestDirectHandoffSurfacesSideEffectsOnMainFlow(t *testing.T) {
	t.Setenv("JINI_STATE_DIR", t.TempDir())
	t.Setenv("JINI_TOOL", "claude-code")
	t.Setenv("JINI_CLI_HANDOFF_SKIP_TRUST_CHECK", "1")
	initSideEffectGitRepo(t)
	writeProviderFakeExecutable(t, t.TempDir(), "claude", strings.Join([]string{
		`printf 'edited\n' >> tracked.txt`,
		`printf 'brand new\n' > created.txt`,
		`echo done`,
	}, "\n"))

	var stdout, stderr bytes.Buffer
	code := runDirectTaskArgsIntake([]string{"append a line to tracked.txt and create created.txt"}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("direct handoff exit %d; stderr:\n%s", code, stderr.String())
	}
	out := stdout.String()
	for _, want := range []string{"Side effects:", "created.txt", "tracked.txt", "Rollback:", "git restore"} {
		if !strings.Contains(out, want) {
			t.Fatalf("main-flow output missing %q; got:\n%s", want, out)
		}
	}
}
