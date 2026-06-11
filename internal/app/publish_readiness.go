package app

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const jiniModuleDeclaration = "module github.com/maridlabsai/jini"

type publishReadinessReport struct {
	SchemaVersion string                    `json:"schema_version"`
	ResultType    string                    `json:"result_type"`
	GeneratedAt   string                    `json:"generated_at"`
	Status        string                    `json:"status"`
	Runtime       publishReadinessRuntime   `json:"runtime"`
	PackCount     int                       `json:"pack_count"`
	BundleCount   int                       `json:"bundle_count"`
	KitCount      int                       `json:"kit_count"`
	TargetCount   int                       `json:"target_count"`
	Sections      []publishReadinessSection `json:"sections"`
}

type publishReadinessRuntime struct {
	Language       string `json:"language"`
	LegacyFallback bool   `json:"legacy_fallback"`
}

type publishReadinessSection struct {
	ID      string                        `json:"id"`
	Label   string                        `json:"label"`
	Status  string                        `json:"status"`
	Checks  []publishReadinessCheck       `json:"checks,omitempty"`
	Details map[string]publishMetricValue `json:"details,omitempty"`
	Claims  []publishEvidenceClaim        `json:"claims,omitempty"`
}

type publishReadinessCheck struct {
	Path             string   `json:"path"`
	Exists           bool     `json:"exists"`
	Status           string   `json:"status"`
	MissingFragments []string `json:"missing_fragments,omitempty"`
}

type publishMetricValue struct {
	Value  int    `json:"value"`
	Status string `json:"status"`
}

type publishEvidenceClaim struct {
	Claim              string `json:"claim"`
	Status             string `json:"status"`
	Evidence           string `json:"evidence"`
	Gap                string `json:"gap"`
	NextCut            string `json:"next_cut"`
	RuntimeImplemented bool   `json:"runtime_implemented"`
}

type publishFragmentRequirement struct {
	checkPath string
	filePath  string
	fragments []string
}

func runPublishReadiness(args []string, stdout, stderr io.Writer) int {
	format, ok := parseOptionalFormatArgs(args)
	if !ok {
		fmt.Fprintln(stderr, "Unsupported publish-readiness format. Try `jini publish-readiness` or `jini publish-readiness --format json`.")
		return 1
	}

	root := discoverSourceRoot()
	report := buildPublishReadinessReport(root)
	if format == "json" {
		encoder := json.NewEncoder(stdout)
		encoder.SetIndent("", "  ")
		encoder.SetEscapeHTML(false)
		if err := encoder.Encode(report); err != nil {
			fmt.Fprintf(stderr, "Could not render publish-readiness report: %v\n", err)
			return 1
		}
		if report.Status == "ok" {
			return 0
		}
		return 1
	}

	renderPublishReadinessText(stdout, report)
	if report.Status == "ok" {
		return 0
	}
	return 1
}

func buildPublishReadinessReport(root string) publishReadinessReport {
	sections := []publishReadinessSection{
		buildPublishDocsSection(root),
		buildPublishHonestAuditSection(root),
		buildPublishAppPlatformSection(root),
		buildPublishOfflineRegressionSection(root),
		buildPublishCompetitivePressureSection(root),
		buildPublishProductivityLearningSection(root),
		buildPublishRuntimeSection(root),
	}
	status := "ok"
	for _, section := range sections {
		if section.Status != "ok" {
			status = "needs-attention"
			break
		}
	}

	return publishReadinessReport{
		SchemaVersion: "0.1.0",
		ResultType:    "JiniPublishReadiness",
		GeneratedAt:   time.Now().UTC().Format(time.RFC3339),
		Status:        status,
		Runtime: publishReadinessRuntime{
			Language:       "go",
			LegacyFallback: false,
		},
		PackCount:   countChildDirs(filepath.Join(root, "packs")),
		BundleCount: countInstallManifestIDs(root, "bundles"),
		KitCount:    countInstallManifestIDs(root, "kits"),
		TargetCount: countInstallManifestIDs(root, "targets"),
		Sections:    sections,
	}
}

