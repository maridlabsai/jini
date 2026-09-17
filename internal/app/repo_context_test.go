package app

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestDefaultRepoContextReadsAgentsAndClaude(t *testing.T) {
	dir := t.TempDir()
	writeTestFile(t, filepath.Join(dir, "AGENTS.md"), "Build: go build ./...\nTest: go test ./...\n")
	writeTestFile(t, filepath.Join(dir, "CLAUDE.md"), "Convention: table-driven tests only.\n")
	t.Chdir(dir)

	got := defaultRepoContext()
	for _, want := range []string{
		"Repository context",
		"--- AGENTS.md ---",
		"Build: go build ./...",
		"--- CLAUDE.md ---",
		"Convention: table-driven tests only.",
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("repo context missing %q; got:\n%s", want, got)
		}
	}
}

func TestDefaultRepoContextEmptyWhenNoFiles(t *testing.T) {
	dir := t.TempDir()
	t.Chdir(dir)
	if got := defaultRepoContext(); got != "" {
		t.Fatalf("expected empty repo context in a bare dir, got:\n%s", got)
	}
}

func TestDefaultRepoContextTruncatesOversized(t *testing.T) {
	dir := t.TempDir()
	big := strings.Repeat("x", repoContextByteCap+5000)
	writeTestFile(t, filepath.Join(dir, "AGENTS.md"), big)
	t.Chdir(dir)

	got := defaultRepoContext()
	if len(got) > repoContextByteCap {
		t.Fatalf("repo context not capped: len=%d cap=%d", len(got), repoContextByteCap)
	}
	if !strings.HasSuffix(got, repoContextTruncatedMarker) {
		t.Fatalf("oversized context must end with the truncation marker; got tail: %q", got[len(got)-20:])
	}
}

// The reader is empty when a file exists but is only whitespace (no noise).
func TestDefaultRepoContextIgnoresBlankFile(t *testing.T) {
	dir := t.TempDir()
	writeTestFile(t, filepath.Join(dir, "AGENTS.md"), "\n\n   \n")
	t.Chdir(dir)
	if got := defaultRepoContext(); got != "" {
		t.Fatalf("blank file should yield empty context, got:\n%s", got)
	}
}

// providerSystemPrompt must include repo context when present, and omit it when
// absent — proving the injection wiring, not just the reader.
func TestProviderSystemPromptIncludesRepoContext(t *testing.T) {
	prev := repoContextReader
	t.Cleanup(func() { repoContextReader = prev })

	repoContextReader = func() string { return "REPO_CTX_SENTINEL" }
	if got := providerSystemPrompt(); !strings.Contains(got, "REPO_CTX_SENTINEL") {
		t.Fatalf("provider system prompt must include repo context; got:\n%s", got)
	}

	repoContextReader = func() string { return "" }
	if got := providerSystemPrompt(); strings.Contains(got, "REPO_CTX_SENTINEL") {
		t.Fatalf("provider system prompt must omit context when none; got:\n%s", got)
	}
}

func TestAgentSystemPromptIncludesRepoContext(t *testing.T) {
	prev := repoContextReader
	t.Cleanup(func() { repoContextReader = prev })
	repoContextReader = func() string { return "REPO_CTX_SENTINEL" }
	if got := agentSystemPrompt("do the thing", nil); !strings.Contains(got, "REPO_CTX_SENTINEL") {
		t.Fatalf("agent system prompt must include repo context; got:\n%s", got)
	}
}

// A subdirectory of a git repo must pick up the repo-root instruction file.
func TestDefaultRepoContextWalksToGitRoot(t *testing.T) {
	root := t.TempDir()
	runGitCommandForInternalTest(t, root, "init", "-q")
	writeTestFile(t, filepath.Join(root, "AGENTS.md"), "Root convention: use go test ./...\n")
	sub := filepath.Join(root, "internal", "deep")
	if err := os.MkdirAll(sub, 0o755); err != nil {
		t.Fatal(err)
	}
	t.Chdir(sub)

	got := defaultRepoContext()
	if !strings.Contains(got, "Root convention: use go test ./...") {
		t.Fatalf("subdir must inherit repo-root AGENTS.md; got:\n%s", got)
	}
}
