package app

import (
	"strings"
	"testing"
)

func kinds(issues []outputIssue) []outputIssueKind {
	var out []outputIssueKind
	for _, is := range issues {
		out = append(out, is.Kind)
	}
	return out
}

func hasKind(issues []outputIssue, want outputIssueKind) bool {
	for _, is := range issues {
		if is.Kind == want {
			return true
		}
	}
	return false
}

func TestAudit_CleanCopyHasNoIssues(t *testing.T) {
	clean := "Route set to claude-code.\nSaved ≈ US$0.01 this task · imputed · jini savings\nTry `jini doctor`."
	if issues := auditUserFacingOutput(clean, auditOptions{}); len(issues) != 0 {
		t.Fatalf("clean copy should pass, got %+v", issues)
	}
}

func TestAudit_ANSIEscapeFlagged(t *testing.T) {
	withColor := "\x1b[31mError:\x1b[0m something failed"
	issues := auditUserFacingOutput(withColor, auditOptions{})
	if !hasKind(issues, issueANSIEscape) {
		t.Fatalf("expected ansi-escape issue, got %v", kinds(issues))
	}
	if !hasSecurityOrAccessibilityIssue(issues) {
		t.Fatal("ansi escape must count as a hard accessibility invariant")
	}
}

func TestAudit_ControlCharFlaggedButTabAllowed(t *testing.T) {
	if issues := auditUserFacingOutput("col1\tcol2\tcol3", auditOptions{}); hasKind(issues, issueControlChar) {
		t.Fatalf("tabs are allowed, got %v", kinds(issues))
	}
	if issues := auditUserFacingOutput("bell\aend", auditOptions{}); !hasKind(issues, issueControlChar) {
		t.Fatalf("bell control char should be flagged, got %v", kinds(issues))
	}
}

func TestAudit_LeakedSecretsFlagged(t *testing.T) {
	cases := []string{
		"key is sk-" + strings.Repeat("A1b2", 6),
		"token ghp_" + strings.Repeat("x", 36),
		"aws AKIA" + strings.Repeat("Z", 12) + "1234",
		"bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIn0.dozjgNryP4gYaBcDeF",
	}
	for _, c := range cases {
		issues := auditUserFacingOutput(c, auditOptions{})
		if !hasKind(issues, issueLeakedSecret) {
			t.Fatalf("expected leaked-secret for %q, got %v", c, kinds(issues))
		}
		// The detector's own message must not reprint the full secret.
		for _, is := range issues {
			if is.Kind == issueLeakedSecret && strings.Contains(is.Detail, strings.Repeat("A1b2", 6)) {
				t.Fatal("detector message leaked the full secret")
			}
		}
	}
}

func TestAudit_PlaceholdersAreNotSecrets(t *testing.T) {
	// The setup wizard and docs use short placeholders — these must NOT trip.
	for _, ok := range []string{"sk-test-key", "sk-…", "sk-<your-key>", "ANTHROPIC_API_KEY=***", "enter your token"} {
		if issues := auditUserFacingOutput(ok, auditOptions{}); hasKind(issues, issueLeakedSecret) {
			t.Fatalf("placeholder %q must not be flagged as a secret", ok)
		}
	}
}

func TestAudit_ToneLexiconButFlagNamesExempt(t *testing.T) {
	if issues := auditUserFacingOutput("This revolutionary tool will unleash productivity!", auditOptions{}); !hasKind(issues, issueFearOrHype) {
		t.Fatalf("hype copy should be flagged, got %v", kinds(issues))
	}
	// The real autonomous-posture flag must never trip the tone guard.
	flagCopy := "Autonomous applies edits and runs commands via `--dangerously-skip-permissions`."
	if issues := auditUserFacingOutput(flagCopy, auditOptions{}); hasKind(issues, issueFearOrHype) {
		t.Fatalf("flag name in backticks must be tone-exempt, got %v", kinds(issues))
	}
	if opts := (auditOptions{SkipTone: true}); len(auditUserFacingOutput("blazing fast!", opts)) != 0 {
		t.Fatal("SkipTone must suppress the tone check")
	}
}

func TestAudit_ReadabilityOverlongVsUnwrappable(t *testing.T) {
	wall := strings.Repeat("word ", 40) // ~200 chars of prose
	if issues := auditUserFacingOutput(wall, auditOptions{}); !hasKind(issues, issueOverlongLine) {
		t.Fatalf("a 200-char prose line should be flagged, got %v", kinds(issues))
	}
	longURL := "See https://example.com/" + strings.Repeat("a", 150)
	if issues := auditUserFacingOutput(longURL, auditOptions{}); hasKind(issues, issueOverlongLine) {
		t.Fatalf("a long URL is unwrappable and must be exempt, got %v", kinds(issues))
	}
	if issues := auditUserFacingOutput(wall, auditOptions{SkipReadab: true}); hasKind(issues, issueOverlongLine) {
		t.Fatal("SkipReadab must suppress readability checks")
	}
}
