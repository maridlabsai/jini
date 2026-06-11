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
)

type shipCheckReport struct {
	SchemaVersion      string                  `json:"schema_version"`
	ResultType         string                  `json:"result_type"`
	GeneratedAt        string                  `json:"generated_at"`
	Status             string                  `json:"status"`
	Workspace          string                  `json:"workspace"`
	InGitRepo          bool                    `json:"in_git_repo"`
	Branch             string                  `json:"branch,omitempty"`
	Upstream           string                  `json:"upstream,omitempty"`
	AheadCount         int                     `json:"ahead_count"`
	BehindCount        int                     `json:"behind_count"`
	DirtyFiles         int                     `json:"dirty_files"`
	UntrackedFiles     int                     `json:"untracked_files"`
	RequiredEvidence   []string                `json:"required_evidence"`
	ReleaseClaimPolicy []string                `json:"release_claim_policy"`
	ConfigIssues       []string                `json:"config_issues,omitempty"`
	EvidenceIssues     []string                `json:"evidence_issues,omitempty"`
	CLIHandoffDogfood  []shipCLIHandoffDogfood `json:"cli_handoff_dogfood"`
	Blockers           []string                `json:"blockers,omitempty"`
	Warnings           []string                `json:"warnings,omitempty"`
	Next               []string                `json:"next"`
}

type shipCLIHandoffDogfood struct {
	RouteID         string   `json:"route_id"`
	Label           string   `json:"label"`
	ReleaseClaimed  bool     `json:"release_claimed"`
	Status          string   `json:"status"`
	SetupStatus     string   `json:"setup_status"`
	DogfoodStatus   string   `json:"dogfood_status"`
	Executable      string   `json:"executable"`
	ArgsTemplate    []string `json:"args_template"`
	RequiredChecks  []string `json:"required_checks"`
	ValidatedChecks []string `json:"validated_checks,omitempty"`
	MissingChecks   []string `json:"missing_checks,omitempty"`
	EvidencePath    string   `json:"evidence_path,omitempty"`
	LastValidatedAt string   `json:"last_validated_at,omitempty"`
	Missing         []string `json:"missing,omitempty"`
}

type shipCLIHandoffDogfoodEvidenceFile struct {
	SchemaVersion string                                     `json:"schema_version"`
	ContextType   string                                     `json:"context_type"`
	Routes        map[string]shipCLIHandoffDogfoodRouteProof `json:"routes"`
}

type routeDogfoodGuideReport struct {
	SchemaVersion      string                            `json:"schema_version"`
	ResultType         string                            `json:"result_type"`
	EvidenceFile       string                            `json:"evidence_file"`
	RequiredChecks     []string                          `json:"required_checks"`
	Routes             []routeDogfoodGuideRoute          `json:"routes"`
	ReleaseClaimPolicy []string                          `json:"release_claim_policy"`
	ConfigIssues       []string                          `json:"config_issues,omitempty"`
	EvidenceIssues     []string                          `json:"evidence_issues,omitempty"`
	ValidationSteps    []string                          `json:"validation_steps"`
	EvidenceRules      []string                          `json:"evidence_rules"`
	EvidenceTemplate   shipCLIHandoffDogfoodEvidenceFile `json:"evidence_template"`
	Next               []string                          `json:"next"`
}

type routeDogfoodGuideRoute struct {
	RouteID         string   `json:"route_id"`
	Label           string   `json:"label"`
	ReleaseClaimed  bool     `json:"release_claimed"`
	SetupStatus     string   `json:"setup_status"`
	DogfoodStatus   string   `json:"dogfood_status"`
	SetupCategory   string   `json:"setup_category,omitempty"`
	SetupHint       string   `json:"setup_hint,omitempty"`
	ValidatedChecks []string `json:"validated_checks,omitempty"`
	MissingChecks   []string `json:"missing_checks,omitempty"`
	LastValidatedAt string   `json:"last_validated_at,omitempty"`
}

type shipCLIHandoffDogfoodRouteProof struct {
	ValidatedAt string   `json:"validated_at"`
	Checks      []string `json:"checks"`
}

