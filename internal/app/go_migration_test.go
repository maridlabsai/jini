package app_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

func gitTrackedFiles(t *testing.T) []string {
	t.Helper()
	output, err := exec.Command("git", "ls-files").Output()
	if err != nil {
		t.Fatalf("git ls-files: %v", err)
	}
	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	if len(lines) == 1 && lines[0] == "" {
		return nil
	}
	return lines
}

func TestRepositoryHasNoTrackedPythonFiles(t *testing.T) {
	for _, path := range gitTrackedFiles(t) {
		if strings.HasSuffix(path, ".py") {
			t.Fatalf("tracked Python file remains after Go migration: %s", path)
		}
	}
}

func TestActiveGoOnlySurfacesDoNotReferencePython(t *testing.T) {
	root := repoRootForMigrationTest(t)
	scanned := []string{
		".github/ISSUE_TEMPLATE/bug_report.yml",
		".github/workflows/ci-installer.yml",
		"Makefile",
		"install.sh",
		"tools/run_required_gates.sh",
	}
	for _, rel := range scanned {
		data, err := os.ReadFile(filepath.Join(root, rel))
		if err != nil {
			t.Fatalf("read %s: %v", rel, err)
		}
		text := string(data)
		for _, forbidden := range []string{"python", "Python", "python3", "jini_validate", "tools/jini.py"} {
			if strings.Contains(text, forbidden) {
				t.Fatalf("Python reference remains in active Go-only surface %s: %s", rel, forbidden)
			}
		}
	}
}

func TestOfficialGoOnlySurfacesDoNotAdvertiseUnsupportedCommands(t *testing.T) {
	root := repoRootForMigrationTest(t)
	files := []string{
		"CHANGELOG.md",
		"PROOF_OF_DIFFERENCE.md",
		"README.md",
		"distribution/install-manifest.yaml",
		"docs/cli.md",
		"docs/index.md",
		"docs/install.md",
		"docs/simple.md",
		"specs/cli-replacement-score-plan.md",
		"specs/cross-surface-session-system-and-dev-design.md",
		"specs/docs-homepage-rewrite-plan.md",
		"specs/install-packaging.md",
		"specs/personal-os.md",
	}
	for _, pattern := range []string{
		"distribution/targets/*/README.md",
		"packs/*/README.md",
	} {
		matches, err := filepath.Glob(filepath.Join(root, pattern))
		if err != nil {
			t.Fatalf("glob %s: %v", pattern, err)
		}
		for _, match := range matches {
			rel, err := filepath.Rel(root, match)
			if err != nil {
				t.Fatalf("relative path for %s: %v", match, err)
			}
			files = append(files, rel)
		}
	}

	allowed := map[string]bool{
		"admin":             true,
		"check":             true,
		"commands":          true,
		"continue":          true,
		"doctor":            true,
		"fix":               true,
		"help":              true,
		"init":              true,
		"memory":            true,
		"new":               true,
		"observe":           true,
		"open":              true,
		"permissions":       true,
		"provider":          true,
		"publish-readiness": true,
		"review":            true,
		"route":             true,
		"run":               true,
		"scorecard-gate":    true,
		"status":            true,
	}
	jiniCommand := regexp.MustCompile(`\bjini\s+([a-z][a-z0-9-]*)`)
	manifestArgs := regexp.MustCompile(`args:\s*\["([^"]+)"`)

	for _, rel := range files {
		data, err := os.ReadFile(filepath.Join(root, rel))
		if err != nil {
			t.Fatalf("read %s: %v", rel, err)
		}
		for lineNumber, line := range strings.Split(string(data), "\n") {
			for _, match := range jiniCommand.FindAllStringSubmatch(line, -1) {
				if !allowed[match[1]] {
					t.Errorf("%s:%d advertises unsupported native Go command: jini %s", rel, lineNumber+1, match[1])
				}
			}
			if rel == "distribution/install-manifest.yaml" {
				for _, match := range manifestArgs.FindAllStringSubmatch(line, -1) {
					if !allowed[match[1]] {
						t.Errorf("%s:%d defines unsupported native Go manifest smoke command: %s", rel, lineNumber+1, match[1])
					}
				}
			}
		}
	}
}

