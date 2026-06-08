package app

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
	"unicode"
)

const scorecardBenchmarkPath = "specs/golden-competitive-benchmark.yaml"

var defaultScorecardGatePolicy = scorecardGatePolicy{
	BenchmarkPath: scorecardBenchmarkPath,
	RequiredCompetitors: []string{
		"claude-code",
		"codex",
		"github-copilot-coding-agent",
		"google-jules",
		"cursor",
		"windsurf",
		"cline",
		"aider",
		"continue",
		"devin",
		"replit-agent",
		"opencode",
		"google-gemini-cli",
		"roo-code",
		"ollama",
		"lm-studio",
		"litellm",
		"openrouter",
		"sourcegraph-amp",
		"tabnine-agent",
		"qodo-merge",
		"ellipsis",
		"langgraph",
		"openai-agents-sdk",
		"pydantic-ai",
		"crewai",
	},
	RequiredPressureVectors: []string{
		"async-background-agents",
		"cross-surface-session-continuity",
		"offline-online-session-stitching",
		"transparent-progress-and-outputs",
		"permissioned-sandbox-execution",
		"skills-hooks-and-context-routing",
		"local-open-model-optionality",
		"token-frugality-p0",
		"throttle-and-power-aware-routing",
		"commit-gated-scorecard-drift",
		"intent-first-cli-parity",
		"familiar-tool-adapter-breadth",
	},
	RequiredOutcomeGates: []string{
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
	},
	MinimumCoreCompetitors:      7,
	MinimumWatchlistCompetitors: 40,
	MinimumScenarios:            8,
}

type scorecardGatePolicy struct {
	BenchmarkPath               string
	RequiredCompetitors         []string
	RequiredPressureVectors     []string
	RequiredOutcomeGates        []string
	MinimumCoreCompetitors      int
	MinimumWatchlistCompetitors int
	MinimumScenarios            int
}

type scorecardGateReport struct {
	SchemaVersion            string                    `json:"schema_version"`
	ResultType               string                    `json:"result_type"`
	GeneratedAt              string                    `json:"generated_at"`
	Status                   string                    `json:"status"`
	ScorecardPath            string                    `json:"scorecard_path"`
	CoreCompetitorCount      int                       `json:"core_competitor_count"`
	WatchlistCompetitorCount int                       `json:"watchlist_competitor_count"`
	ScenarioCount            int                       `json:"scenario_count"`
	RequiredCompetitors      []scorecardPresenceCheck  `json:"required_competitors"`
	PressureVectors          []scorecardPresenceCheck  `json:"pressure_vectors"`
	OutcomeGates             []scorecardPresenceCheck  `json:"outcome_gates"`
	Checks                   []scorecardThresholdCheck `json:"checks"`
}

type scorecardPresenceCheck struct {
	ID      string   `json:"id"`
	Present bool     `json:"present"`
	Status  string   `json:"status"`
	Reasons []string `json:"reasons,omitempty"`
}

type scorecardThresholdCheck struct {
	ID      string `json:"id"`
	Actual  int    `json:"actual"`
	Minimum int    `json:"minimum"`
	Status  string `json:"status"`
}

type scorecardPresenceResult struct {
	Present bool
	Reasons []string
}

type scorecardGateBuilder struct {
	root   string
	policy scorecardGatePolicy
}

func runScorecardGate(args []string, stdout, stderr io.Writer) int {
	format, ok := parseOptionalFormatArgs(args)
	if !ok {
		fmt.Fprintln(stderr, "Unsupported scorecard-gate format. Try `jini scorecard-gate` or `jini scorecard-gate --format json`.")
		return 1
	}

	root := discoverSourceRoot()
	report := buildScorecardGateReport(root)
	if format == "json" {
		encoder := json.NewEncoder(stdout)
		encoder.SetIndent("", "  ")
		encoder.SetEscapeHTML(false)
		if err := encoder.Encode(report); err != nil {
			fmt.Fprintf(stderr, "Could not render scorecard gate report: %v\n", err)
			return 1
		}
		if report.Status == "ok" {
			return 0
		}
		return 1
	}

	renderScorecardGateText(stdout, report)
	if report.Status == "ok" {
		return 0
	}
	return 1
}

