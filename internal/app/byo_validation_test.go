package app

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// Every registered BYO shape must be well-formed. This is the "a shape without
// a passing fixture is not claimed" guard: a shape that can't name its
// credential or build a request would silently misvalidate.
func TestBYOShapesAreWellFormed(t *testing.T) {
	for id, shape := range byoShapes {
		if shape.ID != id {
			t.Errorf("shape %q has mismatched ID %q", id, shape.ID)
		}
		if strings.TrimSpace(shape.Label) == "" {
			t.Errorf("shape %q has empty Label", id)
		}
		if strings.TrimSpace(shape.KeyEnv) == "" {
			t.Errorf("shape %q has empty KeyEnv", id)
		}
		if shape.buildRequest == nil {
			t.Fatalf("shape %q has nil buildRequest", id)
		}
		req, err := shape.buildRequest(context.Background(), shape.defaultBase, "test-key")
		if err != nil {
			t.Errorf("shape %q buildRequest error: %v", id, err)
			continue
		}
		if req.URL == nil || req.URL.Host == "" {
			t.Errorf("shape %q built a request with no host: %v", id, req.URL)
		}
		hasAuth := req.Header.Get("Authorization") != "" || req.Header.Get("x-api-key") != ""
		if !hasAuth {
			t.Errorf("shape %q built a request with no auth header", id)
		}
	}
}

// The xAI (Grok) shape is explicitly required by the PRD trace; its aliases must
// resolve.
func TestGrokAliasesResolve(t *testing.T) {
	for _, name := range []string{"grok", "xai", "x.ai", "x-ai", "Grok", "XAI"} {
		shape, ok := resolveBYOShape(name)
		if !ok {
			t.Fatalf("alias %q did not resolve to a shape", name)
		}
		if shape.ID != "xai" {
			t.Fatalf("alias %q resolved to %q, want xai", name, shape.ID)
		}
	}
	if shape, _ := resolveBYOShape("xai"); shape.KeyEnv != "XAI_API_KEY" {
		t.Fatalf("xai KeyEnv = %q, want XAI_API_KEY", shape.KeyEnv)
	}
}

// One live call, three outcomes: an accepted key, a rejected key, and a
// throttle — each classified through the shared taxonomy. The fixture points
// the shape's base URL at an httptest server so no real network is touched.
func TestValidateBYOCredentialLiveCall(t *testing.T) {
	cases := []struct {
		status     int
		wantOK     bool
		wantKind   providerCredentialErrorKind
		wantMsgHas string
	}{
		{http.StatusOK, true, credentialErrorUnknown, "accepted"},
		{http.StatusUnauthorized, false, credentialErrorInvalid, "XAI_API_KEY"},
		{http.StatusTooManyRequests, false, credentialErrorRateLimited, "HTTP 429"},
	}
	for _, tc := range cases {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Confirm the shape actually sent the credential.
			if r.Header.Get("Authorization") != "Bearer live-key" {
				t.Errorf("missing/wrong auth header: %q", r.Header.Get("Authorization"))
			}
			w.WriteHeader(tc.status)
		}))
		func() {
			defer srv.Close()
			t.Setenv("XAI_API_KEY", "live-key")
			t.Setenv("XAI_BASE_URL", srv.URL)
			shape, _ := resolveBYOShape("grok")
			result := validateBYOCredential(context.Background(), shape, srv.Client())
			if result.OK != tc.wantOK {
				t.Fatalf("status %d: OK = %v, want %v (msg=%q)", tc.status, result.OK, tc.wantOK, result.Message)
			}
			if !result.Configured {
				t.Fatalf("status %d: expected Configured=true", tc.status)
			}
			if !tc.wantOK && result.Kind != tc.wantKind {
				t.Fatalf("status %d: Kind = %v, want %v", tc.status, result.Kind, tc.wantKind)
			}
			if !strings.Contains(result.Message, tc.wantMsgHas) {
				t.Fatalf("status %d: message %q missing %q", tc.status, result.Message, tc.wantMsgHas)
			}
		}()
	}
}

// A pasted Grok key must actually route (not just validate): generateWith
// OpenAICompatible sends the bearer key and default model, and returns the
// reply. This closes the honesty gap — a claimed shape genuinely works.
func TestOpenAICompatibleRoutingGeneratesReply(t *testing.T) {
	var gotAuth, gotModel string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("Authorization")
		var body struct {
			Model string `json:"model"`
		}
		_ = json.NewDecoder(r.Body).Decode(&body)
		gotModel = body.Model
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"choices":[{"message":{"content":"hello from grok"}}]}`))
	}))
	defer srv.Close()
	t.Setenv("XAI_API_KEY", "live-key")
	t.Setenv("XAI_BASE_URL", srv.URL)

	shape, _ := resolveBYOShape("grok")
	text, err := generateWithOpenAICompatible(context.Background(), shape, providerGenerationRequest{}, "sys", "hi")
	if err != nil {
		t.Fatalf("generate failed: %v", err)
	}
	if text != "hello from grok" {
		t.Fatalf("reply = %q, want %q", text, "hello from grok")
	}
	if gotAuth != "Bearer live-key" {
		t.Fatalf("auth header = %q, want Bearer live-key", gotAuth)
	}
	if gotModel != "grok-2-latest" {
		t.Fatalf("model = %q, want default grok-2-latest", gotModel)
	}
}

// A rejected key on the routing path surfaces the typed credential error.
func TestOpenAICompatibleRoutingTypedError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	}))
	defer srv.Close()
	t.Setenv("XAI_API_KEY", "bad-key")
	t.Setenv("XAI_BASE_URL", srv.URL)
	shape, _ := resolveBYOShape("grok")
	_, err := generateWithOpenAICompatible(context.Background(), shape, providerGenerationRequest{}, "sys", "hi")
	var ce *providerCredentialError
	if err == nil || !errors.As(err, &ce) || ce.Kind != credentialErrorInvalid {
		t.Fatalf("expected typed invalid-credential error, got %v", err)
	}
}

// `grok` must resolve through the provider-mode plumbing to the routable xai
// shape, and be ready once its key is set.
func TestGrokResolvesThroughProviderMode(t *testing.T) {
	t.Setenv("JINI_PROVIDER", "grok")
	if mode := configuredProviderMode(); mode != "xai" {
		t.Fatalf("configuredProviderMode() = %q, want xai", mode)
	}
	t.Setenv("XAI_API_KEY", "k")
	if got := detectProviderForMode("xai"); got.Status != "ok" || got.ID != "xai" {
		t.Fatalf("detectProviderForMode(xai) = %+v, want ok/xai", got)
	}
}

// A shape with no configured credential must not make a network call.
func TestValidateBYOCredentialUnconfigured(t *testing.T) {
	called := false
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(200)
	}))
	defer srv.Close()
	t.Setenv("XAI_API_KEY", "")
	t.Setenv("XAI_BASE_URL", srv.URL)
	shape, _ := resolveBYOShape("grok")
	result := validateBYOCredential(context.Background(), shape, srv.Client())
	if result.Configured || result.OK {
		t.Fatalf("unconfigured shape must report not-configured, got %+v", result)
	}
	if called {
		t.Fatal("must not make a live call when no credential is configured")
	}
	if !strings.Contains(result.Message, "XAI_API_KEY") {
		t.Fatalf("message should name the missing env var: %q", result.Message)
	}
}