func TestPublicDocsDoNotTeachPathfulStatus(t *testing.T) {
	root := repoRootForMigrationTest(t)
	files := []string{
		"README.md",
		"docs/cli.md",
		"docs/index.md",
		"docs/install.md",
		"docs/simple.md",
	}
	stalePatterns := []string{
		"jini status /",
		"status /path/to/work",
	}

	for _, rel := range files {
		data, err := os.ReadFile(filepath.Join(root, rel))
		if err != nil {
			t.Fatalf("read %s: %v", rel, err)
		}
		for lineNumber, line := range strings.Split(string(data), "\n") {
			for _, pattern := range stalePatterns {
				if strings.Contains(line, pattern) {
					t.Errorf("%s:%d teaches retired pathful status form: %s", rel, lineNumber+1, pattern)
				}
			}
		}
	}
}

func TestPublicPlanningSpecsDoNotPromoteOperatorOnlyCommands(t *testing.T) {
	root := repoRootForMigrationTest(t)
	files := []string{
		"README.md",
		"docs/cli.md",
		"docs/install.md",
		"docs/simple.md",
		"specs/cli-replacement-score-plan.md",
		"specs/docs-homepage-rewrite-plan.md",
	}
	operatorOnlyCommands := []string{
		"jini check",
		"jini provider doctor",
	}

	for _, rel := range files {
		data, err := os.ReadFile(filepath.Join(root, rel))
		if err != nil {
			t.Fatalf("read %s: %v", rel, err)
		}
		for lineNumber, line := range strings.Split(string(data), "\n") {
			for _, command := range operatorOnlyCommands {
				if strings.Contains(line, command) {
					t.Errorf("%s:%d promotes operator-only command in public planning surface: %s", rel, lineNumber+1, command)
				}
			}
		}
	}
}

func TestGoldenBenchmarkRunnableCommandsUseNativeGoSurface(t *testing.T) {
	root := repoRootForMigrationTest(t)
	rel := "specs/golden-competitive-benchmark.yaml"
	data, err := os.ReadFile(filepath.Join(root, rel))
	if err != nil {
		t.Fatalf("read %s: %v", rel, err)
	}
	text := string(data)
	commandPattern := regexp.MustCompile(`(?ms)^\s+command:\s*(?:\[\s*|\n\s*\[\s*)\n?\s*"([^"]+)"`)
	allowedJSONCommands := map[string]bool{
		"doctor":            true,
		"provider":          true,
		"publish-readiness": true,
		"scorecard-gate":    true,
	}
	for _, match := range commandPattern.FindAllStringSubmatchIndex(text, -1) {
		command := text[match[2]:match[3]]
		if allowedJSONCommands[command] {
			continue
		}
		lineNumber := strings.Count(text[:match[0]], "\n") + 1
		t.Errorf("%s:%d has runnable benchmark command outside native JSON-capable Go surface: %s", rel, lineNumber, command)
	}
}

func TestInstallDocsMatchGoInstallerContract(t *testing.T) {
	root := repoRootForMigrationTest(t)
	data, err := os.ReadFile(filepath.Join(root, "docs/install.md"))
	if err != nil {
		t.Fatalf("read docs/install.md: %v", err)
	}
	text := string(data)
	staleFragments := []string{
		"source runtime",
		"source fallback",
		"support receipt",
		"send version=, source_reason=, release_validation=, next_step=",
		"release validation failed: unsupported-public-command-surface",
	}
	for _, fragment := range staleFragments {
		if strings.Contains(text, fragment) {
			t.Errorf("docs/install.md documents stale installer output: %s", fragment)
		}
	}
}

func repoRootForMigrationTest(t *testing.T) string {
	t.Helper()
	output, err := exec.Command("git", "rev-parse", "--show-toplevel").Output()
	if err != nil {
		t.Fatalf("git rev-parse: %v", err)
	}
	return strings.TrimSpace(string(output))
}
