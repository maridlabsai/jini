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
	} {
		if !strings.Contains(requiredGates, want) {
			t.Fatalf("required commit gate must run %q", want)
		}
	}

	gateMatrix := readRepoFile(t, root, "specs/engineering-gate-matrix.md")
	for _, want := range []string{
		"`git diff --check`",
		"`git diff --cached --check`",
		"staged and unstaged whitespace",
	} {
		if !strings.Contains(gateMatrix, want) {
			t.Fatalf("engineering gate matrix must document %q", want)
		}
	}
}
