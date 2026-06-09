package app

import (
	"strings"
	"testing"
)

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

func TestDiagnosticsTextRedactionRemovesStateHomeAndSecretValues(t *testing.T) {
	t.Setenv("JINI_STATE_DIR", "/tmp/jini-state")
	t.Setenv("HOME", "/Users/example")
	t.Setenv("OPENAI_API_KEY", "sk-test-secret")

	redacted := redactDiagnosticsText("state=/tmp/jini-state home=/Users/example key=sk-test-secret")
	for _, forbidden := range []string{"/tmp/jini-state", "/Users/example", "sk-test-secret"} {
		if strings.Contains(redacted, forbidden) {
			t.Fatalf("expected diagnostics text to redact %q, got %q", forbidden, redacted)
		}
	}
	for _, want := range []string{"$JINI_STATE_DIR", "~", "[redacted]"} {
		if !strings.Contains(redacted, want) {
			t.Fatalf("expected diagnostics redaction marker %q, got %q", want, redacted)
		}
	}
}