type cliHandoffReleaseClaimConfig struct {
	Claims map[string]bool
	Issues []string
}

type shipCLIHandoffDogfoodEvidenceLoad struct {
	Path   string
	Routes map[string]shipCLIHandoffDogfoodRouteProof
	Issues []string
}

func runShipCheck(args []string, stdout, stderr io.Writer) int {
	format, ok := parseOptionalFormatArgs(args)
	if !ok {
		fmt.Fprintln(stderr, "Unsupported check ship format. Try `jini check ship` or `jini check ship --format json`.")
		return 1
	}
	report := buildShipCheckReport()
	if format == "json" {
		encoder := json.NewEncoder(stdout)
		encoder.SetIndent("", "  ")
		encoder.SetEscapeHTML(false)
		if err := encoder.Encode(report); err != nil {
			fmt.Fprintf(stderr, "Could not render ship check report: %v\n", err)
			return 1
		}
	} else {
		renderShipCheckText(stdout, report)
	}
	if report.Status == "ok" {
		return 0
	}
	return 1
}

func buildShipCheckReport() shipCheckReport {
	report := shipCheckReport{
		SchemaVersion: "0.1.0",
		ResultType:    "JiniShipCheck",
		GeneratedAt:   time.Now().UTC().Format(time.RFC3339),
		Status:        "ok",
		RequiredEvidence: []string{
			"bash tools/run_required_gates.sh push",
			"git worktree add",
			"write validation report before push",
			"real installed CLI handoff dogfood for Wave 1 routes",
		},
		ReleaseClaimPolicy: cliHandoffReleaseClaimPolicy(),
		Next: []string{
			"Create an isolated worktree for validation.",
			"Run the required push gates there.",
			"Record installed CLI dogfood evidence in .jini/cli-dogfood.json.",
			"Push only after the validation report is clean.",
		},
	}
	var cliConfigIssues []string
	var evidenceIssues []string
	report.CLIHandoffDogfood, cliConfigIssues, evidenceIssues = buildShipCLIHandoffDogfood()
	report.ConfigIssues = cliConfigIssues
	report.EvidenceIssues = evidenceIssues
	for _, issue := range cliConfigIssues {
		report.Blockers = append(report.Blockers, issue)
	}
	for _, issue := range evidenceIssues {
		report.Blockers = append(report.Blockers, issue)
	}
	readyDogfood, missingDogfood := summarizeShipCLIHandoffDogfood(report.CLIHandoffDogfood)
	validatedDogfood, needsValidationDogfood, setupBlockedDogfood := summarizeShipCLIHandoffValidation(report.CLIHandoffDogfood)
	if missingDogfood > 0 {
		report.Warnings = append(report.Warnings, fmt.Sprintf("CLI handoff setup incomplete: %d executable ready, %d need setup", readyDogfood, missingDogfood))
	}
	if needsValidationDogfood > 0 || setupBlockedDogfood > 0 {
		report.Warnings = append(report.Warnings, fmt.Sprintf("CLI handoff dogfood incomplete: %d validated, %d need validation, %d setup blocked", validatedDogfood, needsValidationDogfood, setupBlockedDogfood))
	}
	report.Blockers = append(report.Blockers, shipCLIHandoffDogfoodBlockers(report.CLIHandoffDogfood)...)
	if len(report.Blockers) > 0 {
		report.Status = "blocked"
	}

	if cwd, ok := runGitOutput("rev-parse", "--show-toplevel"); ok {
		report.Workspace = filepath.Base(strings.TrimSpace(cwd))
	} else {
		report.Workspace = "current directory"
	}
	if inRepo, ok := runGitOutput("rev-parse", "--is-inside-work-tree"); !ok || strings.TrimSpace(inRepo) != "true" {
		report.Status = "blocked"
		report.Blockers = append(report.Blockers, "not inside a git repository")
		return report
	}
	report.InGitRepo = true

	if branch, ok := runGitOutput("branch", "--show-current"); ok {
		report.Branch = strings.TrimSpace(branch)
	}
	if upstream, ok := runGitOutput("rev-parse", "--abbrev-ref", "--symbolic-full-name", "@{u}"); ok {
		report.Upstream = strings.TrimSpace(upstream)
		report.AheadCount, report.BehindCount = gitAheadBehind()
	} else {
		report.Warnings = append(report.Warnings, "no upstream branch configured")
	}

	report.DirtyFiles, report.UntrackedFiles = gitDirtyCounts()
	if report.DirtyFiles > 0 {
		report.Status = "blocked"
		report.Blockers = append(report.Blockers, "working tree has uncommitted changes")
	}
	return report
}