func buildScorecardGateReport(root string) scorecardGateReport {
	return newScorecardGateBuilder(root, defaultScorecardGatePolicy).Build()
}

func newScorecardGateBuilder(root string, policy scorecardGatePolicy) scorecardGateBuilder {
	return scorecardGateBuilder{
		root:   root,
		policy: policy,
	}
}

func (builder scorecardGateBuilder) Build() scorecardGateReport {
	report := scorecardGateReport{
		SchemaVersion: "0.1.0",
		ResultType:    "JiniScorecardGate",
		GeneratedAt:   time.Now().UTC().Format(time.RFC3339),
		Status:        "ok",
		ScorecardPath: builder.policy.BenchmarkPath,
	}

	text, ok := builder.readBenchmark()
	if !ok {
		report.Status = "needs-attention"
		report.Checks = append(report.Checks, scorecardThresholdCheck{
			ID:     "source-scorecard-readable",
			Status: "missing",
		})
		return report
	}

	scorecardGateSection := yamlSection(text, "scorecard_gates")
	normalizedBenchmark := normalizedYAMLIDs(text)
	normalizedScorecardGate := normalizedYAMLIDs(scorecardGateSection)
	outcomeGateEvidence := builder.outcomeGateEvidenceReferences(scorecardGateSection)

	builder.addThresholdChecks(&report, text, scorecardGateSection)
	report.RequiredCompetitors = builder.presenceChecks(builder.policy.RequiredCompetitors, func(id string) bool {
		return normalizedBenchmark[id] && normalizedScorecardGate[id]
	})
	report.PressureVectors = builder.presenceChecks(builder.policy.RequiredPressureVectors, func(id string) bool {
		return normalizedScorecardGate[id]
	})
	report.OutcomeGates = builder.presenceChecksWithReasons(builder.policy.RequiredOutcomeGates, func(id string) scorecardPresenceResult {
		if !normalizedScorecardGate[id] {
			return scorecardPresenceResult{
				Present: false,
				Reasons: []string{
					"required outcome gate is missing from scorecard_gates.required_outcome_gates",
				},
			}
		}
		evidence, ok := outcomeGateEvidence[id]
		if !ok {
			return scorecardPresenceResult{
				Present: false,
				Reasons: []string{
					"required outcome gate has no valid proof_references",
				},
			}
		}
		return evidence
	})

	if !scorecardChecksPass(report) {
		report.Status = "needs-attention"
	}
	return report
}

func (builder scorecardGateBuilder) readBenchmark() (string, bool) {
	if builder.root == "" {
		return "", false
	}
	data, err := os.ReadFile(filepath.Join(builder.root, builder.policy.BenchmarkPath))
	if err != nil {
		return "", false
	}
	return string(data), true
}

func (builder scorecardGateBuilder) addThresholdChecks(report *scorecardGateReport, text, scorecardGateSection string) {
	report.CoreCompetitorCount = countYAMLListItems(yamlSection(text, "core_benchmark_set"))
	report.WatchlistCompetitorCount = countYAMLListItems(yamlSection(text, "watchlist"))
	report.ScenarioCount = countYAMLListItems(yamlSection(text, "scenarios"))
	report.Checks = []scorecardThresholdCheck{
		buildScorecardThresholdCheck("minimum-core-competitors", report.CoreCompetitorCount, readScorecardMinimum(scorecardGateSection, "minimum_core_competitors", builder.policy.MinimumCoreCompetitors)),
		buildScorecardThresholdCheck("minimum-watchlist-competitors", report.WatchlistCompetitorCount, readScorecardMinimum(scorecardGateSection, "minimum_watchlist_competitors", builder.policy.MinimumWatchlistCompetitors)),
		buildScorecardThresholdCheck("minimum-scenarios", report.ScenarioCount, readScorecardMinimum(scorecardGateSection, "minimum_scenarios", builder.policy.MinimumScenarios)),
	}
}

