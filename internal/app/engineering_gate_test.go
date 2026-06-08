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
		"run_product_prd_drift_gate",
		"tools/product_prd_drift_gate.sh",
		"run_cli_ux_regression_gate",
		"tools/cli_ux_regression_gate.sh",
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
		"`bash tools/product_prd_drift_gate.sh`",
		"`bash tools/cli_ux_regression_gate.sh`",
		"`jini scorecard-gate --format json`",
		"staged and unstaged whitespace",
		"protected PRD and product-positioning surfaces cannot drift",
		"direct CLI edit and simple-question flows cannot regress into draft/status frames",
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
		"required_outcome_gates:",
		"async-background-agents",
		"cross-surface-session-continuity",
		"offline-online-session-stitching",
		"local-open-model-optionality",
		"token-frugality-p0",
		"throttle-and-power-aware-routing",
		"commit-gated-scorecard-drift",
		"direct-cwd-file-edit-fixture",
		"simple-question-compact-answer",
		"async-work-receipt-fixture",
		"offline-route-proof-fixture",
		"adversarial-code-review-fixture",
		"competitor-watch-refresh-fixture",
		"commercial-tier-boundary-fixture",
		"cross-surface-continuity-fixture",
		"token-frugality-route-proof-fixture",
		"opencode",
		"sourcegraph-amp",
		"tabnine-agent",
		"qodo-merge",
		"ellipsis",
		"langgraph",
		"openai-agents-sdk",
		"pydantic-ai",
		"crewai",
	} {
		if !strings.Contains(goldenBenchmark, want) {
			t.Fatalf("golden benchmark must document scorecard gate item %q", want)
		}
	}
}

func TestRequiredOutcomeGatesDeclareProofReferences(t *testing.T) {
	root := repoRootForMigrationTest(t)

	goldenBenchmark := readRepoFile(t, root, "specs/golden-competitive-benchmark.yaml")
	requiredProofs := map[string][]string{
		"direct-cwd-file-edit-fixture": {
			"id: cli-ux-regression-direct-edit",
			"kind: executable",
			`ref: "go test ./internal/app -run TestInteractiveLocalTextEditAppendsQuotedLineInsteadOfDrafting"`,
		},
		"simple-question-compact-answer": {
			"id: cli-ux-regression-simple-question",
			"kind: executable",
			`ref: "go test ./internal/app -run 'TestCurrentWorkSimpleFactualQuestionAnswersDirectly|TestDirectArgsSimpleFactualQuestionAnswersDirectly|TestInteractiveSimpleFactualQuestionAnswersDirectlyWithoutCurrentWork'"`,
		},
		"async-work-receipt-fixture": {
			"id: competitive-release-plan-async-work-receipt",
			"kind: named-proof",
			`ref: "specs/competitive-release-plan.md#async-work-receipt"`,
		},
		"offline-route-proof-fixture": {
			"id: publish-readiness-offline-regression",
			"kind: executable",
			`ref: "go test ./internal/app -run TestPublishReadinessIncludesOfflineRegressionGuardrails"`,
		},
		"adversarial-code-review-fixture": {
			"id: competitive-release-plan-code-review-quality",
			"kind: named-proof",
			`ref: "specs/competitive-release-plan.md#code-review-quality-is-now-a-product-bar"`,
		},
		"competitor-watch-refresh-fixture": {
			"id: competitive-release-plan-competitive-watch-packet",
			"kind: named-proof",
			`ref: "specs/competitive-release-plan.md#competitive-watch-packet"`,
		},
		"commercial-tier-boundary-fixture": {
			"id: product-simplicity-tier-boundary",
			"kind: executable",
			`ref: "go test ./internal/app -run TestP1SimplicityPriorityCoversCommandsSkillsAndAgents"`,
		},
		"cross-surface-continuity-fixture": {
			"id: product-resource-cross-surface-continuity",
			"kind: executable",
			`ref: "go test ./internal/app -run TestResourcePolicyPrioritiesAreGated"`,
		},
		"token-frugality-route-proof-fixture": {
			"id: product-resource-token-frugality",
			"kind: executable",
			`ref: "go test ./internal/app -run TestResourcePolicyPrioritiesAreGated"`,
		},
	}
	for id, required := range requiredProofs {
		block := requiredOutcomeGateBlock(t, goldenBenchmark, id)
		if !strings.Contains(block, "proof_references:") {
			t.Fatalf("required outcome gate %q must declare proof_references:\n%s", id, block)
		}
		for _, want := range required {
			if !strings.Contains(block, want) {
				t.Fatalf("required outcome gate %q must include proof reference %q:\n%s", id, want, block)
			}
		}
	}

	gateMatrix := readRepoFile(t, root, "specs/engineering-gate-matrix.md")
	for _, want := range []string{
		"Outcome gates require executable or named proof references, not just competitor or fixture names.",
		"A gate name without a runnable command or named proof reference is planning prose, not evidence.",
	} {
		if !strings.Contains(gateMatrix, want) {
			t.Fatalf("engineering gate matrix must document proof-backed outcome gates %q", want)
		}
	}

	productSettling := readRepoFile(t, root, "specs/product-settling-decisions.md")
	for _, want := range []string{
		"Outcome gates require executable or named proof references; names alone do not satisfy the scorecard.",
		"Each required outcome must point to a runnable check or a named proof surface that can be inspected later.",
	} {
		if !strings.Contains(productSettling, want) {
			t.Fatalf("product settling decisions must document proof-backed outcome gates %q", want)
		}
	}
}

