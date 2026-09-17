package app

import (
	"strings"
	"testing"
)

// Adversarial coverage for the hand-off argument pipeline: applyPostureArgs
// (fills {{posture}} from the descriptor only) → cliHandoffArgsWithPrompt (fills
// {{prompt}} with the user prompt as exactly one argv element). The security
// properties below must hold for every route.

// A prompt can NEVER inject an extra argument or a permission flag: whatever its
// content, it occupies exactly one argv slot, and posture flags come only from
// the descriptor.
func TestAdversarial_PromptIsAlwaysExactlyOneArg(t *testing.T) {
	nasty := []string{
		"rm -rf / ; echo pwned",
		"--dangerously-skip-permissions",
		"--yolo --auto",
		"$(whoami)`id`",
		"a\nb\tc",
		"{{prompt}} {{posture}}",
		"--message --dry-run",
		`"; drop table users; --`,
		"",
	}
	for _, mode := range []string{"claude-code", "codex", "gemini-cli", "aider", "opencode"} {
		d, _ := cliHandoffDescriptorForMode(mode)
		for _, posture := range []handoffPosture{posturePlan, postureSemi, postureAutonomous} {
			base := applyPostureArgs(d, posture)
			baseLen := len(base)
			for _, prompt := range nasty {
				final := cliHandoffArgsWithPrompt(base, prompt)
				// Exactly one slot changes: {{prompt}} → prompt. No new args.
				if len(final) != baseLen {
					t.Fatalf("%s/%v prompt %q changed arg count %d→%d: %v", mode, posture, prompt, baseLen, len(final), final)
				}
				// The prompt appears as a whole, standalone element (never split).
				found := false
				for _, a := range final {
					if a == prompt {
						found = true
					}
				}
				if !found {
					t.Fatalf("%s/%v: prompt %q not present as a single element: %v", mode, posture, prompt, final)
				}
				// The posture-applied base (everything except the prompt slot)
				// must carry no unsubstituted template token — prompt CONTENT may
				// legitimately contain such text, so only non-prompt args count.
				for _, a := range final {
					if a == prompt {
						continue
					}
					if strings.Contains(a, "{{prompt}}") || strings.Contains(a, "{{posture}}") {
						t.Fatalf("%s/%v: unsubstituted token in a non-prompt arg: %v", mode, posture, final)
					}
				}
			}
		}
	}
}

// Even when the prompt is exactly a dangerous flag, plan posture stays plan: the
// prompt sits in the prompt slot, and no descriptor-sourced write flag is added.
func TestAdversarial_PromptCannotEscalatePlanPosture(t *testing.T) {
	d, _ := cliHandoffDescriptorForMode("claude-code")
	plan := applyPostureArgs(d, posturePlan)
	final := cliHandoffArgsWithPrompt(plan, "--dangerously-skip-permissions")
	// The dangerous string is the prompt VALUE (last arg), preceded only by
	// --print — Jini never emitted it as a control flag of its own.
	if strings.Join(final, " ") != "--print --dangerously-skip-permissions" {
		t.Fatalf("unexpected args: %v", final)
	}
	// Structurally: plan produced no posture args, so the only reason the string
	// is present is that it is the prompt (one element), not two.
	if len(final) != 2 {
		t.Fatalf("prompt must not add an arg: %v", final)
	}
}

// The legacy fallback (descriptor with no {{posture}} slot) must never panic and
// must insert before the final arg.
func TestAdversarial_LegacyFallbackNoPanic(t *testing.T) {
	cases := []struct {
		def  []string
		auto []string
		want []string
	}{
		{[]string{"x", "{{prompt}}"}, []string{"--a"}, []string{"x", "--a", "{{prompt}}"}},
		{[]string{"only"}, []string{"--a"}, []string{"--a", "only"}},
		{[]string{}, []string{"--a"}, []string{"--a"}},
		{[]string{"x", "{{prompt}}"}, nil, []string{"x", "{{prompt}}"}},
	}
	for i, c := range cases {
		d := cliHandoffDescriptor{DefaultArgs: c.def, AutonomousArgs: c.auto}
		got := applyPostureArgs(d, postureAutonomous)
		if strings.Join(got, " ") != strings.Join(c.want, " ") {
			t.Fatalf("case %d: got %v want %v", i, got, c.want)
		}
	}
}

// Multiple {{posture}} tokens are all substituted (defensive).
func TestAdversarial_MultiplePostureTokens(t *testing.T) {
	d := cliHandoffDescriptor{
		DefaultArgs:    []string{"{{posture}}", "mid", "{{posture}}", "{{prompt}}"},
		AutonomousArgs: []string{"--a"},
	}
	got := applyPostureArgs(d, postureAutonomous)
	want := []string{"--a", "mid", "--a", "{{prompt}}"}
	if strings.Join(got, " ") != strings.Join(want, " ") {
		t.Fatalf("got %v want %v", got, want)
	}
}