func buildPublishDocsSection(root string) publishReadinessSection {
	if root == "" {
		return publishReadinessSection{
			ID:     "docs",
			Label:  "Public docs",
			Status: "ok",
			Checks: []publishReadinessCheck{{
				Path:   "source checkout",
				Exists: false,
				Status: "not-required-for-installed-binary",
			}},
		}
	}
	required := []string{
		"README.md",
		"WHITEPAPER.md",
		"docs/index.md",
		"docs/install.md",
		"docs/cli.md",
		"specs/app-platform-shipping-playbook.md",
		"specs/platform-offline-strategy.md",
		"specs/lean-platform-doctrine.md",
		"specs/public-repo-boundary.md",
		"distribution/install-manifest.yaml",
	}
	checks := make([]publishReadinessCheck, 0, len(required))
	status := "ok"
	for _, path := range required {
		exists := regularFileExists(filepath.Join(root, path))
		checkStatus := "ok"
		if !exists {
			checkStatus = "missing"
			status = "needs-attention"
		}
		checks = append(checks, publishReadinessCheck{Path: path, Exists: exists, Status: checkStatus})
	}
	return publishReadinessSection{
		ID:     "docs",
		Label:  "Public docs",
		Status: status,
		Checks: checks,
	}
}

func buildPublishHonestAuditSection(root string) publishReadinessSection {
	claims := buildPublishHonestAuditClaims()
	if root == "" {
		return publishReadinessSection{
			ID:      "honest-audit",
			Label:   "Honest system audit guardrails",
			Status:  "ok",
			Details: buildPublishClaimDetails(claims),
			Claims:  claims,
			Checks: []publishReadinessCheck{{
				Path:   "source checkout",
				Exists: false,
				Status: "not-required-for-installed-binary",
			}},
		}
	}

	required := []publishFragmentRequirement{
		{
			checkPath: "specs/honest-system-audit.md#current-implementation-reality",
			filePath:  "specs/honest-system-audit.md",
			fragments: []string{
				"## Current Implementation Reality",
				"Guarded is not implemented.",
				"Configured CLI handoff",
				"P0 competitor watching",
				"P0 compounding user productivity learning",
			},
		},
		{
			checkPath: "specs/honest-system-audit.md#core-feedback-accommodations",
			filePath:  "specs/honest-system-audit.md",
			fragments: []string{
				"## Core Feedback Accommodations",
				"claim, status, evidence, and next cut",
				"No P0 is complete from documentation alone.",
			},
		},
		{
			checkPath: "specs/honest-system-audit.md#machine-readable-evidence-contract",
			filePath:  "specs/honest-system-audit.md",
			fragments: []string{
				"## Machine-Readable Evidence Contract",
				"`claim`",
				"`runtime_implemented`",
				"Readiness `ok` does not mean every claim is implemented.",
			},
		},
		{
			checkPath: "specs/skills-and-delegation-slice.md#tier-boundary",
			filePath:  "specs/skills-and-delegation-slice.md",
			fragments: []string{
				"## Purpose",
				"This public spec is a boundary handoff",
				"Commercial tier owns the agent and skills OS productivity suite.",
				"Free tier must not ship `skills`, `delegate`, developer agents, tester agents,",
				"public docs and readiness gates do not describe the commercial suite as a",
			},
		},
		{
			checkPath: "specs/lean-platform-gate.md#command-surface-discipline",
			filePath:  "specs/lean-platform-gate.md",
			fragments: []string{
				"### 3. Command-Surface Discipline",
				"The free tier must not include a skills-based OS productivity suite.",
				"ships developer agents, tester agents, `skills`, `delegate`, or a skills-based OS productivity suite in the free tier",
			},
		},
	}

	checks, status := buildFragmentChecks(root, required)
	return publishReadinessSection{
		ID:      "honest-audit",
		Label:   "Honest system audit guardrails",
		Status:  status,
		Checks:  checks,
		Details: buildPublishClaimDetails(claims),
		Claims:  claims,
	}
}

