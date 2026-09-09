package app

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// A capable model (correct probe replies) verifies; a broken one does not — the
// gate is deterministic and needs no human/LLM judge.
func TestRunModelCheckVerdicts(t *testing.T) {
	// Good server: answers each probe correctly.
	good := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body := readAllString(r)
		reply := "READY"
		switch {
		case strings.Contains(body, "JSON object"):
			reply = `{"ok": true, "n": 7}`
		case strings.Contains(body, "Go function named add"):
			reply = "func add(a, b int) int { return a + b }"
		case strings.Contains(body, "three fruits"):
			reply = "apple\nbanana\ncherry"
		}
		writeChatCompletion(w, reply)
	}))
	defer good.Close()

	// Broken server: always returns garbage.
	bad := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		writeChatCompletion(w, "lol idk here is a long rambling non-answer that ignores the instruction entirely")
	}))
	defer bad.Close()

	t.Setenv("GROQ_API_KEY", "k")

	t.Setenv("GROQ_BASE_URL", good.URL)
	var out bytes.Buffer
	if code := runModelCheck("groq", &out, &out); code != 0 {
		t.Fatalf("capable model must verify (exit 0), got %d:\n%s", code, out.String())
	}
	if !strings.Contains(out.String(), "VERIFIED — 4/4") {
		t.Fatalf("expected 4/4 VERIFIED, got:\n%s", out.String())
	}

	t.Setenv("GROQ_BASE_URL", bad.URL)
	out.Reset()
	if code := runModelCheck("groq", &out, &out); code != 1 {
		t.Fatalf("broken model must NOT verify (exit 1), got %d:\n%s", code, out.String())
	}
	if !strings.Contains(out.String(), "UNVERIFIED") {
		t.Fatalf("expected UNVERIFIED, got:\n%s", out.String())
	}
}

// A handoff route is not a model-probe target.
func TestRunModelCheckRejectsHandoff(t *testing.T) {
	var out bytes.Buffer
	if code := runModelCheck("claude-code", &out, &out); code != 1 {
		t.Fatalf("handoff route should be rejected, exit %d", code)
	}
}

func readAllString(r *http.Request) string {
	buf := new(bytes.Buffer)
	_, _ = buf.ReadFrom(r.Body)
	return buf.String()
}

func writeChatCompletion(w http.ResponseWriter, content string) {
	w.Header().Set("Content-Type", "application/json")
	body, _ := json.Marshal(map[string]any{
		"choices": []map[string]any{{"message": map[string]any{"content": content}}},
	})
	_, _ = w.Write(body)
}
