package app

import "testing"

func TestEstimateTaskDifficultyLow(t *testing.T) {
	for _, prompt := range []string{
		"fix a typo in the README",
		"add a comment to the add function",
		"print hello",
	} {
		if got := estimateTaskDifficulty(prompt); got != difficultyLow {
			t.Fatalf("%q should be low difficulty, got %s", prompt, got)
		}
	}
}

func TestEstimateTaskDifficultyHigh(t *testing.T) {
	for _, prompt := range []string{
		"refactor the authentication and authorization across the whole codebase to be thread-safe",
		"migrate every module to the new architecture and fix the race condition, this is a breaking change touching all packages",
	} {
		if got := estimateTaskDifficulty(prompt); got != difficultyHigh {
			t.Fatalf("%q should be high difficulty, got %s", prompt, got)
		}
	}
}

func TestEstimateTaskDifficultyMedium(t *testing.T) {
	// one hard signal (+1) plus a long prompt (>160 chars, +1) = 2 → medium,
	// not escalating to high.
	prompt := "Please optimize performance of the parser module — walk through the tokenizer and the lexer and tidy up the hot loop so it allocates less, keeping the existing behavior and its tests exactly as they are today."
	if got := estimateTaskDifficulty(prompt); got != difficultyMedium {
		t.Fatalf("expected medium difficulty, got %s (len=%d)", got, len(prompt))
	}
}

func TestEscalationCapScalesWithDifficulty(t *testing.T) {
	if escalationCapForDifficulty(difficultyLow) != 1 {
		t.Fatalf("a trivial task must cap at a single escalation")
	}
	if escalationCapForDifficulty(difficultyMedium) != verificationEscalationCap {
		t.Fatalf("a medium task uses the default cap")
	}
	if escalationCapForDifficulty(difficultyHigh) <= verificationEscalationCap {
		t.Fatalf("a hard task must earn more escalation budget than the default")
	}
}
