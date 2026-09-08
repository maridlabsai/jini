package app

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestListProviderModelsParsesAndSorts(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer k" {
			t.Errorf("missing bearer auth: %q", r.Header.Get("Authorization"))
		}
		w.Header().Set("Content-Type", "application/json")
		// Deliberately unsorted + a brand-new model id.
		_, _ = w.Write([]byte(`{"data":[{"id":"llama-3.3-70b-versatile"},{"id":"future-model-9"},{"id":"gemma2-9b"}]}`))
	}))
	defer srv.Close()
	t.Setenv("GROQ_API_KEY", "k")
	t.Setenv("GROQ_BASE_URL", srv.URL)

	shape, _ := resolveBYOShape("groq")
	got, err := listProviderModels(context.Background(), shape, srv.Client())
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"future-model-9", "gemma2-9b", "llama-3.3-70b-versatile"}
	if strings.Join(got, ",") != strings.Join(want, ",") {
		t.Fatalf("models = %v, want %v (sorted)", got, want)
	}
}

// A brand-new model a provider lists is usable with zero Jini change — the
// discovery command surfaces it and marks the currently selected one.
func TestRunModelsMarksCurrentAndListsNew(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"data":[{"id":"gpt-6"},{"id":"gpt-4o"}]}`))
	}))
	defer srv.Close()
	t.Setenv("OPENAI_API_KEY", "k")
	t.Setenv("OPENAI_BASE_URL", srv.URL)
	t.Setenv("OPENAI_MODEL", "gpt-6") // user opted into the new model

	var out bytes.Buffer
	if code := runModels([]string{"openai"}, &out, &out); code != 0 {
		t.Fatalf("models exit %d: %s", code, out.String())
	}
	s := out.String()
	if !strings.Contains(s, "gpt-6") || !strings.Contains(s, "gpt-4o") {
		t.Fatalf("expected discovered models listed, got:\n%s", s)
	}
	if !strings.Contains(s, "→ gpt-6") {
		t.Fatalf("current model gpt-6 should be marked, got:\n%s", s)
	}
}

// No key configured => a helpful message, no crash.
func TestRunModelsUnconfigured(t *testing.T) {
	shape, _ := resolveBYOShape("openai")
	t.Setenv(shape.KeyEnv, "")
	var out bytes.Buffer
	if code := runModels([]string{"openai"}, &out, &out); code == 0 {
		t.Fatal("expected non-zero for an unconfigured specific provider")
	}
}
