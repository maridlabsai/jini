package app

import (
	"context"
	"testing"
)

func withCatalog(t *testing.T, entries ...catalogProviderEntry) {
	t.Helper()
	prev := providerCatalogLoader
	providerCatalogLoader = func() []catalogProviderEntry { return entries }
	t.Cleanup(func() { providerCatalogLoader = prev })
}

// A catalog entry with a NEW id adds a routable provider — no code change.
func TestCatalogAddsNewProvider(t *testing.T) {
	withCatalog(t, catalogProviderEntry{
		ID: "together", Label: "Together AI", KeyEnv: "TOGETHER_API_KEY",
		DefaultBase: "https://api.together.xyz", ChatPath: "/v1/chat/completions",
		ModelEnv: "TOGETHER_MODEL", DefaultModel: "meta-llama/Llama-3.3-70B",
	})
	shape, ok := byoShapeByID("together")
	if !ok {
		t.Fatal("catalog provider not registered")
	}
	if !shape.routable() || shape.KeyEnv != "TOGETHER_API_KEY" {
		t.Fatalf("catalog provider not routable/keyed: %+v", shape)
	}
	if s, ok := resolveBYOShape("together"); !ok || s.ID != "together" {
		t.Fatalf("resolveBYOShape(together) failed: %+v ok=%v", s, ok)
	}
}

// A catalog entry matching a BUILT-IN id overrides only the fields it sets —
// the fix for a stale default model without redefining the provider.
func TestCatalogOverridesStaleDefaultModel(t *testing.T) {
	// Sanity: built-in groq default before overlay.
	if s, _ := byoShapeByID("groq"); s.defaultModel == "" {
		t.Fatal("groq built-in should have a default model")
	}
	withCatalog(t, catalogProviderEntry{ID: "groq", DefaultModel: "llama-4-next"})
	s, ok := byoShapeByID("groq")
	if !ok {
		t.Fatal("groq missing after overlay")
	}
	if s.defaultModel != "llama-4-next" {
		t.Fatalf("override default_model = %q, want llama-4-next", s.defaultModel)
	}
	// Other fields are preserved (overlay is field-level, not replacement).
	if s.KeyEnv != "GROQ_API_KEY" || s.chatPath == "" {
		t.Fatalf("overlay clobbered untouched fields: %+v", s)
	}
}

// A catalog entry with no key env is dropped (can't authenticate → useless).
func TestCatalogRejectsKeylessEntry(t *testing.T) {
	withCatalog(t, catalogProviderEntry{ID: "ghost", Label: "Ghost"})
	if _, ok := byoShapeByID("ghost"); ok {
		t.Fatal("keyless catalog entry must be dropped")
	}
}

// A catalog-added provider is discoverable/validatable end to end.
func TestCatalogProviderValidates(t *testing.T) {
	withCatalog(t, catalogProviderEntry{
		ID: "together", Label: "Together AI", KeyEnv: "TOGETHER_API_KEY",
		DefaultBase: "https://api.together.xyz", ChatPath: "/v1/chat/completions",
	})
	t.Setenv("TOGETHER_API_KEY", "")
	shape, _ := byoShapeByID("together")
	res := validateBYOCredential(context.Background(), shape, byoValidationClient())
	if res.Configured {
		t.Fatalf("unconfigured catalog provider should report not-configured: %+v", res)
	}
}