func buildPublishHonestAuditClaims() []publishEvidenceClaim {
	return []publishEvidenceClaim{
		{
			Claim:              "Native Go CLI",
			Status:             "implemented",
			Evidence:           "Go runtime, Go tests, no tracked Python source, required gates",
			Gap:                "CLI still needs first-minute dogfood across more personas",
			NextCut:            "Keep reducing command and launcher friction",
			RuntimeImplemented: true,
		},
		{
			Claim:              "Configured CLI handoff",
			Status:             "implemented",
			Evidence:           "Wave 0 handoff contract, Wave 1 route registry, doctor detection, fake downstream CLI command-shape tests, fail-closed missing/trust checks, route receipts, check ship setup status, signed .jini/cli-smoke.json evidence, validation freshness, and .jini/cli-dogfood.json release checks",
			Gap:                "Each downstream CLI still needs real-world dogfood for command templates, approvals, output shape, route receipt privacy, and signed smoke evidence on tester machines.",
			NextCut:            "Dogfood Wave 1 command templates against real installed CLIs without broadening the first-minute UX.",
			RuntimeImplemented: true,
		},
		{
			Claim:              "Simplicity as UX tenet",
			Status:             "guarded",
			Evidence:           "Canonical PRD, lean gate, product simplicity test",
			Gap:                "The runtime still exposes too much internal vocabulary in some flows",
			NextCut:            "Prefer natural jini intake and progressive disclosure",
			RuntimeImplemented: false,
		},
		{
			Claim:              "Repo review snapshot",
			Status:             "implemented",
			Evidence:           "Direct repo-review test and porcelain parser coverage",
			Gap:                "It is a model-free first pass, not a full code-review agent",
			NextCut:            "Add richer changed-file focus and security/test prompts",
			RuntimeImplemented: true,
		},
		{
			Claim:              "P0 competitor watching",
			Status:             "partial",
			Evidence:           "jini check competitor-watch runtime packet, release-plan ingestion feed, scorecard deltas, stale-requirement risks, PRD, competitive release plan, readiness guard, benchmark coverage",
			Gap:                "Runtime packet and ingestion feed exist, but scheduler automation and live source refresh are not implemented",
			NextCut:            "Add source-backed refresh and scheduled competitor-watch packets without changing active scope automatically",
			RuntimeImplemented: true,
		},
		{
			Claim:              "P0 compounding user productivity learning",
			Status:             "implemented",
			Evidence:           "user-context.json runtime store, jini memory inspect/on/off/forget, route preference signal capture, privacy filtering tests, PRD, learning-system spec, readiness guard, benchmark coverage",
			Gap:                "Runtime learning now has inspect, opt-out, forget, and safe route preference signals; broader habit signals and productivity metrics still need expansion",
			NextCut:            "Expand learning signals across accepted/rejected actions and add productivity impact metrics",
			RuntimeImplemented: true,
		},
		{
			Claim:              "Offline and local model story",
			Status:             "partial",
			Evidence:           "Local SLM profiles, device/runtime gates, offline strategy, tests",
			Gap:                "Local/offline behavior is not proven across shipped macOS, Windows, iOS, and Android apps",
			NextCut:            "Add device runtime smoke fixtures and app-surface proof",
			RuntimeImplemented: false,
		},
		{
			Claim:              "Skills and delegation",
			Status:             "specified",
			Evidence:           "Commercial-tier boundary in skills-and-delegation slice and simplicity gate",
			Gap:                "The agent and skills OS productivity suite is commercial-only and not a free-tier runtime feature",
			NextCut:            "Implement the commercial suite from the commercial PRD without adding skills or delegate commands to the free tier",
			RuntimeImplemented: false,
		},
		{
			Claim:              "App surfaces",
			Status:             "specified",
			Evidence:           "App platform playbook and app-surface PRDs",
			Gap:                "Desktop and mobile apps are not shipped clients yet",
			NextCut:            "Build app-shell proof over the same work object",
			RuntimeImplemented: false,
		},
		{
			Claim:              "Publish readiness",
			Status:             "partial",
			Evidence:           "jini publish-readiness checks docs, specs, benchmark coverage, runtime counts, and claim evidence",
			Gap:                "Many checks are fragment checks, not functional proofs",
			NextCut:            "Require every P0 claim to list implementation evidence or say it is only guarded",
			RuntimeImplemented: true,
		},
	}
}

func buildPublishClaimDetails(claims []publishEvidenceClaim) map[string]publishMetricValue {
	details := map[string]publishMetricValue{
		"total_claims": {Value: len(claims), Status: "ok"},
	}
	for _, claim := range claims {
		key := claim.Status + "_claims"
		metric := details[key]
		metric.Value++
		metric.Status = "ok"
		details[key] = metric
		if claim.RuntimeImplemented {
			runtimeMetric := details["runtime_implemented_claims"]
			runtimeMetric.Value++
			runtimeMetric.Status = "ok"
			details["runtime_implemented_claims"] = runtimeMetric
		}
	}
	return details
}

