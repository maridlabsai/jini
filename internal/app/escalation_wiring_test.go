package app

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// The quote helper is unit-tested in escalation_quote_test.go; this test proves
// the wiring — that a real request routed to a metered provider actually prints
// the quote before the route is called. It is hermetic: the "provider" is a
// local httptest server that rejects the call, so the answer path fails after
// the quote has already been written.
func TestEscalationCostQuotePrintsOnPremiumProviderRoute(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	}))
	defer server.Close()

	// Auto routing with a Claude key and no installed handoff CLI (TestMain
	// pins those to a missing path) resolves to the premium claude-api route.
	t.Setenv("JINI_TOOL", "auto")
	t.Setenv("JINI_PROVIDER", "auto")
	t.Setenv("JINI_MODEL", "auto")
	t.Setenv("ANTHROPIC_API_KEY", "test-key-not-real")
	t.Setenv("ANTHROPIC_BASE_URL", server.URL)

	var stdout bytes.Buffer
	if !maybeHandleStandaloneQuestion("What is a database index?", &stdout) {
		t.Fatal("expected the standalone question path to handle this prompt")
	}

	out := stdout.String()
	for _, want := range []string{"Cost note:", "paid route"} {
		if !strings.Contains(out, want) {
			t.Errorf("routed output missing %q\ngot:\n%s", want, out)
		}
	}
}

// Imperative tasks take runDirectTaskArgsIntake (native-loop / provider paths),
// not the standalone-question path, so the quote needs its own emit there.
func TestEscalationCostQuotePrintsOnDirectTaskIntake(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	}))
	defer server.Close()

	t.Setenv("JINI_TOOL", "auto")
	t.Setenv("JINI_PROVIDER", "auto")
	t.Setenv("JINI_MODEL", "auto")
	t.Setenv("ANTHROPIC_API_KEY", "test-key-not-real")
	t.Setenv("ANTHROPIC_BASE_URL", server.URL)

	var stdout, stderr bytes.Buffer
	runDirectTaskArgsIntake(
		[]string{"add", "exponential", "backoff", "to", "the", "retry", "loop", "in", "the", "payments", "service"},
		&stdout, &stderr,
	)

	out := stdout.String()
	for _, want := range []string{"Cost note:", "paid route"} {
		if !strings.Contains(out, want) {
			t.Errorf("direct task output missing %q\nstdout:\n%s\nstderr:\n%s", want, out, stderr.String())
		}
	}
}
