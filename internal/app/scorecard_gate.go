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
	},
	MinimumCoreCompetitors:      7,
	MinimumWatchlistCompetitors: 30,
	MinimumScenarios:            8,
}

type scorecardGatePolicy struct {
	BenchmarkPath               string
	RequiredCompetitors         []string
	RequiredPressureVectors     []string
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
	Checks                   []scorecardThresholdCheck `json:"checks"`
}

type scorecardPresenceCheck struct {
	ID      string `json:"id"`
	Present bool   `json:"present"`
	Status  string `json:"status"`
}

type scorecardThresholdCheck struct {
	ID      string `json:"id"`
	Actual  int    `json:"actual"`
	Minimum int    `json:"minimum"`
	Status  string `json:"status"`
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

	builder.addThresholdChecks(&report, text, scorecardGateSection)
	report.RequiredCompetitors = builder.presenceChecks(builder.policy.RequiredCompetitors, func(id string) bool {
		return normalizedBenchmark[id] && normalizedScorecardGate[id]
	})
	report.PressureVectors = builder.presenceChecks(builder.policy.RequiredPressureVectors, func(id string) bool {
		return normalizedScorecardGate[id]
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
	}
	fmt.Fprintln(stdout, "PRESSURE VECTORS")
	for _, vector := range report.PressureVectors {
		fmt.Fprintf(stdout, "  %s %s\n", strings.ToUpper(vector.Status), vector.ID)
	}
	fmt.Fprintln(stdout, "THRESHOLDS")
	for _, check := range report.Checks {
		fmt.Fprintf(stdout, "  %s %s %d/%d\n", strings.ToUpper(check.Status), check.ID, check.Actual, check.Minimum)
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