func buildPublishAppPlatformSection(root string) publishReadinessSection {
	if root == "" {
		return publishReadinessSection{
			ID:     "app-platform",
			Label:  "App platform shipping guardrails",
			Status: "ok",
			Checks: []publishReadinessCheck{{
				Path:   "source checkout",
				Exists: false,
				Status: "not-required-for-installed-binary",
			}},
		}
	}

	required := []publishFragmentRequirement{
		{
			checkPath: "specs/app-platform-shipping-playbook.md#default-stack-decision",
			filePath:  "specs/app-platform-shipping-playbook.md",
			fragments: []string{
				"## Default Stack Decision",
				"Next.js App Router",
				"Tauri 2",
				"Expo React Native",
			},
		},
		{
			checkPath: "specs/app-platform-shipping-playbook.md#security-baseline",
			filePath:  "specs/app-platform-shipping-playbook.md",
			fragments: []string{
				"## Security Baseline",
				"OWASP MASVS",
				"sign, notarize, staple",
				"Play Integrity",
				"strict Content Security Policy",
			},
		},
		{
			checkPath: "specs/app-platform-shipping-playbook.md#performance-and-optimization-baseline",
			filePath:  "specs/app-platform-shipping-playbook.md",
			fragments: []string{
				"## Performance And Optimization Baseline",
				"Core Web Vitals",
				"Baseline Profiles",
				"model load time",
			},
		},
		{
			checkPath: "specs/app-platform-shipping-playbook.md#logging-diagnostics-and-observability",
			filePath:  "specs/app-platform-shipping-playbook.md",
			fragments: []string{
				"## Logging, Diagnostics, And Observability",
				"`session_id`",
				"`route_id`",
				"`privacy_redaction_state`",
				"OpenTelemetry",
			},
		},
		{
			checkPath: "specs/app-platform-shipping-playbook.md#update-and-release-policy",
			filePath:  "specs/app-platform-shipping-playbook.md",
			fragments: []string{
				"## Update And Release Policy",
				"Signed updates are mandatory",
				"rollback target",
				"schema version",
			},
		},
		{
			checkPath: "specs/app-platform-shipping-playbook.md#app-shipping-gates",
			filePath:  "specs/app-platform-shipping-playbook.md",
			fragments: []string{
				"## App Shipping Gates",
				"security and privacy gate",
				"observability and diagnostics gate",
				"support bundle gate",
			},
		},
		{
			checkPath: "specs/app-platform-shipping-playbook.md#source-backed-inputs",
			filePath:  "specs/app-platform-shipping-playbook.md",
			fragments: []string{
				"## Source-Backed Inputs",
				"developer.apple.com",
				"developer.android.com",
				"learn.microsoft.com",
				"mas.owasp.org/MASVS",
				"v2.tauri.app/security",
				"opentelemetry.io",
			},
		},
	}

	checks, status := buildFragmentChecks(root, required)
	return publishReadinessSection{
		ID:     "app-platform",
		Label:  "App platform shipping guardrails",
		Status: status,
		Checks: checks,
	}
}

func buildPublishOfflineRegressionSection(root string) publishReadinessSection {
	if root == "" {
		return publishReadinessSection{
			ID:     "offline-regression",
			Label:  "Offline model regression guardrails",
			Status: "ok",
			Checks: []publishReadinessCheck{{
				Path:   "source checkout",
				Exists: false,
				Status: "not-required-for-installed-binary",
			}},
		}
	}

	required := []publishFragmentRequirement{
		{
			checkPath: "specs/local-model-support-matrix.md#registry-contract",
			filePath:  "specs/local-model-support-matrix.md",
			fragments: []string{
				"## Registry Contract",
				"`profile_role`",
				"`status`",
			},
		},
		{
			checkPath: "specs/local-model-support-matrix.md#promotion-loop",
			filePath:  "specs/local-model-support-matrix.md",
			fragments: []string{
				"## Promotion Loop",
				"### 3. Canary",
				"### 5. Promote",
				"successor versions",
			},
		},
		{
			checkPath: "specs/platform-offline-strategy.md#future-update-policy",
			filePath:  "specs/platform-offline-strategy.md",
			fragments: []string{
				"## Future Update Policy",
				"Future model updates should:",
				"preserve route evidence shape",
				"preserve session and artifact identity",
			},
		},
		{
			checkPath: "specs/adapter-benchmark-gate.md#routing-use",
			filePath:  "specs/adapter-benchmark-gate.md",
			fragments: []string{
				"### 4. Routing Use",
				"repeated regression across recent samples",
				"strong recovery after degradation",
			},
		},
	}

	checks, status := buildFragmentChecks(root, required)
	return publishReadinessSection{
		ID:     "offline-regression",
		Label:  "Offline model regression guardrails",
		Status: status,
		Checks: checks,
	}
}