func shipCLIHandoffDogfoodBlockers(items []shipCLIHandoffDogfood) []string {
	var blockers []string
	for _, item := range items {
		switch item.DogfoodStatus {
		case "needs-validation":
			blockers = append(blockers, "CLI handoff dogfood missing validation for installed route: "+item.RouteID)
		case "setup-blocked":
			category := shipCLIHandoffSetupCategory(item.Missing)
			if item.ReleaseClaimed {
				blockers = append(blockers, "CLI handoff setup blocked for claimed route: "+item.RouteID+" ("+category+")")
			} else if category != "missing executable" {
				blockers = append(blockers, "CLI handoff setup blocked for installed route: "+item.RouteID+" ("+category+")")
			}
		}
	}
	return blockers
}

func cliHandoffReleaseClaimPolicy() []string {
	return []string{
		"Installed CLI routes and routes named in JINI_CLI_RELEASE_ROUTES must be trusted and dogfooded before release claims.",
		"Missing optional CLI executables are setup backlog until the release claim names them.",
		"Do not publicly claim a CLI route until it is installed, trusted, and validated in .jini/cli-dogfood.json.",
	}
}

func cliHandoffReleaseClaimSet() map[string]bool {
	return parseCLIHandoffReleaseClaimConfig().Claims
}

func parseCLIHandoffReleaseClaimConfig() cliHandoffReleaseClaimConfig {
	config := cliHandoffReleaseClaimConfig{Claims: map[string]bool{}}
	raw := strings.TrimSpace(configValue("JINI_CLI_RELEASE_ROUTES"))
	if raw == "" {
		return config
	}
	for _, token := range strings.FieldsFunc(raw, func(r rune) bool {
		return r == ',' || r == ';' || r == ' ' || r == '\t' || r == '\n'
	}) {
		rawToken := strings.ToLower(strings.TrimSpace(token))
		if rawToken == "" {
			continue
		}
		if rawToken == "all" {
			for _, mode := range cliHandoffRouteIDs() {
				config.Claims[mode] = true
			}
			continue
		}
		routeID := normalizeCLIHandoffRouteID(token)
		if routeID == "" {
			config.Issues = append(config.Issues, "unknown CLI release claim route: "+rawToken+" (valid: "+strings.Join(cliHandoffRouteIDs(), ", ")+")")
			continue
		}
		if _, ok := cliHandoffDescriptorForMode(routeID); ok {
			config.Claims[routeID] = true
		}
	}
	return config
}

func normalizeCLIHandoffRouteID(value string) string {
	routeID := strings.ToLower(strings.TrimSpace(value))
	routeID = strings.ReplaceAll(routeID, "_", "-")
	routeID = strings.ReplaceAll(routeID, " ", "-")
	if routeID == "" {
		return ""
	}
	if _, ok := cliHandoffDescriptorForMode(routeID); ok {
		return routeID
	}
	return ""
}

func cliHandoffRouteIDs() []string {
	return []string{"codex", "claude-code", "gemini-cli", "aider", "opencode"}
}

