package app

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestSkillNewCreatesReviewableFile(t *testing.T) {
	dir := t.TempDir()
	t.Chdir(dir)
	var stdout, stderr bytes.Buffer
	if code := runSkill([]string{"new", "Run Gates", "run the commit gate"}, &stdout, &stderr); code != 0 {
		t.Fatalf("skill new exit %d; stderr: %s", code, stderr.String())
	}
	path := filepath.Join(dir, ".jini", "skills", "run-gates.md")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("expected skill file at %s: %v", path, err)
	}
	for _, want := range []string{"name: run-gates", "description: run the commit gate", "kind: skill", "## When to use"} {
		if !strings.Contains(string(data), want) {
			t.Fatalf("skill file missing %q:\n%s", want, data)
		}
	}
}

func TestAgentNewCreatesReviewableFile(t *testing.T) {
	dir := t.TempDir()
	t.Chdir(dir)
	var stdout, stderr bytes.Buffer
	if code := runAgent([]string{"new", "reviewer"}, &stdout, &stderr); code != 0 {
		t.Fatalf("agent new exit %d; stderr: %s", code, stderr.String())
	}
	data, err := os.ReadFile(filepath.Join(dir, ".jini", "agents", "reviewer.md"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), "kind: agent") || !strings.Contains(string(data), "## Role") {
		t.Fatalf("agent file missing expected sections:\n%s", data)
	}
}

// Creation runs a secret scrub before writing (PRD §Skills And Agents).
func TestSkillNewRefusesSecret(t *testing.T) {
	dir := t.TempDir()
	t.Chdir(dir)
	var stdout, stderr bytes.Buffer
	code := runSkill([]string{"new", "leaky", "paste sk-abcdefghijklmnopqrstuvwxyz012345 here"}, &stdout, &stderr)
	if code == 0 {
		t.Fatal("expected refusal when content contains a secret")
	}
	if !strings.Contains(stderr.String(), "secret") {
		t.Fatalf("expected a secret-scrub message, got: %s", stderr.String())
	}
	if _, err := os.Stat(filepath.Join(dir, ".jini", "skills", "leaky.md")); !os.IsNotExist(err) {
		t.Fatal("must not write a file that contains a secret")
	}
}

func TestSkillNewRejectsMissingName(t *testing.T) {
	dir := t.TempDir()
	t.Chdir(dir)
	var stdout, stderr bytes.Buffer
	if code := runSkill([]string{"new"}, &stdout, &stderr); code == 0 {
		t.Fatal("expected usage error when no name is given")
	}
}

// The free-tier creation surface must be recognized so the commercial gate lets
// it through, while bare skill/agent stays commercial.
func TestIsFreeSkillOrAgentCreation(t *testing.T) {
	free := [][]string{{"skill", "new", "x"}, {"agent", "new", "y"}}
	notFree := [][]string{{"skill"}, {"agent"}, {"skill", "list"}, {"skills", "new"}, {"provider", "new"}}
	for _, a := range free {
		if !isFreeSkillOrAgentCreation(a) {
			t.Errorf("%v should be free creation", a)
		}
	}
	for _, a := range notFree {
		if isFreeSkillOrAgentCreation(a) {
			t.Errorf("%v should NOT be free creation", a)
		}
	}
}

// The commercial Skills OS / agent fleets stay gated.
func TestBareSkillStaysCommercial(t *testing.T) {
	if _, ok := commercialOnlyCommandAccess("skills"); !ok {
		t.Fatal("bare `skills` (Skills OS) must remain commercial-gated")
	}
	if _, ok := commercialOnlyCommandAccess("agents"); !ok {
		t.Fatal("bare `agents` (agent fleets) must remain commercial-gated")
	}
}
