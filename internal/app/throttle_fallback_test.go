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