func buildShipCLIHandoffDogfood() ([]shipCLIHandoffDogfood, []string, []string) {
	releaseClaimConfig := parseCLIHandoffReleaseClaimConfig()
	evidenceLoad := loadShipCLIHandoffDogfoodEvidence()
	routeIDs := cliHandoffRouteIDs()
	out := make([]shipCLIHandoffDogfood, 0, len(routeIDs))
	for _, mode := range routeIDs {
		descriptor, ok := cliHandoffDescriptorForMode(mode)
		if !ok {
			continue
		}
		command, missing := resolveCLIHandoffCommand(descriptor)
		status := "ready"
		if len(missing) > 0 {
			status = "needs-setup"
		}
		args := command.Args
		if len(args) == 0 {
			args = descriptor.DefaultArgs
		}
		requiredChecks := requiredCLIHandoffDogfoodChecks()
		setupStatus := status
		dogfoodStatus := "setup-blocked"
		var validatedChecks []string
		var missingChecks []string
		var routeEvidencePath string
		var lastValidatedAt string
		if setupStatus == "ready" {
			routeProof := evidenceLoad.Routes[descriptor.Mode]
			validatedChecks, missingChecks = dogfoodCheckCoverage(requiredChecks, routeProof.Checks)
			if timestampIssue := dogfoodValidationTimestampIssue(routeProof.ValidatedAt); timestampIssue != "" {
				missingChecks = append(missingChecks, timestampIssue)
			}
			if len(missingChecks) == 0 {
				dogfoodStatus = "validated"
				routeEvidencePath = evidenceLoad.Path
				lastValidatedAt = strings.TrimSpace(routeProof.ValidatedAt)
			} else {
				dogfoodStatus = "needs-validation"
			}
		}
		out = append(out, shipCLIHandoffDogfood{
			RouteID:         descriptor.Mode,
			Label:           descriptor.Label,
			ReleaseClaimed:  releaseClaimConfig.Claims[descriptor.Mode],
			Status:          status,
			SetupStatus:     setupStatus,
			DogfoodStatus:   dogfoodStatus,
			Executable:      firstNonEmpty(command.Executable, descriptor.DefaultExecutable),
			ArgsTemplate:    append([]string(nil), args...),
			RequiredChecks:  requiredChecks,
			ValidatedChecks: validatedChecks,
			MissingChecks:   missingChecks,
			EvidencePath:    routeEvidencePath,
			LastValidatedAt: lastValidatedAt,
			Missing:         missing,
		})
	}
	return out, releaseClaimConfig.Issues, evidenceLoad.Issues
}

func requiredCLIHandoffDogfoodChecks() []string {
	return []string{
		"auth",
		"approvals",
		"output shape",
		"route receipt privacy",
	}
}

func dogfoodValidationTimestampIssue(value string) string {
	value = strings.TrimSpace(value)
	if value == "" || value == "YYYY-MM-DDTHH:MM:SSZ" {
		return "valid validated_at timestamp"
	}
	if _, err := time.Parse(time.RFC3339, value); err != nil {
		return "valid validated_at timestamp"
	}
	return ""
}

