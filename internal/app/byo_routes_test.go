package app

import "testing"

// BYO providers must be first-class, manually-selectable routes (so route set /
// the paid Autopilot fallback can target them) but NEVER auto-adopted (frugality:
// a stray OPENAI_API_KEY must not silently start spending).
func TestBYOProvidersAreManualRoutes(t *testing.T) {
	reg := adapterRegistry()
	for _, id := range []string{"groq", "cerebras", "deepseek", "xai", "mistral", "openai"} {
		d, ok := reg[id]
		if !ok {
			t.Fatalf("BYO route %q missing from adapter registry", id)
		}
		if d.ProviderMode != id {
			t.Errorf("%q ProviderMode = %q, want %q (so detectProviderForMode resolves it)", id, d.ProviderMode, id)
		}
		if d.SupportsAutoRoute {
			t.Errorf("%q must NOT be auto-routable (frugality: no silent paid adoption)", id)
		}
		if d.CostTier != "byo" {
			t.Errorf("%q CostTier = %q, want byo (keeps the escalation quote from misquoting frontier prices)", id, d.CostTier)
		}
		// Selectable as a route: adapterDescriptorForMode resolves it.
		if _, ok := adapterDescriptorForMode(id); !ok {
			t.Errorf("%q not resolvable via adapterDescriptorForMode", id)
		}
	}
}

// A configured BYO route (key present) resolves to a ready provider — the
// precondition for the throttle-survival switch executor to reach it.
func TestConfiguredBYORouteResolvesReady(t *testing.T) {
	t.Setenv("GROQ_API_KEY", "test-key")
	got := detectProviderForMode("groq")
	if got.ID != "groq" || got.Status != "ok" {
		t.Fatalf("detectProviderForMode(groq) = %+v, want ok/groq", got)
	}
}
