package app

import "strings"

// #1 Phase 3 of the verification design (specs/pre-viral-readiness.md): a cheap,
// objective pre-estimate of task difficulty so the escalation loop spends its
// budget where it matters — a hard change (refactor across a codebase, concurrency,
// security) earns more attempts to reach a verified result, a trivial one caps at a
// single escalation. Heuristic only (no model call): keeps it fast, deterministic,
// and bias-free. It never *starts* on a paid route — it only tunes how many
// consented escalations a failure is allowed, so the free/commercial boundary and
// frugality are untouched.

type taskDifficulty int

const (
	difficultyLow taskDifficulty = iota
	difficultyMedium
	difficultyHigh
)

func (d taskDifficulty) String() string {
	switch d {
	case difficultyHigh:
		return "high"
	case difficultyMedium:
		return "medium"
	default:
		return "low"
	}
}

// hard-work signals: kinds of change that are genuinely error-prone.
var taskDifficultyHardSignals = []string{
	"refactor", "migrate", "migration", "architecture", "redesign", "rewrite",
	"concurren", "goroutine", "race condition", "distributed", "thread-safe",
	"optimize performance", "security", "authentication", "authorization",
	"backward compat", "breaking change",
}

// scope signals: touching many files is harder to get right.
var taskDifficultyScopeSignals = []string{
	"multiple files", "several files", "all files", "every file", "across the",
	"whole codebase", "entire codebase", "each module", "all packages", "everywhere",
}

// ambiguity signals: vague asks are harder to satisfy correctly.
var taskDifficultyAmbiguitySignals = []string{
	"somehow", "figure out", "not sure", "and so on", "etc.", "or something",
}

// estimateTaskDifficulty scores a prompt heuristically.
func estimateTaskDifficulty(prompt string) taskDifficulty {
	p := strings.ToLower(prompt)
	score := 0

	switch {
	case len(prompt) > 400:
		score += 2
	case len(prompt) > 160:
		score++
	}
	score += countSignalHits(p, taskDifficultyHardSignals)
	score += countSignalHits(p, taskDifficultyScopeSignals)
	score += countSignalHits(p, taskDifficultyAmbiguitySignals)

	switch {
	case score >= 4:
		return difficultyHigh
	case score >= 2:
		return difficultyMedium
	default:
		return difficultyLow
	}
}

func countSignalHits(haystack string, signals []string) int {
	hits := 0
	for _, s := range signals {
		if strings.Contains(haystack, s) {
			hits++
		}
	}
	return hits
}

// escalationCapForDifficulty tunes the verification-escalation budget to the
// task's difficulty, bounded so a single task never walks an unbounded ladder.
func escalationCapForDifficulty(d taskDifficulty) int {
	switch d {
	case difficultyHigh:
		return verificationEscalationCap + 1
	case difficultyLow:
		return 1
	default:
		return verificationEscalationCap
	}
}
