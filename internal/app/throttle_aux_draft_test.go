package app

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
)

// Gap G: the auxiliary quality drafts (selective consistency and refine) must be
// dropped when the primary answer already had to survive a throttle — firing
// more calls at a pressured provider risks re-throttling for a non-essential
// refinement. Hermetic: the "provider" is a local httptest server that throttles
// the first call, answers the second, and counts every request it receives.
func TestThrottledPrimaryDropsAuxiliaryDrafts(t *testing.T) {
	withThrottleTestHarness(t) // holds are narrated to a buffer and never really sleep

	var calls int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if atomic.AddInt32(&calls, 1) == 1 {
			w.WriteHeader(http.StatusTooManyRequests)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"content":[{"type":"text","text":"primary answer"}]}`))
	}))
	defer server.Close()

	t.Setenv("ANTHROPIC_API_KEY", "test-key-not-real")
	t.Setenv("ANTHROPIC_BASE_URL", server.URL)
	t.Setenv("JINI_MODEL", "auto")

	decision := routeDecision{
		Active:    true,
		ToolMode:  "claude-api",
		ToolLabel: "Claude API",
		Provider:  providerConfig{ID: "anthropic", Label: "Claude API", Status: "ok"},
	}
	// "exhaustive" classifies as extra high effort, which is exactly what turns
	// the selective consistency check on. Asserting the trigger keeps the test
	// from going quietly vacuous if the classifier changes.
	request := providerGenerationRequest{
		Title:  "Exhaustive architecture review",
		Source: "Do an exhaustive architecture review of the payments service.",
	}
	if shouldCheck, _ := shouldRunSelectiveConsistencyCheck(request, decision); !shouldCheck {
		t.Fatal("precondition: this request must trigger the selective consistency check")
	}

	text, handled, resolved, err := generateWithConfiguredProviderDecision(context.Background(), request, decision)
	if err != nil {
		t.Fatalf("expected the primary answer to survive the throttle, got error: %v", err)
	}
	if !handled || text != "primary answer" {
		t.Fatalf("expected the primary answer to stand, got handled=%v text=%q", handled, text)
	}
	if total := atomic.LoadInt32(&calls); total != 2 {
		t.Fatalf("expected exactly 2 provider calls (throttled attempt + resumed primary), got %d: an auxiliary draft hit the pressured provider", total)
	}
	if resolved.VerificationLevel == "Consistency check" || resolved.VerificationLevel == "Single pass + refine" {
		t.Fatalf("dropped drafts must not be reported as run, got verification level %q", resolved.VerificationLevel)
	}
}
