package app

import "testing"

func TestHiddenAppSidecarServeCommandRequiresExactShape(t *testing.T) {
	if !isHiddenAppSidecarServeCommand([]string{"app", "serve", "--stdio", "--surface", "macos"}) {
		t.Fatalf("expected exact macOS sidecar command to match")
	}

	for _, args := range [][]string{
		{"app"},
		{"app", "serve"},
		{"app", "serve", "--stdio"},
		{"app", "serve", "--surface", "macos", "--stdio"},
		{"app", "serve", "--stdio", "--surface", "ios"},
		{"app", "serve", "--stdio", "--surface", "macos", "--verbose"},
		{"app", "status", "--stdio", "--surface", "macos"},
	} {
		if isHiddenAppSidecarServeCommand(args) {
			t.Fatalf("expected non-exact sidecar command not to match: %#v", args)
		}
	}
}

func TestAppSidecarCommandStaysOutOfPublicCommandCanonicalization(t *testing.T) {
	if got := canonicalTopLevelCommand("app"); got != "" {
		t.Fatalf("expected app not to be a public top-level command, got %q", got)
	}
	if got := canonicalHelpTopic("app"); got != "" {
		t.Fatalf("expected app not to be a public help topic, got %q", got)
	}
}
