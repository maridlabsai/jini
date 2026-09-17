package app

import "testing"

func TestReadyThrottleFallbackModesSurfacesConfiguredBYO(t *testing.T) {
	t.Setenv("GROQ_API_KEY", "test-key")
	got := readyThrottleFallbackModes(providerGenerationRequest{}, "claude-code")
	if len(got) == 0 || got[0] != "groq" {
		t.Fatalf("expected groq as first ready fallback, got %v", got)
	}
}

// The throttled route itself must never be offered as its own fallback.
func TestReadyThrottleFallbackExcludesThrottledRoute(t *testing.T) {
	t.Setenv("GROQ_API_KEY", "test-key")
	got := readyThrottleFallbackModes(providerGenerationRequest{}, "groq")
	for _, m := range got {
		if m == "groq" {
			t.Fatalf("throttled route groq must be excluded, got %v", got)
		}
	}
}

// No configured BYO key => no online fallback surfaced (honest empty).
func TestReadyThrottleFallbackEmptyWithoutKeys(t *testing.T) {
	for _, k := range []string{"GROQ_API_KEY", "DEEPSEEK_API_KEY", "XAI_API_KEY", "MISTRAL_API_KEY", "OPENAI_API_KEY"} {
		t.Setenv(k, "")
	}
	got := readyThrottleFallbackModes(providerGenerationRequest{}, "claude-code")
	for _, m := range got {
		if m == "groq" || m == "deepseek" {
			t.Fatalf("no key configured, must not surface %q: %v", m, got)
		}
	}
}

// Cerebras is the top free rung when configured (fastest inference).
func TestCerebrasRanksFirstInLadder(t *testing.T) {
	t.Setenv("CEREBRAS_API_KEY", "test-key")
	t.Setenv("GROQ_API_KEY", "test-key")
	got := readyThrottleFallbackModes(providerGenerationRequest{}, "claude-code")
	if len(got) < 2 || got[0] != "cerebras" || got[1] != "groq" {
		t.Fatalf("expected [cerebras groq ...] order, got %v", got)
	}
}

// Gemini trains on free-tier prompts, so it must NOT enter the automatic ladder
// without an explicit opt-in — but is always available for manual selection.
func TestGeminiGatedFromAutoLadderWithoutOptIn(t *testing.T) {
	t.Setenv("GEMINI_API_KEY", "test-key")
	t.Setenv("JINI_GEMINI_ALLOW_TRAINING", "") // not opted in

	got := readyThrottleFallbackModes(providerGenerationRequest{}, "claude-code")
	for _, m := range got {
		if m == "gemini-api" {
			t.Fatalf("gemini must be excluded from the auto ladder without opt-in: %v", got)
		}
	}

	t.Setenv("JINI_GEMINI_ALLOW_TRAINING", "1") // opted in
	got = readyThrottleFallbackModes(providerGenerationRequest{}, "claude-code")
	found := false
	for _, m := range got {
		if m == "gemini-api" {
			found = true
		}
	}
	if !found {
		t.Fatalf("gemini must join the ladder once opted in: %v", got)
	}
}

// The byo Gemini API route must not collide with the gemini-cli handoff route.
func TestGeminiRouteNamesDoNotCollide(t *testing.T) {
	if got := normalizeToolMode("gemini"); got != "gemini-cli" {
		t.Fatalf("bare gemini must stay the CLI handoff, got %q", got)
	}
	if got := normalizeToolMode("gemini-api"); got != "gemini-api" {
		t.Fatalf("gemini-api must resolve to the API route, got %q", got)
	}
	if got := normalizeToolMode("gemini api"); got != "gemini-api" {
		t.Fatalf("gemini api must resolve to the API route, got %q", got)
	}
}
