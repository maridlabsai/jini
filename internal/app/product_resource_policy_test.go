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
		"Token frugality is P0.",
		"Default to the cheapest safe route that can complete the task.",
		"Use local/offline routes when they meet the task quality bar.",
		"Escalate to stronger online routes when correctness, codebase scope, or",
		"Preserve enough session state to continue work without replaying stale chat.",
		"Avoiding throttling is P1.",
		"Power awareness is P1.",
		"throttle-aware route switching",
		"powered-mode and low-battery routing",
		"offline local-model quality regression harness",
		"cross-surface session handoff",
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
		"splits offline and online execution into separate transcripts",
		"- `offline-online-session-stitching`",
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
		"id: offline-online-session-stitching",
		"Offline work, online CLI work, local model execution, mobile review,",
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

	offlineStrategy := readResourcePolicyFile(t, root, "specs/platform-offline-strategy.md")
	for _, want := range []string{
		"### Guarantee 4a: Offline And Online Toggle Seamlessly",
		"local model work performed offline",
		"downstream CLI work resumed online",
		"managed-route recovery after throttling or provider limits",
		"configured CLI throttle state",
		"online CLI throttle level and quota pressure",
	} {
		if !strings.Contains(offlineStrategy, want) {
			t.Fatalf("offline strategy must preserve seamless toggle requirement %q", want)
		}
	}

	crossSurfacePRD := readResourcePolicyFile(t, root, "specs/cross-surface-session-platform-prd.md")
	for _, want := range []string{
		"Offline and online are route states, not separate work modes.",
		"cross-navigate between offline and online routes without restarting",
		"configured CLI throttle state",
		"device capability state",
		"battery or thermal posture",
	} {
		if !strings.Contains(crossSurfacePRD, want) {
			t.Fatalf("cross-surface PRD must preserve seamless navigation requirement %q", want)
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