func buildRouteDogfoodGuide() routeDogfoodGuideReport {
	dogfoodRows, configIssues, evidenceIssues := buildShipCLIHandoffDogfood()
	requiredChecks := requiredCLIHandoffDogfoodChecks()
	templateRoutes := map[string]shipCLIHandoffDogfoodRouteProof{}
	routes := make([]routeDogfoodGuideRoute, 0, len(dogfoodRows))
	for _, route := range dogfoodRows {
		if route.SetupStatus != "ready" {
			setupCategory := shipCLIHandoffSetupCategory(route.Missing)
			routes = append(routes, routeDogfoodGuideRoute{
				RouteID:        route.RouteID,
				Label:          route.Label,
				ReleaseClaimed: route.ReleaseClaimed,
				SetupStatus:    route.SetupStatus,
				DogfoodStatus:  route.DogfoodStatus,
				SetupCategory:  setupCategory,
				SetupHint:      routeDogfoodSetupHint(route.RouteID, route.Label, setupCategory),
			})
			continue
		}
		templateRoutes[route.RouteID] = shipCLIHandoffDogfoodRouteProof{
			ValidatedAt: "YYYY-MM-DDTHH:MM:SSZ",
			Checks:      append([]string(nil), requiredChecks...),
		}
		routes = append(routes, routeDogfoodGuideRoute{
			RouteID:         route.RouteID,
			Label:           route.Label,
			ReleaseClaimed:  route.ReleaseClaimed,
			SetupStatus:     route.SetupStatus,
			DogfoodStatus:   route.DogfoodStatus,
			ValidatedChecks: append([]string(nil), route.ValidatedChecks...),
			MissingChecks:   append([]string(nil), route.MissingChecks...),
			LastValidatedAt: route.LastValidatedAt,
		})
	}
	return routeDogfoodGuideReport{
		SchemaVersion:      "0.1.0",
		ResultType:         "JiniRouteDogfoodGuide",
		EvidenceFile:       ".jini/cli-dogfood.json",
		RequiredChecks:     requiredChecks,
		Routes:             routes,
		ReleaseClaimPolicy: cliHandoffReleaseClaimPolicy(),
		ConfigIssues:       configIssues,
		EvidenceIssues:     evidenceIssues,
		ValidationSteps: []string{
			"For each ready route, select that route and run a harmless prompt through Jini using the real installed CLI.",
			"Confirm downstream auth, approval behavior, output shape, and route receipt privacy before editing evidence.",
			"Record the route only after the installed CLI completed successfully on the validation machine.",
		},
		EvidenceRules: []string{
			"Do not use fake CLIs, provider API aliases, skipped trust checks, or stale evidence from an older CLI version.",
			"Do not mark setup-blocked routes validated.",
			"Rerun `jini check ship --format json` after editing .jini/cli-dogfood.json.",
		},
		EvidenceTemplate: shipCLIHandoffDogfoodEvidenceFile{
			SchemaVersion: "0.1.0",
			ContextType:   "JiniCLIHandoffDogfoodEvidence",
			Routes:        templateRoutes,
		},
		Next: []string{
			"Validate only with the real installed CLI, not a provider API alias.",
			"Record validated checks in .jini/cli-dogfood.json.",
			"Run `jini check ship --format json`.",
		},
	}
}

func loadShipCLIHandoffDogfoodEvidence() shipCLIHandoffDogfoodEvidenceLoad {
	path := cliHandoffDogfoodEvidencePath()
	load := shipCLIHandoffDogfoodEvidenceLoad{
		Path:   path,
		Routes: map[string]shipCLIHandoffDogfoodRouteProof{},
	}
	data, err := os.ReadFile(path)
	if err != nil {
		if !os.IsNotExist(err) {
			load.Issues = append(load.Issues, "CLI dogfood evidence file could not be read: "+filepath.Base(path))
		}
		return load
	}
	var payload shipCLIHandoffDogfoodEvidenceFile
	if err := json.Unmarshal(data, &payload); err != nil {
		load.Issues = append(load.Issues, "CLI dogfood evidence file is invalid JSON: "+filepath.Base(path))
		return load
	}
	if strings.TrimSpace(payload.SchemaVersion) != "0.1.0" {
		load.Issues = append(load.Issues, "CLI dogfood evidence file has invalid schema_version: expected 0.1.0")
	}
	if strings.TrimSpace(payload.ContextType) != "JiniCLIHandoffDogfoodEvidence" {
		load.Issues = append(load.Issues, "CLI dogfood evidence file has invalid context_type: expected JiniCLIHandoffDogfoodEvidence")
	}
	if payload.Routes == nil {
		load.Issues = append(load.Issues, "CLI dogfood evidence file has no routes object: "+filepath.Base(path))
		return load
	}
	out := make(map[string]shipCLIHandoffDogfoodRouteProof, len(payload.Routes))
	for routeID, proof := range payload.Routes {
		normalizedRouteID := normalizeCLIHandoffRouteID(routeID)
		if normalizedRouteID == "" {
			load.Issues = append(load.Issues, "CLI dogfood evidence ignored unknown route: "+strings.TrimSpace(routeID)+" (valid: "+strings.Join(cliHandoffRouteIDs(), ", ")+")")
			continue
		}
		out[normalizedRouteID] = proof
	}
	load.Routes = out
	return load
}

func cliHandoffDogfoodEvidencePath() string {
	return filepath.Join(sessionStateRoot(), "cli-dogfood.json")
}