func buildPublishCompetitivePressureSection(root string) publishReadinessSection {
	if root == "" {
		return publishReadinessSection{
			ID:     "competitive-pressure",
			Label:  "Competitive release pressure guardrails",
			Status: "ok",
			Checks: []publishReadinessCheck{{
				Path:   "source checkout",
				Exists: false,
				Status: "not-required-for-installed-binary",
			}},
		}
	}

	required := []publishFragmentRequirement{
		{
			checkPath: "specs/competitive-release-plan.md#competitive-universe",
			filePath:  "specs/competitive-release-plan.md",
			fragments: []string{
				"## Competitive Universe",
				"Direct Replacement Threats",
				"Local And Offline Front Doors",
				"Routing And Gateway Infrastructure",
			},
		},
		{
			checkPath: "specs/competitive-release-plan.md#requirement-rejection-filter",
			filePath:  "specs/competitive-release-plan.md",
			fragments: []string{
				"## Requirement Rejection Filter",
				"adopt, integrate, watch, reject, or delete",
				"delete: remove an existing Jini requirement",
			},
		},
		{
			checkPath: "specs/competitive-release-plan.md#p0-feature-selection-loop",
			filePath:  "specs/competitive-release-plan.md",
			fragments: []string{
				"Competitor watching is a P0 feature-selection loop.",
				"nominates next feature candidates and deletion candidates",
				"replacement-critical claim",
			},
		},
		{
			checkPath: "specs/number-one-platform-prd.md#market-and-learning-guards",
			filePath:  "specs/number-one-platform-prd.md",
			fragments: []string{
				"## Market And Learning Guards",
				"Competitor watch packets can nominate next feature candidates",
				"No competitor finding becomes active scope unless the decision record changes.",
				"copy, integrate, watch, reject, or",
			},
		},
	}

	checks, status := buildFragmentChecks(root, required)
	return publishReadinessSection{
		ID:     "competitive-pressure",
		Label:  "Competitive release pressure guardrails",
		Status: status,
		Checks: checks,
	}
}

func buildPublishProductivityLearningSection(root string) publishReadinessSection {
	if root == "" {
		return publishReadinessSection{
			ID:     "productivity-learning",
			Label:  "Compounding productivity learning guardrails",
			Status: "ok",
			Checks: []publishReadinessCheck{{
				Path:   "source checkout",
				Exists: false,
				Status: "not-required-for-installed-binary",
			}},
		}
	}

	required := []publishFragmentRequirement{
		{
			checkPath: "specs/number-one-platform-prd.md#market-and-learning-guards",
			filePath:  "specs/number-one-platform-prd.md",
			fragments: []string{
				"## Market And Learning Guards",
				"learn stable user context, usage, habits, and repeated patterns",
				"fewer repeated prompts, better defaults, and better route choices",
				"keep learning inspectable and controllable",
			},
		},
		{
			checkPath: "specs/learning-system.md#user-context-productivity-learning",
			filePath:  "specs/learning-system.md",
			fragments: []string{
				"## 2a. User Context Productivity Learning",
				"User productivity learning is a P0 product requirement.",
				"stable user context, usage, habits, and repeated work patterns",
			},
		},
	}

	checks, status := buildFragmentChecks(root, required)
	return publishReadinessSection{
		ID:     "productivity-learning",
		Label:  "Compounding productivity learning guardrails",
		Status: status,
		Checks: checks,
	}
}

func buildFragmentChecks(root string, required []publishFragmentRequirement) ([]publishReadinessCheck, string) {
	checks := make([]publishReadinessCheck, 0, len(required))
	status := "ok"
	for _, requirement := range required {
		file := filepath.Join(root, requirement.filePath)
		data, err := os.ReadFile(file)
		exists := err == nil
		checkStatus := "ok"
		missingFragments := []string(nil)
		if !exists {
			checkStatus = "missing"
			status = "needs-attention"
		} else if missingFragments = missingRequiredFragments(string(data), requirement.fragments); len(missingFragments) > 0 {
			checkStatus = "incomplete"
			status = "needs-attention"
		}
		checks = append(checks, publishReadinessCheck{
			Path:             requirement.checkPath,
			Exists:           exists,
			Status:           checkStatus,
			MissingFragments: missingFragments,
		})
	}
	return checks, status
}

func missingRequiredFragments(text string, fragments []string) []string {
	missing := []string(nil)
	for _, fragment := range fragments {
		if !strings.Contains(text, fragment) {
			missing = append(missing, fragment)
		}
	}
	return missing
}

