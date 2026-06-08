package app_test

import (
	"os"
	"path/filepath"
	"regexp"
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
		"intent/parity golden transcript gate blocks questions, bare entities, and",
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
		"intent-first-cli-parity",
		"direct-cwd-file-edit-fixture",
		"simple-question-compact-answer",
		"intent-first-routing-fixture",
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
		"intent-first-routing-fixture": {
			"id: cli-ux-regression-intent-routing",
			"kind: executable",
			`ref: "go test ./internal/app -run 'TestInteractiveMalformedCapitalQuestionCorrectsWithoutTravelFlow|TestInteractiveBareEntityAsksForIntentWithoutCreatingWork|TestInteractiveExplicitTripChoiceCanUseBareDestination|TestCurrentWorkMalformedCapitalQuestionCorrectsDirectly'"`,
		},
		"async-work-receipt-fixture": {
			"id: competitive-release-plan-async-work-receipt",
			"kind: named-proof",
			`ref: "specs/competitive-release-plan.md"`,
		},
		"offline-route-proof-fixture": {
			"id: publish-readiness-offline-regression",
			"kind: executable",
			`ref: "go test ./internal/app -run TestPublishReadinessIncludesOfflineRegressionGuardrails"`,
		},
		"adversarial-code-review-fixture": {
			"id: competitive-release-plan-code-review-quality",
			"kind: named-proof",
			`ref: "specs/competitive-release-plan.md"`,
		},
		"competitor-watch-refresh-fixture": {
			"id: competitive-release-plan-competitive-watch-packet",
			"kind: named-proof",
			`ref: "specs/competitive-release-plan.md"`,
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
		"Named-proof refs must resolve to existing repository files; executable refs",
		"must name real Go test functions.",
	} {
		if !strings.Contains(gateMatrix, want) {
			t.Fatalf("engineering gate matrix must document proof-backed outcome gates %q", want)
		}
	}

	productSettling := readRepoFile(t, root, "specs/product-settling-decisions.md")
	for _, want := range []string{
		"Outcome gates require executable or named proof references; names alone do not satisfy the scorecard.",
		"Each required outcome must point to a runnable check or a named proof surface that can be inspected later.",
		"Named-proof refs must resolve to existing repository files, and executable refs",
		"must name real Go test functions.",
	} {
		if !strings.Contains(productSettling, want) {
			t.Fatalf("product settling decisions must document proof-backed outcome gates %q", want)
		}
	}
}

func TestRequiredOutcomeGateProofReferencesResolve(t *testing.T) {
	root := repoRootForMigrationTest(t)
	goldenBenchmark := readRepoFile(t, root, "specs/golden-competitive-benchmark.yaml")
	goTests := goTestFunctions(t, root, "internal/app")

	for _, id := range []string{
		"direct-cwd-file-edit-fixture",
		"simple-question-compact-answer",
		"intent-first-routing-fixture",
		"async-work-receipt-fixture",
		"offline-route-proof-fixture",
		"adversarial-code-review-fixture",
		"competitor-watch-refresh-fixture",
		"commercial-tier-boundary-fixture",
		"cross-surface-continuity-fixture",
		"token-frugality-route-proof-fixture",
	} {
		block := requiredOutcomeGateBlock(t, goldenBenchmark, id)
		for _, proof := range proofReferencesFromOutcomeGateBlock(t, block) {
			switch proof.kind {
			case "named-proof":
				assertNamedProofRefResolves(t, root, id, proof.ref)
			case "executable":
				assertExecutableProofRefNamesRealGoTests(t, id, proof.ref, goTests)
			default:
				t.Fatalf("required outcome gate %q has unsupported proof kind %q in ref %q", id, proof.kind, proof.ref)
			}
		}
	}
}

type outcomeGateProofReference struct {
	kind string
	ref  string
}

func proofReferencesFromOutcomeGateBlock(t *testing.T, block string) []outcomeGateProofReference {
	t.Helper()

	var refs []outcomeGateProofReference
	inProofReferences := false
	for _, line := range strings.Split(block, "\n") {
		trimmed := strings.TrimSpace(line)
		if trimmed == "proof_references:" {
			inProofReferences = true
			continue
		}
		if !inProofReferences {
			continue
		}
		if strings.HasPrefix(trimmed, "gate:") {
			break
		}
		if strings.HasPrefix(trimmed, "- id: ") {
			refs = append(refs, outcomeGateProofReference{})
			continue
		}
		if len(refs) == 0 {
			continue
		}
		key, raw, ok := strings.Cut(trimmed, ":")
		if !ok {
			continue
		}
		switch strings.TrimSpace(key) {
		case "kind":
			refs[len(refs)-1].kind = unquoteYAMLScalar(raw)
		case "ref":
			refs[len(refs)-1].ref = unquoteYAMLScalar(raw)
		}
	}
	if len(refs) == 0 {
		t.Fatalf("required outcome gate block has no proof refs:\n%s", block)
	}
	for _, ref := range refs {
		if ref.kind == "" || ref.ref == "" {
			t.Fatalf("required outcome gate block has incomplete proof ref %#v:\n%s", ref, block)
		}
	}
	return refs
}

