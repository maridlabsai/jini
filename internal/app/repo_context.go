package app

import (
	"os"
	"path/filepath"
	"strings"
)

// Repo-scoped project context — the convergent industry pattern (Codex
// AGENTS.md, Kiro steering, Claude Code CLAUDE.md). The PRD wants repo-scoped
// project context (build/test commands, conventions, pitfalls), NOT user-profile
// building. Jini reads these files and injects them into NON-handoff routes
// (local model, BYO API, native loop) so the model follows the repo's rules;
// CLI-handoff routes are left alone because the downstream CLI reads the same
// files itself.

const (
	repoContextByteCap         = 32 * 1024 // matches Codex's AGENTS.md default cap
	repoContextTruncatedMarker = "\n…[truncated]"
)

// repoContextFiles are read in order; all present ones are included.
var repoContextFiles = []string{"AGENTS.md", "CLAUDE.md"}

// repoContextReader is the seam so tests (which run inside this repo, which HAS
// a CLAUDE.md) don't get repo context injected into every provider prompt.
var repoContextReader = defaultRepoContext

func repoContextForCwd() string { return repoContextReader() }

// defaultRepoContext reads the repo instruction files from the current
// directory and, when inside a git work tree, the repo root (if different),
// returning an authoritative-sounding context block or "" when none exist.
func defaultRepoContext() string {
	cwd, err := os.Getwd()
	if err != nil {
		return ""
	}
	dirs := []string{cwd}
	if root, ok := gitRepoRoot(cwd); ok && root != cwd {
		// Root first so cwd-level overrides read last (closest wins on dupes).
		dirs = []string{root, cwd}
	}

	seen := map[string]bool{}
	var sections []string
	for _, dir := range dirs {
		for _, name := range repoContextFiles {
			path := filepath.Join(dir, name)
			if seen[path] {
				continue
			}
			seen[path] = true
			data, err := os.ReadFile(path)
			if err != nil {
				continue
			}
			content := strings.TrimRight(string(data), "\n")
			if strings.TrimSpace(content) == "" {
				continue
			}
			sections = append(sections, "--- "+name+" ---\n"+content)
		}
	}
	if len(sections) == 0 {
		return ""
	}

	block := "Repository context (the project provides these instructions - follow them):\n\n" + strings.Join(sections, "\n\n")
	if len(block) > repoContextByteCap {
		block = block[:repoContextByteCap-len(repoContextTruncatedMarker)] + repoContextTruncatedMarker
	}
	return block
}

// gitRepoRoot returns the git work-tree root containing dir, or ok=false.
func gitRepoRoot(dir string) (string, bool) {
	out, ok := runGitOutput("-C", dir, "rev-parse", "--show-toplevel")
	if !ok {
		return "", false
	}
	root := strings.TrimSpace(out)
	if root == "" {
		return "", false
	}
	return root, true
}
