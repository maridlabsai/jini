package app_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestP1SimplicityPriorityCoversCommandsSkillsAndAgents(t *testing.T) {
	root := repoRootForMigrationTest(t)

	canonicalPRD := readProductSimplicityFile(t, root, "specs/number-one-platform-prd.md")
	for _, want := range []string{
		"Bare `jini` is a task prompt, not a dashboard.",
		"Never require users to learn Jini-specific command vocabulary before value.",
		"no visible `Switch` startup control",
		"no `Start/Keep` interruption model",
		"Free excludes:",
		"developer-agent fleets",
		"tester-agent fleets",
		"commercial skills-based OS productivity suite",
		"Commercial UX still follows the same rule: simple task in, useful result out.",
	} {
		if !strings.Contains(canonicalPRD, want) {
			t.Fatalf("canonical PRD must preserve P1 simplification requirement %q", want)
		}
	}

	skillsSlice := readProductSimplicityFile(t, root, "specs/skills-and-delegation-slice.md")
	for _, want := range []string{
		"## Tier Boundary",
		"Commercial tier owns the agent and skills OS productivity suite.",
		"Free tier must not ship `skills`, `delegate`, developer agents, tester agents,",
		"or a skills-based OS productivity suite.",
		"This is a P1 simplification slice, not a reason to teach another interaction",
		"Natural `jini` intake remains the default path.",
		"Commercial helper capability can be mentioned through progressive-disclosure",
		"controls for users who need them.",
		"without requiring the user to learn the word",
		"It must not show agent trees, role theater, or step-by-step orchestration logs",
		"public docs and readiness gates do not describe the commercial suite as a",
		"free-tier runtime requirement",
	} {
		if !strings.Contains(skillsSlice, want) {
			t.Fatalf("skills slice must preserve simplified delegation rule %q", want)
		}
	}
	for _, forbidden := range []string{
		"Jini should load skills from two roots",
		"The explicit command is:",
		"The discovery command is:",
		"`skills` returns discovered skills",
		"`delegate` must:",
		"Jini supports file-backed project and user skills",
		"`skills` and `delegate` work end-to-end",
	} {
		if strings.Contains(skillsSlice, forbidden) {
			t.Fatalf("skills slice must not keep free-tier implementation requirement %q", forbidden)
		}
	}

	leanGate := readProductSimplicityFile(t, root, "specs/lean-platform-gate.md")
	for _, want := range []string{
		"This discipline also applies to skills and agent interactions.",
		"The free tier must not include a skills-based OS productivity suite.",
		"they must not become a second command tree or visible agent control plane.",
		"- teaches skill or agent vocabulary as a prerequisite to normal use",
		"- ships developer agents, tester agents, `skills`, `delegate`, or a skills-based OS productivity suite in the free tier",
		"- shows agent trees, role theater, or orchestration logs by default",
		"- `skill-agent-interaction-simplicity`",
	} {
		if !strings.Contains(leanGate, want) {
			t.Fatalf("lean platform gate must preserve skill and agent simplification rule %q", want)
		}
	}
}

func readProductSimplicityFile(t *testing.T, root, rel string) string {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(root, rel))
	if err != nil {
		t.Fatalf("read %s: %v", rel, err)
	}
	return string(data)
}
