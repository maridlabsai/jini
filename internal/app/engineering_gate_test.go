package app_test

import (
	"strings"
	"testing"
)

func TestRequiredCommitGateChecksStagedAndUnstagedWhitespace(t *testing.T) {
	root := repoRootForMigrationTest(t)

	requiredGates := readRepoFile(t, root, "tools/run_required_gates.sh")
	for _, want := range []string{
		"git diff --check",
		"git diff --cached --check",
		"run_scorecard_gate",
		"scorecard-gate --format json",
	} {
		if !strings.Contains(requiredGates, want) {
			t.Fatalf("required commit gate must run %q", want)
		}
	}

	gateMatrix := readRepoFile(t, root, "specs/engineering-gate-matrix.md")
	for _, want := range []string{
		"`git diff --check`",
		"`git diff --cached --check`",
		"`jini scorecard-gate --format json`",
		"staged and unstaged whitespace",
		"competitive scorecard drift is blocked before commit",
	} {
		if !strings.Contains(gateMatrix, want) {
			t.Fatalf("engineering gate matrix must document %q", want)
		}
	}
}

func TestScorecardGateIsDocumentedAsCommitGate(t *testing.T) {
	root := repoRootForMigrationTest(t)

	adminHelp := readRepoFile(t, root, "internal/app/app.go")
	for _, want := range []string{
		`case "scorecard-gate":`,
		"- jini scorecard-gate",
	} {
		if !strings.Contains(adminHelp, want) {
			t.Fatalf("native command surface must include %q", want)
		}
	}

	goldenBenchmark := readRepoFile(t, root, "specs/golden-competitive-benchmark.yaml")
	for _, want := range []string{
		"scorecard_gates:",
		"required_watchlist_competitors:",
		"required_pressure_vectors:",
		"async-background-agents",
		"cross-surface-session-continuity",
		"offline-online-session-stitching",
		"local-open-model-optionality",
		"token-frugality-p0",
		"throttle-and-power-aware-routing",
		"commit-gated-scorecard-drift",
	} {
		if !strings.Contains(goldenBenchmark, want) {
			t.Fatalf("golden benchmark must document scorecard gate item %q", want)
		}
	}
}

func TestEngineeringGateMatrixDoesNotListRequiredGatesAsPromotionCandidates(t *testing.T) {
	root := repoRootForMigrationTest(t)

	gateMatrix := readRepoFile(t, root, "specs/engineering-gate-matrix.md")
	promotionCandidateIndex := strings.Index(gateMatrix, "## Promotion Candidates")
	if promotionCandidateIndex == -1 {
		t.Fatal("engineering gate matrix must keep an explicit promotion-candidates section")
	}
	promotionCandidates := gateMatrix[promotionCandidateIndex:]
	for _, alreadyRequired := range []string{
		"go test ./...",
		"git diff --check",
		"git diff --cached --check",
		"security_configuration_gate.sh",
		"scorecard-gate",
	} {
		if strings.Contains(promotionCandidates, alreadyRequired) {
			t.Fatalf("promotion candidates must not relist required gate %q:\n%s", alreadyRequired, promotionCandidates)
		}
	}
}

func TestEngineeringGateMatrixDocumentsMakefileAliases(t *testing.T) {
	root := repoRootForMigrationTest(t)

	makefile := readRepoFile(t, root, "Makefile")
	for _, want := range []string{
		"gates-commit:\n\tbash tools/run_required_gates.sh commit",
		"gates-push:\n\tbash tools/run_required_gates.sh push",
		"gates-release:\n\tbash tools/run_required_gates.sh release",
	} {
		if !strings.Contains(makefile, want) {
			t.Fatalf("Makefile must wire canonical gate alias:\n%s", want)
		}
	}

	gateMatrix := readRepoFile(t, root, "specs/engineering-gate-matrix.md")
	for _, want := range []string{
		"`make gates-commit`",
		"`make gates-push`",
		"`make gates-release`",
	} {
		if !strings.Contains(gateMatrix, want) {
			t.Fatalf("engineering gate matrix must document Makefile alias %q", want)
		}
	}
}