func saveCLIHandoffDogfoodEvidence(routeID string, checks []string, validatedAt string) (string, error) {
	load := loadShipCLIHandoffDogfoodEvidence()
	if len(load.Issues) > 0 {
		return load.Path, fmt.Errorf("%s", strings.Join(load.Issues, "; "))
	}
	payload := shipCLIHandoffDogfoodEvidenceFile{
		SchemaVersion: "0.1.0",
		ContextType:   "JiniCLIHandoffDogfoodEvidence",
		Routes:        load.Routes,
	}
	if payload.Routes == nil {
		payload.Routes = map[string]shipCLIHandoffDogfoodRouteProof{}
	}
	payload.Routes[routeID] = shipCLIHandoffDogfoodRouteProof{
		ValidatedAt: validatedAt,
		Checks:      append([]string(nil), checks...),
	}
	data, err := json.MarshalIndent(payload, "", "  ")
	if err != nil {
		return load.Path, err
	}
	if err := os.MkdirAll(filepath.Dir(load.Path), 0o755); err != nil {
		return load.Path, err
	}
	return load.Path, os.WriteFile(load.Path, append(data, '\n'), 0o600)
}

func dogfoodCheckCoverage(required, actual []string) (validated, missing []string) {
	actualSet := map[string]bool{}
	for _, item := range actual {
		item = strings.ToLower(strings.TrimSpace(item))
		if item != "" {
			actualSet[item] = true
		}
	}
	for _, check := range required {
		normalized := strings.ToLower(strings.TrimSpace(check))
		if actualSet[normalized] {
			validated = append(validated, check)
			continue
		}
		missing = append(missing, check)
	}
	return validated, missing
}

func summarizeShipCLIHandoffDogfood(items []shipCLIHandoffDogfood) (ready, needsSetup int) {
	for _, item := range items {
		if item.SetupStatus == "ready" || item.Status == "ready" {
			ready++
		} else {
			needsSetup++
		}
	}
	return ready, needsSetup
}

func summarizeShipCLIHandoffValidation(items []shipCLIHandoffDogfood) (validated, needsValidation, setupBlocked int) {
	for _, item := range items {
		switch item.DogfoodStatus {
		case "validated":
			validated++
		case "needs-validation":
			needsValidation++
		default:
			setupBlocked++
		}
	}
	return validated, needsValidation, setupBlocked
}

func gitAheadBehind() (int, int) {
	out, ok := runGitOutput("rev-list", "--left-right", "--count", "HEAD...@{u}")
	if !ok {
		return 0, 0
	}
	fields := strings.Fields(out)
	if len(fields) != 2 {
		return 0, 0
	}
	ahead, _ := strconv.Atoi(fields[0])
	behind, _ := strconv.Atoi(fields[1])
	return ahead, behind
}

func gitDirtyCounts() (int, int) {
	out, ok := runGitOutput("status", "--porcelain")
	if !ok {
		return 0, 0
	}
	dirty := 0
	untracked := 0
	for _, line := range strings.Split(out, "\n") {
		if strings.TrimSpace(line) == "" {
			continue
		}
		dirty++
		if strings.HasPrefix(line, "??") {
			untracked++
		}
	}
	return dirty, untracked
}