func (builder scorecardGateBuilder) presenceChecks(ids []string, present func(string) bool) []scorecardPresenceCheck {
	checks := make([]scorecardPresenceCheck, 0, len(ids))
	for _, id := range ids {
		isPresent := present(id)
		checks = append(checks, scorecardPresenceCheck{
			ID:      id,
			Present: isPresent,
			Status:  scorecardPresenceStatus(isPresent),
		})
	}
	return checks
}

func (builder scorecardGateBuilder) presenceChecksWithReasons(ids []string, present func(string) scorecardPresenceResult) []scorecardPresenceCheck {
	checks := make([]scorecardPresenceCheck, 0, len(ids))
	for _, id := range ids {
		result := present(id)
		checks = append(checks, scorecardPresenceCheck{
			ID:      id,
			Present: result.Present,
			Status:  scorecardPresenceStatus(result.Present),
			Reasons: append([]string{}, result.Reasons...),
		})
	}
	return checks
}

func buildScorecardThresholdCheck(id string, actual, minimum int) scorecardThresholdCheck {
	status := "ok"
	if actual < minimum {
		status = "needs-attention"
	}
	return scorecardThresholdCheck{
		ID:      id,
		Actual:  actual,
		Minimum: minimum,
		Status:  status,
	}
}

func scorecardChecksPass(report scorecardGateReport) bool {
	for _, check := range report.Checks {
		if check.Status != "ok" {
			return false
		}
	}
	for _, competitor := range report.RequiredCompetitors {
		if !competitor.Present {
			return false
		}
	}
	for _, vector := range report.PressureVectors {
		if !vector.Present {
			return false
		}
	}
	for _, gate := range report.OutcomeGates {
		if !gate.Present {
			return false
		}
	}
	return true
}

func scorecardPresenceStatus(present bool) string {
	if present {
		return "ok"
	}
	return "missing"
}

func renderScorecardGateText(stdout io.Writer, report scorecardGateReport) {
	fmt.Fprintf(stdout, "STATUS %s\n", report.Status)
	fmt.Fprintln(stdout, "COUNTS")
	fmt.Fprintf(stdout, "  CORE_COMPETITORS %d\n", report.CoreCompetitorCount)
	fmt.Fprintf(stdout, "  WATCHLIST_COMPETITORS %d\n", report.WatchlistCompetitorCount)
	fmt.Fprintf(stdout, "  SCENARIOS %d\n", report.ScenarioCount)
	fmt.Fprintln(stdout, "COMPETITORS")
	for _, competitor := range report.RequiredCompetitors {
		fmt.Fprintf(stdout, "  %s %s\n", strings.ToUpper(competitor.Status), competitor.ID)
		renderScorecardPresenceReasons(stdout, competitor)
	}
	fmt.Fprintln(stdout, "PRESSURE VECTORS")
	for _, vector := range report.PressureVectors {
		fmt.Fprintf(stdout, "  %s %s\n", strings.ToUpper(vector.Status), vector.ID)
		renderScorecardPresenceReasons(stdout, vector)
	}
	fmt.Fprintln(stdout, "OUTCOME GATES")
	for _, gate := range report.OutcomeGates {
		fmt.Fprintf(stdout, "  %s %s\n", strings.ToUpper(gate.Status), gate.ID)
		renderScorecardPresenceReasons(stdout, gate)
	}
	fmt.Fprintln(stdout, "THRESHOLDS")
	for _, check := range report.Checks {
		fmt.Fprintf(stdout, "  %s %s %d/%d\n", strings.ToUpper(check.Status), check.ID, check.Actual, check.Minimum)
	}
}