func requiredOutcomeGateBlock(t *testing.T, benchmark, id string) string {
	t.Helper()
	marker := "    - id: " + id
	start := strings.Index(benchmark, marker)
	if start == -1 {
		t.Fatalf("required outcome gate %q is missing", id)
	}
	rest := benchmark[start+len(marker):]
	end := strings.Index(rest, "\n    - id: ")
	if end == -1 {
		end = strings.Index(rest, "\nmethodology:")
	}
	if end == -1 {
		return benchmark[start:]
	}
	return benchmark[start : start+len(marker)+end]
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
		"product_prd_drift_gate.sh",
		"cli_ux_regression_gate.sh",
		"scorecard-gate",
	} {
		if strings.Contains(promotionCandidates, alreadyRequired) {
			t.Fatalf("promotion candidates must not relist required gate %q:\n%s", alreadyRequired, promotionCandidates)
		}
	}
}

func TestPushGateRunsShipCheckEvidence(t *testing.T) {
	root := repoRootForMigrationTest(t)

	requiredGates := readRepoFile(t, root, "tools/run_required_gates.sh")
	for _, want := range []string{
		"run_ship_check_gate",
		"check ship --format json",
	} {
		if !strings.Contains(requiredGates, want) {
			t.Fatalf("push gate must run ship-check evidence %q", want)
		}
	}

	gateMatrix := readRepoFile(t, root, "specs/engineering-gate-matrix.md")
	for _, want := range []string{
		"`jini check ship --format json`",
		"push gate records local shipping evidence",
		"dirty worktrees are blocked before push",
	} {
		if !strings.Contains(gateMatrix, want) {
			t.Fatalf("engineering gate matrix must document ship-check evidence %q", want)
		}
	}
}

func TestCLIUXRegressionGatePinsIncidentScenarios(t *testing.T) {
	root := repoRootForMigrationTest(t)

	gate := readRepoFile(t, root, "tools/cli_ux_regression_gate.sh")
	for _, want := range []string{
		"TestInteractiveLocalTextEditAppendsQuotedLineInsteadOfDrafting",
		"TestCurrentWorkLocalTextEditExecutesWithoutStartPrompt",
		"TestCurrentWorkSimpleFactualQuestionAnswersDirectly",
		"TestCurrentWorkCapitalQuestionAcceptsNaturalPhrasing",
		"TestDirectArgsSimpleFactualQuestionAnswersDirectly",
		"TestInteractiveSimpleFactualQuestionAnswersDirectlyWithoutCurrentWork",
		"TestCurrentWorkUnknownStandaloneQuestionStaysCompact",
		"TestInteractiveLauncherHandlesUnsureInputWithUsefulPass",
		"TestLauncherStartsAsCompactShellWhenCurrentWorkExists",
		"TestCurrentWorkInteractiveLauncherIsCompactByDefault",
		"TestLauncherShowsOtherActiveWorkWhenMultipleProjectsExist",
		"TestInteractiveLauncherCanResumeNamedActiveProject",
		"TestPublicDocsUseCurrentFirstRunFlow",
		"TestP1SimplicityPriorityCoversCommandsSkillsAndAgents",
		"direct CLI edit and simple-question UX regression gate",
	} {
		if !strings.Contains(gate, want) {
			t.Fatalf("CLI UX regression gate must pin %q", want)
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

func TestProductPRDDriftGateProtectsCanonicalSurfaces(t *testing.T) {
	root := repoRootForMigrationTest(t)

	gate := readRepoFile(t, root, "tools/product_prd_drift_gate.sh")
	for _, want := range []string{
		"SETTLING_DOC=\"specs/product-settling-decisions.md\"",
		"git diff --name-only",
		"git diff --cached --name-only",
		"git ls-files --others --exclude-standard",
		"README.md",
		"specs/number-one-platform-prd.md",
		"specs/product-settling-decisions.md",
		"specs/launcher-intake-design.md",
		"specs/number-one-development-plan.md",
		"specs/client-surfaces-and-free-tier.md",
		"specs/platform-offline-strategy.md",
		"specs/skills-and-delegation-slice.md",
		"specs/competitive-release-plan.md",
		"specs/travel-curated-experience-framework.md",
		"Product PRD drift gate failed.",
	} {
		if !strings.Contains(gate, want) {
			t.Fatalf("product PRD drift gate must contain %q", want)
		}
	}
}

func TestProductPRDDriftGateAvoidsEmptyArrayExpansion(t *testing.T) {
	root := repoRootForMigrationTest(t)

	gate := readRepoFile(t, root, "tools/product_prd_drift_gate.sh")
	if strings.Contains(gate, "changed=()") || strings.Contains(gate, `"${changed[@]}"`) {
		t.Fatal("product PRD drift gate must stream changed files so a clean worktree does not fail under set -u")
	}
	for _, want := range []string{
		"while IFS= read -r path; do",
		"done < <(changed_files)",
	} {
		if !strings.Contains(gate, want) {
			t.Fatalf("product PRD drift gate must keep streaming changed-file loop %q", want)
		}
	}
}