func renderShipCheckText(w io.Writer, report shipCheckReport) {
	fmt.Fprintf(w, "Ship check %s\n", report.Status)
	fmt.Fprintf(w, "Branch: %s\n", firstNonEmpty(report.Branch, "detached or unnamed"))
	fmt.Fprintf(w, "Workspace: %s\n", firstNonEmpty(report.Workspace, "current directory"))
	fmt.Fprintf(w, "Dirty files: %d\n", report.DirtyFiles)
	if report.AheadCount > 0 || report.BehindCount > 0 {
		fmt.Fprintf(w, "Upstream delta: ahead %d, behind %d\n", report.AheadCount, report.BehindCount)
	}
	for _, blocker := range report.Blockers {
		fmt.Fprintf(w, "Blocker: %s\n", blocker)
	}
	for _, warning := range report.Warnings {
		fmt.Fprintf(w, "Warning: %s\n", warning)
	}
	renderRouteDogfoodListSection(w, "Config issues:", report.ConfigIssues)
	renderRouteDogfoodListSection(w, "Evidence issues:", report.EvidenceIssues)
	readyDogfood, missingDogfood := summarizeShipCLIHandoffDogfood(report.CLIHandoffDogfood)
	validatedDogfood, needsValidationDogfood, setupBlockedDogfood := summarizeShipCLIHandoffValidation(report.CLIHandoffDogfood)
	fmt.Fprintf(w, "CLI handoff setup: %d executable ready, %d need setup\n", readyDogfood, missingDogfood)
	fmt.Fprintf(w, "CLI handoff dogfood: %d validated, %d need validation, %d setup blocked\n", validatedDogfood, needsValidationDogfood, setupBlockedDogfood)
	renderShipCLIHandoffDogfoodText(w, report.CLIHandoffDogfood)
	renderRouteDogfoodListSection(w, "Release claim policy:", report.ReleaseClaimPolicy)
	fmt.Fprintln(w, "Dogfood before release: verify auth, approvals, output shape, and route receipt privacy on real installed CLIs.")
	fmt.Fprintln(w, "Evidence file: .jini/cli-dogfood.json")
	fmt.Fprintln(w, "Evidence checks: auth, approvals, output shape, route receipt privacy")
	fmt.Fprintln(w, "Run before push: bash tools/run_required_gates.sh push")
	fmt.Fprintln(w, "Safe lane: create an isolated worktree, run gates, then push only after evidence is clean.")
}

func renderRouteDogfoodGuide(w io.Writer, report routeDogfoodGuideReport) {
	fmt.Fprintln(w, "CLI dogfood")
	fmt.Fprintf(w, "Evidence file: %s\n", report.EvidenceFile)
	fmt.Fprintf(w, "Required checks: %s\n", strings.Join(report.RequiredChecks, ", "))
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Routes:")
	for _, route := range report.Routes {
		routeLabel := formatCLIHandoffRouteLabel(route.RouteID, route.ReleaseClaimed)
		switch {
		case route.DogfoodStatus == "validated":
			fmt.Fprintf(w, "- %s: validated at %s\n", routeLabel, firstNonEmpty(route.LastValidatedAt, "unknown time"))
		case route.SetupStatus == "ready":
			fmt.Fprintf(w, "- %s: ready; missing %s\n", routeLabel, strings.Join(route.MissingChecks, ", "))
		default:
			fmt.Fprintf(w, "- %s: setup blocked (%s)\n", routeLabel, firstNonEmpty(route.SetupCategory, "see doctor"))
		}
	}
	renderRouteDogfoodListSection(w, "Config issues:", report.ConfigIssues)
	renderRouteDogfoodListSection(w, "Evidence issues:", report.EvidenceIssues)
	renderRouteDogfoodSetupHints(w, report.Routes)
	renderRouteDogfoodListSection(w, "Release claim policy:", report.ReleaseClaimPolicy)
	renderRouteDogfoodListSection(w, "Validation steps:", report.ValidationSteps)
	renderRouteDogfoodListSection(w, "Evidence rules:", report.EvidenceRules)
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Template:")
	renderRouteDogfoodEvidenceTemplate(w, report.EvidenceTemplate, report.Routes)
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Do not mark a route validated until you used the real installed CLI.")
	fmt.Fprintln(w, "Then run `jini check ship --format json`.")
}

func renderRouteDogfoodListSection(w io.Writer, title string, items []string) {
	if len(items) == 0 {
		return
	}
	fmt.Fprintln(w)
	fmt.Fprintln(w, title)
	for _, item := range items {
		if strings.TrimSpace(item) == "" {
			continue
		}
		fmt.Fprintf(w, "- %s\n", item)
	}
}

func renderRouteDogfoodSetupHints(w io.Writer, routes []routeDogfoodGuideRoute) {
	hasHints := false
	for _, route := range routes {
		if strings.TrimSpace(route.SetupHint) != "" {
			hasHints = true
			break
		}
	}
	if !hasHints {
		return
	}
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Setup fixes:")
	for _, route := range routes {
		if strings.TrimSpace(route.SetupHint) == "" {
			continue
		}
		fmt.Fprintf(w, "- %s: %s\n", formatCLIHandoffRouteLabel(route.RouteID, route.ReleaseClaimed), route.SetupHint)
	}
}