func renderScorecardPresenceReasons(stdout io.Writer, check scorecardPresenceCheck) {
	for _, reason := range check.Reasons {
		fmt.Fprintf(stdout, "    REASON %s\n", reason)
	}
}

func readScorecardMinimum(section, key string, fallback int) int {
	for _, line := range strings.Split(section, "\n") {
		trimmed := strings.TrimSpace(line)
		if !strings.HasPrefix(trimmed, key+":") {
			continue
		}
		raw := strings.TrimSpace(strings.TrimPrefix(trimmed, key+":"))
		if value, err := strconv.Atoi(raw); err == nil {
			return value
		}
	}
	return fallback
}

type scorecardProofReference struct {
	ID   string
	Kind string
	Ref  string
}

func (builder scorecardGateBuilder) outcomeGateEvidenceReferences(scorecardGateSection string) map[string]scorecardPresenceResult {
	outcomeGateSection := yamlSection(scorecardGateSection, "required_outcome_gates")
	evidenceByID := map[string]scorecardPresenceResult{}
	for _, block := range yamlListItemBlocks(outcomeGateSection) {
		id := yamlBlockID(block)
		if id == "" {
			continue
		}
		evidenceByID[id] = builder.yamlBlockEvidenceResult(block)
	}
	return evidenceByID
}

func yamlListItemBlocks(section string) []string {
	var blocks []string
	var current []string
	itemIndent := -1
	for _, line := range strings.Split(section, "\n") {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "- ") {
			indent := leadingSpaceCount(line)
			if itemIndent == -1 {
				itemIndent = indent
			}
			if indent == itemIndent {
				if len(current) > 0 {
					blocks = append(blocks, strings.Join(current, "\n"))
				}
				current = []string{line}
				continue
			}
		}
		if len(current) > 0 {
			current = append(current, line)
		}
	}
	if len(current) > 0 {
		blocks = append(blocks, strings.Join(current, "\n"))
	}
	return blocks
}

func yamlBlockID(block string) string {
	for _, line := range strings.Split(block, "\n") {
		value := strings.TrimSpace(line)
		if strings.HasPrefix(value, "- ") {
			value = strings.TrimSpace(strings.TrimPrefix(value, "- "))
		}
		key, raw, ok := splitYAMLKeyValue(value)
		if !ok || key != "id" {
			continue
		}
		return normalizeScorecardID(strings.Trim(raw, `"'[]`))
	}
	return ""
}

func splitYAMLKeyValue(value string) (string, string, bool) {
	key, raw, ok := strings.Cut(value, ":")
	if !ok {
		return "", "", false
	}
	key = strings.TrimSpace(key)
	if key == "" || strings.ContainsAny(key, " []{}") {
		return "", "", false
	}
	return key, strings.TrimSpace(raw), true
}

func isOutcomeGateEvidenceKey(key string) bool {
	normalized := normalizeScorecardID(key)
	return strings.Contains(normalized, "evidence") ||
		strings.Contains(normalized, "proof") ||
		strings.Contains(normalized, "reference")
}

func (builder scorecardGateBuilder) yamlBlockHasEvidenceReference(block string) bool {
	return builder.yamlBlockEvidenceResult(block).Present
}

func (builder scorecardGateBuilder) yamlBlockEvidenceResult(block string) scorecardPresenceResult {
	references := yamlBlockProofReferences(block)
	if len(references) == 0 {
		return scorecardPresenceResult{
			Present: false,
			Reasons: []string{
				"missing proof_references",
			},
		}
	}
	var reasons []string
	for _, reference := range references {
		if reason := builder.scorecardProofReferenceFailureReason(reference); reason != "" {
			reasons = append(reasons, reason)
		}
	}
	if len(reasons) > 0 {
		return scorecardPresenceResult{
			Present: false,
			Reasons: reasons,
		}
	}
	return scorecardPresenceResult{Present: true}
}

