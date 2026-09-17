package app

import (
	"testing"
	"time"
)

func TestStandaloneAttemptTimeout_RouteAware(t *testing.T) {
	// A local/provider route keeps the short local budget.
	local := routeDecision{ToolMode: "local-slm"}
	if got := standaloneAttemptTimeout(local); got != 10*time.Second {
		t.Fatalf("local route budget = %v, want 10s", got)
	}
	// A CLI hand-off gets the generous budget so `claude --print` isn't killed.
	handoff := routeDecision{ToolMode: "claude-code"}
	if got := standaloneAttemptTimeout(handoff); got != 3*time.Minute {
		t.Fatalf("hand-off budget = %v, want 3m", got)
	}
	// An explicit override wins for any route, including a hand-off, so a caller
	// can still bound a slow CLI.
	t.Setenv("JINI_STANDALONE_QUESTION_TIMEOUT", "20ms")
	if got := standaloneAttemptTimeout(handoff); got != 20*time.Millisecond {
		t.Fatalf("explicit override must win for hand-offs, got %v", got)
	}
}

func TestCLIHandoffAttemptTimeout_EnvOverride(t *testing.T) {
	t.Setenv("JINI_CLI_HANDOFF_TIMEOUT", "90s")
	if got := cliHandoffAttemptTimeout(); got != 90*time.Second {
		t.Fatalf("override = %v, want 90s", got)
	}
	t.Setenv("JINI_CLI_HANDOFF_TIMEOUT", "garbage")
	if got := cliHandoffAttemptTimeout(); got != 3*time.Minute {
		t.Fatalf("invalid override should fall back to 3m, got %v", got)
	}
}
