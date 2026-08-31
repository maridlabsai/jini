package app

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// writeProbeFake writes an executable shell fake that acts on the posture flags
// it receives (standing in for a real CLI honoring each posture).
func writeProbeFake(t *testing.T, dir, name, body string) string {
	t.Helper()
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte("#!/bin/sh\n"+body), 0o755); err != nil {
		t.Fatal(err)
	}
	return path
}

// A well-behaved CLI: plan writes nothing, semi applies the edit, autonomous
// applies the edit and runs the command. The harness must verify it.
func TestValidateRoutePosture_WellBehaved(t *testing.T) {
	dir := t.TempDir()
	fake := writeProbeFake(t, dir, "claude", `for a in "$@"; do
  case "$a" in
    --dangerously-skip-permissions) printf OK > PROOF.txt; printf done > CMD.txt ;;
    acceptEdits) printf OK > PROOF.txt ;;
  esac
done
`)
	t.Setenv("JINI_CLAUDE_CODE_CLI", fake)

	d, _ := cliHandoffDescriptorForMode("claude-code")
	probes, verified, issues := validateRoutePosture(context.Background(), d)
	if !verified {
		t.Fatalf("well-behaved CLI must verify: issues=%v probes=%+v", issues, probes)
	}
	if probes[0].Edited || probes[0].CommandRan {
		t.Fatalf("plan must be read-only: %+v", probes[0])
	}
	if !probes[1].Edited || probes[1].CommandRan {
		t.Fatalf("semi must edit with no command: %+v", probes[1])
	}
	if !probes[2].Edited || !probes[2].CommandRan {
		t.Fatalf("autonomous must edit and run a command: %+v", probes[2])
	}
}

// SAFETY: a CLI that writes even in plan posture must fail validation.
func TestValidateRoutePosture_CatchesPlanWrite(t *testing.T) {
	dir := t.TempDir()
	fake := writeProbeFake(t, dir, "claude", "printf OK > PROOF.txt\n") // always writes
	t.Setenv("JINI_CLAUDE_CODE_CLI", fake)

	d, _ := cliHandoffDescriptorForMode("claude-code")
	_, verified, issues := validateRoutePosture(context.Background(), d)
	if verified {
		t.Fatal("a CLI that writes in plan must NOT verify")
	}
	found := false
	for _, s := range issues {
		if strings.Contains(s, "read-only") {
			found = true
		}
	}
	if !found {
		t.Fatalf("expected a plan read-only issue, got %v", issues)
	}
}

// A CLI that ignores its escalation flag (never edits) must fail — the claimed
// semi/autonomous posture isn't actually honored.
func TestValidateRoutePosture_CatchesNoEscalation(t *testing.T) {
	dir := t.TempDir()
	fake := writeProbeFake(t, dir, "claude", ":\n") // no-op: never edits
	t.Setenv("JINI_CLAUDE_CODE_CLI", fake)

	d, _ := cliHandoffDescriptorForMode("claude-code")
	_, verified, issues := validateRoutePosture(context.Background(), d)
	if verified {
		t.Fatal("a CLI that never applies edits must NOT verify semi/autonomous")
	}
	if len(issues) == 0 {
		t.Fatal("expected escalation issues")
	}
}