func yamlBlockProofReferences(block string) []scorecardProofReference {
	var references []scorecardProofReference
	proofReferencesSection := nestedYAMLSection(block, "proof_references")
	for _, proofBlock := range yamlListItemBlocks(proofReferencesSection) {
		references = append(references, scorecardProofReference{
			ID:   yamlBlockScalarValue(proofBlock, "id"),
			Kind: yamlBlockScalarValue(proofBlock, "kind"),
			Ref:  yamlBlockScalarValue(proofBlock, "ref"),
		})
	}
	for _, line := range strings.Split(block, "\n") {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			continue
		}
		if strings.HasPrefix(trimmed, "- ") {
			trimmed = strings.TrimSpace(strings.TrimPrefix(trimmed, "- "))
		}
		key, raw, ok := splitYAMLKeyValue(trimmed)
		if !ok || normalizeScorecardID(key) == "proof-references" || !isOutcomeGateEvidenceKey(key) {
			continue
		}
		if hasScorecardReferenceValue(raw) {
			references = append(references, scorecardProofReference{Ref: raw})
		}
	}
	return references
}

func yamlBlockScalarValue(block, wantKey string) string {
	for _, line := range strings.Split(block, "\n") {
		value := strings.TrimSpace(line)
		if strings.HasPrefix(value, "- ") {
			value = strings.TrimSpace(strings.TrimPrefix(value, "- "))
		}
		key, raw, ok := splitYAMLKeyValue(value)
		if ok && key == wantKey {
			return scorecardScalarValue(raw)
		}
	}
	return ""
}

func nestedYAMLSection(text, key string) string {
	lines := strings.Split(text, "\n")
	start := -1
	for i, line := range lines {
		if strings.TrimSpace(line) == key+":" {
			start = i
			break
		}
	}
	if start == -1 {
		return ""
	}
	baseIndent := leadingSpaceCount(lines[start])
	end := len(lines)
	for i := start + 1; i < len(lines); i++ {
		trimmed := strings.TrimSpace(lines[i])
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			continue
		}
		if leadingSpaceCount(lines[i]) <= baseIndent {
			end = i
			break
		}
	}
	return strings.Join(lines[start:end], "\n")
}

func (builder scorecardGateBuilder) scorecardProofReferenceExists(reference scorecardProofReference) bool {
	return builder.scorecardProofReferenceFailureReason(reference) == ""
}

func (builder scorecardGateBuilder) scorecardProofReferenceFailureReason(reference scorecardProofReference) string {
	ref := scorecardScalarValue(reference.Ref)
	label := scorecardProofReferenceLabel(reference)
	if !hasScorecardReferenceValue(ref) {
		return fmt.Sprintf("proof reference %s is missing ref", label)
	}
	kind := normalizeScorecardID(reference.Kind)
	switch kind {
	case "named-proof":
		return builder.namedProofReferenceFailureReason(label, ref)
	case "executable":
		return builder.executableProofReferenceFailureReason(label, ref)
	case "":
		if looksLikeNamedProofReference(ref) {
			return builder.namedProofReferenceFailureReason(label, ref)
		}
		if isInternalAppGoTestRunReference(ref) {
			return builder.executableProofReferenceFailureReason(label, ref)
		}
		return fmt.Sprintf("proof reference %s has no supported kind or recognizable reference", label)
	default:
		return fmt.Sprintf("proof reference %s has unsupported kind %q", label, scorecardScalarValue(reference.Kind))
	}
}

func scorecardProofReferenceLabel(reference scorecardProofReference) string {
	if id := scorecardScalarValue(reference.ID); id != "" {
		return id
	}
	if ref := scorecardScalarValue(reference.Ref); ref != "" {
		return ref
	}
	if kind := scorecardScalarValue(reference.Kind); kind != "" {
		return kind + " proof"
	}
	return "unnamed proof reference"
}