func routeDogfoodSetupHint(routeID, label, category string) string {
	descriptor, _ := cliHandoffDescriptorForMode(routeID)
	toolName := strings.TrimSuffix(label, " handoff")
	switch category {
	case "macOS Gatekeeper":
		return "reinstall from a trusted source, open the CLI once, approve it in macOS Privacy & Security if prompted, then rerun `jini route dogfood`."
	case "missing executable":
		envVar := firstNonEmpty(descriptor.ExecutableEnv, "the route executable env var")
		return "install " + toolName + " or set " + envVar + ", then rerun `jini route dogfood`."
	case "invalid args":
		envVar := firstNonEmpty(descriptor.ArgsEnv, "the route args env var")
		return "fix " + envVar + " quoting, then rerun `jini route dogfood`."
	case "local trust check":
		return "resolve the local trust check in `jini doctor`, then rerun `jini route dogfood`."
	default:
		return "run `jini doctor` for setup details, then rerun `jini route dogfood`."
	}
}

func renderRouteDogfoodEvidenceTemplate(w io.Writer, template shipCLIHandoffDogfoodEvidenceFile, routes []routeDogfoodGuideRoute) {
	fmt.Fprintln(w, "{")
	fmt.Fprintf(w, "  \"schema_version\": %q,\n", template.SchemaVersion)
	fmt.Fprintf(w, "  \"context_type\": %q,\n", template.ContextType)
	fmt.Fprintln(w, "  \"routes\": {")
	readyRoutes := make([]routeDogfoodGuideRoute, 0, len(routes))
	for _, route := range routes {
		if route.SetupStatus == "ready" {
			readyRoutes = append(readyRoutes, route)
		}
	}
	for i, route := range readyRoutes {
		comma := ","
		if i == len(readyRoutes)-1 {
			comma = ""
		}
		fmt.Fprintf(w, "    %q: {\n", route.RouteID)
		fmt.Fprintln(w, "      \"validated_at\": \"YYYY-MM-DDTHH:MM:SSZ\",")
		fmt.Fprintf(w, "      \"checks\": [%s]\n", quotedStringList(requiredCLIHandoffDogfoodChecks()))
		fmt.Fprintf(w, "    }%s\n", comma)
	}
	fmt.Fprintln(w, "  }")
	fmt.Fprintln(w, "}")
}

func quotedStringList(items []string) string {
	quoted := make([]string, 0, len(items))
	for _, item := range items {
		quoted = append(quoted, strconv.Quote(item))
	}
	return strings.Join(quoted, ", ")
}

func renderShipCLIHandoffDogfoodText(w io.Writer, items []shipCLIHandoffDogfood) {
	if len(items) == 0 {
		return
	}
	fmt.Fprintln(w, "CLI handoff routes:")
	for _, item := range items {
		routeLabel := formatCLIHandoffRouteLabel(item.RouteID, item.ReleaseClaimed)
		if item.SetupStatus == "ready" {
			if item.DogfoodStatus == "validated" {
				fmt.Fprintf(w, "- %s: executable ready, dogfood validated\n", routeLabel)
				continue
			}
			fmt.Fprintf(w, "- %s: executable ready, dogfood needs validation\n", routeLabel)
			continue
		}
		fmt.Fprintf(w, "- %s: needs setup (%s)\n", routeLabel, shipCLIHandoffSetupCategory(item.Missing))
	}
}

func formatCLIHandoffRouteLabel(routeID string, releaseClaimed bool) string {
	if releaseClaimed {
		return routeID + " (claimed)"
	}
	return routeID
}

func shipCLIHandoffSetupCategory(missing []string) string {
	for _, item := range missing {
		normalized := strings.ToLower(item)
		switch {
		case strings.Contains(normalized, "gatekeeper rejected"):
			return "macOS Gatekeeper"
		case strings.Contains(normalized, "requires an installed cli executable"):
			return "missing executable"
		case strings.Contains(normalized, "invalid quoting"):
			return "invalid args"
		case strings.Contains(normalized, "trust checks"):
			return "local trust check"
		}
	}
	return "see JSON details"
}
