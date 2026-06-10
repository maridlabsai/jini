package app_test

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestPublicDocsUseCurrentFirstRunFlow(t *testing.T) {
	root := repoRootForMigrationTest(t)
	files := []string{"README.md"}
	docFiles, err := filepath.Glob(filepath.Join(root, "docs", "*.md"))
	if err != nil {
		t.Fatalf("scan docs markdown: %v", err)
	}
	if len(docFiles) == 0 {
		t.Fatal("expected public docs markdown files")
	}
	for _, path := range docFiles {
		rel, err := filepath.Rel(root, path)
		if err != nil {
			t.Fatalf("make docs path relative: %v", err)
		}
		files = append(files, filepath.ToSlash(rel))
	}

	forbidden := []string{
		"Paste notes or type what you want finished",
		"paste the work you want finished",
		"paste the work you already have",
		"pasting the work they want finished",
		"work they want finished",
		"Your first draft is ready",
		"Working Draft",
		"first draft",
		"Before Jini starts",
		"Use Auto",
		"type <code>Auto</code>",
		"<span>Auto</span>",
		"<h3>Auto if needed</h3>",
		"<h3>Auto</h3>",
		"<h3>Auto route</h3>",
		"Full Auto",
		"Auto mode prefers",
		"Open Missing Pieces",
		"Open Build-readiness",
		"Use <code>Open</code>",
		"use <code>Plan</code>",
		"<span>Open</span>",
		"<span>Missing</span>",
		"<span>Plan</span>",
		"<span>Switch</span>",
		"What “Open” Should Feel Like",
		"Describe the task.\nType `help` for examples and commands.",
	}

	for _, rel := range files {
		content := readRepoFile(t, root, rel)
		for _, pattern := range forbidden {
			if strings.Contains(content, pattern) {
				t.Fatalf("%s contains retired public flow copy %q", rel, pattern)
			}
		}
	}

	readme := readRepoFile(t, root, "README.md")
	for _, want := range []string{
		"The first screen should stay light: no saved-work dashboard, no tutorial block,",
		"describe the task or paste the notes, files, screenshot, transcript, or rough ask",
		"`auto` means: Jini picks the cheapest suitable route by default",
		"Durable work details belong behind explicit commands such as",
		"`jini status`, `jini continue`, `jini open`, and `jini route`",
	} {
		if !strings.Contains(readme, want) {
			t.Fatalf("README must teach current public flow %q", want)
		}
	}

	quickstart := readRepoFile(t, root, "docs/simple.md")
	for _, want := range []string{
		"Describe the task, then use <code>jini open</code>, <code>jini continue</code>, or <code>jini status</code>",
		"<pre><code class=\"language-text\">Jini\n&gt; what is the capital of france\nParis.",
		"<span>jini open</span>",
		"<span>jini continue</span>",
		"<span>jini status</span>",
	} {
		if !strings.Contains(quickstart, want) {
			t.Fatalf("Quickstart must teach current public flow %q", want)
		}
	}
}