func (builder scorecardGateBuilder) namedProofReferenceExists(ref string) bool {
	return builder.namedProofReferenceFailureReason(scorecardScalarValue(ref), ref) == ""
}

func (builder scorecardGateBuilder) namedProofReferenceFailureReason(label, ref string) string {
	if builder.root == "" {
		return fmt.Sprintf("named-proof %s cannot resolve without source root", label)
	}
	path, ok := scorecardRepoFilePath(ref)
	if !ok {
		return fmt.Sprintf("named-proof %s has invalid repository file ref %q", label, ref)
	}
	info, err := os.Stat(filepath.Join(builder.root, path))
	if err != nil {
		return fmt.Sprintf("named-proof %s references missing file %s", label, path)
	}
	if info.IsDir() {
		return fmt.Sprintf("named-proof %s references directory %s", label, path)
	}
	return ""
}

func scorecardRepoFilePath(ref string) (string, bool) {
	path, _, _ := strings.Cut(scorecardScalarValue(ref), "#")
	path = strings.TrimSpace(path)
	if path == "" || filepath.IsAbs(path) || strings.Contains(path, "://") {
		return "", false
	}
	cleaned := filepath.Clean(path)
	if cleaned == "." || cleaned == ".." || strings.HasPrefix(cleaned, ".."+string(filepath.Separator)) {
		return "", false
	}
	return cleaned, true
}

func looksLikeNamedProofReference(ref string) bool {
	path, anchor, hasAnchor := strings.Cut(scorecardScalarValue(ref), "#")
	return hasAnchor && strings.TrimSpace(path) != "" && strings.TrimSpace(anchor) != ""
}

func (builder scorecardGateBuilder) executableProofReferenceExists(ref string) bool {
	return builder.executableProofReferenceFailureReason(scorecardScalarValue(ref), ref) == ""
}

func (builder scorecardGateBuilder) executableProofReferenceFailureReason(label, ref string) string {
	pattern, ok := internalAppGoTestRunPattern(ref)
	if !ok {
		return fmt.Sprintf("executable proof %s must use go test ./internal/app -run ...", label)
	}
	testNames := scorecardTestNamesFromRunPattern(pattern)
	if len(testNames) == 0 {
		return fmt.Sprintf("executable proof %s does not name any Test... functions", label)
	}
	for _, testName := range testNames {
		if !builder.scorecardTestFunctionExists(testName) {
			return fmt.Sprintf("executable proof %s names missing Go test function %s", label, testName)
		}
	}
	return ""
}

func isInternalAppGoTestRunReference(ref string) bool {
	_, ok := internalAppGoTestRunPattern(ref)
	return ok
}

func internalAppGoTestRunPattern(ref string) (string, bool) {
	fields := strings.Fields(scorecardScalarValue(ref))
	if len(fields) < 4 || fields[0] != "go" || fields[1] != "test" || fields[2] != "./internal/app" {
		return "", false
	}
	for i := 3; i < len(fields); i++ {
		if fields[i] == "-run" && i+1 < len(fields) {
			return scorecardScalarValue(fields[i+1]), true
		}
		if strings.HasPrefix(fields[i], "-run=") {
			return scorecardScalarValue(strings.TrimPrefix(fields[i], "-run=")), true
		}
	}
	return "", false
}

func scorecardTestNamesFromRunPattern(pattern string) []string {
	pattern = strings.Trim(scorecardScalarValue(pattern), "^$")
	parts := strings.Split(pattern, "|")
	names := make([]string, 0, len(parts))
	for _, part := range parts {
		name := strings.Trim(strings.TrimSpace(part), "()^$")
		if beforeSlash, _, ok := strings.Cut(name, "/"); ok {
			name = beforeSlash
		}
		if isGoTestFunctionName(name) {
			names = append(names, name)
		}
	}
	return names
}

