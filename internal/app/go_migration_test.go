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

func TestRequiredGatesDoNotInvokePython(t *testing.T) {
	root := repoRootForMigrationTest(t)
	scanned := []string{
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
		for _, forbidden := range []string{"python3", "jini_validate", "tools/jini.py"} {
			if strings.Contains(text, forbidden) {
				t.Fatalf("Python invocation remains in %s: %s", rel, forbidden)
			}
		}
	}
}

func TestOfficialGoOnlySurfacesDoNotAdvertiseUnsupportedCommands(t *testing.T) {
	root := repoRootForMigrationTest(t)
	files := []string{
		"CHANGELOG.md",
		"PROOF_OF_DIFFERENCE.md",
		"distribution/install-manifest.yaml",
		"docs/cli.md",
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

func repoRootForMigrationTest(t *testing.T) string {
	t.Helper()
	output, err := exec.Command("git", "rev-parse", "--show-toplevel").Output()
	if err != nil {
		t.Fatalf("git rev-parse: %v", err)
	}
	return strings.TrimSpace(string(output))
}
