package app

import (
	"strings"
	"testing"
)

// Deep simulation of the posture→args mapping for every hand-off CLI under every
// posture, asserting the exact "by the book" argument vector each CLI expects
// (specs/cli-handoff-compatibility-premortem.md) plus the cross-cutting safety
// invariant: plan posture must never carry a write-enabling flag, and a
// write-by-default CLI (aider) must be forced read-only in plan.

func TestCLICompat_PostureArgMatrix(t *testing.T) {
	want := map[string]map[handoffPosture][]string{
		"claude-code": {
			posturePlan:       {"--print", "{{prompt}}"},
			postureSemi:       {"--print", "--permission-mode", "acceptEdits", "{{prompt}}"},
			postureAutonomous: {"--print", "--dangerously-skip-permissions", "{{prompt}}"},
		},
		"codex": {
			posturePlan:       {"exec", "{{prompt}}"},
			postureSemi:       {"exec", "{{prompt}}"}, // no clean edits-only mode → degrades to plan args
			postureAutonomous: {"exec", "--dangerously-bypass-approvals-and-sandbox", "{{prompt}}"},
		},
		"gemini-cli": {
			posturePlan:       {"-p", "{{prompt}}"},
			postureSemi:       {"--approval-mode", "auto_edit", "-p", "{{prompt}}"},
			postureAutonomous: {"--yolo", "-p", "{{prompt}}"},
		},
		"aider": {
			posturePlan:       {"--dry-run", "--message", "{{prompt}}"},
			postureSemi:       {"--yes-always", "--message", "{{prompt}}"},
			postureAutonomous: {"--yes-always", "--message", "{{prompt}}"},
		},
		"opencode": {
			posturePlan:       {"run", "{{prompt}}"},
			postureSemi:       {"run", "{{prompt}}"}, // no clean semi flag → degrades to plan args
			postureAutonomous: {"run", "--auto", "{{prompt}}"},
		},
	}

	for mode, byPosture := range want {
		d, ok := cliHandoffDescriptorForMode(mode)
		if !ok {
			t.Fatalf("descriptor %q missing", mode)
		}
		for posture, expect := range byPosture {
			got := applyPostureArgs(d, posture)
			if strings.Join(got, " ") != strings.Join(expect, " ") {
				t.Fatalf("%s / %v: got %v want %v", mode, posture, got, expect)
			}
			// No {{posture}} token may survive; exactly one {{prompt}}.
			if containsArg(got, "{{posture}}") {
				t.Fatalf("%s / %v: {{posture}} token not substituted: %v", mode, posture, got)
			}
			if n := countArg(got, "{{prompt}}"); n != 1 {
				t.Fatalf("%s / %v: expected exactly one {{prompt}}, got %d in %v", mode, posture, n, got)
			}
		}
	}
}

// writeEnablingFlags are flags that let a CLI modify files or run commands.
var writeEnablingFlags = []string{
	"--dangerously-skip-permissions", "--dangerously-bypass-approvals-and-sandbox",
	"--yolo", "--yes-always", "--auto", "acceptEdits", "auto_edit", "--full-auto",
}

func TestCLICompat_PlanIsAlwaysReadOnly(t *testing.T) {
	for _, mode := range []string{"claude-code", "codex", "gemini-cli", "aider", "opencode"} {
		d, _ := cliHandoffDescriptorForMode(mode)
		plan := applyPostureArgs(d, posturePlan)
		for _, w := range writeEnablingFlags {
			if containsArg(plan, w) {
				t.Fatalf("SAFETY: %s plan posture carries write-enabling flag %q: %v", mode, w, plan)
			}
		}
	}
	// A write-by-default CLI must be forced read-only in plan.
	aider, _ := cliHandoffDescriptorForMode("aider")
	if !containsArg(applyPostureArgs(aider, posturePlan), "--dry-run") {
		t.Fatal("SAFETY: aider plan posture must force --dry-run (it writes+commits by default)")
	}
}

func countArg(args []string, want string) int {
	n := 0
	for _, a := range args {
		if a == want {
			n++
		}
	}
	return n
}