func isGoTestFunctionName(name string) bool {
	if !strings.HasPrefix(name, "Test") || len(name) == len("Test") {
		return false
	}
	for _, r := range name {
		if r != '_' && !unicode.IsLetter(r) && !unicode.IsDigit(r) {
			return false
		}
	}
	return true
}

func (builder scorecardGateBuilder) scorecardTestFunctionExists(testName string) bool {
	if builder.root == "" {
		return false
	}
	appDir := filepath.Join(builder.root, "internal", "app")
	found := false
	_ = filepath.WalkDir(appDir, func(path string, entry os.DirEntry, err error) error {
		if err != nil || found || entry.IsDir() || filepath.Ext(path) != ".go" {
			return nil
		}
		data, readErr := os.ReadFile(path)
		if readErr != nil {
			return nil
		}
		if strings.Contains(string(data), "func "+testName+"(") {
			found = true
		}
		return nil
	})
	return found
}

func hasScorecardReferenceValue(value string) bool {
	value = scorecardScalarValue(value)
	switch strings.ToLower(value) {
	case "", "[]", "{}", ">", "|", "null", "none", "n/a":
		return false
	default:
		return true
	}
}

func scorecardScalarValue(value string) string {
	return strings.TrimSpace(strings.Trim(strings.TrimSpace(value), `"'`))
}

func yamlSection(text, key string) string {
	lines := strings.Split(text, "\n")
	start := -1
	for i, line := range lines {
		if strings.TrimSpace(line) == key+":" {
			start = i
			break
		}
	}
	if start == -1 {
		return ""
	}
	baseIndent := leadingSpaceCount(lines[start])
	end := len(lines)
	for i := start + 1; i < len(lines); i++ {
		trimmed := strings.TrimSpace(lines[i])
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			continue
		}
		if leadingSpaceCount(lines[i]) <= baseIndent && strings.HasSuffix(trimmed, ":") {
			end = i
			break
		}
	}
	return strings.Join(lines[start:end], "\n")
}

func countYAMLListItems(section string) int {
	count := 0
	for _, line := range strings.Split(section, "\n") {
		if strings.HasPrefix(line, "    - ") || strings.HasPrefix(line, "  - ") {
			count++
		}
	}
	return count
}

func normalizedYAMLIDs(text string) map[string]bool {
	ids := map[string]bool{}
	for _, line := range strings.Split(text, "\n") {
		value := strings.TrimSpace(line)
		listItem := false
		if strings.HasPrefix(value, "- ") {
			value = strings.TrimSpace(strings.TrimPrefix(value, "- "))
			listItem = true
		}
		if strings.HasPrefix(value, "id:") {
			value = strings.TrimSpace(strings.TrimPrefix(value, "id:"))
		} else if strings.HasPrefix(value, "label:") {
			value = strings.TrimSpace(strings.TrimPrefix(value, "label:"))
		} else if strings.HasPrefix(value, "contains:") {
			value = strings.TrimSpace(strings.TrimPrefix(value, "contains:"))
		} else if !listItem {
			continue
		}
		value = strings.Trim(value, `"'[]`)
		if id := normalizeScorecardID(value); id != "" {
			ids[id] = true
		}
	}
	return ids
}

func normalizeScorecardID(value string) string {
	var builder strings.Builder
	lastDash := false
	for _, r := range strings.ToLower(strings.TrimSpace(value)) {
		switch {
		case unicode.IsLetter(r) || unicode.IsDigit(r):
			builder.WriteRune(r)
			lastDash = false
		default:
			if !lastDash && builder.Len() > 0 {
				builder.WriteByte('-')
				lastDash = true
			}
		}
	}
	return strings.Trim(builder.String(), "-")
}

func leadingSpaceCount(line string) int {
	count := 0
	for _, r := range line {
		if r != ' ' {
			break
		}
		count++
	}
	return count
}
