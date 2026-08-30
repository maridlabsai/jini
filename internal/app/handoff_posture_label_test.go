package app

import "testing"

func TestHandoffPostureLabel(t *testing.T) {
	if got := handoffPostureLabel("claude-code"); got != "posture verified" {
		t.Fatalf("claude-code should be verified, got %q", got)
	}
	for _, m := range []string{"codex", "gemini-cli", "aider", "opencode"} {
		if got := handoffPostureLabel(m); got != "posture experimental (doc-verified)" {
			t.Fatalf("%s should be experimental, got %q", m, got)
		}
	}
	// Non-hand-off route → no label.
	if got := handoffPostureLabel("local-slm"); got != "" {
		t.Fatalf("non-handoff should have no posture label, got %q", got)
	}
}
