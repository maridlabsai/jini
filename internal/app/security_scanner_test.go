package app_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestSecurityScannersAreWiredIntoCIAndRequiredGates(t *testing.T) {
	root := repoRootForMigrationTest(t)

	securityWorkflow := readRepoFile(t, root, ".github/workflows/security.yml")
	for _, want := range []string{
		"github/codeql-action/init@v4",
		"github/codeql-action/analyze@v4",
		"golang/govulncheck-action@v1",
		"google/osv-scanner-action/.github/workflows/osv-scanner-reusable-pr.yml@v2.3.8",
		"google/osv-scanner-action/.github/workflows/osv-scanner-reusable.yml@v2.3.8",
		"trufflesecurity/trufflehog@v3.95.5",
		`version: "3.95.5"`,
		"extra_args: --results=verified,unknown",
		"fetch-depth: 0",
		"security-events: write",
		"go-version-file: go.mod",
	} {
		if !strings.Contains(securityWorkflow, want) {
			t.Fatalf("security workflow must include %q", want)
		}
	}

	dependabot := readRepoFile(t, root, ".github/dependabot.yml")
	for _, want := range []string{
		`package-ecosystem: "gomod"`,
		`package-ecosystem: "github-actions"`,
		`directory: "/"`,
		"interval: \"weekly\"",
	} {
		if !strings.Contains(dependabot, want) {
			t.Fatalf("dependabot config must include %q", want)
		}
	}

	securityGate := readRepoFile(t, root, "tools/security_configuration_gate.sh")
	for _, want := range []string{
		".github/workflows/security.yml",
		".github/dependabot.yml",
		"github/codeql-action/init@v4",
		"golang/govulncheck-action@v1",
		"google/osv-scanner-action/.github/workflows/osv-scanner-reusable-pr.yml@v2.3.8",
		"trufflesecurity/trufflehog@v3.95.5",
		`version: "3.95.5"`,
	} {
		if !strings.Contains(securityGate, want) {
			t.Fatalf("security configuration gate must check %q", want)
		}
	}

	requiredGates := readRepoFile(t, root, "tools/run_required_gates.sh")
	if !strings.Contains(requiredGates, "tools/security_configuration_gate.sh") {
		t.Fatalf("required gates must run the security configuration gate")
	}

	gateMatrix := readRepoFile(t, root, "specs/engineering-gate-matrix.md")
	for _, want := range []string{
		"tools/security_configuration_gate.sh",
		"CodeQL",
		"govulncheck",
		"OSV-Scanner",
		"TruffleHog",
		"Dependabot",
	} {
		if !strings.Contains(gateMatrix, want) {
			t.Fatalf("engineering gate matrix must document %q", want)
		}
	}
}

func readRepoFile(t *testing.T, root, rel string) string {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(root, rel))
	if err != nil {
		t.Fatalf("read %s: %v", rel, err)
	}
	return string(data)
}
