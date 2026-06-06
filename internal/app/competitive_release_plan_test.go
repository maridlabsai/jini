package app_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCompetitiveReleasePlanIsWiredIntoReleasePlanning(t *testing.T) {
	root := repoRootForMigrationTest(t)

	plan := readCompetitivePlanFile(t, root, "specs/competitive-release-plan.md")
	for _, want := range []string{
		"Claude Code",
		"OpenAI Codex",
		"GitHub Copilot coding agent",
		"Google Jules",
		"Cursor",
		"Windsurf",
		"Kiro",
		"JetBrains Junie",
		"Google Gemini CLI",
		"Cline",
		"Continue",
		"Aider",
		"OpenHands",
		"Roomote",
		"Devin",
		"Replit Agent",
		"Ollama",
		"LM Studio",
		"LiteLLM",
		"OpenRouter",
		"Competitive watch packet",
		"Async work receipt",
		"Cross-surface session proof",
		"Local model host adapters",
		"Gateway adapter boundary",
		"Requirement Rejection Filter",
		"adopt, integrate, watch, reject, or delete",
	} {
		if !strings.Contains(plan, want) {
			t.Fatalf("competitive release plan must include %q", want)
		}
	}

	canonicalPRD := readCompetitivePlanFile(t, root, "specs/number-one-platform-prd.md")
	for _, want := range []string{
		"competitive-release-plan.md",
		"P0.10 Competitive release pressure",
		"reject, downgrade, or delete requirements",
		"copy, integrate, watch, reject, delete",
	} {
		if !strings.Contains(canonicalPRD, want) {
			t.Fatalf("canonical PRD must include %q", want)
		}
	}

	developmentPlan := readCompetitivePlanFile(t, root, "specs/number-one-development-plan.md")
	if !strings.Contains(developmentPlan, "competitive-release-plan.md") {
		t.Fatalf("development plan must reference the competitive release plan")
	}

	benchmark := readCompetitivePlanFile(t, root, "specs/golden-competitive-benchmark.yaml")
	for _, want := range []string{
		"GitHub Copilot coding agent",
		"Google Jules",
		"Ollama",
		"LiteLLM",
	} {
		if !strings.Contains(benchmark, want) {
			t.Fatalf("golden benchmark watchlist must include %q", want)
		}
	}
}

func readCompetitivePlanFile(t *testing.T, root, rel string) string {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(root, rel))
	if err != nil {
		t.Fatalf("read %s: %v", rel, err)
	}
	return string(data)
}
