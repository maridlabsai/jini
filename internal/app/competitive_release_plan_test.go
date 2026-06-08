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
		"OpenCode",
		"Sourcegraph Amp",
		"Cline",
		"Continue",
		"Aider",
		"OpenHands",
		"Tabnine Agent",
		"Qodo Merge",
		"Ellipsis",
		"Roomote",
		"Devin",
		"Replit Agent",
		"Ollama",
		"LM Studio",
		"LiteLLM",
		"OpenRouter",
		"LangGraph",
		"OpenAI Agents SDK",
		"Pydantic AI",
		"CrewAI",
		"Base44",
		"Competitive watch packet",
		"Async work receipt",
		"Cross-surface session proof",
		"Local model host adapters",
		"Gateway adapter boundary",
		"Requirement Rejection Filter",
		"adopt, integrate, watch, reject, or delete",
		"Competitor watching is a P0 feature-selection loop.",
		"nominates next feature candidates and deletion candidates",
		"Compounding user productivity learning",
		"Outcome-gated competitive scorecard",
		"Code review quality is now a product bar",
		"Agent frameworks are implementation pressure, not product scope",
	} {
		if !strings.Contains(plan, want) {
			t.Fatalf("competitive release plan must include %q", want)
		}
	}

	canonicalPRD := readCompetitivePlanFile(t, root, "specs/number-one-platform-prd.md")
	for _, want := range []string{
		"competitive-release-plan.md",
		"## Market And Learning Guards",
		"Competitor watching is a P0 feature-selection loop",
		"Competitor watch packets can nominate next feature candidates",
		"learn stable user context, usage, habits, and repeated patterns",
		"No competitor finding becomes active scope unless the decision record changes.",
		"copy, integrate, watch, reject, or",
		"delete",
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
		"Base44",
		"Ollama",
		"LiteLLM",
		"OpenCode",
		"Sourcegraph Amp",
		"Tabnine Agent",
		"Qodo Merge",
		"Ellipsis",
		"LangGraph",
		"OpenAI Agents SDK",
		"Pydantic AI",
		"CrewAI",
	} {
		if !strings.Contains(benchmark, want) {
			t.Fatalf("golden benchmark watchlist must include %q", want)
		}
	}

	kpis := readCompetitivePlanFile(t, root, "specs/competitive-kpis.yaml")
	for _, want := range []string{
		"GitHub Copilot coding agent",
		"Google Jules",
		"Base44",
		"Ollama",
		"LiteLLM",
		"OpenCode",
		"Sourcegraph Amp",
		"Tabnine Agent",
		"Qodo Merge",
		"Ellipsis",
		"LangGraph",
		"OpenAI Agents SDK",
		"Pydantic AI",
		"CrewAI",
	} {
		if !strings.Contains(kpis, want) {
			t.Fatalf("competitive KPI watchlist/comparison set must include %q", want)
		}
	}

	learning := readCompetitivePlanFile(t, root, "specs/learning-system.md")
	for _, want := range []string{
		"## 2a. User Context Productivity Learning",
		"User productivity learning is a P0 product requirement.",
		"stable user context, usage, habits, and repeated work patterns",
		"fewer repeated prompts",
	} {
		if !strings.Contains(learning, want) {
			t.Fatalf("learning system must include %q", want)
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
