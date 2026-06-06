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
}

type publishReadinessCheck struct {
	Path   string `json:"path"`
	Exists bool   `json:"exists"`
	Status string `json:"status"`
}

type publishMetricValue struct {
	Value  int    `json:"value"`
	Status string `json:"status"`
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
		buildPublishOfflineRegressionSection(root),
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

	required := []struct {
		checkPath string
		filePath  string
		fragments []string
	}{
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

	checks := make([]publishReadinessCheck, 0, len(required))
	status := "ok"
	for _, requirement := range required {
		file := filepath.Join(root, requirement.filePath)
		data, err := os.ReadFile(file)
		exists := err == nil
		checkStatus := "ok"
		if !exists {
			checkStatus = "missing"
			status = "needs-attention"
		} else if missingRequiredFragment(string(data), requirement.fragments) {
			checkStatus = "incomplete"
			status = "needs-attention"
		}
		checks = append(checks, publishReadinessCheck{
			Path:   requirement.checkPath,
			Exists: exists,
			Status: checkStatus,
		})
	}

	return publishReadinessSection{
		ID:     "offline-regression",
		Label:  "Offline model regression guardrails",
		Status: status,
		Checks: checks,
	}
}

func missingRequiredFragment(text string, fragments []string) bool {
	for _, fragment := range fragments {
		if !strings.Contains(text, fragment) {
			return true
		}
	}
	return false
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
