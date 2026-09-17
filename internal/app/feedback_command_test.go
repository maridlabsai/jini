package app

import (
	"bytes"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestFeedbackGitHubLinks(t *testing.T) {
	title, body, submit, browse := feedbackGitHubLinks("bug", "route set groq fails", "jini 0.1.2 (source)")
	if !strings.HasPrefix(title, "Bug: ") {
		t.Fatalf("bug title prefix, got %q", title)
	}
	if !strings.Contains(body, "route set groq fails") || !strings.Contains(body, "jini 0.1.2") {
		t.Fatalf("body missing message/context: %q", body)
	}
	// The submit URL must carry the message, the bug label, and the title.
	u, err := url.Parse(submit)
	if err != nil {
		t.Fatal(err)
	}
	q := u.Query()
	if q.Get("labels") != "bug" || !strings.Contains(q.Get("body"), "route set groq fails") {
		t.Fatalf("submit URL missing label/body: %s", submit)
	}
	if !strings.Contains(browse, "reactions") || !strings.Contains(browse, "label%3Abug") {
		t.Fatalf("browse URL should sort by reactions and filter the label: %s", browse)
	}

	// Asks are labelled enhancement.
	_, _, askSubmit, _ := feedbackGitHubLinks("ask", "add cerebras", "ctx")
	if !strings.Contains(askSubmit, "labels=enhancement") {
		t.Fatalf("ask should be labelled enhancement: %s", askSubmit)
	}
}

func TestRunFeedbackWritesLogAndPrintsLink(t *testing.T) {
	t.Setenv("JINI_STATE_DIR", t.TempDir())
	var out, errB bytes.Buffer
	if code := runFeedback([]string{"bug", "it", "broke"}, &out, &errB); code != 0 {
		t.Fatalf("exit %d: %s", code, errB.String())
	}
	if !strings.Contains(out.String(), "github.com/maridlabsai/jini/issues/new") {
		t.Fatalf("expected a prefilled issue link, got:\n%s", out.String())
	}
	data, err := os.ReadFile(filepath.Join(sessionStateRoot(), "feedback.log"))
	if err != nil {
		t.Fatalf("feedback log not written: %v", err)
	}
	if !strings.Contains(string(data), "\tbug\t") || !strings.Contains(string(data), "it broke") {
		t.Fatalf("log missing kind/message: %s", data)
	}
}

func TestRunFeedbackRejectsEmpty(t *testing.T) {
	var out bytes.Buffer
	if code := runFeedback([]string{}, &out, &out); code != 1 {
		t.Fatalf("empty feedback should exit 1, got %d", code)
	}
}
