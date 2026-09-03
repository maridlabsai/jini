package app

import (
	"strings"
	"testing"
)

func TestEscalationCostQuoteLinePremiumRouteQuotesCost(t *testing.T) {
	line := escalationCostQuoteLine(
		routeDecision{ToolMode: "claude-api", ToolLabel: "Claude API"},
		"system prompt text",
		"user prompt text",
	)
	if line == "" {
		t.Fatal("premium route: want a cost quote, got empty string")
	}
	for _, want := range []string{"Claude API", "paid route", "est ~$"} {
		if !strings.Contains(line, want) {
			t.Errorf("premium quote missing %q\ngot: %s", want, line)
		}
	}
}

func TestEscalationCostQuoteLineStandardRouteQuotesCost(t *testing.T) {
	line := escalationCostQuoteLine(routeDecision{ToolMode: "chatgpt"}, "sys", "user")
	if line == "" {
		t.Fatal("standard route: want a cost quote, got empty string")
	}
	if !strings.Contains(line, "paid route") {
		t.Errorf("standard quote missing %q\ngot: %s", "paid route", line)
	}
}

// Routes that cost the user nothing at the point of use must stay silent, as
// must an unknown mode — a quote there would be a guess.
func TestEscalationCostQuoteLineFreeRoutesQuoteNothing(t *testing.T) {
	for _, mode := range []string{"claude-code", "local-preview", "local-fast", "not-a-real-mode", ""} {
		if line := escalationCostQuoteLine(routeDecision{ToolMode: mode}, "sys", "user"); line != "" {
			t.Errorf("mode %q: want no quote, got: %s", mode, line)
		}
	}
}

// The quote falls back to the raw mode when the decision carries no label.
func TestEscalationCostQuoteLineFallsBackToToolMode(t *testing.T) {
	line := escalationCostQuoteLine(routeDecision{ToolMode: "chatgpt"}, "sys", "user")
	if !strings.Contains(line, "chatgpt") {
		t.Errorf("want fallback to tool mode\ngot: %s", line)
	}
}
