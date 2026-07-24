package app

import (
	"math"
	"testing"
)

func TestBaselineForRoute(t *testing.T) {
	cases := []struct {
		class, label string
		want         string
		ok           bool
	}{
		{savingsRouteSubscription, "Claude Code CLI handoff", "claude-frontier", true},
		{savingsRouteSubscription, "Codex CLI handoff", "gpt-frontier", true},
		{savingsRouteLocal, "local-fast", "claude-frontier", true},
		{savingsRouteMetered, "OpenAI API", "", false},
		{"unknown", "whatever", "", false},
	}
	for _, c := range cases {
		got, ok := baselineForRoute(c.class, c.label)
		if got != c.want || ok != c.ok {
			t.Errorf("baselineForRoute(%q,%q)=%q,%v want %q,%v", c.class, c.label, got, ok, c.want, c.ok)
		}
	}
}

func TestSavingsUSDUnknownBaselineIsZero(t *testing.T) {
	if v := savingsUSD("nonexistent", 1000, 1000); v != 0 {
		t.Fatalf("unknown baseline must impute 0, got %v", v)
	}
}

func TestSavingsUSDComputesFromPostedPrices(t *testing.T) {
	// claude-frontier: in 3.00/1M, out 15.00/1M. 1M in + 1M out = 3 + 15 = 18.
	got := savingsUSD("claude-frontier", 1_000_000, 1_000_000)
	if math.Abs(got-18.0) > 1e-9 {
		t.Fatalf("expected 18.0, got %v", got)
	}
}

func TestEstTokensFromChars(t *testing.T) {
	if got := estTokensFromChars(400); got != 100 {
		t.Fatalf("400 chars / 4 = 100, got %d", got)
	}
	if got := estTokensFromChars(0); got != 0 {
		t.Fatalf("0 chars = 0 tokens, got %d", got)
	}
	if got := estTokensFromChars(-5); got != 0 {
		t.Fatalf("negative chars = 0 tokens, got %d", got)
	}
}
