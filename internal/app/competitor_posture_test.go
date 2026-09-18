package app_test

import (
	"strings"
	"testing"
)

// The competitive posture spec (specs/competitor-strengths-weaknesses.md) is the
// forum-grounded source of truth for what each rival does well (a bar Jini must
// match or honestly concede) and badly (a wedge Jini's architecture presses).
// This test keeps it honest: it must cover every major competitor and explicitly
// name the must-match strengths and the exploitable weaknesses, so the posture
// cannot silently go stale as the market moves.
func TestCompetitorPostureCoversStrengthsAndWeaknesses(t *testing.T) {
	root := repoRootForMigrationTest(t)
	spec := strings.ToLower(readRepoFile(t, root, "specs/competitor-strengths-weaknesses.md"))

	// Every major competitor must be analyzed by name.
	for _, competitor := range []string{"codex", "kiro", "cursor", "windsurf", "claude code"} {
		if !strings.Contains(spec, competitor) {
			t.Fatalf("competitor posture must analyze %q", competitor)
		}
	}

	// Each rival's genuine strength must be acknowledged — a bar Jini must match.
	// If we drop one, we risk shipping a product that concedes ground blindly.
	for _, strength := range []string{
		"spec-driven",       // Kiro's real edge
		"autocomplete",      // Cursor Tab
		"multi-model",       // Cursor per-task model switching
		"cascade",           // Windsurf coherent multi-file context
		"agentic autonomy",  // Claude Code
		"context engineering", // the cross-cutting differentiator
	} {
		if !strings.Contains(spec, strength) {
			t.Fatalf("competitor posture must acknowledge the strength %q (a bar Jini must match)", strength)
		}
	}

	// Each exploitable weakness — Jini's wedge — must be named, so the message and
	// roadmap keep pressing where rivals actually hurt.
	for _, weakness := range []string{
		"throttle",   // Codex/Claude/Kiro walls
		"credit",     // opaque/punishing credit economics
		"byok",       // no bring-your-own-key (Cursor)
		"lock-in",    // single-vendor lock-in
		"hallucinat", // hallucination control now beats raw capability
	} {
		if !strings.Contains(spec, weakness) {
			t.Fatalf("competitor posture must name the exploitable weakness %q (Jini's wedge)", weakness)
		}
	}

	// The synthesis sections must be explicit, not left implied.
	for _, section := range []string{"must-match strengths", "exploitable weaknesses"} {
		if !strings.Contains(spec, section) {
			t.Fatalf("competitor posture must carry an explicit %q section", section)
		}
	}
}
