package app

import (
	"strings"
	"testing"
)

func TestPostureDisclosureLine(t *testing.T) {
	if postureDisclosureLine(posturePlan, "Claude Code CLI handoff", "/repo") != "" {
		t.Fatal("plan must not disclose")
	}
	semi := postureDisclosureLine(postureSemi, "Claude Code CLI handoff", "/repo")
	if !strings.Contains(semi, "applying edits") || !strings.Contains(semi, "no commands run") || !strings.Contains(semi, "Claude Code CLI handoff") {
		t.Fatalf("semi disclosure wrong: %q", semi)
	}
	auto := postureDisclosureLine(postureAutonomous, "Codex CLI handoff", "/repo")
	if !strings.Contains(auto, "applying edits and running commands") || !strings.Contains(auto, "Codex CLI handoff") {
		t.Fatalf("autonomous disclosure wrong: %q", auto)
	}
	// Neutral tone: no caps fear framing.
	for _, line := range []string{semi, auto} {
		for _, banned := range []string{"WARNING", "DANGER", "IRREVERSIBLE"} {
			if strings.Contains(line, banned) {
				t.Fatalf("disclosure has fear framing %q: %q", banned, line)
			}
		}
	}
}

func TestPostureDegradedHint_OnlyWhenTrustedButUnhonored(t *testing.T) {
	withExecutionModeHome(t)
	if err := saveExecutionMode(executionModeAuto); err != nil {
		t.Fatal(err)
	}
	dir := t.TempDir()
	unverified := cliHandoffDescriptor{Mode: "unverified"} // no posture args

	// Untrusted → no hint (no per-task nagging).
	if h := postureDegradedHintForDir(unverified, "Unverified CLI", dir); h != "" {
		t.Fatalf("untrusted dir must not hint: %q", h)
	}

	// Trusted autonomous but route unverified → degraded to plan → hint.
	if err := saveTrustGrant(dir, trustLevelAutonomous, "edits+commands"); err != nil {
		t.Fatal(err)
	}
	h := postureDegradedHintForDir(unverified, "Unverified CLI", dir)
	if !strings.Contains(h, "doesn't support") || !strings.Contains(h, "autonomous") {
		t.Fatalf("degraded hint wrong: %q", h)
	}

	// Verified route honoring the grant → no hint (it escalated, not degraded).
	claude, _ := cliHandoffDescriptorForMode("claude-code")
	if h := postureDegradedHintForDir(claude, "Claude Code CLI handoff", dir); h != "" {
		t.Fatalf("honored grant must not show degraded hint: %q", h)
	}
}

func TestPostureDegradedHint_SilentInAskMode(t *testing.T) {
	withExecutionModeHome(t)
	if err := saveExecutionMode(executionModeAsk); err != nil {
		t.Fatal(err)
	}
	dir := t.TempDir()
	if err := saveTrustGrant(dir, trustLevelAutonomous, "edits+commands"); err != nil {
		t.Fatal(err)
	}
	if h := postureDegradedHintForDir(cliHandoffDescriptor{Mode: "unverified"}, "Unverified CLI", dir); h != "" {
		t.Fatalf("Ask mode must not hint (it's always plan by design): %q", h)
	}
}
