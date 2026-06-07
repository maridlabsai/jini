package app_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestResourcePolicyPrioritiesAreGated(t *testing.T) {
	root := repoRootForMigrationTest(t)

	canonicalPRD := readResourcePolicyFile(t, root, "specs/number-one-platform-prd.md")
	for _, want := range []string{
		"### 2a. Token frugality is P0",
		"Being frugal with tokens is a P0 product goal",
		"Every route, memory, skill, agent, and adapter decision must prefer the smallest",
		"Token savings must come from architecture and routing discipline",
		"#### P0.4 Token frugality and local-first cost discipline",
		"- token frugality is a P0 product goal",
		"- compact continuation must reuse saved state instead of replaying full",
		"- throttle avoidance, including preemptive route switching",
		"- powered-mode full power execution",
		"- low-battery and thermal-aware execution",
		"- `token-efficiency >= 9.5`",
		"- `throttle-avoided-interruption-rate >= 85%`",
		"- `battery-aware-route-regret <= 5%`",
	} {
		if !strings.Contains(canonicalPRD, want) {
			t.Fatalf("canonical PRD must preserve resource policy requirement %q", want)
		}
	}

	leanGate := readResourcePolicyFile(t, root, "specs/lean-platform-gate.md")
	for _, want := range []string{
		"Token frugality is P0 and must be treated as a first-order gate",
		"- `token-frugality-p0`",
		"- `power-and-battery-aware-routing`",
		"increases token load, transcript replay, or verbose output without measurable",
		"removes or weakens powered-mode full power execution",
		"removes or weakens low-battery or thermal-aware execution",
	} {
		if !strings.Contains(leanGate, want) {
			t.Fatalf("lean platform gate must preserve resource policy gate %q", want)
		}
	}

	tierPolicy := readResourcePolicyFile(t, root, "specs/client-surfaces-and-free-tier.md")
	for _, want := range []string{
		"- free must be structurally token-frugal by default",
		"- free must show enough token-saving and context-reuse evidence",
		"- paid can run full power mode when powered",
		"- paid can enforce battery-conscious mode",
	} {
		if !strings.Contains(tierPolicy, want) {
			t.Fatalf("tier policy must preserve resource policy boundary %q", want)
		}
	}

	benchmark := readResourcePolicyFile(t, root, "specs/golden-competitive-benchmark.yaml")
	for _, want := range []string{
		"id: token-frugality-p0",
		"Token frugality is P0",
		"id: throttle-and-power-aware-routing",
		"Throttling avoidance, powered-mode full power execution, and",
		"low-battery or thermal-aware route selection are P1 route-policy",
	} {
		if !strings.Contains(benchmark, want) {
			t.Fatalf("golden benchmark must preserve resource pressure vector %q", want)
		}
	}
}

func readResourcePolicyFile(t *testing.T, root, rel string) string {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(root, rel))
	if err != nil {
		t.Fatalf("read %s: %v", rel, err)
	}
	return string(data)
}