func buildPublishRuntimeSection(root string) publishReadinessSection {
	details := map[string]publishMetricValue{
		"packs":   {Value: countChildDirs(filepath.Join(root, "packs")), Status: "ok"},
		"bundles": {Value: countInstallManifestIDs(root, "bundles"), Status: "ok"},
		"kits":    {Value: countInstallManifestIDs(root, "kits"), Status: "ok"},
		"targets": {Value: countInstallManifestIDs(root, "targets"), Status: "ok"},
	}
	return publishReadinessSection{
		ID:      "go-runtime",
		Label:   "Native Go runtime",
		Status:  "ok",
		Details: details,
	}
}

func renderPublishReadinessText(stdout io.Writer, report publishReadinessReport) {
	fmt.Fprintf(stdout, "STATUS %s\n", report.Status)
	fmt.Fprintln(stdout, "RUNTIME")
	fmt.Fprintf(stdout, "  LANGUAGE %s\n", report.Runtime.Language)
	fmt.Fprintf(stdout, "  LEGACY_FALLBACK %t\n", report.Runtime.LegacyFallback)
	fmt.Fprintln(stdout, "COUNTS")
	fmt.Fprintf(stdout, "  PACKS %d\n", report.PackCount)
	fmt.Fprintf(stdout, "  BUNDLES %d\n", report.BundleCount)
	fmt.Fprintf(stdout, "  KITS %d\n", report.KitCount)
	fmt.Fprintf(stdout, "  TARGETS %d\n", report.TargetCount)
	fmt.Fprintln(stdout, "SECTIONS")
	for _, section := range report.Sections {
		fmt.Fprintf(stdout, "  %s %s\n", strings.ToUpper(section.ID), section.Status)
		for _, check := range section.Checks {
			fmt.Fprintf(stdout, "    %s %s\n", strings.ToUpper(check.Status), check.Path)
			for _, fragment := range check.MissingFragments {
				fmt.Fprintf(stdout, "      MISSING %s\n", fragment)
			}
		}
		for _, claim := range section.Claims {
			fmt.Fprintf(stdout, "    CLAIM %s STATUS %s RUNTIME %t\n", claim.Claim, claim.Status, claim.RuntimeImplemented)
		}
	}
}

func discoverSourceRoot() string {
	if configured := strings.TrimSpace(os.Getenv("JINI_SOURCE_DIR")); configured != "" && isJiniSourceRoot(configured) {
		return configured
	}
	if cwd, err := os.Getwd(); err == nil {
		if root, ok := findJiniSourceRoot(cwd); ok {
			return root
		}
	}
	if exe, err := os.Executable(); err == nil {
		if root, ok := findJiniSourceRoot(filepath.Dir(exe)); ok {
			return root
		}
	}
	return ""
}

func findJiniSourceRoot(start string) (string, bool) {
	current, err := filepath.Abs(start)
	if err != nil {
		return "", false
	}
	for {
		if isJiniSourceRoot(current) {
			return current, true
		}
		parent := filepath.Dir(current)
		if parent == current {
			return "", false
		}
		current = parent
	}
}

func isJiniSourceRoot(root string) bool {
	data, err := os.ReadFile(filepath.Join(root, "go.mod"))
	if err != nil {
		return false
	}
	for _, line := range strings.Split(string(data), "\n") {
		if strings.TrimSpace(line) == jiniModuleDeclaration {
			return true
		}
	}
	return false
}

func countChildDirs(path string) int {
	entries, err := os.ReadDir(path)
	if err != nil {
		return 0
	}
	count := 0
	for _, entry := range entries {
		if entry.IsDir() && !strings.HasPrefix(entry.Name(), ".") {
			count++
		}
	}
	return count
}

func regularFileExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}

func countInstallManifestIDs(root, sectionName string) int {
	data, err := os.ReadFile(filepath.Join(root, "distribution", "install-manifest.yaml"))
	if err != nil {
		return 0
	}
	inSection := false
	count := 0
	for _, raw := range strings.Split(string(data), "\n") {
		line := strings.TrimRight(raw, " \t")
		if line == sectionName+":" {
			inSection = true
			continue
		}
		if inSection && strings.TrimSpace(line) != "" && !strings.HasPrefix(line, " ") && !strings.HasPrefix(line, "\t") {
			break
		}
		if inSection && strings.HasPrefix(line, "  - id:") {
			count++
		}
	}
	return count
}