func unquoteYAMLScalar(raw string) string {
	value := strings.TrimSpace(raw)
	if len(value) < 2 {
		return value
	}
	first := value[0]
	last := value[len(value)-1]
	if (first == '"' && last == '"') || (first == '\'' && last == '\'') {
		return value[1 : len(value)-1]
	}
	return value
}

func assertNamedProofRefResolves(t *testing.T, root, gateID, ref string) {
	t.Helper()
	if filepath.IsAbs(ref) || strings.Contains(ref, "#") {
		t.Fatalf("required outcome gate %q named-proof ref must be a repository file path without a fragment: %q", gateID, ref)
	}
	info, err := os.Stat(filepath.Join(root, ref))
	if err != nil {
		t.Fatalf("required outcome gate %q named-proof ref must resolve to an existing file: %q: %v", gateID, ref, err)
	}
	if info.IsDir() {
		t.Fatalf("required outcome gate %q named-proof ref must resolve to a file, got directory: %q", gateID, ref)
	}
}

func assertExecutableProofRefNamesRealGoTests(t *testing.T, gateID, ref string, goTests map[string]bool) {
	t.Helper()
	runPattern := goTestRunPattern(t, gateID, ref)
	for _, testName := range strings.Split(runPattern, "|") {
		testName = strings.Trim(testName, "^$")
		if !strings.HasPrefix(testName, "Test") || !goTests[testName] {
			t.Fatalf("required outcome gate %q executable ref must name a real Go test function, got %q in %q", gateID, testName, ref)
		}
	}
}

func goTestRunPattern(t *testing.T, gateID, ref string) string {
	t.Helper()
	match := regexp.MustCompile(`(?:^|\s)-run\s+('([^']+)'|"([^"]+)"|([^\s]+))`).FindStringSubmatch(ref)
	if match == nil {
		t.Fatalf("required outcome gate %q executable ref must include go test -run pattern: %q", gateID, ref)
	}
	for _, group := range match[2:] {
		if group != "" {
			return group
		}
	}
	t.Fatalf("required outcome gate %q executable ref has empty go test -run pattern: %q", gateID, ref)
	return ""
}

func goTestFunctions(t *testing.T, root, rel string) map[string]bool {
	t.Helper()
	files, err := filepath.Glob(filepath.Join(root, rel, "*_test.go"))
	if err != nil {
		t.Fatalf("glob Go test files: %v", err)
	}
	functions := map[string]bool{}
	functionPattern := regexp.MustCompile(`(?m)^func (Test[A-Za-z0-9_]+)\s*\(`)
	for _, file := range files {
		data, err := os.ReadFile(file)
		if err != nil {
			t.Fatalf("read Go test file %s: %v", file, err)
		}
		for _, match := range functionPattern.FindAllStringSubmatch(string(data), -1) {
			functions[match[1]] = true
		}
	}
	if len(functions) == 0 {
		t.Fatal("expected at least one Go test function under internal/app")
	}
	return functions
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
		"TestInteractiveMalformedCapitalQuestionCorrectsWithoutTravelFlow",
		"TestInteractiveBareEntityAsksForIntentWithoutCreatingWork",
		"TestInteractiveExplicitTripChoiceCanUseBareDestination",
		"TestCurrentWorkMalformedCapitalQuestionCorrectsDirectly",
		"TestCurrentWorkUnknownStandaloneQuestionStaysCompact",
		"TestInteractiveLauncherHandlesUnsureInputWithUsefulPass",
		"TestLauncherStartsAsCompactShellWhenCurrentWorkExists",
		"TestCurrentWorkInteractiveLauncherIsCompactByDefault",
		"TestLauncherShowsOtherActiveWorkWhenMultipleProjectsExist",
		"TestInteractiveLauncherCanResumeNamedActiveProject",
		"TestPublicDocsUseCurrentFirstRunFlow",
		"TestP1SimplicityPriorityCoversCommandsSkillsAndAgents",
		"direct CLI edit, simple-question, and intent-first UX regression gate",
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
